// routes/todoRoutes.ts - Definición de rutas
import { Router } from 'express';
import { TodoController } from '../controllers/todoController';
import { authMiddleware } from '../middleware/auth';

const router = Router();
const controller = new TodoController();

router.post('/', authMiddleware, controller.handleCreate);
router.get('/', authMiddleware, controller.handleList);
router.get('/:id', authMiddleware, controller.handleGet);
router.put('/:id', authMiddleware, controller.handleUpdate);
router.delete('/:id', authMiddleware, controller.handleDelete);
router.patch('/:id/complete', authMiddleware, controller.handleComplete);

export default router;