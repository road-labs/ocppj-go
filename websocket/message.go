package websocket

import "github.com/gorilla/websocket"

type Message struct {
	Type MessageType
	Data []byte
}

type MessageType int

const (
	MessageTypeText  MessageType = websocket.TextMessage
	MessageTypeClose MessageType = websocket.CloseMessage
	MessageTypePing  MessageType = websocket.PingMessage
	MessageTypePong  MessageType = websocket.PongMessage
)

func (m MessageType) IsControl() bool {
	return m == MessageTypeClose ||
		m == MessageTypePing ||
		m == MessageTypePong
}

func TextMessage(data []byte) Message {
	return Message{
		Type: MessageTypeText,
		Data: data,
	}
}
