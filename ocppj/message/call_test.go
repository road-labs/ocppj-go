package message

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCall_MarshalJSON(t *testing.T) {
	call := Call{
		MessageID: "1",
		Action:    "2",
		Payload:   []byte(`{"foo": "bar"}`),
	}

	b, err := call.MarshalJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `[2,"1","2",{"foo":"bar"}]`
	if string(b) != expected {
		t.Errorf("expected %s, got %s", expected, string(b))
	}
}

func TestParseCall(t *testing.T) {
	var parts []json.RawMessage
	msg := []byte(`[2,"1","2",{"foo":"bar"}]`)
	err := json.Unmarshal(msg, &parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	messageParsed, err := ParseCall(parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := Call{
		MessageID: "1",
		Action:    "2",
		Payload:   []byte(`{"foo":"bar"}`),
	}

	if !cmp.Equal(expected, messageParsed) {
		t.Errorf("expected %v, got %v", expected, messageParsed)
	}
}
