package message

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCallResultError_MarshalJSON(t *testing.T) {
	call := CallResultError{
		MessageID:        "1",
		ErrorCode:        "2",
		ErrorDescription: "3",
		ErrorDetails:     []byte(`{"foo": "bar"}`),
	}

	b, err := call.MarshalJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `[5,"1","2","3",{"foo":"bar"}]`
	if string(b) != expected {
		t.Errorf("expected %s, got %s", expected, string(b))
	}
}

func TestParseCallResultError(t *testing.T) {
	var parts []json.RawMessage
	msg := []byte(`[5,"1","2","3",{"foo":"bar"}]`)
	err := json.Unmarshal(msg, &parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	messageParsed, err := ParseCallResultError(parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := CallResultError{
		MessageID:        "1",
		ErrorCode:        "2",
		ErrorDescription: "3",
		ErrorDetails:     []byte(`{"foo":"bar"}`),
	}

	if !cmp.Equal(expected, messageParsed) {
		t.Errorf("expected %v, got %v", expected, messageParsed)
	}
}
