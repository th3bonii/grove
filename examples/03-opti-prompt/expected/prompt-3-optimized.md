# Prompt 3 Optimizado: API REST con Node.js

## Contexto
Eres un desarrollador backend senior. El usuario necesita crear una API REST con Node.js.

## Requisitos
- Framework: Express.js o Fastify
- Lenguaje: TypeScript
- Base de datos: PostgreSQL (usar pg o knex)
- Estructura: MVC o Clean Architecture

## Especificaciones
Crea una estructura de proyecto completa con:
1. `package.json` con dependencias necesarias
2. `tsconfig.json` configurado
3. Archivo principal `src/app.ts` o `src/index.ts`
4. Estructura de carpetas sugerida

## Restricciones
- NO usar frameworks como NestJS o Sequelize (mantener simple)
- Usar variables de entorno para configuración
- Incluir manejo de errores básico
- Documentar endpoints con JSDoc o comentarios

## Output esperado
Estructura de directorios:
```
src/
├── app.ts
├── config/
├── controllers/
├── routes/
├── services/
├── repositories/
└── middleware/
```

## Ejemplo de código base (incluir en respuesta)
```typescript
// src/app.ts - ejemplo básico
import express from 'express';
const app = express();
export default app;
```

## Notas adicionales
- Proporcionar comandos para instalar dependencias y ejecutar
- Sugerir rutas de endpoints básicos (CRUD)
- Incluir ejemplo de middleware de autenticación básico