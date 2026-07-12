# Especificación Técnica Refinada (Hard Spec) - [ID_Feature] - [Nombre]

## 1. Casos de Uso y Flujos de Negocio
- **Flujo Principal (Happy Path):** [Paso a paso de la funcionalidad]
- **Flujos Alternativos / Validaciones:** [Qué ocurre si fallan las condiciones previas]

## 2. Modelo de Datos e Impacto en Base de Datos (jOOQ)
- **Tablas Modificadas/Creadas:** [Esquema SQL, tipos de datos, constraints, índices]
- **Tipos jOOQ Requeridos:** [Clases generadas por jOOQ que se emplearán en el backend]

## 3. Contrato de API Spring REST
- **Endpoints:**
  - `VERBO /api/v1/...`
- **Payload Petición (Request):** [JSON schema]
- **Payload Respuesta Exitosa (Response):** [JSON schema, código HTTP]
- **Manejo de Errores (Error Responses):** [JSON de error, códigos HTTP, ej: 400 Bad Request, 404 Not Found]

## 4. Comportamiento y Componentes de Interfaz (React 19 / shadcn)
- **Componentes Creados/Modificados:** [Estructura, props, estado local]
- **Control de Errores e Indicadores de Carga:** [Feedback visual en la UI en caso de carga o peticiones fallidas]

## 5. Matriz de Internacionalización e Idiomas (i18n)
*Todos los textos fijos de la pantalla deben mapearse a claves y traducirse íntegramente:*

| Clave de Traducción | Español (es) | Gallego (gl) |
|---------------------|--------------|--------------|
| `title_feature`     | ...          | ...          |
| `error_validation`  | ...          | ...          |
