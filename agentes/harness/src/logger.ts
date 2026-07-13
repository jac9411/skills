import { red, green, yellow, cyan, gray, bold } from 'colorette';

export function formatState(state: string): string {
  const greenStates = ['spec_approved', 'designed', 'gherkin_generated', 'human_approved', 'tdd_completed', 'audit_passed', 'mutation_passed', 'completed', 'done'];
  const redStates = ['pending', 'audit_failed', 'mutation_failed'];
  
  if (greenStates.includes(state)) {
    return bold(green(state));
  } else if (redStates.includes(state)) {
    return bold(red(state));
  }
  return bold(yellow(state));
}

export function logInfo(message: string): void {
  console.log(`${cyan('ℹ')} ${message}`);
}

export function logSuccess(message: string): void {
  console.log(`${green('✔')} ${bold(green(message))}`);
}

export function logWarning(message: string): void {
  console.log(`${yellow('⚠')} ${message}`);
}

export function logError(message: string): void {
  console.log(`${red('✖')} ${bold(red(message))}`);
}

export function logPending(message: string): void {
  console.log(`${red('⏳')} ${bold(red(message))}`);
}

export function logTestResult(testName: string, passed: boolean, details?: string): void {
  if (passed) {
    console.log(`${green('✔')} PRUEBA SUPERADA: [${green(testName)}]`);
  } else {
    console.log(`${red('✖')} PRUEBA FALLIDA: [${red(testName)}]`);
    if (details) {
      console.log(gray(details.split('\n').map(line => `  ${line}`).join('\n')));
    }
  }
}

export function logHeader(title: string): void {
  const line = '='.repeat(60);
  console.log(`\n${cyan(line)}`);
  console.log(`${bold(cyan(`        ${title.toUpperCase()}        `))}`);
  console.log(`${cyan(line)}\n`);
}

export function logStep(agentName: string, stepDescription: string): void {
  console.log(`[${cyan(agentName)}] ${gray('→')} ${stepDescription}`);
}
