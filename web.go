// /*********************************************************************************
// * Projeto:     Batedor
// * Componente:  Web - Servidor Web e WebSocket
// *********************************************************************************/
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Permite todas as origens
}

// Hub mantém o conjunto de clientes ativos.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case message := <-h.broadcast:
			for conn := range h.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("error: %v", err)
					h.unregister <- conn
				}
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hub.register <- conn

	// Garante que o cliente seja desregistrado ao sair.
	defer func() { hub.unregister <- conn }()
	for {
		// Apenas mantém a conexão viva. Não precisamos ler nada do cliente.
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

// startWebServer inicializa as rotas e o servidor.
func startWebServer(hub *Hub) {
	// Rota para servir os arquivos estáticos (index.html)
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Rota para a conexão WebSocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Dashboard web iniciado em http://localhost:9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("Falha ao iniciar servidor web: %v", err)
	}
}