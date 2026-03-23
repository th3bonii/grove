// app.ts - Entry point de la Todo API
import express from 'express';
import dotenv from 'dotenv';
import todoRoutes from './routes/todoRoutes';

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(express.json());

// Rutas
app.use('/api/todos', todoRoutes);

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});

export default app;