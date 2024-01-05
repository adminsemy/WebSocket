package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/adminsemy/WebSocket/internal/client"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients   map[*client.WebSocketClient]bool
	delete    chan *client.WebSocketClient
	messageCh chan []byte
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients:   make(map[*client.WebSocketClient]bool),
		delete:    make(chan *client.WebSocketClient),
		messageCh: make(chan []byte),
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	slog.Info("Serving Websocket request")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade websocket", err)
	}
	client := client.NewWebSocketClient(conn)
	m.AddClient(client)
	go client.ReadMessages(m.delete, m.messageCh)
	go client.WriteMessage(m.delete)
	go m.RemoveClient()
	go m.WriteMessageToAll()
}

func (m *Manager) AddClient(client *client.WebSocketClient) {
	m.Lock()
	m.clients[client] = true
	m.Unlock()
}

func (m *Manager) RemoveClient() {
	var client *client.WebSocketClient
	for {
		client = <-m.delete
		m.Lock()
		delete(m.clients, client)
		m.Unlock()
		slog.Info("Client removed", "client", *client)
	}
}

func (m *Manager) WriteMessageToAll() {
	for {
		message := <-m.messageCh
		m.RLock()
		slog.Info("Writing message to all clients", "clients", m.clients)
		for client := range m.clients {
			client.WriteChan <- message
		}
		m.RUnlock()
	}
}
