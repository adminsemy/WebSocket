package event

import (
	"encoding/json"
	"time"

	"github.com/adminsemy/WebSocket/internal/client"
)

var (
	EventSendMessage = "send_message"
	EvenNewMessage   = "new_message"
	ChangeChatRoom   = "change_chat_room"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event *Event, c *client.WebSocketClient) error

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type NewSendMessage struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

type ChangeChatRoomEvent struct {
	Name string `json:"name"`
}
