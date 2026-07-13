import * as fs from 'fs';
import * as path from 'path';
import * as childProcess from 'child_process';
import { logInfo, logError, logStep } from './logger';

export async function runAgent(
  agentName: string,
  agentSpecFilePath: string,
  stateFilePath: string,
  initialStateStr: string
): Promise<void> {
  if (!fs.existsSync(agentSpecFilePath)) {
    throw new Error(`No se encontró el archivo de especificación del agente: ${agentSpecFilePath}`);
  }
  const systemInstruction = fs.readFileSync(agentSpecFilePath, 'utf-8');

  // Extraer el nombre del agente registrado desde el Frontmatter de YAML (campo 'name')
  const frontmatterMatch = systemInstruction.match(/^---\s*\n([\s\S]*?)\n---\s*\n/);
  let registeredAgentName: string | null = null;
  if (frontmatterMatch) {
    const frontmatter = frontmatterMatch[1];
    const nameMatch = frontmatter.match(/^name:\s*([^\n\r]+)/m);
    if (nameMatch) {
      registeredAgentName = nameMatch[1].replace(/['"]/g, '').trim();
    }
  }

    const featureFolder = path.dirname(stateFilePath);
  let userMessage = '';

  const VALID_OUTPUTS: Record<string, string> = {
    'spec-partner': 'spec_approved',
    'spec partner': 'spec_approved',
    'gherkin-author': 'gherkin_generated',
    'gherkin author': 'gherkin_generated',
    'tdd-craftsman': 'tdd_completed',
    'tdd craftsman': 'tdd_completed',
    'judge': 'audit_passed o audit_failed',
    'mutation-tester': 'mutation_passed o mutation_failed',
    'mutation tester': 'mutation_passed o mutation_failed'
  };

  const currentKey = (registeredAgentName || agentName).toLowerCase();
  const allowedTransition = VALID_OUTPUTS[currentKey] || 'done';

  if (registeredAgentName) {
    logInfo(`Iniciando agente registrado de Gemini CLI [@${registeredAgentName}]...`);
    userMessage = `@${registeredAgentName}
Hola. Tu espacio de trabajo está en la carpeta de la feature: "${featureFolder}".

Por favor, lee el archivo de estado "${stateFilePath}" para entender la tarea actual, los archivos involucrados y el contexto del handoff o errores anteriores.
Procede a realizar tus tareas utilizando tus herramientas predeterminadas. 
IMPORTANTE: Al concluir tus tareas, DEBES actualizar el campo "* **Estado Actual:**" en el archivo de estado "${stateFilePath}" estrictamente a uno de tus estados de transición válidos: [${allowedTransition}]. No uses "done", "completed" u otros términos genéricos a menos que estén listados como válidos para tu rol. Cuando termines y hayas guardado el archivo, despídete y detente.`;
  } else {
    logInfo(`Iniciando agente nativo de Gemini CLI [${agentName}] pasando instrucciones del sistema...`);
    userMessage = `Hola. Eres el agente "${agentName}". Tu espacio de trabajo está en la carpeta de la feature: "${featureFolder}".

Tu comportamiento base estricto se define en la siguiente especificación:
${systemInstruction}

Por favor, lee el archivo de estado "${stateFilePath}" para entender la tarea actual, los archivos involucrados y el contexto del handoff o errores anteriores.
Procede a realizar tus tareas utilizando tus herramientas predeterminadas. 
IMPORTANTE: Al concluir tus tareas, DEBES actualizar el campo "* **Estado Actual:**" en el archivo de estado "${stateFilePath}" estrictamente a uno de tus estados de transición válidos: [${allowedTransition}]. No uses "done", "completed" u otros términos genéricos a menos que estén listados como válidos para tu rol. Cuando termines y hayas guardado el archivo, despídete y detente.`;
  }

  logStep(agentName, "Delegando control al proceso hijo (Gemini CLI YOLO Mode)...");
  
  try {
    // execFileSync runs the command directly without a shell, avoiding all shell-escaping, backtick, and command injection issues!
    childProcess.execFileSync('gemini', ['-y', '-p', userMessage], { stdio: 'inherit' });
  } catch (error: any) {
    logError(`El proceso subagente Gemini CLI finalizó con código de salida u error: ${error.message}`);
    // We swallow the exception to allow the orchestrator loop to re-evaluate the state in current.md
  }
}
