package message

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCallResult_MarshalJSON(t *testing.T) {
	call := CallResult{
		MessageID: "1",
		Payload:   []byte(`{"foo": "bar"}`),
	}

	b, err := call.MarshalJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := `[3,"1",{"foo":"bar"}]`
	if string(b) != expected {
		t.Errorf("expected %s, got %s", expected, string(b))
	}
}

func TestParseCallResult(t *testing.T) {
	var parts []json.RawMessage
	msg := []byte(`[3,"1",{"foo":"bar"}]`)
	err := json.Unmarshal(msg, &parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	messageParsed, err := ParseCallResult(parts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := CallResult{
		MessageID: "1",
		Payload:   []byte(`{"foo":"bar"}`),
	}

	if !cmp.Equal(expected, messageParsed) {
		t.Errorf("expected %v, got %v", expected, messageParsed)
	}
}
