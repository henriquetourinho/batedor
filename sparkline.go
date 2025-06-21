// /*********************************************************************************
// * Projeto:     Batedor
// * Autor:       Carlos Henrique Tourinho Santana (https://github.com/henriquetourinho)
// * Versão:      4.0 - Splash Screen
// * Data:        21 de Junho de 2025
// *********************************************************************************/
package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Sparkline é nosso widget customizado que desenha um gráfico em tempo real.
type Sparkline struct {
	*tview.Box
	mu         sync.RWMutex
	data       []float64 // Histórico de dados a serem plotados (ex: uso de CPU em %).
	capacity   int       // Quantos pontos de dados queremos manter no histórico.
	label      string
	labelColor tcell.Color
}

// NewSparkline cria um novo widget de gráfico.
func NewSparkline(label string) *Sparkline {
	return &Sparkline{
		Box:        tview.NewBox().SetBorder(true).SetTitle(label),
		capacity:   100, // Vamos guardar os últimos 100 pontos de dados.
		data:       make([]float64, 0, 100),
		label:      label,
		labelColor: tcell.ColorYellow,
	}
}

// AddData adiciona um novo ponto de dado ao gráfico e atualiza o título.
func (s *Sparkline) AddData(value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, value)

	if len(s.data) > s.capacity {
		s.data = s.data[1:]
	}

	s.SetTitle(fmt.Sprintf(" %s: %.2f%% ", s.label, value))
}

// SetLabelColor define a cor do texto do valor percentual.
func (s *Sparkline) SetLabelColor(color tcell.Color) *Sparkline {
	s.labelColor = color
	return s
}

// Draw é a função mágica chamada pelo tview para desenhar o widget na tela.
func (s *Sparkline) Draw(screen tcell.Screen) {
	s.Box.Draw(screen)
	s.mu.RLock()
	defer s.mu.RUnlock()

	x, y, width, height := s.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	if s.capacity != width {
		s.capacity = width
		if len(s.data) > width {
			s.data = s.data[len(s.data)-width:]
		}
	}

	sparkChars := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	for i := 0; i < len(s.data); i++ {
		value := s.data[i]
		charIndex := int(value / (100.0 / float64(len(sparkChars))))
		if charIndex >= len(sparkChars) {
			charIndex = len(sparkChars) - 1
		}
		if charIndex < 0 {
			charIndex = 0
		}

		valueText := fmt.Sprintf("%.1f%%", value)
		if i == len(s.data)-1 {
			tview.Print(screen, valueText, x+width-len(valueText)-1, y+height/2, width, tview.AlignLeft, s.labelColor)
		}

		screen.SetContent(x+i, y+height-1, sparkChars[charIndex], nil, tcell.StyleDefault.Foreground(s.labelColor))
	}
}