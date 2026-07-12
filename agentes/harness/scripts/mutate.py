#!/usr/bin/env python3
import os
import sys
import re
import subprocess
import shutil
import argparse

# Mapeo de operadores para mutación
MUTATION_MAP = {
    r'==': '!=',
    r'!=': '==',
    r'>=': '<',
    r'<=': '>',
    r'>': '<=',  # Evitamos que coincida dentro de >=
    r'<': '>=',  # Evitamos que coincida dentro de <=
    r'&&': '||',
    r'\|\|': '&&',
    r'\btrue\b': 'false',
    r'\bfalse\b': 'true',
    r'\bTRUE\b': 'FALSE',
    r'\bFALSE\b': 'TRUE'
}

def print_banner():
    print("=" * 60)
    print("     MUTATION TESTING ENGINE - HARNESS (Java/TypeScript)     ")
    print("=" * 60)

def restore_file(backup_path, target_path):
    if os.path.exists(backup_path):
        shutil.copyfile(backup_path, target_path)

def run_tests(test_command, dir_path=None):
    """
    Ejecuta el comando de pruebas y retorna True si pasan con éxito,
    o False si fallan (lo que significa que el mutante fue matado).
    """
    try:
        # Ejecutamos con output silenciado para no saturar la pantalla
        result = subprocess.run(
            test_command,
            shell=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            cwd=dir_path
        )
        return result.returncode == 0
    except Exception as e:
        # Si hay error de ejecución del proceso, asumimos que los tests no pasaron
        return False

def generate_mutations(content):
    """
    Genera posiciones y reemplazos de mutantes basados en expresiones regulares.
    """
    mutants = []
    # Buscamos cada operador soportado
    for pattern, replacement in MUTATION_MAP.items():
        for match in re.finditer(pattern, content):
            start, end = match.span()
            # Validamos que no sea parte de un comentario o un import
            line_start = content.rfind('\n', 0, start) + 1
            line = content[line_start:content.find('\n', start)]
            trimmed = line.strip()
            if trimmed.startswith('//') or trimmed.startswith('*') or trimmed.startswith('/*') or trimmed.startswith('import '):
                continue

            # Evitar mutar líneas estéticas, de clases CSS (className), estilos en línea o iconos
            if 'className' in trimmed or 'style={' in trimmed or 'style=' in trimmed or 'style:' in trimmed or 'Icon' in trimmed:
                continue
            
            # Evitar mutar tags HTML/JSX
            if pattern == r'<':
                next_char = content[end:end+1] if end < len(content) else ''
                if next_char.isalpha() or next_char in ['/', '!', '?']:
                    continue
            elif pattern == r'>':
                prev_char = content[start-1:start] if start > 0 else ''
                if prev_char.isalnum() or prev_char in ['/', '"', "'", '=', '-']:
                    continue
            
            mutants.append({
                'pattern': pattern,
                'replacement': replacement,
                'start': start,
                'end': end,
                'line_num': content.count('\n', 0, start) + 1,
                'line_text': trimmed
            })
    return mutants

def apply_mutation(content, mutant):
    return content[:mutant['start']] + mutant['replacement'] + content[mutant['end']:]

def main():
    parser = argparse.ArgumentParser(description="Harness Mutation Testing Engine para Java, Spring, jOOQ y TS React")
    parser.add_argument('--files', nargs='+', required=True, help="Lista de rutas de archivos de producción a mutar")
    parser.add_argument('--test-cmd', required=True, help="Comando exacto para correr las pruebas (ej: './gradlew test' o 'npm run test')")
    parser.add_argument('--dir', default=None, help="Directorio base para la ejecución del comando de pruebas")
    parser.add_argument('--out', default='specs/mutation_results.json', help="Ruta de destino del archivo JSON de resultados")
    
    args = parser.parse_args()
    print_banner()
    
    total_mutants = 0
    killed_mutants = 0
    survived_mutants = []
    
    for file_path in args.files:
        if not os.path.exists(file_path):
            print(f"[!] Error: El archivo {file_path} no existe.")
            continue
            
        print(f"\n[*] Analizando archivo: {file_path}")
        
        # Copia de seguridad absoluta para restauración segura
        backup_path = file_path + ".bak"
        shutil.copyfile(file_path, backup_path)
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                original_content = f.read()
                
            mutants = generate_mutations(original_content)
            print(f"[+] Se identificaron {len(mutants)} posibles candidatos de mutación.")
            
            for idx, mutant in enumerate(mutants, 1):
                total_mutants += 1
                print(f"  -> Aplicando Mutante #{total_mutants} (Línea {mutant['line_num']}): '{original_content[mutant['start']:mutant['end']]}' -> '{mutant['replacement']}'")
                
                # Crear el contenido mutado y escribirlo
                mutated_content = apply_mutation(original_content, mutant)
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write(mutated_content)
                    
                # Ejecutar pruebas
                tests_passed = run_tests(args.test_cmd, args.dir)
                
                if tests_passed:
                    print(f"     [!] ¡SUPERVIVIENTE! Las pruebas pasaron con éxito. El mutante no fue detectado.")
                    survived_mutants.append({
                        'mutant_id': f"M{total_mutants}",
                        'file': file_path,
                        'line': mutant['line_num'],
                        'original': original_content[mutant['start']:mutant['end']],
                        'mutated': mutant['replacement'],
                        'line_text': mutant['line_text']
                    })
                else:
                    print(f"     [✓] MUERTO. Las pruebas fallaron (o compilación fallida). Mutante anulado.")
                    killed_mutants += 1
                    
                # Restaurar inmediatamente
                restore_file(backup_path, file_path)
                
        finally:
            # Restaurar el archivo original siempre en caso de excepción
            restore_file(backup_path, file_path)
            if os.path.exists(backup_path):
                os.remove(backup_path)

    print("\n" + "=" * 60)
    print("                      RESULTADOS FINALES                      ")
    print("=" * 60)
    print(f"Total de Mutantes Generados: {total_mutants}")
    print(f"Mutantes Muertos (Tests correctos): {killed_mutants}")
    print(f"Mutantes Supervivientes (Falta cobertura): {len(survived_mutants)}")
    
    if total_mutants > 0:
        kill_rate = (killed_mutants / total_mutants) * 100
        print(f"Tasa de Matanza (Kill Rate): {kill_rate:.2f}%")
    else:
        print("Tasa de Matanza (Kill Rate): N/A (0 mutantes)")
        
    print("=" * 60)
    
    # Escribir resultado a JSON estructurado para consumo fácil por el agente Mutation Tester
    import json
    report_data = {
        'total_mutants': total_mutants,
        'killed': killed_mutants,
        'survived_count': len(survived_mutants),
        'kill_rate': (killed_mutants / total_mutants * 100) if total_mutants > 0 else 0,
        'survived': survived_mutants
    }
    # Asegurar que el directorio de salida existe
    out_dir = os.path.dirname(args.out)
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)
        
    with open(args.out, 'w', encoding='utf-8') as rf:
        json.dump(report_data, rf, indent=2, ensure_ascii=False)
    print(f"[*] Resultados estructurados exportados a {args.out}")
    
    if len(survived_mutants) > 0:
        sys.exit(1) # Código de error si hay supervivientes
    else:
        sys.exit(0)

if __name__ == "__main__":
    main()
