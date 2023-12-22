package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/adminsemy/WebSocket/internal/client"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients map[*client.WebSocketClient]bool
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[*client.WebSocketClient]bool),
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
	go client.ReadMessages()
}

func (m *Manager) AddClient(client *client.WebSocketClient) {
	m.Lock()
	m.clients[client] = true
	m.Unlock()
}

func (m *Manager) RemoveClient(client *client.WebSocketClient) {
	m.Lock()
	delete(m.clients, client)
	m.Unlock()
}
