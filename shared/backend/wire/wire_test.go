package wire

import (
	"encoding/json"
	"testing"
)

type view struct {
	You string `json:"you"`
}

type event struct {
	Type string `json:"type"`
}

func TestViewMessageAlwaysCarriesEvents(t *testing.T) {
	for _, m := range []ServerMessage[view, event]{
		ViewMessage[view, event](view{You: "p1"}, nil),
		{Type: MsgView, View: &view{You: "p1"}},
	} {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != `{"type":"view","view":{"you":"p1"},"events":[]}` {
			t.Fatalf("got %s", got)
		}
	}
	b, _ := json.Marshal(ViewMessage(view{You: "p1"}, []event{{Type: "x"}}))
	if got := string(b); got != `{"type":"view","view":{"you":"p1"},"events":[{"type":"x"}]}` {
		t.Fatalf("got %s", got)
	}
}

func TestBareMessages(t *testing.T) {
	for _, tt := range []struct {
		name string
		mk   func() ([]byte, error)
		want string
	}{
		{"presence nil", func() ([]byte, error) { return PresenceMessage(nil) }, `{"type":"presence","online":[]}`},
		{"presence", func() ([]byte, error) { return PresenceMessage([]string{"a", "b"}) }, `{"type":"presence","online":["a","b"]}`},
		{"ack", func() ([]byte, error) { return AckMessage("7") }, `{"type":"ack","requestId":"7"}`},
		{"error", func() ([]byte, error) { return ErrorMessage("", ErrorBody{Code: "bad_request", Message: "m"}) },
			`{"type":"error","error":{"code":"bad_request","message":"m"}}`},
	} {
		b, err := tt.mk()
		if err != nil || string(b) != tt.want {
			t.Errorf("%s: got %s, %v; want %s", tt.name, b, err, tt.want)
		}
	}
}

func TestDecodeIntoGenericMessage(t *testing.T) {
	var m ServerMessage[view, event]
	if err := json.Unmarshal([]byte(`{"type":"view","view":{"you":"p2"},"events":[{"type":"y"}]}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.View == nil || m.View.You != "p2" || len(m.Events) != 1 || m.Events[0].Type != "y" {
		t.Fatalf("decoded %+v", m)
	}
}
