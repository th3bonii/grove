// models/todoModel.ts - Tipos para el modelo Todo
export interface Todo {
  id: string;
  title: string;
  description: string | null;
  completed: boolean;
  userId: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateTodoDTO {
  title: string;
  description?: string;
}

export interface UpdateTodoDTO {
  title?: string;
  description?: string;
  completed?: boolean;
}

export interface TodoFilter {
  status?: 'pending' | 'completed';
}