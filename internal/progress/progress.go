package progress

import (
	"fmt"
	"strings"
)

// ProgressBar representa una barra de progreso visual en la terminal.
type ProgressBar struct {
	width   int
	current int
	total   int
	message string
}

// NewProgressBar crea una nueva barra de progreso.
// Por defecto el ancho es 40 caracteres.
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		width:   40,
		current: 0,
		total:   total,
		message: "",
	}
}

// Update actualiza el progreso actual y muestra la barra.
// Si message está vacía, mantiene el mensaje anterior.
func (p *ProgressBar) Update(current int, message string) {
	if current < 0 {
		current = 0
	}
	if current > p.total {
		current = p.total
	}
	p.current = current
	if message != "" {
		p.message = message
	}
	p.render()
}

// Finish marca la barra de progreso como completada.
func (p *ProgressBar) Finish(message string) {
	p.current = p.total
	p.message = message
	p.render()
}

// Reset reinicia la barra de progreso con un nuevo total.
// Si total es 0, mantiene el total anterior.
func (p *ProgressBar) Reset(total int) {
	if total > 0 {
		p.total = total
	}
	p.current = 0
	p.message = ""
}

// render genera la representación visual de la barra de progreso.
func (p *ProgressBar) render() {
	// Calcular porcentaje
	percent := 0
	if p.total > 0 {
		percent = (p.current * 100) / p.total
	}

	// Calcular cuántos caracteresfilled vs empty
	filled := (p.width * p.current) / p.total
	if p.current == p.total {
		filled = p.width
	}
	empty := p.width - filled

	// Construir la barra
	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < filled; i++ {
		bar.WriteString("█")
	}
	for i := 0; i < empty; i++ {
		bar.WriteString("░")
	}
	bar.WriteString("]")

	// Construir mensaje final
	output := fmt.Sprintf("%s %2d%% %s", bar.String(), percent, p.message)
	fmt.Println(output)
}
