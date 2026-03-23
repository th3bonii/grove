# Prompt Optimizado: Componente de Lista de Usuarios

## Prompt Original
"Crea un componente de React para mostrar una lista de usuarios"

## Prompt Optimizado

Eres un desarrollador senior de React con experiencia en TypeScript y diseño de componentes.

**Contexto**: Estás desarrollando una aplicación web de dashboard administrativo que necesita mostrar una lista de usuarios con sus datos básicos.

**Objetivo**: Crear un componente de React que renderice una lista de usuarios.

**Requisitos Técnicos**:
- Usar TypeScript
- Componente funcional con React Hooks
- Props bien tipadas
- Manejo de estados: loading, empty, error, success
- Diseño responsive (mobile-first)

**Estructura de Datos Esperada**:
```typescript
interface User {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  role: 'admin' | 'user' | 'guest';
  status: 'active' | 'inactive' | 'pending';
  createdAt: string;
}
```

**Props del Componente**:
```typescript
interface UserListProps {
  users: User[];
  onUserClick?: (user: User) => void;
  onDeleteUser?: (userId: string) => void;
  isLoading?: boolean;
  showActions?: boolean;
}
```

**Casos de Borde a Manejar**:
- Lista vacía → mostrar mensaje "No hay usuarios"
- Error de carga → mostrar mensaje de error con retry button
- many users (100+) → implementar virtual scrolling o paginación
- Imagen de avatar no disponible → mostrar inicial del nombre

**Estilo**: UI consistente con componentes de dashboard, usar CSS modules o styled-components. Include accessibility (a11y) - aria-labels, keyboard navigation.

**Output**: Provide the complete component code in TypeScript with comments explaining key decisions.