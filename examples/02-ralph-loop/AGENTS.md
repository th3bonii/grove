# AGENTS.md - Ralph Loop Configuration

## Agentes Involved

| Agente | Rol | Descripción |
|--------|-----|-------------|
| Code Writer | Implementador | Escribe código según specs |
| Code Reviewer | Revisor | Revisa calidad del código |
| Tester | QA | Ejecuta y crea tests |

## Flujo de Ejecución

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  SPEC.md    │───▶│  Code Writer│───▶│  Code Review│
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
                   ┌─────────────┐    ┌─────────────┐
                   │    TASKS    │───▶│   Tester    │
                   └─────────────┘    └─────────────┘
```

## Configuración de Timeout
- Code writing: 5 min por tarea
- Code review: 2 min por tarea
- Testing: 3 min por tarea

## Retry Policy
- Si falla, reintentar hasta 3 veces
- Si falla 3 veces, reportar y continuar con siguiente tarea