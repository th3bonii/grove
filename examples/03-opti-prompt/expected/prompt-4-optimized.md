# Prompt 4 Optimizado: Tests para función de login

## Contexto
Eres un QA engineer y experto en testing. El usuario necesita escribir tests para una función de autenticación.

## Requisitos
- Framework de testing: Jest o Vitest
- Lenguaje: TypeScript
- Tipos de tests: Unitarios y de integración

## Función a testear (simulada)
```typescript
interface LoginInput {
  email: string;
  password: string;
}

interface LoginOutput {
  success: boolean;
  token?: string;
  user?: {
    id: string;
    email: string;
    name: string;
  };
  error?: string;
}

async function login(input: LoginInput): Promise<LoginOutput> {
  // Implementación que valida credenciales
  // Retorna token JWT si es exitoso
}
```

## Casos de prueba a cubrir

### Tests unitarios
- ✅ Login exitoso con credenciales válidas
- ✅ Login fallido con email inválido
- ✅ Login fallido con contraseña incorrecta
- ✅ Login con email o password vacíos
- ✅ Login con formato de email inválido
- ✅ Login con password menor a 8 caracteres

### Tests de integración (opcional)
- ✅ Múltimos 3 intentos fallidos desde misma IP
- ✅ Rate limiting después de 5 intentos fallidos

## Restricciones
- Usar describe/it o test (sintaxis moderna)
- Mockear dependencias externas (base de datos, servicios)
- Incluir aserciones significativas
- Nombrar tests de forma descriptiva

## Formato de respuesta
```typescript
// __tests__/login.test.ts
import { login } from '../src/auth/login';

describe('login', () => {
  // tests aquí
});
```

## Notas adicionales
- Explicar mocking de dependencias
- Sugerir configuración de coverage mínimo (80%)
- Incluir examples de uso con datos de prueba