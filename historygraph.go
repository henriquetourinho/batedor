// /*********************************************************************************
// * Projeto:     Batedor
// * Componente:  HistoryGraph - Widget para visualização de dados históricos (Corrigido)
// *********************************************************************************/
package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HistoryGraph é o widget para desenhar o gráfico histórico.
type HistoryGraph struct {
	*tview.Box
	mu         sync.RWMutex
	data       []MetricRecord
	metricName string
	maxVal     float64
	minVal     float64
}

func NewHistoryGraph() *HistoryGraph {
	return &HistoryGraph{
		Box:        tview.NewBox().SetBorder(true),
		metricName: "CPU",
	}
}

// LoadData busca os dados do banco e prepara o widget para desenhar.
func (h *HistoryGraph) LoadData() {
	h.mu.Lock()
	defer h.mu.Unlock()

	metricToFetch := "cpu_usage"
	if h.metricName == "Memória" {
		metricToFetch = "mem_usage"
	}

	data, err := getMetricsForLast24h(metricToFetch)
	if err != nil {
		h.data = []MetricRecord{}
		return
	}

	h.data = data
	// MUDANÇA: O texto aqui foi simplificado e corrigido.
	h.SetTitle(fmt.Sprintf(" Histórico de Uso de %s (Últimas 24h) | [C]/[M] Trocar | [Q] Sair ", h.metricName))

	if len(h.data) > 0 {
		h.maxVal = h.data[0].Value
		h.minVal = h.data[0].Value
		for _, rec := range h.data {
			if rec.Value > h.maxVal {
				h.maxVal = rec.Value
			}
			if rec.Value < h.minVal {
				h.minVal = rec.Value
			}
		}
	} else {
		h.maxVal = 100
		h.minVal = 0
	}
}

// ToggleMetric alterna entre CPU e Memória.
func (h *HistoryGraph) ToggleMetric() {
	h.mu.Lock()
	if h.metricName == "CPU" {
		h.metricName = "Memória"
	} else {
		h.metricName = "CPU"
	}
	h.mu.Unlock()
	h.LoadData()
}

// Draw desenha o gráfico na tela.
func (h *HistoryGraph) Draw(screen tcell.Screen) {
	h.Box.Draw(screen)
	h.mu.RLock()
	defer h.mu.RUnlock()

	x, y, width, height := h.GetInnerRect()
	if width <= 2 || height <= 2 || len(h.data) == 0 {
		tview.Print(screen, "Coletando dados históricos... (Aguarde alguns minutos)", x+1, y+(height/2), width-2, tview.AlignCenter, tcell.ColorYellow)
		return
	}

	yAxisLabelMax := fmt.Sprintf("%.0f%%", h.maxVal)
	yAxisLabelMin := fmt.Sprintf("%.0f%%", h.minVal)
	tview.Print(screen, yAxisLabelMax, x, y, width-2, tview.AlignLeft, tcell.ColorYellow)
	tview.Print(screen, yAxisLabelMin, x, y+height-1, width-2, tview.AlignLeft, tcell.ColorYellow)

	xAxisLabelStart := h.data[0].Timestamp.Format("15:04")
	xAxisLabelEnd := h.data[len(h.data)-1].Timestamp.Format("15:04")
	tview.Print(screen, xAxisLabelStart, x+3, y+height-1, width-2, tview.AlignLeft, tcell.ColorYellow)
	tview.Print(screen, xAxisLabelEnd, x+width-len(xAxisLabelEnd)-2, y+height-1, width-2, tview.AlignLeft, tcell.ColorYellow)

	valRange := h.maxVal - h.minVal
	if valRange == 0 {
		valRange = 1
	}

	for i := 0; i < width-4; i++ {
		dataIndex := int(float64(i) / float64(width-4) * float64(len(h.data)))
		if dataIndex >= len(h.data) {
			dataIndex = len(h.data) - 1
		}

		val := h.data[dataIndex].Value
		yPos := int(float64(height-2) * (1 - (val-h.minVal)/valRange))

		char := '•'
		color := tcell.ColorAqua
		if h.metricName == "CPU" {
			color = tcell.ColorGreen
		}

		screen.SetContent(x+i+2, y+yPos, char, nil, tcell.StyleDefault.Foreground(color))
	}
}