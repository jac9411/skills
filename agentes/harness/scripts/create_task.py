#!/usr/bin/env python3
import os
import sys
import re
import argparse
from datetime import datetime

TEMPLATE_CURRENT = """# Estado del Desarrollo (Handoff)

* **Estado Actual:** pending
* **Último Agente Activo:** Ninguno
* **Fecha/Hora de Actualización:** {datetime}

## Componentes y Archivos de Trabajo
- **ID y Nombre de Feature:** {feature_id_name}
- **Archivo de Estado (Handoff):** specs/{feature_id_name}/current.md
- **Especificación Hard Spec:** specs/{feature_id_name}/hard_spec.md
- **Contrato Gherkin:** features/{feature_id_name}.feature
- **Log de TDD:** specs/{feature_id_name}/tdd_log.md
- **Reporte de Auditoría:** specs/{feature_id_name}/audit_report.md
- **Reporte de Mutación:** specs/{feature_id_name}/mutation_report.md

## Contexto del Handoff / Errores Recientes
- [Inicio del proyecto. Esperando lanzamiento del Spec Partner para comenzar el debate de la feature {feature_id_name}.]
"""

TEMPLATE_SUBTASK = """# Estado del Desarrollo (Handoff)

* **Estado Actual:** pending
* **Último Agente Activo:** Ninguno
* **Fecha/Hora de Actualización:** {datetime}

## Componentes y Archivos de Trabajo
- **ID y Nombre de Feature:** {parent_id_name}/{subtask_name}
- **Archivo de Estado (Handoff):** specs/{parent_id_name}/{subtask_name}/current.md
- **Especificación Hard Spec:** specs/{parent_id_name}/hard_spec.md
- **Contrato Gherkin:** features/{parent_id_name}.feature
- **Log de TDD:** specs/{parent_id_name}/{subtask_name}/tdd_log.md
- **Reporte de Auditoría:** specs/{parent_id_name}/{subtask_name}/audit_report.md
- **Reporte de Mutación:** specs/{parent_id_name}/{subtask_name}/mutation_results.json

## Contexto del Handoff / Errores Recientes
- [Inicio del proyecto de subtarea. Esperando lanzamiento del desarrollador en el sub-entorno de la feature {parent_id_name}/{subtask_name}.]
"""

def clean_name(name):
    # Convertir a minúsculas, reemplazar espacios y guiones por guiones bajos
    name = name.lower().strip()
    name = re.sub(r'[\s\-]+', '_', name)
    # Eliminar caracteres especiales no amigables para nombres de archivo
    name = re.sub(r'[^a-z0-9_áéíóúüñ]', '', name)
    return name

def get_next_sequence():
    specs_dir = 'specs'
    if not os.path.exists(specs_dir):
        os.makedirs(specs_dir, exist_ok=True)
        return 1
        
    folders = os.listdir(specs_dir)
    max_seq = 0
    for folder in folders:
        match = re.match(r'^(\d{4})[-_]', folder)
        if match:
            seq = int(match.group(1))
            if seq > max_seq:
                max_seq = seq
    return max_seq + 1

def find_parent_directory(seq_num):
    specs_dir = 'specs'
    if not os.path.exists(specs_dir):
        return None
    prefix = f"{seq_num:04d}-"
    for folder in os.listdir(specs_dir):
        if folder.startswith(prefix) and os.path.isdir(os.path.join(specs_dir, folder)):
            return folder
    return None

def main():
    parser = argparse.ArgumentParser(description="Inicializador de nuevas funcionalidades para el entorno Harness.")
    parser.add_argument('name', nargs='?', default=None, help="Nombre de la nueva feature (ej: 'Pago con tarjeta' o 'gestion_usuarios')")
    parser.add_argument('--subtask', type=str, default=None, help="Número de secuencia de la tarea padre para crear una subtarea (ej: 1 o 0001)")
    args = parser.parse_args()
    
    feature_name = args.name
    if not feature_name:
        try:
            feature_name = input("[?] Introduce el nombre de la nueva funcionalidad (feature): ").strip()
        except KeyboardInterrupt:
            print("\n[!] Operación cancelada por el usuario.")
            sys.exit(0)
            
    if not feature_name:
        print("[!] Error: El nombre de la funcionalidad no puede estar vacío.")
        sys.exit(1)
        
    cleaned_name = clean_name(feature_name)
    now_str = datetime.now().strftime('%Y-%m-%d %H:%M')
    
    if args.subtask:
        try:
            seq_val = int(args.subtask)
        except ValueError:
            print(f"[!] Error: El valor de --subtask debe ser un número entero (recibido: '{args.subtask}')")
            sys.exit(1)
            
        parent_folder = find_parent_directory(seq_val)
        if not parent_folder:
            print(f"[!] Error: No se encontró ninguna tarea raíz con el prefijo {seq_val:04d}- en la carpeta 'specs/'.")
            print("    Asegúrate de inicializar primero la tarea raíz.")
            sys.exit(1)
            
        feature_dir = os.path.join('specs', parent_folder, cleaned_name)
        os.makedirs(feature_dir, exist_ok=True)
        
        current_md_path = os.path.join(feature_dir, 'current.md')
        current_content = TEMPLATE_SUBTASK.format(
            datetime=now_str,
            parent_id_name=parent_folder,
            subtask_name=cleaned_name
        )
    else:
        next_seq = get_next_sequence()
        feature_id_name = f"{next_seq:04d}-{cleaned_name}"
        feature_dir = os.path.join('specs', feature_id_name)
        os.makedirs(feature_dir, exist_ok=True)
        
        current_md_path = os.path.join(feature_dir, 'current.md')
        current_content = TEMPLATE_CURRENT.format(
            datetime=now_str,
            feature_id_name=feature_id_name
        )
        
    with open(current_md_path, 'w', encoding='utf-8') as f:
        f.write(current_content)
        
    os.makedirs('features', exist_ok=True)
    
    print("=" * 60)
    print("        ¡FUNCIONALIDAD INICIALIZADA CON ÉXITO!        ")
    print("=" * 60)
    print(f"[+] Carpeta de Especificaciones:  {feature_dir}/")
    print(f"[+] Archivo de Estado (Handoff):  {current_md_path}")
    if args.subtask:
        print(f"[+] Especificación Hard Spec (Padre): specs/{parent_folder}/hard_spec.md")
        print(f"[+] Ruta Compartida Gherkin (Padre):  features/{parent_folder}.feature")
    else:
        print(f"[+] Ruta Futura de Gherkin:       features/{feature_id_name}.feature")
    print("-" * 60)
    print("[*] Siguiente Paso:")
    if args.subtask:
        print(f"    Lanza al agente correspondiente (ej: TDD Craftsman) sobre")
        print(f"    esta subtarea para comenzar con el desarrollo independiente.")
    else:
        print(f"    Lanza al agente Spec Partner sobre la carpeta de esta feature")
        print("    para debatir y refinar el diseño del comportamiento.")
    print("=" * 60)

if __name__ == "__main__":
    main()