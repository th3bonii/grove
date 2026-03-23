# Ejemplo: GROVE Opti Prompt

## Escenario
Tienes prompts crudos en lenguaje natural que quieres optimizar para obtener mejores resultados de los agentes de IA.

## Estructura del Ejemplo

```
03-opti-prompt/
├── input/
│   └── prompts.txt          # Prompts originales (sin optimizar)
├── expected/                 # Output esperado
│   ├── prompt-1-optimized.md # Prompt 1 optimizado
│   └── prompt-2-optimized.md # Prompt 2 optimizado
└── README.md                 # Este archivo
```

## Input (prompts.txt)
```
Prompt 1:
Crea un componente de React para mostrar una lista de usuarios

Prompt 2:
Haz una función que valide un email
```

## Cómo ejecutar

Este ejemplo simula el flujo de `grove opti`. El contenido de `expected/` representa el resultado que deberías obtener después de optimizar.

```bash
cd examples/03-opti-prompt

# Ver los prompts originales
cat prompts.txt

# Ver los prompts optimizados
cat expected/prompt-1-optimized.md
cat expected/prompt-2-optimized.md
```

## Para probar con grove-opti real

```bash
cd examples/03-opti-prompt
grove opti --input ./prompts.txt --output ./output
```

Esto debería generar archivos optimizados similares a `expected/`.

## Proceso de Optimización

El proceso de optimización incluye:

1. **Análisis del prompt original** - Identificar ambigüedades
2. **Adición de contexto** - Incluir información relevante
3. **Estructuración** - Agregar formato clear y secciones
4. **Ejemplos** - Añadir casos de uso de ejemplo
5. **Restricciones** - Definir límites y requisitos

## Criterios de Validación

- [ ] El prompt optimizado es más específico y detallado
- [ ] Incluye ejemplos de input/output
- [ ] Define restricciones claras
- [ ] El resultado esperado está bien definido

## Ejemplo de Optimización

**Original:**
```
Crea un componente de React para mostrar una lista de usuarios
```

**Optimizado (expected/prompt-1-optimized.md):**
Incluye:
- Contexto del framework (React + TypeScript)
- Props esperadas
- Estilos (CSS modules / Tailwind)
- Estados de carga/vacío/error
- Ejemplos de datos de entrada/salida