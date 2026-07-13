import * as fs from 'fs';
import * as path from 'path';
import * as childProcess from 'child_process';
import * as readline from 'readline';
import { Type, FunctionDeclaration } from '@google/genai';
import { logInfo, logSuccess, logError, logTestResult } from './logger';

export interface CommandResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

// 1. Filesystem Tools
export function readFileTool(filePath: string): string {
  logInfo(`Herramienta de Agente: Leyendo archivo "${filePath}"`);
  const resolved = path.resolve(filePath);
  if (!fs.existsSync(resolved)) {
    throw new Error(`El archivo no existe: ${filePath}`);
  }
  return fs.readFileSync(resolved, 'utf-8');
}

export function writeFileTool(filePath: string, content: string): string {
  logInfo(`Herramienta de Agente: Escribiendo archivo "${filePath}"`);
  const resolved = path.resolve(filePath);
  const dir = path.dirname(resolved);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  fs.writeFileSync(resolved, content, 'utf-8');
  return `Archivo escrito correctamente en ${filePath}`;
}

export function replaceInFileTool(filePath: string, oldString: string, newString: string): string {
  logInfo(`Herramienta de Agente: Reemplazando contenido en "${filePath}"`);
  const resolved = path.resolve(filePath);
  if (!fs.existsSync(resolved)) {
    throw new Error(`File does not exist: ${filePath}`);
  }
  const content = fs.readFileSync(resolved, 'utf-8');
  if (!content.includes(oldString)) {
    throw new Error(`La cadena objetivo a reemplazar no se encontró en ${filePath}. Asegúrate de haber proporcionado una coincidencia exacta.`);
  }
  
  // Replace only the first occurrence to avoid collateral changes
  const newContent = content.replace(oldString, newString);
  fs.writeFileSync(resolved, newContent, 'utf-8');
  return `Contenido reemplazado correctamente en ${filePath}`;
}

export function listFilesTool(dirPath: string): string[] {
  logInfo(`Herramienta de Agente: Listando directorio "${dirPath}"`);
  const resolved = path.resolve(dirPath);
  if (!fs.existsSync(resolved) || !fs.statSync(resolved).isDirectory()) {
    throw new Error(`No es un directorio: ${dirPath}`);
  }
  return fs.readdirSync(resolved);
}

// 2. Command Tool
export function runCommandTool(command: string): CommandResult {
  logInfo(`Herramienta de Agente: Ejecutando comando de terminal "${command}"`);
  try {
    const result = childProcess.execSync(command, { encoding: 'utf-8', stdio: 'pipe' });
    const isTest = command.includes('test') || command.includes('mutate') || command.includes('vitest');
    if (isTest) {
      logTestResult(command, true);
    }
    return {
      stdout: result,
      stderr: '',
      exitCode: 0
    };
  } catch (err: any) {
    const isTest = command.includes('test') || command.includes('mutate') || command.includes('vitest');
    if (isTest) {
      logTestResult(command, false, err.stderr || err.message);
    }
    return {
      stdout: err.stdout || '',
      stderr: err.stderr || err.message || '',
      exitCode: err.status ?? 1
    };
  }
}

// 3. Human Interaction Tool
export function askHumanTool(question: string): Promise<string> {
  logInfo(`Herramienta de Agente: Solicitando entrada al humano en la terminal...`);
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  return new Promise((resolve) => {
    rl.question(`\n[?] ${question}\n> `, (answer) => {
      rl.close();
      resolve(answer.trim());
    });
  });
}

// Gemini Function Declarations Schema for @google/genai
export const toolDeclarations: FunctionDeclaration[] = [
  {
    name: 'readFile',
    description: 'Reads the exact text content of a file.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        filePath: { type: Type.STRING, description: 'The relative path of the file to read.' }
      },
      required: ['filePath']
    }
  },
  {
    name: 'writeFile',
    description: 'Writes the given content to a file, creating any parent directory if necessary.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        filePath: { type: Type.STRING, description: 'The relative path of the file to write.' },
        content: { type: Type.STRING, description: 'The full text content to write.' }
      },
      required: ['filePath', 'content']
    }
  },
  {
    name: 'replaceInFile',
    description: 'Surgically replaces a specific exact string with another string inside a file.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        filePath: { type: Type.STRING, description: 'The relative path of the file.' },
        oldString: { type: Type.STRING, description: 'The exact literal text to replace.' },
        newString: { type: Type.STRING, description: 'The exact literal text to insert instead.' }
      },
      required: ['filePath', 'oldString', 'newString']
    }
  },
  {
    name: 'listFiles',
    description: 'Lists all files and directories inside a given folder.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        dirPath: { type: Type.STRING, description: 'The relative path of the folder.' }
      },
      required: ['dirPath']
    }
  },
  {
    name: 'runCommand',
    description: 'Executes a bash shell command inside the workspace and returns its stdout, stderr, and exit code. Useful for compiling, testing, etc.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        command: { type: Type.STRING, description: 'The shell command to execute.' }
      },
      required: ['command']
    }
  },
  {
    name: 'askHuman',
    description: 'Pauses execution to ask the human user a question or clarify requirements and waits for their response in the terminal.',
    parameters: {
      type: Type.OBJECT,
      properties: {
        question: { type: Type.STRING, description: 'The clear, detailed question to ask the human.' }
      },
      required: ['question']
    }
  }
];
