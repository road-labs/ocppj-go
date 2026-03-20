package message_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/e-flux-platform/ocppj-go/ocppj/message"
)

func TestBuildMessageFromJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid Call message", func(t *testing.T) {
		t.Parallel()

		data := []byte(`[2,"1","GetConfiguration",{"foo":"bar"}]`)
		msg, err := message.FromJSON(data)
		assert.Nil(t, err)
		assert.Equal(t, msg.(message.Call), message.Call{
			Action:    "GetConfiguration",
			MessageID: "1",
			Payload:   []byte(`{"foo":"bar"}`),
		})
	})

	t.Run("valid CallResult message", func(t *testing.T) {
		t.Parallel()

		data := []byte(`[3,"1",{"foo":"bar"}]`)
		msg, err := message.FromJSON(data)
		assert.Nil(t, err)
		assert.Equal(t, msg.(message.CallResult), message.CallResult{
			MessageID: "1",
			Payload:   []byte(`{"foo":"bar"}`),
		})
	})

	t.Run("valid CallError message", func(t *testing.T) {
		t.Parallel()

		data := []byte(`[4,"1","GenericError","error",{"foo":"bar"}]`)
		msg, err := message.FromJSON(data)
		assert.Nil(t, err)
		assert.Equal(t, msg.(message.CallError), message.CallError{
			MessageID:        "1",
			ErrorCode:        "GenericError",
			ErrorDescription: "error",
			ErrorDetails:     []byte(`{"foo":"bar"}`),
		})
	})

	t.Run("valid CallResultError message", func(t *testing.T) {
		t.Parallel()

		data := []byte(`[5,"1","GenericError","error",{"foo":"bar"}]`)
		msg, err := message.FromJSON(data)
		assert.Nil(t, err)
		assert.Equal(t, msg.(message.CallResultError), message.CallResultError{
			MessageID:        "1",
			ErrorCode:        "GenericError",
			ErrorDescription: "error",
			ErrorDetails:     []byte(`{"foo":"bar"}`),
		})
	})

	t.Run("valid Send message", func(t *testing.T) {
		t.Parallel()

		data := []byte(`[6,"1","DataTransfer",{"foo":"bar"}]`)
		msg, err := message.FromJSON(data)
		assert.Nil(t, err)
		assert.Equal(t, msg.(message.Send), message.Send{
			MessageID: "1",
			Action:    "DataTransfer",
			Payload:   []byte(`{"foo":"bar"}`),
		})
	})

	t.Run("invalid messages", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			data          []byte
			expectedError error
		}{
			{
				name:          "invalid UTF-8",
				data:          []byte(`[2,"1","GetConfiguration",{"foo":"bar` + "\x8A" + `"}]`),
				expectedError: message.ErrInvalidUTF8,
			},
			{
				name:          "invalid JSON",
				data:          []byte(`[2,"1","GetConfiguration",{"foo":None}]`),
				expectedError: message.ErrInvalidJSON,
			},
			{
				name:          "empty array",
				data:          []byte(`[]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "invalid message type",
				data:          []byte(`["abc","1","GetConfiguration",{"foo":"bar"}]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "invalid message id",
				data:          []byte(`[2,123,"GetConfiguration",{"foo":"bar"}]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "missing element of call",
				data:          []byte(`[2,"1","GetConfiguration"]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "missing element of CallResultError",
				data:          []byte(`[5,"1","GenericError","error"]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "missing element of Send",
				data:          []byte(`[6,"1","DataTransfer"]`),
				expectedError: message.ErrInvalidPayload,
			},
			{
				name:          "message type unknown",
				data:          []byte(`[999,"1","GetConfiguration",{"foo":"bar"}]`),
				expectedError: message.ErrUnknownMessageType,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				_, err := message.FromJSON(tc.data)
				assert.ErrorIs(t, err, tc.expectedError)
			})
		}
	})
}
