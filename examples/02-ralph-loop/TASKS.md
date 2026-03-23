# TASKS.md - Todo API

## Fase 1: Configuración del Proyecto
- [ ] 1.1 Inicializar npm y TypeScript
- [ ] 1.2 Instalar dependencias: express, pg, jsonwebtoken, zod, dotenv
- [ ] 1.3 Configurar tsconfig.json

## Fase 2: Estructura Base
- [ ] 2.1 Crear app.ts con Express
- [ ] 2.2 Configurar middleware JSON parser
- [ ] 2.3 Crear config/env.ts

## Fase 3: Capa de Datos
- [ ] 3.1 Crear tabla todos en PostgreSQL
- [ ] 3.2 Crear modelo todoModel.ts (tipos TypeScript)
- [ ] 3.3 Implementar todoRepository.ts

## Fase 4: Capa de Servicios
- [ ] 4.1 Implementar todoService.ts
  - [ ] createTodo(title, description, userId)
  - [ ] getTodos(userId, status?)
  - [ ] getTodoById(id, userId)
  - [ ] updateTodo(id, userId, data)
  - [ ] deleteTodo(id, userId)
  - [ ] toggleComplete(id, userId)

## Fase 5: Controladores
- [ ] 5.1 Implementar todoController.ts
  - [ ] handleCreate
  - [ ] handleList
  - [ ] handleGet
  - [ ] handleUpdate
  - [ ] handleDelete
  - [ ] handleComplete

## Fase 6: Rutas
- [ ] 6.1 Crear routes/todoRoutes.ts
- [ ] 6.2 Montar en app.ts bajo /api/todos

## Fase 7: Middleware de Auth
- [ ] 7.1 Crear middleware/auth.ts para validar JWT

## Fase 8: Testing
- [ ] 8.1 Tests unitarios para todoService
- [ ] 8.2 Tests de integración para endpoints