package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/adminsemy/WebSocket/internal/client"
	"github.com/adminsemy/WebSocket/internal/client/event"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients   map[*client.WebSocketClient]bool
	delete    chan *client.WebSocketClient
	messageCh chan []byte
	sync.RWMutex
	handlers map[string]event.EventHandler
}

func NewManager() *Manager {
	m := &Manager{
		clients:   make(map[*client.WebSocketClient]bool),
		delete:    make(chan *client.WebSocketClient),
		messageCh: make(chan []byte),
		handlers:  make(map[string]event.EventHandler),
	}

	return m
}

func (m *Manager) SetupHandlers() {
	m.handlers[event.EventSendMessage] = func(e *event.Event, c *client.WebSocketClient) error {
		fmt.Println(e)
		return nil
	}
}

func (m *Manager) routeEvent(e event.Event, c *client.WebSocketClient) error {
	if handler, ok := m.handlers[e.Type]; ok {
		if err := handler(&e, c); err != nil {
			return err
		}
		return nil
	}
	return errors.New("event handler not found")
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
