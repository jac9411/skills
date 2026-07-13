# Guía del Entorno de Desarrollo Multi-Agente y Orquestador TypeScript

Este directorio contiene la especificación completa del ciclo de vida multi-agente, junto con el **Orquestador CLI nativo en TypeScript**, diseñado para coordinar de forma automática y visual a los distintos agentes especialistas en el desarrollo BDD/TDD, auditoría de código y pruebas de mutación.

---

## 🛠️ Stack Tecnológico de Destino (Reglas para la IA)

Cualquier IA que asuma un rol de desarrollo o auditoría en este proyecto debe ceñirse estrictamente a las tecnologías del proyecto:

- **Backend:** Java 17+, Spring Framework (Boot, Security, etc.), jOOQ (acceso a datos seguro con tipos generados).
  - Herramienta de compilación: Gradle (`./gradlew classes`, `./gradlew test`).
- **Frontend:** TypeScript, React 19, componentes de `shadcn/ui` (con estilos CSS Vainilla preferentemente, evitando TailwindCSS a menos que se solicite).
  - Herramienta de construcción: npm (`npm run build`, `npm run test` / `vitest run`, `tsc`).
  - **Internacionalización Obligatoria (i18n):** Todo componente visual debe traducir dinámicamente sus textos literales utilizando archivos de localización dinámicos. Quedan prohibidos los textos estáticos. Idiomas obligatorios: **Español (es) y Gallego (gl)**.

---

## 🔄 Flujo Operativo y Máquina de Estados (Ciclo de Vida)

El ciclo de vida progresará de manera rigurosa coordinado automáticamente por el orquestador (**Craftsman Lead**):

```text
[Inicio] -> 1. Spec Partner (Debate) -> Genera specs/[XXXX]/hard_spec.md
                               |
                               v
                     2. Gherkin Author -> Genera features/[XXXX].feature
                               |
                               v
                     3. PUERTA HUMANA (Confirmación interactiva en terminal)
                               | (Aprobado)
                               v
   +---------------> 4. TDD Craftsman (Ciclo incremental TDD)
   |                           |
   |                           v
   |                 5. Judge (Auditoría de correspondencia y refactor)
   |                           |
   |            (Fallo)        +---> [audit_passed]
   |  <------------------------+           |
   |                                       v
   |                             6. Mutation Tester (Resistencia de Tests)
   |                                       |
   |            (Supervivientes)           +---> [Cero Supervivientes]
   +---------------------------------------+           |
                                                       v
                                                    [DONE]
```

### Tabla de Transición de Estados (`current.md`):

| Etapa            | Estado en `current.md`                | Último Agente Activo | Acción / Artefacto Generado                                                                                                                                    |
| :--------------- | :------------------------------------ | :------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **0. Inicio**    | `pending`                             | Ninguno              | Feature creada en estado inicial de espera.                                                                                                                    |
| **1. Debate**    | `spec_approved`                       | `Spec Partner`       | Refina requisitos en Grill Mode; guarda la especificación en `specs/[XXXX]/hard_spec.md`.                                                                      |
| **2. Contrato**  | `gherkin_generated`                   | `Gherkin Author`     | Traduce la spec a escenarios Dado-Cuando-Entonces numerados (`@S1`, `@S2`...) en `features/[XXXX].feature`.                                                    |
| **3. Puerta**    | `human_approved`                      | `Craftsman Lead`     | Detiene el flujo y pide confirmación manual (`Aprobar`) al desarrollador en terminal.                                                                          |
| **4. TDD**       | `tdd_completed`                       | `TDD Craftsman`      | Desarrolla incrementalmente el código y tests mediante las tres leyes de TDD; registra los pasos en `specs/[XXXX]/tdd_log.md`.                                 |
| **5. Auditoría** | `audit_passed` / `audit_failed`       | `Judge`              | Audita correspondencia 1:1 de escenarios con tests, i18n, refactorización real y cero hacks. Genera `specs/[XXXX]/audit_report.md`. Si falla, retrocede a TDD. |
| **6. Mutación**  | `mutation_passed` / `mutation_failed` | `Mutation Tester`    | Ejecuta el script `./harness/scripts/mutate.py`. Si hay supervivientes lógicos (tests débiles), retrocede a TDD con detalles de corrección.                    |
| **7. Éxito**     | `done`                                | `Craftsman Lead`     | Feature finalizada con éxito, cero supervivientes y código de la máxima calidad técnica.                                                                       |

---

## ⚡ Orquestador CLI en TypeScript (Instrucciones de Uso)

El orquestador automatiza por completo las llamadas de agentes y transiciones de estado, interactuando con el LLM (Gemini) y proveyendo herramientas locales a los sub-agentes para la lectura/escritura de archivos y ejecución de pruebas.

### 📋 Prerrequisitos y Configuración

1. Instalar las dependencias del proyecto:
   ```bash
   npm install
   ```
2. Asegurarse de tener una clave de API válida configurada en tu entorno:
   ```bash
   export GEMINI_API_KEY="tu-api-key"
   ```
   *(Opcional) Puedes definir un archivo `.env` en la raíz de este directorio con la variable `GEMINI_API_KEY=tu-api-key`.*

---

### 🚀 Comandos del Orquestador

El orquestador se ejecuta utilizando `npx tsx` directamente sobre los módulos de entrada:

#### 1. Inicializar una Funcionalidad (`init`)
Crea de manera automática la estructura de carpetas en `specs/` e inicializa el archivo `current.md` en estado `pending`.
```bash
npx tsx src/cli.ts init "Nombre de la Funcionalidad"
```
**Ejemplo de salida en terminal:**
```text
ℹ Inicializando funcionalidad "Pago con tarjeta" mediante create_task.py...
============================================================
        ¡FUNCIONALIDAD INICIALIZADA CON ÉXITO!        
============================================================
[+] Carpeta de Especificaciones:  specs/0002-pago_con_tarjeta/
[+] Archivo de Estado (Handoff):  specs/0002-pago_con_tarjeta/current.md
[+] Ruta Futura de Gherkin:       features/0002-pago_con_tarjeta.feature
------------------------------------------------------------
```

*Para tareas Frontend y Backend independientes en paralelo, puedes bifurcar la subtarea:*
```bash
# Para el Backend:
npx tsx src/cli.ts init "Pago Backend" --subtask 2

# Para el Frontend:
npx tsx src/cli.ts init "Pago Frontend" --subtask 2
```

#### 2. Arrancar y Correr la Orquestación (`run`)
Inicia el bucle de la máquina de estados. Lee la situación actual en `current.md`, selecciona e invoca al agente correspondiente con el LLM y gestiona las retroalimentaciones.
```bash
npx tsx src/cli.ts run specs/0002-pago_con_tarjeta
```
*(También puedes pasar la ruta directa al archivo `current.md`: `npx tsx src/cli.ts run specs/0002-pago_con_tarjeta/current.md`)*

---

### 🎨 Consola Coloreada y Test Feedback
El orquestador analiza la terminal de forma interactiva y tiñe las salidas para una inspección rápida en vivo:
- ⏳ **Rojo en Negrita (`pending` / `audit_failed` / `mutation_failed`):** Denota estados en espera, devoluciones o errores del sistema.
- ✔️ **Verde en Negrita (`spec_approved` / `tdd_completed` / `done` / etc.):** Denota etapas finalizadas con éxito.
- ✖️ **TEST FAILED / ✔️ TEST PASSED:** Los resultados de las compilaciones y suites de pruebas ejecutadas por los sub-agentes se interceptan en consola y se pintan con colores específicos.

---

### 🛡️ Reglas y Protecciones del Orquestador

1. **Puerta Humana Interactiva:** Al generarse el contrato Gherkin, la CLI detiene el flujo y requiere que el operador apruebe explícitamente (`Aprobar` / `si`) para habilitar el desarrollo TDD.
2. **Protección de Bucles Infinitos:** Si se detecta un ciclo repetitivo de fallos entre el TDD Craftsman, el Judge o el Mutation Tester más de **3 veces consecutivas**, el orquestador bloquea la ejecución automática y solicita asistencia humana en consola.
3. **Credenciales:** Protege de forma estricta las claves de API (nunca se imprimen ni se loguean en consola).

---

## ⚡ Scripts Auxiliares de Soporte

### 1. Pruebas de Resistencia y Mutación (`mutate.py`)
Utilizado automáticamente por el agente **Mutation Tester** para alterar operadores lógicos y certificar la dureza de los tests:
```bash
./harness/scripts/mutate.py --files src/main/java/Servicio.java --test-cmd "./gradlew test" --out specs/0001-gestion_usuario/mutation_results.json
```

---

## 🌐 Integración Unificada con Herdr

Si utilizas el orquestador dentro de un entorno gestionado por **Herdr** (`HERDR_ENV=1`), el Craftsman Lead de forma nativa:
- Crea pestañas ordenadas automáticamente con el formato `"{Prefijo_Numérico} - {Nombre_Agente}"`.
- Reporta el estado dinámico del panel actual (ej. `working`, `blocked` o `idle`) con un texto de actividad descriptivo en la cabecera del multiplexor.
- Actualiza dinámicamente los metadatos y colores del título del terminal del panel hermano.

Para más detalles consulta: 👉 **[Guía de Integración de Herdr (harness/herdr.md)](./herdr.md)**
