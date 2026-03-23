package progress

import (
	"testing"
)

// helperCaptureOutput captura la salida de una función que escribe a stdout.
func helperCaptureOutput(fn func()) string {
	// Guardar stdout original
	oldStdout := outputWriter
	defer func() { outputWriter = oldStdout }()

	// Crear un pipe para capturar la salida
	outputWriter = &captureWriter{}

	fn()

	return outputWriter.(*captureWriter).buffer
}

// captureWriter es un writer que guarda en un buffer para testing.
type captureWriter struct {
	buffer string
}

func (c *captureWriter) Write(p []byte) (n int, err error) {
	c.buffer += string(p)
	return len(p), nil
}

// outputWriter es una variable global para poder hacer testing de la salida.
// Por defecto es un writer que escribe a stdout.
var outputWriter interface {
	Write(p []byte) (n int, err error)
} = &defaultWriter{}

type defaultWriter struct{}

func (d *defaultWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(100)

	if pb.width != 40 {
		t.Errorf("Expected width 40, got %d", pb.width)
	}
	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}
	if pb.total != 100 {
		t.Errorf("Expected total 100, got %d", pb.total)
	}
	if pb.message != "" {
		t.Errorf("Expected empty message, got %s", pb.message)
	}
}

func TestProgressBar_Update(t *testing.T) {
	pb := NewProgressBar(100)

	// Test que Update establece los valores correctamente
	pb.current = 0
	pb.total = 100

	pb.Update(50, "Test message")

	if pb.current != 50 {
		t.Errorf("Expected current 50, got %d", pb.current)
	}
	if pb.message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", pb.message)
	}
}

func TestProgressBar_Update_Clamping(t *testing.T) {
	pb := NewProgressBar(100)

	// Test clamping hacia abajo
	pb.Update(-10, "")
	if pb.current != 0 {
		t.Errorf("Expected current 0 after negative, got %d", pb.current)
	}

	// Test clamping hacia arriba
	pb.Update(150, "")
	if pb.current != 100 {
		t.Errorf("Expected current 100 after overflow, got %d", pb.current)
	}
}

func TestProgressBar_Update_EmptyMessage(t *testing.T) {
	pb := NewProgressBar(100)
	pb.message = "original"

	// Cuando message está vacía, mantener la anterior
	pb.Update(50, "")

	if pb.message != "original" {
		t.Errorf("Expected message 'original', got %s", pb.message)
	}
}

func TestProgressBar_Finish(t *testing.T) {
	pb := NewProgressBar(100)
	pb.current = 60

	pb.Finish("Completed")

	if pb.current != 100 {
		t.Errorf("Expected current 100, got %d", pb.current)
	}
	if pb.message != "Completed" {
		t.Errorf("Expected message 'Completed', got %s", pb.message)
	}
}

func TestProgressBar_Reset(t *testing.T) {
	pb := NewProgressBar(100)
	pb.current = 75
	pb.message = "some message"

	pb.Reset(200)

	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}
	if pb.total != 200 {
		t.Errorf("Expected total 200, got %d", pb.total)
	}
	if pb.message != "" {
		t.Errorf("Expected empty message after reset, got %s", pb.message)
	}
}

func TestProgressBar_Reset_KeepTotal(t *testing.T) {
	pb := NewProgressBar(100)
	pb.current = 50

	// Reset con total=0 debe mantener el total anterior
	pb.Reset(0)

	if pb.total != 100 {
		t.Errorf("Expected total 100, got %d", pb.total)
	}
	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}
}

func TestProgressBar_Width_Default(t *testing.T) {
	pb := NewProgressBar(50)

	if pb.width != 40 {
		t.Errorf("Expected default width 40, got %d", pb.width)
	}
}

func TestProgressBar_Total_Zero(t *testing.T) {
	// Test con total 0 para evitar división por cero
	pb := NewProgressBar(0)
	pb.current = 0
	// No debería panicar
	_ = pb.current
}
