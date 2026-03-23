# TASKS.md - Sistema de Autenticación JWT

## Fase 1: Configuración Inicial

- [ ] 1.1 Crear proyecto Node.js con TypeScript
- [ ] 1.2 Instalar dependencias: express, jsonwebtoken, bcryptjs, pg/mongoose, dotenv, joi/zod, helmet, cors
- [ ] 1.3 Configurar TypeScript (tsconfig.json)
- [ ] 1.4 Crear estructura de carpetas: src/{controllers,services,middlewares,models,utils}
- [ ] 1.5 Configurar variables de entorno (.env.example)

## Fase 2: Modelos de Base de Datos

- [ ] 2.1 Crear modelo User (schema de validación con Zod)
- [ ] 2.2 Crear modelo RefreshToken
- [ ] 2.3 Crear modelo PasswordReset
- [ ] 2.4 Implementar migraciones o scripts de setup para PostgreSQL/MongoDB

## Fase 3: Servicios Core

- [ ] 3.1 Implementar PasswordService con bcrypt
  - [ ] 3.1.1 hash(password) → string
  - [ ] 3.1.2 compare(password, hash) → boolean

- [ ] 3.2 Implementar JWTService
  - [ ] 3.2.1 generateAccessToken(userId) → string
  - [ ] 3.2.2 generateRefreshToken(userId) → string
  - [ ] 3.2.3 verifyAccessToken(token) → payload
  - [ ] 3.2.4 verifyRefreshToken(token) → payload

- [ ] 3.3 Implementar AuthService
  - [ ] 3.3.1 register(email, password, name) → user
  - [ ] 3.3.2 login(email, password) → { accessToken, refreshToken }
  - [ ] 3.3.3 refresh(refreshToken) → { accessToken }
  - [ ] 3.3.4 logout(refreshToken) → void
  - [ ] 3.3.5 forgotPassword(email) → void
  - [ ] 3.3.6 resetPassword(token, newPassword) → void

## Fase 4: Controladores

- [ ] 4.1 Crear AuthController con endpoints:
  - [ ] 4.1.1 POST /auth/register
  - [ ] 4.1.2 POST /auth/login
  - [ ] 4.1.3 POST /auth/refresh
  - [ ] 4.1.4 POST /auth/logout
  - [ ] 4.1.5 POST /auth/forgot-password
  - [ ] 4.1.6 POST /auth/reset-password

- [ ] 4.2 Implementar validación de requests con Zod schemas

## Fase 5: Middlewares

- [ ] 5.1 Crear authMiddleware para verificar JWT
- [ ] 5.2 Crear rateLimiter para protección contra fuerza bruta
- [ ] 5.3 Crear errorHandler para manejo centralizado de errores

## Fase 6: API Routes

- [ ] 6.1 Configurar rutas en app.ts/index.ts
- [ ] 6.2 Integrar middlewares (helmet, cors, rateLimiter)
- [ ] 6.3 Montar rutas de autenticación

## Fase 7: Documentación

- [ ] 7.1 Generar documentación OpenAPI/Swagger
- [ ] 7.2 Documentar errores posibles con códigos

## Fase 8: Testing

- [ ] 8.1 Escribir tests unitarios para AuthService (mínimo 80% coverage)
- [ ] 8.2 Escribir tests unitarios para JWTService
- [ ] 8.3 Escribir tests de integración para endpoints

## Fase 9: Seguridad Adicional

- [ ] 9.1 Implementar logging de intentos de login fallidos
- [ ] 9.2 Agregar validación de email con regex
- [ ] 9.3 Proteger contra SQL injection con parameterized queries

---

## Notas de Implementación

- Usar PostgreSQL o MongoDB según preferencia del equipo
- Considerar usar JWT con jose en lugar de jsonwebtoken para mejor rendimiento
- Implementar refresh token rotation para mayor seguridad
- Almacenar refresh tokens en memoria Redis para mejor rendimiento en producción