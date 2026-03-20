package message

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSend_MarshalJSON(t *testing.T) {
	send := Send{
		MessageID: "1",
		Action:    "2",
		Payload:   []byte(`{"foo": "bar"}`),
	}

	b, err := send.MarshalJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `[6,"1","2",{"foo":"bar"}]`
	if string(b) != expected {
		t.Errorf("expected %s, got %s", expected, string(b))
	}
}

func TestParseSend(t *testing.T) {
	var parts []json.RawMessage
	msg := []byte(`[6,"1","2",{"foo":"bar"}]`)
	err := json.Unmarshal(msg, &parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	messageParsed, err := ParseSend(parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := Send{
		MessageID: "1",
		Action:    "2",
		Payload:   []byte(`{"foo":"bar"}`),
	}

	if !cmp.Equal(expected, messageParsed) {
		t.Errorf("expected %v, got %v", expected, messageParsed)
	}
}
