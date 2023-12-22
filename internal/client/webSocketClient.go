package client

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	connection *websocket.Conn
}

func NewWebSocketClient(connection *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		connection: connection,
	}
}

func (c *WebSocketClient) ReadMessages() {
	for {
		messageType, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("error reading message: %v", err)
			}
			break
		}
		slog.Info("Message Type: ", messageType)
		slog.Info("Payload: ", payload)
	}
}
