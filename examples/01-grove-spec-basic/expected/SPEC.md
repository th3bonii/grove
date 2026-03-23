# SPEC.md - Sistema de Autenticación JWT

## 1. Resumen del Proyecto

**Nombre**: Sistema de Autenticación JWT
**Tipo**: Feature/Módulo backend
**Descripción**: Implementar autenticación completa con JWT incluyendo login, registro, logout, refresh tokens y recuperación de contraseña.
**Usuario objetivo**: Desarrolladores que necesitan integrar autenticación segura en sus aplicaciones.

---

## 2. Requisitos Funcionales

### 2.1 Autenticación de Usuarios

| ID | Requisito | Descripción | Prioridad |
|----|-----------|-------------|-----------|
| RF-001 | Login con email/password | Los usuarios pueden iniciar sesión con email y contraseña. La contraseña debe ser hasheada con bcrypt. | Alta |
| RF-002 | Registro de usuarios | Los usuarios nuevos pueden registrarse proporcionando email, contraseña y nombre. | Alta |
| RF-003 | Logout | Los usuarios pueden cerrar sesión invalidando su token de acceso. | Media |
| RF-004 | Tokens de refresh | Los usuarios pueden obtener un nuevo token de acceso usando un refresh token sin volver a autenticarse. | Alta |
| RF-005 | Recuperación de contraseña | Los usuarios pueden solicitar un enlace para restablecer su contraseña vía email. | Media |

### 2.2 Validación y Seguridad

| ID | Requisito | Descripción | Prioridad |
|----|-----------|-------------|-----------|
| RF-006 | Validación de email | El sistema debe validar que el email tenga formato correcto antes de almacenarlo. | Alta |
| RF-007 | Longitud mínima de contraseña | Las contraseñas deben tener al menos 8 caracteres. | Alta |
| RF-008 | Protección contra fuerza bruta | Limitar intentos de login a 5 por minuto por IP. | Media |

---

## 3. Requisitos No Funcionales

### 3.1 Rendimiento
- El tiempo de respuesta para login no debe exceder 200ms.
- El sistema debe soportar 1000 запросов/segundo.

### 3.2 Seguridad
- Las contraseñas deben almacenarse con bcrypt (cost factor 12).
- Los tokens JWT deben tener una vigencia máxima de 15 minutos para access tokens.
- Los refresh tokens deben tener una vigencia de 7 días.
- Usar HTTPS en producción.

### 3.3 Compatibilidad
- Backend desarrollado en Node.js con Express.
- Base de datos: PostgreSQL o MongoDB.

---

## 4. Criterios de Aceptación

- [ ] Un usuario puede registrarse con email, contraseña y nombre.
- [ ] Un usuario puede iniciar sesión y recibir un JWT válido.
- [ ] Un usuario puede cerrar sesión invalidando su token.
- [ ] Un usuario puede obtener un nuevo access token usando refresh token.
- [ ] Un usuario puede solicitar recuperación de contraseña.
- [ ] Los endpoints de autenticación están documentados con OpenAPI/Swagger.
- [ ] Las pruebas unitarias cubren al menos 80% del código.