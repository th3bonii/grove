// controllers/todoController.ts - Manejo de requests/responses
import { Request, Response } from 'express';
import { TodoService } from '../services/todoService';
import { CreateTodoDTO, UpdateTodoDTO } from '../models/todoModel';

export class TodoController {
  private service: TodoService;

  constructor() {
    this.service = new TodoService();
  }

  handleCreate = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = req.body.user?.id;
      if (!userId) {
        res.status(401).json({ error: 'No autorizado' });
        return;
      }

      const data: CreateTodoDTO = {
        title: req.body.title,
        description: req.body.description,
      };

      const todo = await this.service.createTodo(userId, data);
      res.status(201).json(todo);
    } catch (error: any) {
      res.status(400).json({ error: error.message });
    }
  };

  handleList = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = (req as any).user?.id;
      if (!userId) {
        res.status(401).json({ error: 'No autorizado' });
        return;
      }

      const status = req.query.status as 'pending' | 'completed' | undefined;
      const todos = await this.service.getTodos(userId, status ? { status } : undefined);
      res.json(todos);
    } catch (error: any) {
      res.status(500).json({ error: error.message });
    }
  };

  handleGet = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = (req as any).user?.id;
      const todo = await this.service.getTodoById(req.params.id, userId);
      res.json(todo);
    } catch (error: any) {
      res.status(404).json({ error: error.message });
    }
  };

  handleUpdate = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = (req as any).user?.id;
      const data: UpdateTodoDTO = {
        title: req.body.title,
        description: req.body.description,
        completed: req.body.completed,
      };
      const todo = await this.service.updateTodo(req.params.id, userId, data);
      res.json(todo);
    } catch (error: any) {
      res.status(404).json({ error: error.message });
    }
  };

  handleDelete = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = (req as any).user?.id;
      await this.service.deleteTodo(req.params.id, userId);
      res.status(204).send();
    } catch (error: any) {
      res.status(404).json({ error: error.message });
    }
  };

  handleComplete = async (req: Request, res: Response): Promise<void> => {
    try {
      const userId = (req as any).user?.id;
      const todo = await this.service.toggleComplete(req.params.id, userId);
      res.json(todo);
    } catch (error: any) {
      res.status(404).json({ error: error.message });
    }
  };
}