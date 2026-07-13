import * as fs from 'fs';
import * as path from 'path';

export interface StateData {
  currentState: string;
  lastActiveAgent: string;
  updateDateTime: string;
  featureIdName: string;
  productionFiles: string[];
  testFiles: string[];
  contextAndErrors: string[];
}

export function parseStateFile(filePath: string): StateData {
  if (!fs.existsSync(filePath)) {
    throw new Error(`State file does not exist: ${filePath}`);
  }

  const content = fs.readFileSync(filePath, 'utf-8');
  
  let currentState = 'pending';
  let lastActiveAgent = 'Ninguno';
  let updateDateTime = '';
  let featureIdName = '';
  const productionFiles: string[] = [];
  const testFiles: string[] = [];
  const contextAndErrors: string[] = [];

  // Parse properties
  const stateMatch = content.match(/\*\s*\*\*Estado Actual:\*\*\s*(.+)/i);
  if (stateMatch) currentState = stateMatch[1].trim();

  const agentMatch = content.match(/\*\s*\*\*Último Agente Activo:\*\*\s*(.+)/i);
  if (agentMatch) lastActiveAgent = agentMatch[1].trim();

  const dateMatch = content.match(/\*\s*\*\*Fecha\/Hora de Actualización:\*\*\s*(.+)/i);
  if (dateMatch) updateDateTime = dateMatch[1].trim();

  const featureMatch = content.match(/-\s*\*\*ID y Nombre de Feature:\*\*\s*(.+)/i);
  if (featureMatch) featureIdName = featureMatch[1].trim();

  // Parsing Modified Files section
  const modifiedFilesIndex = content.indexOf('## Archivos Modificados en este Ciclo');
  const contextIndex = content.indexOf('## Contexto del Handoff / Errores Recientes');

  if (modifiedFilesIndex !== -1) {
    const endOfSection = contextIndex !== -1 ? contextIndex : content.length;
    const sectionText = content.substring(modifiedFilesIndex, endOfSection);
    
    // Parse Production files
    const prodMatches = sectionText.matchAll(/-\s*\*\*Código de Producción:\*\*\s*(.+)/gi);
    for (const match of prodMatches) {
      const file = match[1].trim();
      if (file && file !== '[Ruta del archivo]') {
        productionFiles.push(file);
      }
    }

    // Parse Test files
    const testMatches = sectionText.matchAll(/-\s*\*\*Código de Pruebas:\*\*\s*(.+)/gi);
    for (const match of testMatches) {
      const file = match[1].trim();
      if (file && file !== '[Ruta del archivo]') {
        testFiles.push(file);
      }
    }
  }

  // Parsing Context section
  if (contextIndex !== -1) {
    const sectionText = content.substring(contextIndex);
    const lines = sectionText.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed.startsWith('- ') && !trimmed.startsWith('- **')) {
        const item = trimmed.substring(2).trim();
        if (item) {
          contextAndErrors.push(item);
        }
      }
    }
  }

  return {
    currentState,
    lastActiveAgent,
    updateDateTime,
    featureIdName,
    productionFiles,
    testFiles,
    contextAndErrors
  };
}

export function writeStateFile(filePath: string, data: StateData): void {
  const nowStr = new Date().toISOString().replace('T', ' ').substring(0, 16);
  
  // Format production files and test files
  const prodLines = data.productionFiles.length > 0 
    ? data.productionFiles.map(f => `- **Código de Producción:** ${f}`).join('\n')
    : '- **Código de Producción:** [Ruta del archivo]';

  const testLines = data.testFiles.length > 0 
    ? data.testFiles.map(f => `- **Código de Pruebas:** ${f}`).join('\n')
    : '- **Código de Pruebas:** [Ruta del archivo]';

  const contextLines = data.contextAndErrors.length > 0
    ? data.contextAndErrors.map(e => `- ${e}`).join('\n')
    : '- [Sin novedades en este ciclo.]';

  const newContent = `# Estado del Desarrollo (Handoff)

* **Estado Actual:** ${data.currentState}
* **Último Agente Activo:** ${data.lastActiveAgent}
* **Fecha/Hora de Actualización:** ${nowStr}

## Componentes y Archivos de Trabajo
- **ID y Nombre de Feature:** ${data.featureIdName}
- **Archivo de Estado (Handoff):** specs/${data.featureIdName}/current.md
- **Especificación Hard Spec:** specs/${data.featureIdName}/hard_spec.md
- **Contrato Gherkin:** features/${data.featureIdName}.feature
- **Log de TDD:** specs/${data.featureIdName}/tdd_log.md
- **Reporte de Auditoría:** specs/${data.featureIdName}/audit_report.md
- **Resultados de Mutación:** specs/${data.featureIdName}/mutation_results.json

## Archivos Modificados en este Ciclo
${prodLines}
${testLines}

## Contexto del Handoff / Errores Recientes
${contextLines}
`;

  fs.writeFileSync(filePath, newContent, 'utf-8');
}
