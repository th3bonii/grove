# Ejemplo: GROVE Ralph Loop

## Escenario
Tienes una especificación completa (SPEC.md, DESIGN.md, TASKS.md) y quieres implementar automáticamente todo el código.

## Estructura del Ejemplo

```
02-ralph-loop/
├── input/                    # Archivos de entrada (specs)
│   ├── SPEC.md               # Especificación funcional
│   ├── DESIGN.md             # Diseño técnico
│   └── TASKS.md              # Checklist de tareas
├── expected/                  # Output esperado después de Ralph Loop
│   └── src/                  # Código fuente generado
│       ├── app.ts
│       ├── config/env.ts
│       ├── models/todoModel.ts
│       ├── repositories/todoRepository.ts
│       ├── services/todoService.ts
│       ├── controllers/todoController.ts
│       ├── routes/todoRoutes.ts
│       └── middleware/auth.ts
├── AGENTS.md                  # Configuración de agentes
└── README.md                  # Este archivo
```

## Input (input/)
Los archivos en `input/` representan lo que tendrías después de ejecutar `grove-spec`:
- **SPEC.md**: API REST de Tareas con autenticación JWT
- **DESIGN.md**: Arquitectura Express + PostgreSQL
- **TASKS.md**: Checklist de 8 fases con tareas específicas

## Cómo ejecutar

Este ejemplo simula el flujo de `grove loop`. El contenido de `expected/` representa el resultado que你应该 obtener después de ejecutar Ralph Loop.

```bash
cd examples/02-ralph-loop

# Ver el input (specs generadas por grove-spec)
cat input/SPEC.md
cat input/DESIGN.md
cat input/TASKS.md

# Ver el output esperado (código generado)
ls -la expected/src/
cat expected/src/app.ts
```

## Para probar con grove-loop real

```bash
cd examples/02-ralph-loop
grove loop ./input/SPEC.md --tasks ./input/TASKS.md --output ./output
```

Esto debería generar código similar a `expected/src/`.

## Criterios de Validación

- [ ] Todas las tareas de TASKS.md tienen código correspondiente
- [ ] La arquitectura sigue el DESIGN.md (layers: controller → service → repository)
- [ ] Los archivos de expected/ son funcionales y compilan
- [ ] Las rutas incluyen autenticación JWT

## Notas

- El código en `expected/` es un ejemplo de lo que Ralph Loop puede generar
- El estilo sigue patrones de arquitectura limpia (Clean Architecture)
- Cada archivo tiene una responsabilidad única