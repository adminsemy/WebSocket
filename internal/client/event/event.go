package event

import (
	"encoding/json"

	"github.com/adminsemy/WebSocket/internal/client"
)

var EventSendMessage = "send_message"

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event *Event, c *client.WebSocketClient) error

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
