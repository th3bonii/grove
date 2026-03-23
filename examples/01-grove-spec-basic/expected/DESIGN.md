# DESIGN.md - Sistema de Autenticación JWT

## 1. Arquitectura General

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENTE (Frontend)                        │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HTTP/HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      EXPRESS REST API                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ POST /auth   │  │ POST /auth   │  │ POST /auth   │          │
│  │ /login       │  │ /register    │  │ /refresh     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│  ┌──────────────┐  ┌──────────────┐                            │
│  │ POST /auth   │  │ POST /auth   │                            │
│  │ /logout      │  │ /forgot      │                            │
│  └──────────────┘  └──────────────┘                            │
└─────────────────────────────┬───────────────────────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
         ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   JWT        │    │   bcrypt     │    │   Logger     │
│   Service    │    │   Service    │    │   Service    │
└──────────────┘    └──────────────┘    └──────────────┘
         │                                       │
         └───────────────┬───────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DATABASE (PostgreSQL/MongoDB)                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ users        │  │ refresh_    │  │ password_    │          │
│  │ table        │  │ tokens      │  │ resets       │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Componentes del Sistema

### 2.1 Auth Controller
**Responsabilidad**: Manejar las solicitudes HTTP entrantes y responder al cliente.
- `POST /auth/register` - Crear nuevo usuario
- `POST /auth/login` - Autenticar usuario
- `POST /auth/refresh` - Renovar access token
- `POST /auth/logout` - Invalidar tokens
- `POST /auth/forgot-password` - Solicitar recuperación

### 2.2 Auth Service (Lógica de negocio)
**Responsabilidad**: Implementar la lógica de autenticación.
- `register(email, password, name)` → user
- `login(email, password)` → { accessToken, refreshToken }
- `refresh(refreshToken)` → { accessToken }
- `logout(refreshToken)` → void
- `forgotPassword(email)` → void
- `resetPassword(token, newPassword)` → void

### 2.3 JWT Service
**Responsabilidad**: Manejar la creación y verificación de tokens.
- `generateAccessToken(userId)` → string (15 min)
- `generateRefreshToken(userId)` → string (7 días)
- `verifyAccessToken(token)` → payload
- `verifyRefreshToken(token)` → payload
- `blacklistToken(token)` → void

### 2.4 Password Service
**Responsabilidad**: Hashear y verificar contraseñas.
- `hash(password)` → string (bcrypt cost 12)
- `compare(password, hash)` → boolean
- `generateResetToken()` → string
- `verifyResetToken(token)` → boolean

---

## 3. Estructura de Datos

### Tabla: users
| Campo | Tipo | Restricciones |
|-------|------|---------------|
| id | UUID | PK, auto-generated |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| name | VARCHAR(100) | NOT NULL |
| created_at | TIMESTAMP | DEFAULT NOW() |
| updated_at | TIMESTAMP | DEFAULT NOW() |

### Tabla: refresh_tokens
| Campo | Tipo | Restricciones |
|-------|------|---------------|
| id | UUID | PK, auto-generated |
| user_id | UUID | FK → users.id, NOT NULL |
| token | TEXT | NOT NULL, indexed |
| expires_at | TIMESTAMP | NOT NULL |
| created_at | TIMESTAMP | DEFAULT NOW() |

### Tabla: password_resets
| Campo | Tipo | Restricciones |
|-------|------|---------------|
| id | UUID | PK, auto-generated |
| user_id | UUID | FK → users.id, NOT NULL |
| token | VARCHAR(255) | NOT NULL, indexed |
| used | BOOLEAN | DEFAULT false |
| expires_at | TIMESTAMP | NOT NULL |
| created_at | TIMESTAMP | DEFAULT NOW() |

---

## 4. Flujos Principales

### 4.1 Flujo de Login
```
1. Cliente envía POST /auth/login con {email, password}
2. AuthController busca usuario por email
3. PasswordService verifica contraseña
4. JWTService genera access token y refresh token
5. AuthService guarda refresh token en DB
6. Devuelve {accessToken, refreshToken} al cliente
```

### 4.2 Flujo de Refresh Token
```
1. Cliente envía POST /auth/refresh con {refreshToken}
2. JWTService verifica refresh token
3. AuthService busca token en DB y verifica que no esté vencido
4. JWTService genera nuevo access token
5. Devuelve {accessToken} al cliente
```

---

## 5. Middlewares Requeridos

| Middleware | Función |
|------------|---------|
| `authMiddleware` | Verifica JWT en headers Authorization |
| `rateLimiter` | Limita requests por IP (100/min) |
| `validateRequest` | Valida schema de request con Joi/Zod |
| `errorHandler` | Manejo centralizado de errores |

---

## 6. Configuración de Variables de Entorno

```env
# JWT
JWT_SECRET=your-secret-key-min-32-chars
JWT_EXPIRES_IN=15m
REFRESH_TOKEN_EXPIRES_IN=7d

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/auth_db

# Password Reset
PASSWORD_RESET_URL=https://tu-app.com/reset
PASSWORD_RESET_TOKEN_EXPIRES=1h

# Rate Limiting
RATE_LIMIT_WINDOW_MS=60000
RATE_LIMIT_MAX_REQUESTS=100
```

---

## 7. Consideraciones de Seguridad

1. **Hash de contraseñas**: Usar bcrypt con cost factor 12
2. **Tokens en blacklist**: Implementar logout invalidando tokens
3. **HTTPOnly cookies**: Considerar usar cookies para tokens
4. **Validación de input**: Usar Zod para validación de schemas
5. **Logging**: Registrar intentos de login fallidos sin exponer datos sensibles
6. **HTTPS**: Forzar en producción