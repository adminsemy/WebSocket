package handler

import (
	"encoding/json"
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
	go m.WriteMessageToAll()
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
		event, err := m.toReadMessage(message)
		if err != nil {
			slog.Error("Failed to parse message", "error", err)
			continue
		}
		sendMessage, err := m.toSendMessage(event)
		if err != nil {
			slog.Error("Failed to parse message", "error", err)
			continue
		}
		m.RLock()
		slog.Info("Writing message to all clients", "clients", m.clients, "message", string(sendMessage))
		for client := range m.clients {
			client.WriteChan <- sendMessage
		}
		m.RUnlock()
	}
}

func (m *Manager) toReadMessage(data []byte) (event.Event, error) {
	var e event.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return e, err
	}
	return e, nil
}

func (m *Manager) toSendMessage(e event.Event) ([]byte, error) {
	sendMessage := event.SendMessageEvent{
		Message: string(e.Payload),
		From:    "Server",
	}
	b, err := json.Marshal(sendMessage)
	if err != nil {
		return nil, err
	}
	return b, nil
}
