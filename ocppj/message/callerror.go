package message

import (
	"encoding/json"
	"fmt"
)

type Error string

const (
	// GenericError is a generic error code for callerror
	GenericError Error = "GenericError"
	// FormatViolationError is an error indicating action payload validation issue
	FormatViolationError Error = "FormatViolation"
	// InternalError is an error indicating there was an internal problem while handling of a call
	InternalError Error = "InternalError"
	// NotSupportedError is an error indicating that requested action is not supported
	NotSupportedError Error = "NotSupportedError"
	// SecurityError is an error indicating that the requested action is not allowed due to security reasons
	SecurityError Error = "SecurityError"
	// OccurrenceConstraintViolation indicates the payload is syntactically correct but at least one of the fields
	// violates occurrence constraints
	OccurrenceConstraintViolation Error = "OccurrenceConstraintViolation"
)

type CallError struct {
	MessageID        string
	ErrorCode        string
	ErrorDescription string
	ErrorDetails     json.RawMessage
}

type CallErrorPayload struct {
	ErrorCode        string          `json:"errorCode"`
	ErrorDescription string          `json:"errorDescription"`
	ErrorDetails     json.RawMessage `json:"errorDetails"`
}

func (c CallError) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		TypeCallError,
		c.MessageID,
		c.ErrorCode,
		c.ErrorDescription,
		c.ErrorDetails,
	})
}

func ParseCallError(parts []json.RawMessage) (CallError, error) {
	if len(parts) != 5 {
		return CallError{}, fmt.Errorf("expected CallError to contain 5 elements, received %d", len(parts))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return CallError{}, err
	}
	if messageTypeID != TypeCallError {
		return CallError{}, fmt.Errorf("invalid message type id: %v", messageTypeID)
	}

	var callError CallError
	if err := json.Unmarshal(parts[1], &callError.MessageID); err != nil {
		return CallError{}, err
	}
	if err := json.Unmarshal(parts[2], &callError.ErrorCode); err != nil {
		return CallError{}, err
	}
	if err := json.Unmarshal(parts[3], &callError.ErrorDescription); err != nil {
		return CallError{}, err
	}
	callError.ErrorDetails = parts[4]

	return callError, nil
}
