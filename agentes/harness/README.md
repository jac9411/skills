# Guía del Entorno de Desarrollo Multi-Agente

Este directorio contiene las especificaciones, perfiles y System Prompts para cada uno de los agentes especializados en el ciclo de desarrollo guiado por comportamiento (BDD), desarrollo guiado por pruebas (TDD) en micro-ciclos, auditoría de cobertura automática y resistencia de mutación.

---

## 🛠️ Stack Tecnológico de Destino (Reglas para la IA)

Cualquier IA que asuma un rol de desarrollo o auditoría en este proyecto debe ceñirse estrictamente a las tecnologías del proyecto, buscará en su documentación las decisiones técnicas sobre la tecnología adoptada, en caso de no existir esa documentación por defecto usará:

- **Backend:** Java 17+, Spring Framework (Boot, Security, etc.), jOOQ (acceso a datos seguro con tipos generados).
  - Herramienta de compilación: Gradle (`./gradlew classes`, `./gradlew test`).
- **Frontend:** TypeScript, React 19, componentes de `shadcn/ui` (con estilos CSS Vainilla preferentemente, evitando TailwindCSS a menos que se solicite).
  - Herramienta de construcción: npm (`npm run build`, `npm run test` / `vitest run`, `tsc`).
  - **Internacionalización Obligatoria (i18n):** Todo componente visual debe traducir dinámicamente sus textos literales utilizando archivos de localización dinámicos. Quedan prohibidos los textos estáticos. Idiomas obligatorios: **Español (es) y Gallego (gl)**.

---

## 🔄 Flujo Operativo y Máquina de Estados (Ciclo de Vida)

El ciclo de vida de desarrollo de cualquier funcionalidad (feature) progresa de manera rigurosa de extremo a extremo a través de las siguientes etapas obligatorias, coordinadas por el **Craftsman Lead**:

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
| **0. Inicio**    | `pending`                             | Ninguno              | feature creada en estado inicial de espera.                                                                                                                    |
| **1. Debate**    | `spec_approved`                       | `Spec Partner`       | Refina requisitos en Grill Mode; guarda la especificación técnica refinada en `specs/[XXXX]/hard_spec.md`.                                                     |
| **2. Contrato**  | `gherkin_generated`                   | `Gherkin Author`     | Traduce la spec a escenarios de negocio Dado-Cuando-Entonces numerados (`@S1`, `@S2`...) en `features/[XXXX].feature`.                                         |
| **3. Puerta**    | `human_approved`                      | `Craftsman Lead`     | Detiene el flujo y pide confirmación manual ("Aprobar") al desarrollador en terminal.                                                                          |
| **4. TDD**       | `tdd_completed`                       | `TDD Craftsman`      | Desarrolla incrementalmente el código y tests mediante las tres leyes de TDD; registra los pasos en `specs/[XXXX]/tdd_log.md`.                                 |
| **5. Auditoría** | `audit_passed` / `audit_failed`       | `Judge`              | Audita correspondencia 1:1 de escenarios con tests, i18n, refactorización real y cero hacks. Genera `specs/[XXXX]/audit_report.md`. Si falla, retrocede a TDD. |
| **6. Mutación**  | `mutation_passed` / `mutation_failed` | `Mutation Tester`    | Ejecuta el script `./harness/scripts/mutate.py`. Si hay supervivientes lógicos (tests débiles), retrocede a TDD con detalles de corrección.                    |
| **7. Éxito**     | `done`                                | `Craftsman Lead`     | Feature finalizada con éxito, cero supervivientes y código de la máxima calidad técnica.                                                                       |

---

## ⚡ Automatización de Tareas (Scripts de Apoyo)

### 1. Inicialización de Tareas (`create_task.py`)

El script automatiza la inicialización limpia de nuevas características, escaneando secuencialmente el directorio `specs/` y configurando el handoff de control:

```bash
./harness/scripts/create_task.py "Nombre de la Feature"
```

_Impacto:_ Crea de forma automática la subcarpeta en `specs/` e inicializa `current.md` en estado `pending`.

### 2. Pruebas de Resistencia y Mutación de Operadores (`mutate.py`)

El motor de mutación altera intencionalmente operadores lógicos (`==`, `!=`, `<`, `>`, `&&`, `||`, `true`, `false`) para validar la solidez de las pruebas del proyecto.

```bash
./harness/scripts/mutate.py --files src/main/java/Servicio.java --test-cmd "./gradlew test" --out specs/0001-gestion_usuario/mutation_results.json
```

_Reglas de optimización para la IA:_

- **Filtro Inteligente de UI:** El script descarta mutar líneas estéticas (propiedades `className`, estilos inline o importaciones de iconos) para ahorrar tiempo y evitar falsos supervivientes.
- **Tests Quirúrgicos:** Para acelerar las pruebas, configure siempre el comando exacto del test unitario de la clase mutada (ej: `npm run test -- Componente.test.tsx` o `./gradlew test --tests "*ServicioTest*"`), reduciendo el tiempo de mutación de minutos a segundos.
