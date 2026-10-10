package apidto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
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
	v := FromView(domain.PlayerView{
		Turn: &domain.TurnView{},
		Hand: []domain.HandCard{{Card: domain.NumberCard(5)}},
	})
	s := mustJSON(t, v)
	for _, want := range []string{`"players":[]`, `"piles":[]`, `"plays":[]`, `"legalPiles":[]`, `"result":null`,
		`"activeCommands":[]`, `"skullPiles":[]`, `"exactCount":null`, `"boundPile":null`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	s = mustJSON(t, FromView(domain.PlayerView{}))
	for _, want := range []string{`"hand":[]`, `"turn":null`, `"result":null`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if got := mustJSON(t, FromEvents(nil)); got != "[]" {
		t.Errorf("FromEvents(nil) = %s", got)
	}
}

func TestFromViewMapsEveryField(t *testing.T) {
	top := domain.NumberCard(42)
	nbt := domain.Card{Value: 59, Kind: domain.CardKindNumber, Command: domain.CommandNoBackwardsTrick}
	bound := 1
	in := domain.PlayerView{
		GameID:   "K7MQ2",
		Revision: 9,
		Phase:    domain.PhaseFinished,
		You:      "p2",
		HostID:   "p1",
		Options:  domain.Options{Variant: domain.VariantClassic, Pro: true, ReducedHandSize: true},
		Rules:    domain.RulesView{MinCardsPerTurn: 3, MinCardsDrawPileEmpty: 1, HandSize: 5, MinPlayers: 1, MaxPlayers: 5},
		Players: []domain.PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, HandCount: 5, IsHost: true, IsActive: true},
			{ID: "p2", Name: "Ben", Seat: 1, HandCount: 4, IsYou: true},
		},
		Hand: []domain.HandCard{
			{Card: domain.NumberCard(7), LegalPiles: []int{0, 2}},
			{Card: domain.Card{Value: 81, Kind: domain.CardKindNumber, Command: domain.CommandExactlyThree}, Blocked: fmt.Errorf("%w: start hand 2", domain.ErrExactCountUnreachable)},
			{Card: domain.NumberCard(50), Blocked: errors.New("odd")},
		},
		Piles: []domain.PileView{
			{Index: 0, Direction: domain.DirectionUp, StartValue: 1, TopValue: 42, TopCard: &top, CardCount: 3},
			{Index: 2, Direction: domain.DirectionDown, StartValue: 100, TopValue: 59, TopCard: &nbt, CardCount: 4, CommandActive: true},
		},
		DrawPileCount: 60,
		Turn: &domain.TurnView{
			Number: 4, ActivePlayerID: "p1", CardsPlayed: 1, RequiredMinimum: 3, CanEndTurn: true,
			Plays:       []domain.Play{{Card: top, Pile: 0}},
			Constraints: domain.TurnConstraints{ExactCount: 3, SinglePile: true, BoundPile: &bound, SkullPiles: []int{0}, StopPlayed: true},
		},
		CommunicationRestricted: true,
		ActiveCommands:          []domain.ActiveCommand{{Command: domain.CommandNoBackwardsTrick, Pile: 2}},
		Result: &domain.ResultView{
			Outcome: domain.OutcomeLost, Score: 8, Excellent: true, CardsLeftInHands: 5, CardsLeftInDrawPile: 3, StuckPlayerID: "p1",
		},
	}
	wantTop := Card{Value: 42, Kind: "number"}
	wantNBT := Card{Value: 59, Kind: "number", Command: "noBackwardsTrick"}
	wantExact, wantBound := 3, 1
	want := PlayerView{
		GameID: "K7MQ2", Revision: 9, Phase: "finished", You: "p2", HostID: "p1",
		Options: GameOptions{Variant: "classic", Pro: true, ReducedHandSize: true},
		Rules:   Rules{MinCardsPerTurn: 3, MinCardsDrawPileEmpty: 1, HandSize: 5, MinPlayers: 1, MaxPlayers: 5},
		Players: []PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, HandCount: 5, IsHost: true, IsActive: true},
			{ID: "p2", Name: "Ben", Seat: 1, HandCount: 4, IsYou: true},
		},
		Hand: []HandCard{
			{Card: Card{Value: 7, Kind: "number"}, LegalPiles: []int{0, 2}},
			{Card: Card{Value: 81, Kind: "number", Command: "exactlyThree"}, LegalPiles: []int{}, Blocked: "exact_count_unreachable"},
			{Card: Card{Value: 50, Kind: "number"}, LegalPiles: []int{}},
		},
		Piles: []Pile{
			{Index: 0, Direction: "up", StartValue: 1, TopValue: 42, TopCard: &wantTop, CardCount: 3},
			{Index: 2, Direction: "down", StartValue: 100, TopValue: 59, TopCard: &wantNBT, CardCount: 4, CommandActive: true},
		},
		DrawPileCount: 60,
		Turn: &Turn{
			Number: 4, ActivePlayerID: "p1", CardsPlayed: 1, RequiredMinimum: 3, CanEndTurn: true,
			Plays:       []Play{{Card: wantTop, Pile: 0}},
			Constraints: TurnConstraints{ExactCount: &wantExact, SinglePile: true, BoundPile: &wantBound, SkullPiles: []int{0}, StopPlayed: true},
		},
		CommunicationRestricted: true,
		ActiveCommands:          []ActiveCommand{{Command: "noBackwardsTrick", Pile: 2}},
		Result: &Result{
			Outcome: "lost", Score: 8, Excellent: true, CardsLeftInHands: 5, CardsLeftInDrawPile: 3, StuckPlayerID: "p1",
		},
	}
	got := FromView(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromView mismatch:\n got %+v\nwant %+v", got, want)
	}
	// HandCard embeds Card: JSON is flat.
	if s := mustJSON(t, got.Hand[0]); s != `{"value":7,"kind":"number","legalPiles":[0,2]}` {
		t.Errorf("hand card JSON = %s", s)
	}
	// Mapped LegalPiles must not alias the domain slice.
	in.Hand[0].LegalPiles[0] = 3
	if got.Hand[0].LegalPiles[0] != 0 {
		t.Error("LegalPiles aliases domain slice")
	}
}

func TestFromSummary(t *testing.T) {
	got := FromSummary(domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseLobby, HostName: "Anna", PlayerCount: 2, MaxPlayers: 5, Joinable: true})
	want := GameSummary{GameID: "ABCDE", Phase: "lobby", HostName: "Anna", PlayerCount: 2, MaxPlayers: 5, Joinable: true}
	if got != want {
		t.Fatalf("FromSummary = %+v", got)
	}
}

func TestGameOptionsToDomain(t *testing.T) {
	tests := []struct {
		in   GameOptions
		want domain.Options
	}{
		{GameOptions{}, domain.Options{Variant: domain.VariantClassic}},
		{GameOptions{Variant: "classic", Pro: true, ReducedHandSize: true}, domain.Options{Variant: domain.VariantClassic, Pro: true, ReducedHandSize: true}},
		{GameOptions{Variant: "extreme"}, domain.Options{Variant: domain.VariantExtreme}},
	}
	for _, tt := range tests {
		if got := tt.in.ToDomain(); got != tt.want {
			t.Errorf("%+v.ToDomain() = %+v, want %+v", tt.in, got, tt.want)
		}
	}
	if got := FromOptions(domain.Options{Variant: domain.VariantClassic, Pro: true}); got != (GameOptions{Variant: "classic", Pro: true}) {
		t.Errorf("FromOptions = %+v", got)
	}
}

func TestFromEvents(t *testing.T) {
	evs := []domain.Event{
		domain.PlayerJoined{PlayerID: "p2", Name: "Ben"},
		domain.OptionsChanged{Options: domain.DefaultOptions()},
		domain.GameStarted{},
		domain.FirstPlayerChosen{PlayerID: "p1"},
		domain.TurnStarted{PlayerID: "p1", Number: 1},
		domain.CardPlayed{PlayerID: "p1", Card: domain.NumberCard(12), Pile: 0},
		domain.TurnEnded{PlayerID: "p1", CardsDrawn: 0, Auto: true},
		domain.PlayerSkipped{PlayerID: "p2"},
		domain.GameFinished{Outcome: domain.OutcomeWon, Score: 0},
		domain.GameFinished{Outcome: domain.OutcomeLost, Score: 12, StuckPlayerID: "p2"},
	}
	want := []string{
		`{"type":"playerJoined","playerId":"p2","name":"Ben"}`,
		`{"type":"optionsChanged"}`,
		`{"type":"gameStarted"}`,
		`{"type":"firstPlayerChosen","playerId":"p1"}`,
		`{"type":"turnStarted","playerId":"p1","number":1}`,
		`{"type":"cardPlayed","playerId":"p1","card":{"value":12,"kind":"number"},"pile":0}`,
		`{"type":"turnEnded","playerId":"p1","cardsDrawn":0,"auto":true}`,
		`{"type":"playerSkipped","playerId":"p2"}`,
		`{"type":"gameFinished","outcome":"won","score":0}`,
		`{"type":"gameFinished","outcome":"lost","score":12,"stuckPlayerId":"p2"}`,
	}
	got := FromEvents(evs)
	if len(got) != len(want) {
		t.Fatalf("len = %d", len(got))
	}
	for i := range want {
		if s := mustJSON(t, got[i]); s != want[i] {
			t.Errorf("event %d = %s, want %s", i, s, want[i])
		}
	}
}

func TestServerMessageJSON(t *testing.T) {
	tests := []struct {
		name string
		msg  ServerMessage
		want string
	}{
		{"view nil events", ServerMessage{Type: MsgView, View: &PlayerView{}}, `"events":[]`},
		{"view helper", ViewMessage(PlayerView{}, nil), `"events":[]`},
		{"view with events", ViewMessage(PlayerView{}, []Event{{Type: EventGameStarted}}), `"events":[{"type":"gameStarted"}]`},
		{"presence empty", PresenceMessage(nil), `{"type":"presence","online":[]}`},
		{"presence", PresenceMessage([]string{"p1"}), `{"type":"presence","online":["p1"]}`},
		{"ack", AckMessage("7"), `{"type":"ack","requestId":"7"}`},
		{"error no request", ErrorMessage("", ErrorBody{Code: "bad_request", Message: "x"}), `{"type":"error","error":{"code":"bad_request","message":"x"}}`},
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
	if err := json.Unmarshal([]byte(`{"type":"playCard","requestId":"4","card":42,"pile":0}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != CmdPlayCard || m.RequestID != "4" || m.Card == nil || *m.Card != 42 || m.Pile == nil || *m.Pile != 0 {
		t.Fatalf("decoded %+v", m)
	}
}

func TestErrorFor(t *testing.T) {
	var syntaxErr error
	{
		var v any
		syntaxErr = json.Unmarshal([]byte("{"), &v)
		if syntaxErr == nil {
			syntaxErr = json.Unmarshal([]byte("{x"), &v)
		}
	}
	var typeErr error
	{
		var v PlayCardRequest
		typeErr = json.Unmarshal([]byte(`{"card":"x"}`), &v)
	}
	tests := []struct {
		name       string
		err        error
		status     int
		code       string
		messageHas string
	}{
		{"bad request", BadRequest("missing %s", "card"), 400, "bad_request", "missing card"},
		{"json syntax", fmt.Errorf("decode: %w", syntaxErr), 400, "bad_request", ""},
		{"json type", typeErr, 400, "bad_request", ""},
		{"max bytes", &http.MaxBytesError{Limit: 4096}, 400, "bad_request", ""},
		{"unauthorized", ports.ErrUnauthorized, 401, "unauthorized", ""},
		{"not host", domain.ErrNotHost, 403, "not_host", ""},
		{"game not found", fmt.Errorf("load: %w", ports.ErrGameNotFound), 404, "game_not_found", ""},
		{"wrong phase", domain.ErrWrongPhase, 409, "wrong_phase", ""},
		{"game full", domain.ErrGameFull, 409, "game_full", ""},
		{"name taken", fmt.Errorf("%w: Anna", domain.ErrNameTaken), 409, "name_taken", "Anna"},
		{"not your turn", domain.ErrNotYourTurn, 409, "not_your_turn", ""},
		{"conflict", ports.ErrVersionConflict, 409, "conflict", ""},
		{"invalid name", domain.ErrInvalidName, 422, "invalid_name", ""},
		{"invalid options", fmt.Errorf("svc: %w", fmt.Errorf("%w: reduced", domain.ErrInvalidOptions)), 422, "invalid_options", "reduced"},
		{"unknown player", domain.ErrUnknownPlayer, 422, "unknown_player", ""},
		{"card not in hand", domain.ErrCardNotInHand, 422, "card_not_in_hand", ""},
		{"invalid pile", domain.ErrInvalidPile, 422, "invalid_pile", ""},
		{"illegal placement", fmt.Errorf("%w: 5 on pile 0", domain.ErrIllegalPlacement), 422, "illegal_placement", "5 on pile 0"},
		{"minimum", domain.ErrMinimumNotReached, 422, "minimum_not_reached", ""},
		{"rematch", domain.ErrRematchExists, 409, "rematch_exists", ""},
		{"backwards trick", domain.ErrBackwardsTrickBlocked, 422, "backwards_trick_blocked", ""},
		{"single pile", fmt.Errorf("%w: pile 1", domain.ErrSinglePileViolation), 422, "single_pile_violation", "pile 1"},
		{"exact exceeded", fmt.Errorf("%w: 4th card", domain.ErrExactCountExceeded), 422, "exact_count_exceeded", "4th card"},
		{"exact unreachable", fmt.Errorf("%w: hand 2", domain.ErrExactCountUnreachable), 422, "exact_count_unreachable", "hand 2"},
		{"exact required", fmt.Errorf("%w: 2 of 3", domain.ErrExactCountRequired), 422, "exact_count_required", "2 of 3"},
		{"stop too early", fmt.Errorf("%w: 2nd card", domain.ErrStopTooEarly), 422, "stop_too_early", "2nd card"},
		{"skull uncovered", fmt.Errorf("STOP would end the turn: %w", domain.ErrSkullUncovered), 422, "skull_uncovered", "STOP would end"},
		{"skull uncoverable", fmt.Errorf("svc: %w: x", domain.ErrSkullUncoverable), 422, "skull_uncoverable", "x"},
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

func TestCommandCardJSON(t *testing.T) {
	stop := FromCard(domain.Card{Value: 95, Kind: domain.CardKindNumber, Command: domain.CommandStop})
	if s := mustJSON(t, stop); s != `{"value":95,"kind":"number","command":"stop"}` {
		t.Errorf("command card JSON = %s", s)
	}
	if s := mustJSON(t, FromCard(domain.NumberCard(7))); s != `{"value":7,"kind":"number"}` {
		t.Errorf("plain card JSON = %s", s)
	}
	v := FromView(domain.PlayerView{Hand: []domain.HandCard{
		{Card: domain.NumberCard(7)},
		{Card: domain.NumberCard(8), Blocked: domain.ErrSkullUncoverable},
	}})
	if s := mustJSON(t, v.Hand[0]); strings.Contains(s, "blocked") {
		t.Errorf("unblocked hand card has blocked key: %s", s)
	}
	if s := mustJSON(t, v.Hand[1]); !strings.Contains(s, `"blocked":"skull_uncoverable"`) {
		t.Errorf("blocked hand card JSON = %s", s)
	}
}
