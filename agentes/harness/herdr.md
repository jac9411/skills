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

## 🛡️ Reglas de Seguridad y Coordinación de Paneles

- Utiliza siempre `--no-focus` para tareas en segundo plano a menos que el usuario solicite explícitamente cambiar de contexto.
- Extrae siempre los identificadores reales de las salidas JSON recibidas.
- No cierres o alteres espacios de trabajo, pestañas o paneles que tu agente no haya creado.
- **NUNCA** ejecutes `herdr server stop` o detengas el servicio de Herdr unificado a menos que el usuario lo indique expresamente.
