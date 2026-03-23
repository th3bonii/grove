# Ejemplo: GROVE Opti Prompt

## Escenario
Tienes prompts crudos que quieres optimizar para obtener mejores resultados de la IA.

## Input (prompts.txt)
Prompts en lenguaje natural que describen lo que necesitas.

## Cómo ejecutar
```bash
cd examples/03-opti-prompt
grove opti --input ./prompts.txt --output ./expected
```

## Proceso
1. Opti Prompt analiza el prompt original
2. Identifica ambigüedades y falta de contexto
3. Añade estructura, restricciones y ejemplos
4. Genera un prompt optimizado listo para usar

## Output esperado
- Prompt-1-optimized.md
- Prompt-2-optimized.md