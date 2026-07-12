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

def main():
    parser = argparse.ArgumentParser(description="Inicializador de nuevas funcionalidades para el entorno Harness.")
    parser.add_argument('name', nargs='?', default=None, help="Nombre de la nueva feature (ej: 'Pago con tarjeta' o 'gestion_usuarios')")
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
    next_seq = get_next_sequence()
    feature_id_name = f"{next_seq:04d}-{cleaned_name}"
    
    feature_dir = os.path.join('specs', feature_id_name)
    os.makedirs(feature_dir, exist_ok=True)
    
    current_md_path = os.path.join(feature_dir, 'current.md')
    now_str = datetime.now().strftime('%Y-%m-%d %H:%M')
    
    current_content = TEMPLATE_CURRENT.format(
        datetime=now_str,
        feature_id_name=feature_id_name
    )
    
    with open(current_md_path, 'w', encoding='utf-8') as f:
        f.write(current_content)
        
    # Crear también la carpeta features por comodidad si no existe
    os.makedirs('features', exist_ok=True)
    
    print("=" * 60)
    print("        ¡FUNCIONALIDAD INICIALIZADA CON ÉXITO!        ")
    print("=" * 60)
    print(f"[+] Carpeta de Especificaciones:  {feature_dir}/")
    print(f"[+] Archivo de Estado (Handoff):  {current_md_path}")
    print(f"[+] Ruta Futura de Gherkin:       features/{feature_id_name}.feature")
    print("-" * 60)
    print("[*] Siguiente Paso:")
    print(f"    Lanza al agente Spec Partner sobre la carpeta de esta feature")
    print("    para debatir y refinar el diseño del comportamiento.")
    print("=" * 60)

if __name__ == "__main__":
    main()
