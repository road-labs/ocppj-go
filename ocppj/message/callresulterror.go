package message

import (
	"encoding/json"
	"fmt"
)

type CallResultError struct {
	MessageID        string
	ErrorCode        string
	ErrorDescription string
	ErrorDetails     json.RawMessage
}

func (c CallResultError) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		TypeCallResultError,
		c.MessageID,
		c.ErrorCode,
		c.ErrorDescription,
		c.ErrorDetails,
	})
}

func ParseCallResultError(parts []json.RawMessage) (CallResultError, error) {
	if len(parts) != 5 {
		return CallResultError{}, fmt.Errorf("expected CallResultError to contain 5 elements, received %d", len(parts))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return CallResultError{}, err
	}
	if messageTypeID != TypeCallResultError {
		return CallResultError{}, fmt.Errorf("invalid message type id: %v", messageTypeID)
	}

	var callResultError CallResultError
	if err := json.Unmarshal(parts[1], &callResultError.MessageID); err != nil {
		return CallResultError{}, err
	}
	if err := json.Unmarshal(parts[2], &callResultError.ErrorCode); err != nil {
		return CallResultError{}, err
	}
	if err := json.Unmarshal(parts[3], &callResultError.ErrorDescription); err != nil {
		return CallResultError{}, err
	}
	callResultError.ErrorDetails = parts[4]

	return callResultError, nil
}
