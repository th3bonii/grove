# Ejemplo: GROVE Ralph Loop

## Escenario
Tienes una especificación completa y quieres implementarla automáticamente.

## Input (SPEC.md, DESIGN.md, TASKS.md)
Ya tienes los artefactos generados por grove-spec que describen exactamente qué construir.

## Cómo ejecutar
```bash
cd examples/02-ralph-loop
grove loop ./SPEC.md --tasks ./TASKS.md --output ./src
```

## Proceso
1. Ralph Loop lee SPEC.md y TASKS.md
2. Genera código automáticamente para cada tarea
3. Implementa la arquitectura definida en DESIGN.md
4. Ejecuta tests unitarios automáticamente
5. Reporta progreso y errores

## Output esperado
- Código fuente completo en ./src
- Tests unitarios ejecutados
- Informe de completion