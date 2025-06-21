// /*********************************************************************************
// * Projeto:     Batedor
// * Script:      main.go
// * Autor:       Carlos Henrique Tourinho Santana (https://github.com/henriquetourinho/batedor)
// * Versão:      1.0
// * Data:        21 de Junho de 2025
// *
// * Descrição:   Este é o arquivo principal da aplicação Batedor.
// * Ele inicializa a interface do usuário com tview, gerencia as diferentes
// * telas (principal, histórico, ajuda), atualiza os widgets com dados
// * do sistema em tempo real (CPU, memória, rede, disco, processos)
// * e trata a interação do usuário via teclado. Também integra a
// * funcionalidade de log de métricas em um banco de dados SQLite
// * para visualização histórica.
// *********************************************************************************/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	gopsNet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// --- ESTRUTURAS DE DADOS ---
type AppState struct {
	processSortBy        string
	netBytesSentStart    uint64
	netBytesRecvStart    uint64
	lastNetBytesSent     uint64
	lastNetBytesRecv     uint64
	lastNetCheck         time.Time
	motherboardInfo      string
	publicIP             string
	latency              int64
	lastGlobalNetCheck   time.Time
	primaryInterfaceName string
	primaryInterfaceIP   string
	cpuUsage             float64
	memUsage             float64
}

type App struct {
	app           *tview.Application
	pages         *tview.Pages
	splash        *tview.TextView
	grid          *tview.Grid
	history       *HistoryGraph
	help          *tview.TextView
	confirmation  *tview.Modal
	cpuBox        *CPUBox
	memBox        *Sparkline
	netBox        *NetBox
	diskBox       *tview.TextView
	sysInfoBox    *tview.TextView
	processTable  *tview.Table
	processFilter *tview.InputField
	sortInfo      *tview.TextView
	state         AppState
}

type WebData struct {
	CPU   CPUData    `json:"CPU"`
	Mem   MemData    `json:"Mem"`
	Net   NetDataWeb `json:"Net"`
	Procs []ProcData `json:"Procs"`
}
type CPUData struct {
	Cores []float64 `json:"Cores"`
}
type MemData struct {
	UsedPercent float64 `json:"UsedPercent"`
}
type NetDataWeb struct {
	DownloadRate string `json:"DownloadRate"`
	UploadRate   string `json:"UploadRate"`
	PublicIP     string `json:"PublicIP"`
	Latency      int64  `json:"Latency"`
}
type ProcData struct {
	PID     int32   `json:"PID"`
	User    string  `json:"User"`
	CPU     float64 `json:"CPU"`
	Mem     float32 `json:"Mem"`
	Command string  `json:"Command"`
}

var webHub *Hub

// --- FUNÇÃO PRINCIPAL (main) ---
func main() {
	webFlag := flag.Bool("web", false, "Ativa o dashboard web na porta 9090")
	flag.Parse()

	if err := initDatabase(); err != nil {
		log.Fatalf("Falha ao inicializar banco de dados: %v", err)
	}

	if *webFlag {
		log.Println("Modo web ativado.")
		webHub = newHub()
		go webHub.run()
		go startWebServer(webHub)
	}

	app := NewApp()
	if err := app.Start(); err != nil {
		log.Fatalf("Erro ao iniciar aplicação TUI: %v", err)
	}
}

// --- FUNÇÕES AUXILIARES DE COLETA DE DADOS ---
func getPublicIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "N/A"
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "N/A"
	}
	return string(ip)
}

func getLatency() int64 {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", "8.8.8.8:53", 2*time.Second)
	if err != nil {
		return -1
	}
	defer conn.Close()
	return time.Since(start).Milliseconds()
}

func getPrimaryInterfaceInfo() (string, string) {
	ifaces, err := gopsNet.Interfaces()
	if err != nil {
		return "N/A", "N/A"
	}
	for _, iface := range ifaces {
		isUp := strings.Contains(strings.Join(iface.Flags, ","), "up")
		isLoopback := strings.Contains(strings.Join(iface.Flags, ","), "loopback")
		if isUp && !isLoopback && len(iface.Addrs) > 0 {
			localIP := strings.Split(iface.Addrs[0].Addr, "/")[0]
			return iface.Name, localIP
		}
	}
	return "N/A", "N/A"
}

func getMotherboardInfo() string {
	cmd := exec.Command("sh", "-c", "dmidecode -s baseboard-manufacturer && dmidecode -s baseboard-product-name")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "N/A"
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) >= 2 {
		return fmt.Sprintf("%s %s", lines[0], lines[1])
	}
	return strings.TrimSpace(out.String())
}

// --- LÓGICA DA APLICAÇÃO ---
func NewApp() *App {
	logo := `
██████╗  █████╗ ████████╗███████╗██████╗  ██████╗ ██████╗ 
██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██╔══██╗██╔═══██╗██╔══██╗
██████╔╝███████║   ██║   █████╗  ██████╔╝██║   ██║██████╔╝
██╔══██╗██╔══██║   ██║   ██╔══╝  ██╔══██╗██║   ██║██╔══██╗
██████╔╝██║  ██║   ██║   ███████╗██║  ██║╚██████╔╝██║  ██║
╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝
`
	autor := "Criado por: Carlos Henrique Tourinho Santana (github.com/henriquetourinho)"

	splashScreen := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("\n\n[green]%s\n\n\n[yellow]%s", logo, autor))

	helpText := `
[yellow]Batedor - Comandos do Teclado[-]

[green]Tela Principal:[-]
  [white]C[-]:      Ordernar processos por uso de CPU.
  [white]M[-]:      Ordernar processos por uso de Memória.
  [white]P[-]:      Ordernar processos por PID.
  [white]K[-]:      Encerrar o processo selecionado (pede confirmação).
  [white]H[-]:      Abrir a tela com o Histórico de uso de CPU/Memória.
  [white]F1[-]:     Exibir esta tela de Ajuda.
  [white]Q[-]:      Sair do Batedor.
  (Use as setas para cima/baixo para navegar na lista de processos)

[green]Tela de Histórico:[-]
  [white]C / M[-]:  Alternar entre o gráfico de CPU e Memória.
  [white]Q[-]:      Voltar para a tela principal.

[green]Tela de Ajuda:[-]
  (Pressione qualquer tecla para voltar)
`
	helpWidget := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetText(helpText)
	helpWidget.SetBorder(true).SetTitle(" Ajuda ")

	cpuWidget := NewCPUBox()
	memWidget := NewSparkline("Memória").SetLabelColor(tcell.ColorAqua)
	netWidget := NewNetBox()
	historyWidget := NewHistoryGraph()

	initialNetCounters, _ := gopsNet.IOCounters(false)
	var sentStart, recvStart uint64
	if len(initialNetCounters) > 0 {
		sentStart, recvStart = initialNetCounters[0].BytesSent, initialNetCounters[0].BytesRecv
	}
	ifaceName, ifaceIP := getPrimaryInterfaceInfo()

	a := &App{
		app:           tview.NewApplication(),
		pages:         tview.NewPages(),
		splash:        splashScreen,
		history:       historyWidget,
		help:          helpWidget,
		cpuBox:        cpuWidget,
		memBox:        memWidget,
		netBox:        netWidget,
		diskBox:       tview.NewTextView().SetDynamicColors(true),
		sysInfoBox:    tview.NewTextView().SetDynamicColors(true),
		processTable:  tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		processFilter: tview.NewInputField().SetLabel("Filtrar Processos (Nome): ").SetLabelColor(tcell.ColorYellow),
		sortInfo:      tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter),
		state: AppState{
			processSortBy:        "cpu",
			lastNetCheck:         time.Now(),
			motherboardInfo:      getMotherboardInfo(),
			netBytesSentStart:    sentStart,
			netBytesRecvStart:    recvStart,
			lastNetBytesSent:     sentStart,
			lastNetBytesRecv:     recvStart,
			lastGlobalNetCheck:   time.Now(),
			primaryInterfaceName: ifaceName,
			primaryInterfaceIP:   ifaceIP,
		},
	}

	go func() {
		a.state.publicIP = getPublicIP()
		a.state.latency = getLatency()
	}()

	a.diskBox.SetBorder(true).SetTitle("Uso de Disco")
	a.sysInfoBox.SetBorder(true).SetTitle("Informações do Sistema")
	a.processTable.SetBorder(true).SetTitle("Processos ([P]ID / [K]ill / [H]istórico / [F1] Ajuda)")
	a.sortInfo.SetBorder(true).SetTitle("Ordenação")

	a.confirmation = tview.NewModal().
		AddButtons([]string{"Sim", "Não"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) { a.pages.HidePage("confirmation") })

	a.grid = tview.NewGrid().
		SetRows(1, 0, 10, 0).
		SetColumns(0, 0, 0).
		SetBorders(true).
		AddItem(a.processFilter, 0, 0, 1, 3, 0, 0, true).
		AddItem(a.cpuBox, 1, 0, 1, 1, 0, 0, false).
		AddItem(a.memBox, 1, 1, 1, 1, 0, 0, false).
		AddItem(a.netBox, 1, 2, 1, 1, 0, 0, false).
		AddItem(a.diskBox, 2, 0, 1, 1, 0, 0, false).
		AddItem(a.sysInfoBox, 2, 1, 1, 1, 0, 0, false).
		AddItem(a.sortInfo, 2, 2, 1, 1, 0, 0, false).
		AddItem(a.processTable, 3, 0, 1, 3, 0, 0, true)

	a.pages.AddPage("splash", a.splash, true, true)
	a.pages.AddPage("main", a.grid, true, false)
	a.pages.AddPage("history", a.history, true, false)
	a.pages.AddPage("help", a.help, true, false)
	a.pages.AddPage("confirmation", a.confirmation, true, false)

	return a
}

func (a *App) Start() error {
	go func() {
		logTicker := time.NewTicker(1 * time.Minute)
		defer logTicker.Stop()
		for {
			<-logTicker.C
			logMetric("cpu_usage", a.state.cpuUsage)
			logMetric("mem_usage", a.state.memUsage)
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			a.collectAndDistributeData()
			<-ticker.C
		}
	}()

	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		frontPage, _ := a.pages.GetFrontPage()
		if frontPage == "help" {
			a.pages.SwitchToPage("main")
			return nil
		}
		if frontPage == "history" {
			switch event.Rune() {
			case 'q', 'Q':
				a.pages.SwitchToPage("main")
			case 'c', 'C', 'm', 'M':
				a.history.ToggleMetric()
			}
			return event
		}
		if frontPage != "main" {
			return event
		}
		switch event.Key() {
		case tcell.KeyF1:
			a.pages.SwitchToPage("help")
			return nil
		case tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}
		switch event.Rune() {
		case 'q', 'Q':
			a.app.Stop()
		case 'k', 'K':
			a.showKillConfirmation()
			return nil
		case 'h', 'H':
			a.history.LoadData()
			a.pages.SwitchToPage("history")
			return nil
		case 'c', 'C':
			a.state.processSortBy = "cpu"
		case 'm', 'M':
			a.state.processSortBy = "mem"
		case 'p', 'P':
			a.state.processSortBy = "pid"
		}
		return event
	})

	go func() {
		time.Sleep(3 * time.Second)
		a.app.QueueUpdateDraw(func() {
			a.pages.SwitchToPage("main")
		})
	}()

	return a.app.SetRoot(a.pages, true).Run()
}

func (a *App) collectAndDistributeData() {
	allCores, _ := cpu.Percent(0, true)
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Usage("/")
	hostInfo, _ := host.Info()
	netCounters, _ := gopsNet.IOCounters(false)
	procs, _ := process.Processes()

	now := time.Now()
	duration := now.Sub(a.state.lastNetCheck).Seconds()
	var recvRate, sentRate uint64
	var currentNet gopsNet.IOCountersStat
	if duration > 0.1 && len(netCounters) > 0 {
		currentNet = netCounters[0]
		recvRate = uint64(float64(currentNet.BytesRecv-a.state.lastNetBytesRecv) / duration)
		sentRate = uint64(float64(currentNet.BytesSent-a.state.lastNetBytesSent) / duration)
		a.state.lastNetBytesRecv = currentNet.BytesRecv
		a.state.lastNetBytesSent = currentNet.BytesSent
	}
	a.state.lastNetCheck = now

	if time.Since(a.state.lastGlobalNetCheck) > 30*time.Second {
		go func() {
			a.state.publicIP = getPublicIP()
			a.state.latency = getLatency()
		}()
		a.state.lastGlobalNetCheck = time.Now()
	}

	var totalCPU float64
	for _, c := range allCores {
		totalCPU += c
	}
	if len(allCores) > 0 {
		a.state.cpuUsage = totalCPU / float64(len(allCores))
	}
	if memInfo != nil {
		a.state.memUsage = memInfo.UsedPercent
	}

	a.app.QueueUpdateDraw(func() {
		a.updateAllTUIWidgets(allCores, memInfo, diskInfo, hostInfo, &currentNet, recvRate, sentRate, procs)
	})

	if webHub != nil {
		webData := a.prepareWebData(allCores, memInfo, recvRate, sentRate, procs)
		jsonData, err := json.Marshal(webData)
		if err == nil {
			webHub.broadcast <- jsonData
		}
	}
}

func (a *App) updateAllTUIWidgets(allCores []float64, memInfo *mem.VirtualMemoryStat, diskInfo *disk.UsageStat, hostInfo *host.InfoStat, currentNet *gopsNet.IOCountersStat, recvRate, sentRate uint64, procs []*process.Process) {
	if len(allCores) > 0 {
		a.cpuBox.Update(allCores)
	}
	if memInfo != nil {
		a.memBox.AddData(memInfo.UsedPercent)
	}

	if diskInfo != nil {
		a.diskBox.SetText(fmt.Sprintf("[yellow]Total: [white]%.2f GB\n[green]Usado: [white]%.2f GB (%.2f%%)\n[blue]Livre: [white]%.2f GB",
			float64(diskInfo.Total)/1e9, float64(diskInfo.Used)/1e9, diskInfo.UsedPercent, float64(diskInfo.Free)/1e9))
	}
	if hostInfo != nil {
		uptimeString := (time.Duration(hostInfo.Uptime) * time.Second).String()
		a.sysInfoBox.SetText(fmt.Sprintf("[yellow]Hostname: [white]%s\n[yellow]SO: [white]%s\n[yellow]Placa-Mãe: [white]%s\n[yellow]Atividade: [white]%s",
			hostInfo.Hostname, hostInfo.Platform, a.state.motherboardInfo, uptimeString))
	}

	if currentNet != nil {
		netInfoData := NetInfo{
			DownloadRate:    recvRate,
			UploadRate:      sentRate,
			DownloadSession: currentNet.BytesRecv - a.state.netBytesRecvStart,
			UploadSession:   currentNet.BytesSent - a.state.netBytesSentStart,
			PublicIP:        a.state.publicIP,
			Latency:         a.state.latency,
			InterfaceName:   a.state.primaryInterfaceName,
			LocalIP:         a.state.primaryInterfaceIP,
		}
		a.netBox.Update(netInfoData)
	}

	a.updateProcessTable(procs)
	a.sortInfo.SetText(fmt.Sprintf("Ordenando por: [yellow]%s", strings.ToUpper(a.state.processSortBy)))
}

func (a *App) prepareWebData(cores []float64, memInfo *mem.VirtualMemoryStat, recvRate, sentRate uint64, procs []*process.Process) WebData {
	procDataList := []ProcData{}
	filter := strings.ToLower(a.processFilter.GetText())

	for _, p := range procs {
		cmd, _ := p.Name()
		if filter != "" && !strings.Contains(strings.ToLower(cmd), filter) {
			continue
		}
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		user, _ := p.Username()

		if cpuPercent > 0.01 || memPercent > 0.1 {
			procDataList = append(procDataList, ProcData{
				PID:     p.Pid,
				User:    user,
				CPU:     cpuPercent,
				Mem:     memPercent,
				Command: cmd,
			})
		}
	}

	sort.Slice(procDataList, func(i, j int) bool {
		switch a.state.processSortBy {
		case "mem":
			return procDataList[i].Mem > procDataList[j].Mem
		case "pid":
			return procDataList[i].PID < procDataList[j].PID
		default:
			return procDataList[i].CPU > procDataList[j].CPU
		}
	})

	if len(procDataList) > 50 {
		procDataList = procDataList[:50]
	}

	return WebData{
		CPU: CPUData{Cores: cores},
		Mem: MemData{UsedPercent: memInfo.UsedPercent},
		Net: NetDataWeb{
			DownloadRate: formatBytes(recvRate),
			UploadRate:   formatBytes(sentRate),
			PublicIP:     a.state.publicIP,
			Latency:      a.state.latency,
		},
		Procs: procDataList,
	}
}

func (a *App) updateProcessTable(procs []*process.Process) {
	filter := strings.ToLower(a.processFilter.GetText())
	type ProcInfo struct {
		pid     int32
		user    string
		command string
		cpu     float64
		mem     float32
	}
	var procList []ProcInfo
	for _, p := range procs {
		cmd, _ := p.Name()
		if filter != "" && !strings.Contains(strings.ToLower(cmd), filter) {
			continue
		}
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		if cpuPercent < 0.01 && memPercent < 0.01 {
			continue
		}
		user, _ := p.Username()
		procList = append(procList, ProcInfo{p.Pid, user, cmd, cpuPercent, memPercent})
	}
	sort.Slice(procList, func(i, j int) bool {
		switch a.state.processSortBy {
		case "mem":
			return procList[i].mem > procList[j].mem
		case "pid":
			return procList[i].pid < procList[j].pid
		default:
			return procList[i].cpu > procList[j].cpu
		}
	})

	a.processTable.Clear()
	headers := []string{"PID", "Usuário", "CPU%", "MEM%", "Comando"}
	for i, header := range headers {
		a.processTable.SetCell(0, i, tview.NewTableCell(header).SetTextColor(tcell.ColorYellow).SetSelectable(false))
	}
	for i, p := range procList {
		row := i + 1
		a.processTable.SetCell(row, 0, tview.NewTableCell(strconv.Itoa(int(p.pid))).SetTextColor(tcell.ColorWhite))
		a.processTable.SetCell(row, 1, tview.NewTableCell(p.user).SetTextColor(tcell.ColorBlue))
		a.processTable.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.2f", p.cpu)).SetTextColor(tcell.ColorGreen))
		a.processTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%.2f", p.mem)).SetTextColor(tcell.ColorGreen))
		a.processTable.SetCell(row, 4, tview.NewTableCell(p.command).SetTextColor(tcell.ColorWhite))
	}
}

func (a *App) showKillConfirmation() {
	row, _ := a.processTable.GetSelection()
	if row <= 0 {
		return
	}
	pidCell := a.processTable.GetCell(row, 0)
	cmdCell := a.processTable.GetCell(row, 4)
	pid, err := strconv.Atoi(pidCell.Text)
	if err != nil {
		return
	}
	a.confirmation.SetText(fmt.Sprintf("Você tem certeza que deseja encerrar o processo %d (%s)?", pid, cmdCell.Text))
	a.confirmation.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Sim" {
			p, err := os.FindProcess(pid)
			if err == nil {
				p.Signal(syscall.SIGTERM)
			}
		}
		a.pages.HidePage("confirmation")
		a.app.SetFocus(a.processTable)
	})
	a.pages.ShowPage("confirmation")
}