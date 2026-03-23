# SPEC.md - API REST de Tareas (Todo API)

## 1. Resumen del Proyecto

**Nombre**: Todo API
**Tipo**: Backend REST API
**Descripción**: API REST para gestión de tareas/todos con autenticación JWT.
**Usuario objetivo**: Aplicaciones frontend que necesitan gestionar tareas.

---

## 2. Requisitos Funcionales

### 2.1 Gestión de Tareas

| ID | Requisito | Descripción | Prioridad |
|----|-----------|-------------|-----------|
| RF-001 | Crear tarea | POST /todos - Crear una nueva tarea | Alta |
| RF-002 | Listar tareas | GET /todos - Obtener todas las tareas del usuario | Alta |
| RF-003 | Obtener tarea | GET /todos/:id - Obtener una tarea específica | Alta |
| RF-004 | Actualizar tarea | PUT /todos/:id - Actualizar una tarea existente | Alta |
| RF-005 | Eliminar tarea | DELETE /todos/:id - Eliminar una tarea | Alta |
| RF-006 | Marcar completada | PATCH /todos/:id/complete - Marcar tarea como completada | Media |
| RF-007 | Filtrar por estado | GET /todos?status=pending - Filtrar tareas por completada/pendiente | Media |

### 2.2 Autenticación (Reutilizar módulo existente)

| ID | Requisito | Descripción | Prioridad |
|----|-----------|-------------|-----------|
| RF-008 | Autenticación JWT | Todas las rutas de /todos requieren JWT válido | Alta |

---

## 3. Estructura de Datos

### Modelo: Todo
| Campo | Tipo | Descripción |
|-------|------|-------------|
| id | UUID | Identificador único |
| title | STRING | Título de la tarea (requerido, max 200 chars) |
| description | TEXT? | Descripción opcional |
| completed | BOOLEAN | Estado de la tarea (default: false) |
| userId | UUID | FK al usuario owner |
| createdAt | TIMESTAMP | Fecha de creación |
| updatedAt | TIMESTAMP | Fecha de última actualización |

---

## 4. Criterios de Aceptación

- [ ] Usuario autenticado puede crear tareas
- [ ] Usuario solo ve sus propias tareas
- [ ] Se pueden filtrar tareas por estado
- [ ] La API retorna códigos HTTP apropiados (200, 201, 404, 401)
- [ ] Validación de input: título requerido, máximo 200 caracteres
- [ ] Tests unitarios covering core logic