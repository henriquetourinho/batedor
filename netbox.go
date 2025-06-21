// /*********************************************************************************
// * Projeto:     Batedor
// * Componente:  NetBox - Widget de rede avançado (Corrigido)
// *********************************************************************************/
package main

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NOVO: Copiamos a função para cá para que o NetBox possa usá-la.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B/s", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB/s", float64(b)/float64(div), "KMGTPE"[exp])
}

// NetInfo armazena todos os dados que nosso widget vai exibir.
type NetInfo struct {
	DownloadRate    uint64
	UploadRate      uint64
	DownloadSession uint64
	UploadSession   uint64
	InterfaceName   string
	LocalIP         string
	PublicIP        string
	Latency         int64
}

// NetBox é nosso widget customizado para a rede.
type NetBox struct {
	*tview.Box
	mu   sync.RWMutex
	info NetInfo
}

func NewNetBox() *NetBox {
	return &NetBox{
		Box: tview.NewBox().SetBorder(true).SetTitle("Rede"),
	}
}

// Update atualiza as informações de rede a serem exibidas.
func (n *NetBox) Update(info NetInfo) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.info = info
}

// formatBytesNetBox é uma versão local para formatar totais (sem o "/s").
func formatBytesNetBox(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Draw desenha o widget na tela.
func (n *NetBox) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	n.mu.RLock()
	defer n.mu.RUnlock()

	x, y, width, height := n.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Linha 1: Velocidades (CORRIGIDO para usar a função certa)
	downRateStr := formatBytes(n.info.DownloadRate)
	upRateStr := formatBytes(n.info.UploadRate)
	line1 := fmt.Sprintf("[red]Down: [white]%-15s [green]Up: [white]%s", downRateStr, upRateStr)
	tview.Print(screen, line1, x+1, y, width-2, tview.AlignLeft, tcell.ColorDefault)

	// Linha 2: Totais da Sessão
	downSessStr := formatBytesNetBox(n.info.DownloadSession)
	upSessStr := formatBytesNetBox(n.info.UploadSession)
	line2 := fmt.Sprintf("[red]Total D: [white]%-12s [green]Total U: [white]%s", downSessStr, upSessStr)
	tview.Print(screen, line2, x+1, y+1, width-2, tview.AlignLeft, tcell.ColorDefault)

	// Linha 3: IP Público e Latência
	latColor := "green"
	if n.info.Latency > 100 {
		latColor = "yellow"
	}
	if n.info.Latency > 200 {
		latColor = "red"
	}
	line3 := fmt.Sprintf("[yellow]IP Público: [white]%-15s [%s]Ping: [white]%dms", n.info.PublicIP, latColor, n.info.Latency)
	tview.Print(screen, line3, x+1, y+2, width-2, tview.AlignLeft, tcell.ColorDefault)

	// Linha 4: IP Local
	line4 := fmt.Sprintf("[yellow]IP Local:   [white]%s (%s)", n.info.LocalIP, n.info.InterfaceName)
	tview.Print(screen, line4, x+1, y+3, width-2, tview.AlignLeft, tcell.ColorDefault)
}