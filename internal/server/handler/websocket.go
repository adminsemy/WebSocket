package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/adminsemy/WebSocket/internal/client"
	"github.com/adminsemy/WebSocket/internal/client/event"
	"github.com/adminsemy/WebSocket/internal/client/oneTimePassword"
	"github.com/gorilla/websocket"
)

type Manager struct {
	clients   map[*client.WebSocketClient]bool
	delete    chan *client.WebSocketClient
	messageCh chan *client.Message
	sync.RWMutex
	handlers map[string]event.EventHandler
	opts     oneTimePassword.RetentionMap
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:   make(map[*client.WebSocketClient]bool),
		delete:    make(chan *client.WebSocketClient),
		messageCh: make(chan *client.Message),
		handlers:  make(map[string]event.EventHandler),
		opts:      oneTimePassword.NewRetentionMap(context.Background(), 5*time.Minute),
	}
	m.SetupHandlers()
	go m.ReadMessage()
	return m
}

func (m *Manager) SetupHandlers() {
	m.handlers[event.EventSendMessage] = func(e *event.Event, c *client.WebSocketClient) error {
		var sendMessage event.SendMessageEvent

		err := json.Unmarshal(e.Payload, &sendMessage)
		if err != nil {
			return err
		}
		newMessage := event.NewSendMessage{
			SendMessageEvent: sendMessage,
			Send:             time.Now(),
		}
		data, err := json.Marshal(newMessage)
		if err != nil {
			return err
		}
		m.RLock()
		slog.Info("Writing message to all clients", "clients", m.clients, "message", string(data))
		for client := range m.clients {
			client.WriteChan <- data
		}
		m.RUnlock()
		return nil
	}
	m.handlers[event.ChangeChatRoom] = func(e *event.Event, c *client.WebSocketClient) error {
		var chatRoomEvent event.ChangeChatRoomEvent
		if err := json.Unmarshal(e.Payload, &chatRoomEvent); err != nil {
			return err
		}
		c.ChatRoom = chatRoomEvent.Name
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

func (m *Manager) Login(w http.ResponseWriter, r *http.Request) {
	slog.Info("Serving Login request")
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var user LoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		slog.Error("Failed to parse login request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user.Username == "admin" && user.Password == "admin" {
		type response struct {
			OTP string `json:"otp"`
		}

		otp := m.opts.NewOTP()

		resp := response{
			OTP: otp.OneTimePassword,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			slog.Error("Failed to marshal response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	otp := r.URL.Query().Get("otp")
	slog.Info("Serving Websocket otp", "otp", otp)
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !m.opts.Verify(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
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

func (m *Manager) ReadMessage() {
	for {
		message := <-m.messageCh
		event, err := m.toReadMessage(message.Payload)
		if err != nil {
			slog.Error("Failed to parse message", "error", err)
			continue
		}
		err = m.routeEvent(event, message.Client)
		if err != nil {
			slog.Error("Failed to route message", "error", err)
			continue
		}

	}
}

func (m *Manager) toReadMessage(data []byte) (event.Event, error) {
	var e event.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return e, err
	}
	return e, nil
}
