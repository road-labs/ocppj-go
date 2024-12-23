package message

import (
	"encoding/json"
	"fmt"
)

type CallResult struct {
	MessageID string
	Payload   json.RawMessage
}

func (c CallResult) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		TypeCallResult,
		c.MessageID,
		c.Payload,
	})
}

func ParseCallResult(parts []json.RawMessage) (CallResult, error) {
	if len(parts) != 3 {
		return CallResult{}, fmt.Errorf("expected CallResult to contain 3 elements, received %d", len(parts))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return CallResult{}, err
	}
	if messageTypeID != TypeCallResult {
		return CallResult{}, fmt.Errorf("invalid message type id: %v", messageTypeID)
	}

	var cr CallResult
	if err := json.Unmarshal(parts[1], &cr.MessageID); err != nil {
		return CallResult{}, err
	}
	cr.Payload = parts[2]
	return cr, nil
}
