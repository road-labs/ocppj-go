package message

import (
	"encoding/json"
	"fmt"
)

type Call struct {
	MessageID string
	Action    string
	Payload   json.RawMessage
}

func (c Call) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		TypeCall,
		c.MessageID,
		c.Action,
		c.Payload,
	})
}

func ParseCall(parts []json.RawMessage) (Call, error) {
	if len(parts) != 4 {
		return Call{}, fmt.Errorf("expected Call to contain 4 elements, received %d", len(parts))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return Call{}, err
	}
	if messageTypeID != TypeCall {
		return Call{}, fmt.Errorf("invalid message type id: %v", messageTypeID)
	}

	var call Call
	if err := json.Unmarshal(parts[1], &call.MessageID); err != nil {
		return Call{}, err
	}
	if err := json.Unmarshal(parts[2], &call.Action); err != nil {
		return Call{}, err
	}
	call.Payload = parts[3]

	return call, nil
}
