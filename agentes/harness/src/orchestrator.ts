import * as fs from 'fs';
import * as path from 'path';
import { parseStateFile, writeStateFile, StateData } from './state';
import { runAgent } from './agent-runner';
import { askHumanTool } from './tools';
import { 
  logHeader, 
  logInfo, 
  logSuccess, 
  logError, 
  logPending, 
  logWarning, 
  logStep,
  formatState
} from './logger';
import * as childProcess from 'child_process';

const AGENT_MAP: Record<string, { name: string; spec: string }> = {
  'pending': { name: 'Spec Partner', spec: 'spec_partner.md' },
  'spec_approved': { name: 'Gherkin Author', spec: 'gherkin_author.md' },
  'human_approved': { name: 'TDD Craftsman', spec: 'tdd_craftsman.md' },
  'audit_failed': { name: 'TDD Craftsman', spec: 'tdd_craftsman.md' },
  'mutation_failed': { name: 'TDD Craftsman', spec: 'tdd_craftsman.md' },
  'tdd_completed': { name: 'Judge', spec: 'judge.md' },
  'audit_passed': { name: 'Mutation Tester', spec: 'mutation_tester.md' }
};

// Herdr Integration Status Reporting
function reportHerdr(agentName: string, state: 'working' | 'idle' | 'blocked', statusText: string, title?: string) {
  if (process.env.HERDR_ENV !== '1' || !process.env.HERDR_PANE_ID) return;
  try {
    const paneId = process.env.HERDR_PANE_ID;
    const cmdAgent = `herdr pane report-agent "${paneId}" --source "gemini" --agent "${agentName}" --state "${state}" --custom-status "${statusText.replace(/"/g, '\\"')}"`;
    childProcess.execSync(cmdAgent, { stdio: 'ignore' });

    if (title) {
      const cmdMetadata = `herdr pane report-metadata "${paneId}" --source "gemini" --title "${title.replace(/"/g, '\\"')}"`;
      childProcess.execSync(cmdMetadata, { stdio: 'ignore' });
    }
  } catch (err) {
    // Fail silently if herdr command fails
  }
}

export async function runOrchestrator(stateFilePath: string): Promise<void> {
  const resolvedStatePath = path.resolve(stateFilePath);
  if (!fs.existsSync(resolvedStatePath)) {
    logError(`No se encontró el archivo de estado en la ruta: ${stateFilePath}`);
    return;
  }

  logHeader('Iniciando Orquestación Craftsman Lead');
  logInfo(`Cargando estado desde: ${resolvedStatePath}`);

  let fallbackCount = 0;
  const maxFallbacks = 3;

  while (true) {
    const stateData = parseStateFile(resolvedStatePath);
    const currentState = stateData.currentState;
    const featureName = stateData.featureIdName;

    logInfo(`Estado actual de la Feature [${featureName}]: [${formatState(currentState)}]`);

    // Check terminal Done status
    if (currentState === 'done' || currentState === 'mutation_passed') {
      logSuccess(`¡ÉXITO! El desarrollo de la feature "${featureName}" ha finalizado correctamente (Estado: ${currentState}).`);
      
      // Update state to done if it was mutation_passed or completed
      if (currentState === 'mutation_passed') {
        stateData.currentState = 'done';
        stateData.lastActiveAgent = 'Craftsman Lead';
        writeStateFile(resolvedStatePath, stateData);
      }
      
      reportHerdr('Craftsman Lead', 'idle', 'Completado', 'Feature: DONE');
      break;
    }

    // Interactive Human Approval Step (Puerta Humana)
    if (currentState === 'gherkin_generated' || currentState === 'designed') {
      logPending('PUERTA HUMANA: Se requiere aprobación del contrato Gherkin para proceder.');
      reportHerdr('Craftsman Lead', 'blocked', 'Esperando aprobación humana', 'Puerta Humana');

      const featureFolder = path.dirname(resolvedStatePath);
      const featureFile = path.join('features', `${featureName}.feature`);
      logInfo(`Por favor, revisa el archivo de escenarios Gherkin generado en: ${featureFile}`);

      // Herdr Integration: Open nano in new tabs for both Gherkin and Hard Spec
      if (process.env.HERDR_ENV === '1' && process.env.HERDR_PANE_ID) {
        try {
          const gherkinPath = path.resolve(featureFile);
          const hardSpecFile = path.join('specs', featureName, 'hard_spec.md');
          const hardSpecPath = path.resolve(hardSpecFile);

          // 1. Open Hard Spec Tab (if it exists)
          if (fs.existsSync(hardSpecPath)) {
            logInfo('Detectado entorno Herdr. Abriendo Hard Spec en una nueva pestaña con nano...');
            const specTabCmd = `herdr tab create --label "Spec: ${featureName.substring(0, 15)}" --focus`;
            const specTabJson = childProcess.execSync(specTabCmd, { encoding: 'utf8' });
            const specTabResult = JSON.parse(specTabJson);
            const specPaneId = specTabResult?.result?.root_pane?.pane_id;
            if (specPaneId) {
              const runNanoSpec = `herdr pane run "${specPaneId}" "LC_ALL=C.UTF-8 LANG=C.UTF-8 nano '${hardSpecPath}'"`;
              childProcess.execSync(runNanoSpec);
              logSuccess('Pestaña de Hard Spec abierta con nano.');
            } else {
              logWarning('No se pudo determinar el pane_id de la pestaña de Hard Spec en Herdr.');
            }
          }

          // 2. Open Gherkin Tab (if it exists)
          if (fs.existsSync(gherkinPath)) {
            logInfo('Abriendo contrato Gherkin en una nueva pestaña con nano...');
            const gherkinTabCmd = `herdr tab create --label "Gherkin: ${featureName.substring(0, 15)}" --focus`;
            const gherkinTabJson = childProcess.execSync(gherkinTabCmd, { encoding: 'utf8' });
            const gherkinTabResult = JSON.parse(gherkinTabJson);
            const gherkinPaneId = gherkinTabResult?.result?.root_pane?.pane_id;
            if (gherkinPaneId) {
              const runNanoGherkin = `herdr pane run "${gherkinPaneId}" "LC_ALL=C.UTF-8 LANG=C.UTF-8 nano '${gherkinPath}'"`;
              childProcess.execSync(runNanoGherkin);
              logSuccess('Pestaña de Gherkin abierta con nano.');
            } else {
              logWarning('No se pudo determinar el pane_id de la pestaña de Gherkin en Herdr.');
            }
          }
        } catch (err) {
          logError(`Error al abrir archivos en pestañas de Herdr: ${err.message}`);
        }
      }
      
      const answer = await askHumanTool('Escribe "Aprobar" para dar el visto bueno al contrato Gherkin y comenzar el desarrollo TDD:');
      
      if (answer.toLowerCase() === 'aprobar' || answer.toLowerCase() === 'si' || answer.toLowerCase() === 'yes') {
        logSuccess('¡Contrato Gherkin aprobado por el humano!');
        
        stateData.currentState = 'human_approved';
        stateData.lastActiveAgent = 'Craftsman Lead';
        stateData.contextAndErrors.push('Puerta humana superada. Comienza fase de desarrollo TDD.');
        writeStateFile(resolvedStatePath, stateData);
        
        reportHerdr('Craftsman Lead', 'working', 'Aprobado. Iniciando TDD', 'Puerta Superada');
        continue; // Re-eval loop to launch TDD Craftsman
      } else {
        logWarning('El contrato no fue aprobado. El flujo se mantendrá en espera.');
        break;
      }
    }

    // Check if the current state is mapped to an agent
    const agentConfig = AGENT_MAP[currentState];
    if (!agentConfig) {
      logError(`Estado desconocido o no gestionado automáticamente: "${currentState}"`);
      break;
    }

    // Prevent Infinite Loop on successive failures
    if (currentState === 'audit_failed' || currentState === 'mutation_failed') {
      fallbackCount++;
      logWarning(`Se ha activado un retorno (fallback) debido a fallos previos (Intento de fallback: ${fallbackCount}/${maxFallbacks}).`);
      
      if (fallbackCount > maxFallbacks) {
        logError(`BUCLE DETECTADO: El desarrollo ha fallado consecutivamente más de ${maxFallbacks} veces en las fases de auditoría o mutación.`);
        logPending('Deteniendo el flujo automático para evitar bucle infinito. Por favor, asiste al agente manualmente.');
        reportHerdr('Craftsman Lead', 'blocked', 'Detenido por bucle infinito', 'Bucle Infinito');
        break;
      }
    } else if (currentState === 'tdd_completed') {
      // If we got to tdd_completed, it means the developer successfully submitted their work. 
      // We don't reset the fallback count here because if the Judge fails immediately, we want to accumulate.
      // But if the Judge or Mutation passed, the fallback counts will resolve when moving to success states.
    } else {
      // For forward stages, reset the fallback tracker
      if (currentState !== 'audit_failed' && currentState !== 'mutation_failed') {
        fallbackCount = 0;
      }
    }

    // Determine the agent specification file path
    const agentSpecPath = path.resolve(__dirname, '..', agentConfig.spec);
    if (!fs.existsSync(agentSpecPath)) {
      logError(`No se encontró el archivo de especificación del agente en: ${agentSpecPath}`);
      break;
    }

    // Herdr state reporting
    const featurePrefix = featureName.split('-')[0] || 'Task';
    const tabTitle = `${featurePrefix} - ${agentConfig.name}`;
    reportHerdr(agentConfig.name, 'working', `Ejecutando fase [${currentState}]`, tabTitle);

    logHeader(`Lanzando Agente: ${agentConfig.name}`);
    logInfo(`Fase: ${currentState} → Siguiente Paso`);

    try {
      // Run the agent loop
      await runAgent(agentConfig.name, agentSpecPath, resolvedStatePath, currentState);
    } catch (error: any) {
      logError(`Error durante la ejecución del agente [${agentConfig.name}]: ${error.message}`);
      reportHerdr(agentConfig.name, 'blocked', `Error: ${error.message}`);
      break;
    }
  }
}
