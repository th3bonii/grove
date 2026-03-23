# Prompt Optimizado: Validación de Email

## Prompt Original
"Haz una función que valide un email"

## Prompt Optimizado

Eres un desarrollador backend Senior especializado en validación de datos y seguridad.

**Contexto**: Estás implementando validación de input para un formulario de registro de usuarios en una API REST Node.js/TypeScript.

**Objetivo**: Crear una función de validación de email robusta y segura.

**Requisitos**:
- TypeScript con tipos estrictos
- Retornar boolean o throwing error con mensaje descriptivo
- RFC 5322 compliant (soportar emails válidos estándar)
- NO permitir caracteres peligrosos (SQL injection, XSS)
- Manejar edge cases: null, undefined, empty string, whitespace

**Firma de Función Esperada**:
```typescript
function validateEmail(email: string): ValidationResult;

interface ValidationResult {
  isValid: boolean;
  error?: string;
  normalized?: string; // lowercase, trimmed
}
```

**Reglas de Validación**:
1. No puede estar vacío o ser solo whitespace
2. Debe contener exactamente un @ 
3. Dominio debe tener al menos un punto (ej: example.com)
4. Local part (antes del @): max 64 caracteres
5. Dominio: max 255 caracteres
6. No permitir caracteres especiales peligrosos: `< > ( ) [ ] : ; , \ / " | ? =`
7. Solo permitir caracteres alfanuméricos, puntos, guiones y underscores en local part

**Nohacer**:
- No usar regex simplista tipo `/^[^\s@]+@[^\s@]+\.[^\s@]+$/`
- No aceptar emails con múltiples @ o domains malformed
- No usar blacklist de emails (manejar en capa diferente)

**Output**: Provide la función completa con JSDoc, ejemplos de uso, y casos de test recomendados.