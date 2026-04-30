package message

import (
	"encoding/json"
	"fmt"
)

type Send struct {
	MessageID string
	Action    string
	Payload   json.RawMessage
}

func (c Send) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		TypeSend,
		c.MessageID,
		c.Action,
		c.Payload,
	})
}

func ParseSend(parts []json.RawMessage) (Send, error) {
	if len(parts) != 4 {
		return Send{}, fmt.Errorf("expected Send to contain 4 elements, received %d", len(parts))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return Send{}, err
	}
	if messageTypeID != TypeSend {
		return Send{}, fmt.Errorf("invalid message type id: %v", messageTypeID)
	}

	var send Send
	if err := json.Unmarshal(parts[1], &send.MessageID); err != nil {
		return Send{}, err
	}
	if err := json.Unmarshal(parts[2], &send.Action); err != nil {
		return Send{}, err
	}
	send.Payload = parts[3]

	return send, nil
}
