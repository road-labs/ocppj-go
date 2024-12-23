package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

type Type int

var (
	// ErrInvalidUTF8 is returned when the message contains invalid UTF-8 byte sequences.
	ErrInvalidUTF8 = errors.New("message contains invalid UTF-8")
	// ErrInvalidJSON is returned if the message contains malformed json
	ErrInvalidJSON = errors.New("message contains invalid JSON")
	// ErrInvalidPayload is returned if the message is valid JSON, but is otherwise malformed (e.g. wrong number of
	// fields, or fields containing the wrong type).
	ErrInvalidPayload = errors.New("message format is not valid")
	// ErrUnknownMessageType is returned when the message type of OCPP is unknown
	ErrUnknownMessageType = errors.New("unknown message type")
)

const (
	TypeCall       Type = 2
	TypeCallResult Type = 3
	TypeCallError  Type = 4
)

type Message any

// FromJSON builds a message from JSON data
// the first return value is the message struct that will need to be type asserted. It can be any of the following:
// - Call
// - CallResult
// - CallError
func FromJSON(data []byte) (Message, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("message cannot be parsed: %w", ErrInvalidUTF8)
	}

	var parts []json.RawMessage
	if err := json.Unmarshal(data, &parts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", errors.Join(ErrInvalidJSON, err))
	}

	if len(parts) == 0 {
		return nil, errors.Join(ErrInvalidPayload, errors.New("message type ID required"))
	}

	var messageTypeID Type
	if err := json.Unmarshal(parts[0], &messageTypeID); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message type ID: %w", errors.Join(ErrInvalidPayload, err))
	}

	var (
		msg Message
		err error
	)
	switch messageTypeID {
	case TypeCall:
		msg, err = ParseCall(parts)
	case TypeCallResult:
		msg, err = ParseCallResult(parts)
	case TypeCallError:
		msg, err = ParseCallError(parts)
	default:
		return nil, fmt.Errorf("failed to handle message type %v: %w", messageTypeID, ErrUnknownMessageType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", errors.Join(ErrInvalidPayload, err))
	}

	return msg, nil
}
