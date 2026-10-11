package apidto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFromViewEmptySlicesMarshalAsArrays(t *testing.T) {
	s := mustJSON(t, FromView(domain.PlayerView{}))
	if !strings.Contains(s, `"players":[]`) {
		t.Errorf("missing empty players array in %s", s)
	}
	if got := mustJSON(t, FromEvents(nil)); got != "[]" {
		t.Errorf("FromEvents(nil) = %s", got)
	}
}

func TestFromViewLobby(t *testing.T) {
	g, err := domain.NewGame("K7MQ2", "p1", "Anna")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Ben"); err != nil {
		t.Fatal(err)
	}
	v, err := g.ViewFor("p2")
	if err != nil {
		t.Fatal(err)
	}
	want := PlayerView{
		GameID: "K7MQ2", Revision: 2, Phase: "lobby", You: "p2", HostID: "p1",
		Players: []PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, IsHost: true},
			{ID: "p2", Name: "Ben", Seat: 1, IsYou: true},
		},
	}
	if got := FromView(v); !reflect.DeepEqual(got, want) {
		t.Fatalf("FromView mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestFromSummary(t *testing.T) {
	created := time.Date(2026, 10, 10, 12, 30, 5, 0, time.UTC)
	tests := []struct {
		name string
		in   domain.GameSummary
		want GameSummary
	}{
		{
			"lobby",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseLobby, HostName: "Anna", PlayerCount: 1, MaxPlayers: 5, Joinable: true},
			GameSummary{GameID: "ABCDE", Phase: "lobby", HostName: "Anna", PlayerCount: 1, MaxPlayers: 5, Joinable: true},
		},
		{
			"created and ended",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseEnded, PlayerCount: 2, MaxPlayers: 5, CreatedAt: created.In(time.FixedZone("x", 3600))},
			GameSummary{GameID: "ABCDE", Phase: "ended", PlayerCount: 2, MaxPlayers: 5, CreatedAt: "2026-10-10T12:30:05Z"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromSummary(tt.in); got != tt.want {
				t.Fatalf("FromSummary = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFromEvents(t *testing.T) {
	evs := []domain.Event{
		domain.PlayerJoined{PlayerID: "p2", Name: "Ben"},
		domain.PlayerLeft{PlayerID: "p2", HostID: "p1"},
		domain.GameEnded{ByPlayerID: "p1"},
	}
	want := []string{
		`{"type":"playerJoined","playerId":"p2","name":"Ben"}`,
		`{"type":"playerLeft","playerId":"p2","hostId":"p1"}`,
		`{"type":"gameEnded","byPlayerId":"p1"}`,
	}
	got := FromEvents(evs)
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if s := mustJSON(t, got[i]); !sameJSON(s, want[i]) {
			t.Errorf("event %d = %s, want %s", i, s, want[i])
		}
	}
}

// sameJSON compares two JSON objects regardless of key order.
func sameJSON(a, b string) bool {
	var x, y any
	return json.Unmarshal([]byte(a), &x) == nil && json.Unmarshal([]byte(b), &y) == nil && reflect.DeepEqual(x, y)
}

func TestServerMessageJSON(t *testing.T) {
	tests := []struct {
		name string
		msg  ServerMessage
		want string
	}{
		{"view nil events", ServerMessage{Type: MsgView, View: &PlayerView{}}, `"events":[]`},
		{"view helper", ViewMessage(PlayerView{}, nil), `"events":[]`},
		{"view with events", ViewMessage(PlayerView{}, []Event{{Type: EventGameEnded}}), `"events":[{"type":"gameEnded"}]`},
		// presence, ack and error messages are tested in shared/backend/wire.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mustJSON(t, tt.msg)
			if !strings.Contains(s, tt.want) {
				t.Fatalf("got %s, want it to contain %s", s, tt.want)
			}
			if strings.Count(s, `"events"`) > 1 || strings.Count(s, `"online"`) > 1 {
				t.Fatalf("duplicate key in %s", s)
			}
		})
	}
	// Pointer marshal uses the same method.
	if s := mustJSON(t, &ServerMessage{Type: MsgView}); !strings.Contains(s, `"events":[]`) {
		t.Fatalf("pointer marshal = %s", s)
	}
	// Round trip.
	var back ServerMessage
	if err := json.Unmarshal([]byte(mustJSON(t, ViewMessage(PlayerView{GameID: "X"}, nil))), &back); err != nil {
		t.Fatal(err)
	}
	if back.Type != MsgView || back.View == nil || back.View.GameID != "X" || back.Events == nil {
		t.Fatalf("round trip = %+v", back)
	}
}

func TestClientMessageDecode(t *testing.T) {
	var m ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"endGame","requestId":"4"}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != CmdEndGame || m.RequestID != "4" {
		t.Fatalf("decoded %+v", m)
	}
}

func TestErrorFor(t *testing.T) {
	var syntaxErr error
	{
		var v any
		syntaxErr = json.Unmarshal([]byte("{x"), &v)
	}
	var typeErr error
	{
		var v CreateGameRequest
		typeErr = json.Unmarshal([]byte(`{"name":5}`), &v)
	}
	tests := []struct {
		name       string
		err        error
		status     int
		code       string
		messageHas string
	}{
		{"bad request", BadRequest("missing %s", "name"), 400, "bad_request", "missing name"},
		{"json syntax", fmt.Errorf("decode: %w", syntaxErr), 400, "bad_request", ""},
		{"json type", typeErr, 400, "bad_request", ""},
		{"max bytes", &http.MaxBytesError{Limit: 4096}, 400, "bad_request", ""},
		{"rate limited", ErrRateLimited, 429, "rate_limited", ""},
		{"unauthorized", ports.ErrUnauthorized, 401, "unauthorized", ""},
		{"not host", domain.ErrNotHost, 403, "not_host", ""},
		{"game not found", fmt.Errorf("load: %w", ports.ErrGameNotFound), 404, "game_not_found", ""},
		{"wrong phase", domain.ErrWrongPhase, 409, "wrong_phase", ""},
		{"game full", domain.ErrGameFull, 409, "game_full", ""},
		{"name taken", fmt.Errorf("%w: Anna", domain.ErrNameTaken), 409, "name_taken", "Anna"},
		{"cannot leave", domain.ErrCannotLeave, 409, "cannot_leave", ""},
		{"conflict", ports.ErrVersionConflict, 409, "conflict", ""},
		{"invalid name", domain.ErrInvalidName, 422, "invalid_name", ""},
		{"unknown player", domain.ErrUnknownPlayer, 422, "unknown_player", ""},
		{"not enough players", domain.ErrNotEnoughPlayers, 409, "not_enough_players", ""},
		{"duplicate player", domain.ErrDuplicatePlayer, 500, "internal", "internal error"},
		{"duplicate game id", ports.ErrDuplicateGameID, 500, "internal", "internal error"},
		{"other", errors.New("disk on fire"), 500, "internal", "internal error"},
		{"nil", nil, 500, "internal", "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := ErrorFor(tt.err)
			if status != tt.status || body.Code != tt.code {
				t.Fatalf("ErrorFor(%v) = %d %+v, want %d %s", tt.err, status, body, tt.status, tt.code)
			}
			if body.Message == "" || !strings.Contains(body.Message, tt.messageHas) {
				t.Fatalf("message %q, want it to contain %q", body.Message, tt.messageHas)
			}
			if tt.code == "internal" && body.Message != InternalMessage {
				t.Fatalf("internal message leaked: %q", body.Message)
			}
		})
	}
}
