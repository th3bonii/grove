// repositories/todoRepository.ts - Capa de acceso a datos
import { Pool } from 'pg';
import { config } from '../config/env';
import { Todo, CreateTodoDTO, UpdateTodoDTO, TodoFilter } from '../models/todoModel';

const pool = new Pool(config.database);

export class TodoRepository {
  async create(userId: string, data: CreateTodoDTO): Promise<Todo> {
    const result = await pool.query(
      `INSERT INTO todos (title, description, user_id) 
       VALUES ($1, $2, $3) 
       RETURNING *`,
      [data.title, data.description || null, userId]
    );
    return this.mapToTodo(result.rows[0]);
  }

  async findByUser(userId: string, filter?: TodoFilter): Promise<Todo[]> {
    let query = 'SELECT * FROM todos WHERE user_id = $1';
    const params: any[] = [userId];

    if (filter?.status) {
      query += filter.status === 'completed' 
        ? ' AND completed = true' 
        : ' AND completed = false';
    }

    query += ' ORDER BY created_at DESC';

    const result = await pool.query(query, params);
    return result.rows.map(this.mapToTodo);
  }

  async findById(id: string, userId: string): Promise<Todo | null> {
    const result = await pool.query(
      'SELECT * FROM todos WHERE id = $1 AND user_id = $2',
      [id, userId]
    );
    return result.rows[0] ? this.mapToTodo(result.rows[0]) : null;
  }

  async update(id: string, userId: string, data: UpdateTodoDTO): Promise<Todo | null> {
    const updates: string[] = [];
    const params: any[] = [];
    let paramIndex = 1;

    if (data.title !== undefined) {
      updates.push(`title = $${paramIndex++}`);
      params.push(data.title);
    }
    if (data.description !== undefined) {
      updates.push(`description = $${paramIndex++}`);
      params.push(data.description);
    }
    if (data.completed !== undefined) {
      updates.push(`completed = $${paramIndex++}`);
      params.push(data.completed);
    }

    if (updates.length === 0) return null;

    params.push(id, userId);
    const result = await pool.query(
      `UPDATE todos SET ${updates.join(', ')}, updated_at = NOW() 
       WHERE id = $${paramIndex++} AND user_id = $${paramIndex} 
       RETURNING *`,
      params
    );
    return result.rows[0] ? this.mapToTodo(result.rows[0]) : null;
  }

  async delete(id: string, userId: string): Promise<boolean> {
    const result = await pool.query(
      'DELETE FROM todos WHERE id = $1 AND user_id = $2',
      [id, userId]
    );
    return (result.rowCount ?? 0) > 0;
  }

  private mapToTodo(row: any): Todo {
    return {
      id: row.id,
      title: row.title,
      description: row.description,
      completed: row.completed,
      userId: row.user_id,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
    };
  }
}