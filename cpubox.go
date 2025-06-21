// /*********************************************************************************
// * Projeto:     Batedor
// * Componente:  CPUBox - Widget para exibir todos os núcleos da CPU
// *********************************************************************************/
package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CPUBox é um widget customizado para exibir o uso de múltiplos núcleos de CPU.
type CPUBox struct {
	*tview.Box
	mu    sync.RWMutex
	cores []float64 // Armazena o uso percentual de cada core.
}

// NewCPUBox cria um novo widget CPUBox.
func NewCPUBox() *CPUBox {
	return &CPUBox{
		Box:   tview.NewBox().SetBorder(true).SetTitle("Uso de CPU (por Núcleo)"),
		cores: []float64{},
	}
}

// Update atualiza os dados de uso dos cores.
func (c *CPUBox) Update(cores []float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cores = cores
}

// Draw desenha o widget na tela.
func (c *CPUBox) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)
	c.mu.RLock()
	defer c.mu.RUnlock()

	x, y, width, height := c.GetInnerRect()
	if width <= 0 || height <= 0 || len(c.cores) == 0 {
		return
	}

	// Caracteres para desenhar a barra de uso.
	barChars := []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}
	
	// Desenha uma linha para cada core.
	for i, coreUsage := range c.cores {
		// Não desenha mais linhas do que a altura permite.
		if i >= height {
			break
		}

		// Texto do label, ex: "Core 0:  25.7%"
		labelText := fmt.Sprintf("Núcleo %-2d: [%5.1f%%]", i, coreUsage)
		
		// Calcula o comprimento da barra de uso.
		barMaxWidth := width - len(labelText) - 2 // Deixa espaço para o label e um padding
		if barMaxWidth < 0 {
			barMaxWidth = 0
		}
		
		barLen := int(coreUsage / 100.0 * float64(barMaxWidth))

		// Cria a string da barra.
		var bar strings.Builder
		for k := 0; k < barLen; k++ {
			bar.WriteRune(barChars[len(barChars)-1]) // Caractere de bloco cheio
		}

		// Monta a string final da linha.
		fullLine := fmt.Sprintf("%s [%s]", labelText, bar.String())

		// Define a cor da barra com base no uso.
		barColor := tcell.ColorGreen
		if coreUsage > 75 {
			barColor = tcell.ColorRed
		} else if coreUsage > 50 {
			barColor = tcell.ColorYellow
		}
		
		// Imprime a linha na tela.
		tview.Print(screen, fullLine, x+1, y+i, width-2, tview.AlignLeft, barColor)
	}
}