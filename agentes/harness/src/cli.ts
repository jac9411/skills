import { runOrchestrator } from './orchestrator';
import { logError, logInfo, logSuccess } from './logger';
import * as childProcess from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

// Load environment variables from .env
import 'dotenv/config';

function printUsage() {
  console.log(`
Uso del Arnés Multi-Agente (Harness CLI):

  harness init "<Nombre de la Feature>" [--subtask <parent_seq>]
    Inicializa una nueva funcionalidad (feature) o subtarea.
    
  harness run <Ruta-a-specs-carpeta-o-current.md>
    Lanza el orquestador principal (Craftsman Lead) para gestionar la funcionalidad.

Ejemplos:
  harness init "Pago con tarjeta"
  harness init "Pago Backend" --subtask 1
  harness run specs/0001-pago_con_tarjeta/current.md
  harness run specs/0001-pago_con_tarjeta
`);
}

async function main() {
  const args = process.argv.slice(2);
  if (args.length === 0) {
    printUsage();
    process.exit(0);
  }

  const command = args[0];

  if (command === 'help' || command === '/help' || command === '--help' || command === '-h') {
    printUsage();
    process.exit(0);
  }

  if (command === 'init') {
    const featureName = args[1];
    if (!featureName) {
      logError('Falta el nombre de la funcionalidad.');
      printUsage();
      process.exit(1);
    }

    let subtaskArg = '';
    const subtaskIndex = args.indexOf('--subtask');
    if (subtaskIndex !== -1 && args[subtaskIndex + 1]) {
      subtaskArg = ` --subtask "${args[subtaskIndex + 1]}"`;
    }

    logInfo(`Inicializando funcionalidad "${featureName}" mediante create_task.py...`);
    try {
      // Execute the python script to keep numbering and template structures consistent
      const scriptPath = path.resolve(__dirname, '../scripts/create_task.py');
      const cmd = `python3 "${scriptPath}" "${featureName}"${subtaskArg}`;
      const output = childProcess.execSync(cmd, { encoding: 'utf-8' });
      console.log(output);
      logSuccess('Feature inicializada correctamente.');
    } catch (err: any) {
      logError(`Error inicializando la feature: ${err.stderr || err.message}`);
      process.exit(1);
    }

  } else if (command === 'run') {
    const targetPath = args[1];
    if (!targetPath) {
      logError('Falta la ruta de la feature o del archivo current.md.');
      printUsage();
      process.exit(1);
    }

    let stateFilePath = path.resolve(targetPath);
    // If the path is a folder, try to find current.md inside it
    if (fs.existsSync(stateFilePath) && fs.statSync(stateFilePath).isDirectory()) {
      stateFilePath = path.join(stateFilePath, 'current.md');
    }

    if (!fs.existsSync(stateFilePath)) {
      logError(`No se encontró el archivo current.md en la ruta provista: ${stateFilePath}`);
      process.exit(1);
    }

    try {
      await runOrchestrator(stateFilePath);
    } catch (err: any) {
      logError(`Error fatal de ejecución: ${err.message}`);
      process.exit(1);
    }

  } else {
    logError(`Comando no reconocido: "${command}"`);
    printUsage();
    process.exit(1);
  }
}

main().catch(err => {
  logError(`Excepción no controlada: ${err.message}`);
  process.exit(1);
});
