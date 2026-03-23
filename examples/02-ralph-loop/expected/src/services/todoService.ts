// services/todoService.ts - Lógica de negocio
import { TodoRepository } from '../repositories/todoRepository';
import { CreateTodoDTO, UpdateTodoDTO, TodoFilter, Todo } from '../models/todoModel';

export class TodoService {
  private repository: TodoRepository;

  constructor() {
    this.repository = new TodoRepository();
  }

  async createTodo(userId: string, data: CreateTodoDTO): Promise<Todo> {
    if (!data.title || data.title.trim().length === 0) {
      throw new Error('El título es requerido');
    }
    if (data.title.length > 200) {
      throw new Error('El título no puede exceder 200 caracteres');
    }
    return this.repository.create(userId, {
      title: data.title.trim(),
      description: data.description?.trim(),
    });
  }

  async getTodos(userId: string, filter?: TodoFilter): Promise<Todo[]> {
    return this.repository.findByUser(userId, filter);
  }

  async getTodoById(id: string, userId: string): Promise<Todo> {
    const todo = await this.repository.findById(id, userId);
    if (!todo) {
      throw new Error('Tarea no encontrada');
    }
    return todo;
  }

  async updateTodo(id: string, userId: string, data: UpdateTodoDTO): Promise<Todo> {
    if (data.title !== undefined) {
      if (data.title.trim().length === 0) {
        throw new Error('El título no puede estar vacío');
      }
      if (data.title.length > 200) {
        throw new Error('El título no puede exceder 200 caracteres');
      }
    }

    const todo = await this.repository.update(id, userId, {
      ...data,
      title: data.title?.trim(),
      description: data.description?.trim(),
    });

    if (!todo) {
      throw new Error('Tarea no encontrada');
    }
    return todo;
  }

  async deleteTodo(id: string, userId: string): Promise<void> {
    const deleted = await this.repository.delete(id, userId);
    if (!deleted) {
      throw new Error('Tarea no encontrada');
    }
  }

  async toggleComplete(id: string, userId: string): Promise<Todo> {
    const todo = await this.repository.findById(id, userId);
    if (!todo) {
      throw new Error('Tarea no encontrada');
    }
    return this.repository.update(id, userId, { completed: !todo.completed });
  }
}