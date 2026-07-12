# Integración de Herdr en el Harness Multi-Agente

Esta especificación detalla cómo los agentes dentro del harness pueden interactuar con **Herdr** (un multiplexor de terminales y entorno de ejecución diseñado para agentes de codificación) para realizar tareas en paralelo, controlar paneles y coordinar comandos.

---

## ⚠️ Requisitos y Validación de Entorno

Para interactuar con Herdr, el entorno debe tener activa la variable `HERDR_ENV=1`. Antes de realizar cualquier acción de control, se debe validar si el agente está corriendo dentro de un panel gestionado por Herdr:

```bash
test "${HERDR_ENV:-}" = 1
```

- **Si falla:** El agente debe detenerse e informar que no se encuentra dentro de Herdr.
- **Si tiene éxito:** El binario `herdr` estará disponible en el `PATH` para comunicarse con la sesión activa.

---

## 🔍 Comandos de Autodescubrimiento (CLI)

El binario instalado es la autoridad para la sintaxis de comandos. Utiliza estos comandos para inspeccionar la CLI y descubrir subcomandos:

```bash
herdr --help
herdr pane
herdr workspace
herdr worktree
herdr tab
herdr wait
herdr terminal
herdr notification
herdr integration
herdr session
```

*Nota: La mayoría de los comandos devuelven respuestas en formato JSON. Se deben leer los IDs y estados desde estas respuestas en lugar de predecirlos.*

---

## 📌 Identificadores y Contexto Activo

Los IDs públicos son cadenas opacas y estables (ej: `w1` para workspaces, `w1:t1` para tabs, `w1:p1` para panes, `term_...` para terminales).
Herdr inyecta las siguientes variables de entorno para ubicar el contexto activo:

```bash
printf '%s\n' "$HERDR_WORKSPACE_ID" "$HERDR_TAB_ID" "$HERDR_PANE_ID"
```

Para descubrir el estado en vivo:
```bash
herdr workspace list
herdr tab list --workspace "$HERDR_WORKSPACE_ID"
herdr pane current --current
herdr pane list --workspace "$HERDR_WORKSPACE_ID"
```

---

## 🔄 Flujo de Trabajo: Lanzamiento Interactivo de Agentes

Para lanzar un agente interactivo en un panel hermano en el directorio actual sin perder el foco del panel original:

1. **Evaluar geometría del panel actual**:
   ```bash
   herdr pane layout --pane "$HERDR_PANE_ID"
   ```
2. **Dividir el panel** (derecha para pantallas anchas, abajo para pantallas estrechas):
   ```bash
   herdr pane split --current --direction right --no-focus
   ```
3. **Renombrar e iniciar el agente** (ej. usando `codex`, `claude`, etc.):
   ```bash
   herdr pane rename <returned-pane-id> "reviewer"
   herdr pane run <returned-pane-id> "codex"
   ```
4. **Esperar a que esté listo (estado `idle`) y enviar la tarea**:
   ```bash
   herdr wait agent-status <returned-pane-id> --status idle --timeout 30000
   herdr pane run <returned-pane-id> "Review the current diff and report only actionable findings."
   ```
5. **Esperar la finalización y leer los resultados**:
   ```bash
   herdr wait agent-status <returned-pane-id> --status working --timeout 30000
   herdr wait agent-status <returned-pane-id> --status done --timeout 120000
   herdr pane read <returned-pane-id> --source recent-unwrapped --lines 120
   ```

---

## ⚙️ Flujo de Trabajo: Ejecutar Comandos Ordinarios en Segundo Plano

Para ejecutar comandos rápidos o de verificación (como compilaciones o tests) en un panel hermano separado:

1. **Dividir el panel sin cambiar el foco**:
   ```bash
   herdr pane split --current --direction right --no-focus
   ```
2. **Ejecutar el comando y esperar salida**:
   ```bash
   herdr pane run <returned-pane-id> "just test"
   herdr wait output <returned-pane-id> --match "test result" --timeout 120000
   herdr pane read <returned-pane-id> --source recent-unwrapped --lines 120
   ```

---

## 📢 Feedback en Tiempo Real y Reporte de Estado (TUI/API)

Para ofrecer visibilidad en tiempo real al usuario de lo que los agentes están haciendo en paralelo (especialmente útil en la TUI visual de Herdr), los agentes deben utilizar los comandos de reporte dinámico utilizando la variable de entorno `$HERDR_PANE_ID` que inyecta Herdr de forma nativa en cada panel:

### 1. Reportar el Estado de Actividad del Agente (`report-agent`)
Permite definir el nombre del agente, su estado interno (`working`, `blocked`, `idle`) y una descripción de texto corta de lo que está haciendo (se actualiza dinámicamente en el encabezado de Herdr):
```bash
herdr pane report-agent "$HERDR_PANE_ID" \
  --source "gemini" \
  --agent "Backend-Agent" \
  --state "working" \
  --custom-status "TDD: Escribiendo Tests unitarios..."
```

### 2. Actualizar los Metadatos y Título del Panel (`report-metadata`)
Permite cambiar el título dinámico del panel de la terminal en Herdr para denotar fases del ciclo de vida (ej: TDD, auditoría, mutaciones) y complementar la visualización:
```bash
herdr pane report-metadata "$HERDR_PANE_ID" \
  --source "gemini" \
  --title "TDD Backend [ROJO]" \
  --custom-status "Compilando con Gradle..."
```

### 💡 Ejemplo de Estructura de Script con Feedback Dinámico:
```bash
#!/bin/bash
# Reporte Fase Inicial
herdr pane report-agent "$HERDR_PANE_ID" --source "gemini" --agent "TDD-Agent" --state "working" --custom-status "Investigando..."
herdr pane report-metadata "$HERDR_PANE_ID" --source "gemini" --title "Fase: Investigación"

# Ejecutar proceso
gemini --prompt "Analizar especificaciones..." --approval-mode yolo

# Reporte de TDD Activo
herdr pane report-agent "$HERDR_PANE_ID" --source "gemini" --agent "TDD-Agent" --state "working" --custom-status "Fase Verde: Codificando..."
herdr pane report-metadata "$HERDR_PANE_ID" --source "gemini" --title "Fase: TDD [VERDE]"

# Finalizar
herdr pane report-agent "$HERDR_PANE_ID" --source "gemini" --agent "TDD-Agent" --state "idle" --custom-status "Trabajo finalizado al 100%"
herdr pane report-metadata "$HERDR_PANE_ID" --source "gemini" --title "Estado: COMPLETADO"
```

---

## 🛡️ Reglas de Seguridad y Coordinación de Paneles

- Utiliza siempre `--no-focus` para tareas en segundo plano a menos que el usuario solicite explícitamente cambiar de contexto.
- Extrae siempre los identificadores reales de las salidas JSON recibidas.
- No cierres o alteres espacios de trabajo, pestañas o paneles que tu agente no haya creado.
- **NUNCA** ejecutes `herdr server stop` o detengas el servicio de Herdr unificado a menos que el usuario lo indique expresamente.
