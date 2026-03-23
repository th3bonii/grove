# Ejemplo: GROVE Spec - Basic

## Escenario
Tienes una idea vaga: "Agregar autenticación JWT" y quieres convertirla en una especificación completa.

## Estructura del Ejemplo

```
01-grove-spec-basic/
├── input/
│   └── idea.md           # Tu idea inicial en formato libre
├── expected/             # Output esperado después de ejecutar
│   ├── SPEC.md           # Especificación funcional completa
│   ├── DESIGN.md         # Diseño técnico y arquitectura
│   └── TASKS.md          # Checklist de implementación
└── README.md             # Este archivo
```

## Input (idea.md)
```
- Login con email/password
- Registro de usuarios
- Logout
- Tokens de refresh
- Recuperación de contraseña
```

## Cómo ejecutar

Este ejemplo simula el flujo de `grove spec`. El contenido de `expected/` representa el resultado que deberías obtener.

```bash
# Simular ejecución (los archivos ya están生成ados en expected/)
cd examples/01-grove-spec-basic

# Ver el input inicial
cat input/idea.md

# Ver el output esperado (SPEC)
cat expected/SPEC.md

# Ver el diseño técnico
cat expected/DESIGN.md

# Ver las tareas
cat expected/TASKS.md
```

## Para probar con grove-spec real

Si tienes grove instalado:

```bash
cd examples/01-grove-spec-basic
grove spec --input ./input/idea.md --output ./output
```

Esto debería generar archivos similares a los de `expected/`.

## Criterios de Validación

- [ ] SPEC.md contiene todos los requisitos funcionales
- [ ] DESIGN.md define la arquitectura técnica
- [ ] TASKS.md tiene un checklist ejecutable
- [ ] Los archivos de expected/ son realistas y completos