# DESIGN.md - Todo API

## 1. Arquitectura

```
┌──────────────────┐
│   Frontend       │
└────────┬─────────┘
         │ HTTP
         ▼
┌──────────────────┐
│  Express.js      │
│  ───────────     │
│  Routes          │
│  Controllers     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Services        │
│  ───────────     │
│  TodoService     │
│  AuthService     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Repository      │
│  ───────────     │
│  TodoRepository  │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  PostgreSQL      │
└──────────────────┘
```

## 2. Estructura de Archivos

```
src/
├── app.ts              # Entry point
├── config/
│   └── env.ts          # Environment variables
├── controllers/
│   └── todoController.ts
├── services/
│   └── todoService.ts
├── repositories/
│   └── todoRepository.ts
├── models/
│   └── todoModel.ts
├── middleware/
│   └── auth.ts
└── utils/
    └── errors.ts
```

## 3. Endpoints

| Método | Path | Descripción | Auth |
|--------|------|-------------|------|
| POST | /api/todos | Crear tarea | JWT |
| GET | /api/todos | Listar tareas | JWT |
| GET | /api/todos/:id | Obtener tarea | JWT |
| PUT | /api/todos/:id | Actualizar tarea | JWT |
| DELETE | /api/todos/:id | Eliminar tarea | JWT |
| PATCH | /api/todos/:id/complete | Completar tarea | JWT |

## 4. Dependencias

- express
- jsonwebtoken
- pg (postgres)
- zod (validation)
- dotenv