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

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
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
	for _, want := range []string{`"players":[]`, `"hand":[]`, `"enemy":null`, `"turn":null`, `"result":null`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if strings.Contains(s, "revealedHand") || strings.Contains(s, "nextGameId") {
		t.Errorf("optional keys present in %s", s)
	}
	s = mustJSON(t, FromView(domain.PlayerView{
		Enemy: &domain.EnemyView{Plays: []domain.PlayedSetView{{By: "p1", Cards: nil, BlockedSuits: nil}}},
	}))
	for _, want := range []string{`"plays":[{"playerId":"p1","cards":[],"value":0,"damage":0,"doubled":false,"blockedSuits":[],"jester":false}]`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	s = mustJSON(t, FromView(domain.PlayerView{Enemy: &domain.EnemyView{}}))
	if !strings.Contains(s, `"plays":[]`) {
		t.Errorf("enemy plays not an array: %s", s)
	}
	if got := mustJSON(t, FromEvents(nil)); got != "[]" {
		t.Errorf("FromEvents(nil) = %s", got)
	}
}

func TestFromViewMapsEveryField(t *testing.T) {
	jack := domain.NewCard(domain.SuitSpades, domain.RankJack)
	c8 := domain.NewCard(domain.SuitClubs, 8)
	in := domain.PlayerView{
		GameID:   "K7MQ2",
		Revision: 9,
		Phase:    domain.PhasePlaying,
		You:      "p2",
		HostID:   "p1",
		Rules:    domain.RulesView{MinPlayers: 1, MaxPlayers: 4, MaxHandSize: 7, TavernJesters: 0, SoloJesters: 2},
		Players: []domain.PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, HandCount: 5, IsHost: true, IsActive: true, YieldedLastTurn: true, Signal: domain.SignalNone},
			{ID: "p2", Name: "Ben", Seat: 1, HandCount: 4, IsYou: true, Signal: domain.SignalGoodPlay},
		},
		Hand:         []domain.Card{domain.NewCard(domain.SuitHearts, 3), domain.Jester()},
		Enemy:        &domain.EnemyView{Card: jack, Attack: 10, Health: 20, Damage: 16, Shield: 4, EffectiveAttack: 6, Immunity: domain.SuitSpades, ImmunityCancelled: true, Plays: []domain.PlayedSetView{{By: "p1", Cards: []domain.Card{c8}, Value: 8, Damage: 16, Doubled: true, BlockedSuits: []domain.Suit{domain.SuitSpades}, Jester: false}}},
		CastleCount:  8,
		TavernCount:  20,
		DiscardCount: 3,
		Turn:         &domain.TurnView{Number: 4, ActivePlayerID: "p1", Step: domain.StepDiscard, CanYield: false, DamageToCover: 6},
		NextGameID:   "NEXT1",
		Result:       &domain.ResultView{Outcome: domain.PhaseLost, Reason: domain.LossCannotCover, PlayerID: "p1", Grade: domain.GradeNone, EnemiesDefeated: 3, SoloJestersUsed: 1},

		SoloJestersLeft: 1,
	}
	in.Players[0].RevealedHand = []domain.Card{c8}
	want := PlayerView{
		GameID: "K7MQ2", Revision: 9, Phase: "playing", You: "p2", HostID: "p1",
		Rules: Rules{MinPlayers: 1, MaxPlayers: 4, MaxHandSize: 7, SoloJesters: 2},
		Players: []PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, HandCount: 5, IsHost: true, IsActive: true, YieldedLastTurn: true, Signal: "none", RevealedHand: []Card{{Suit: "clubs", Rank: 8}}},
			{ID: "p2", Name: "Ben", Seat: 1, HandCount: 4, IsYou: true, Signal: "goodPlay"},
		},
		Hand: []Card{{Suit: "hearts", Rank: 3}, {Suit: "", Rank: 0}},
		Enemy: &Enemy{
			Card: Card{Suit: "spades", Rank: 11}, Attack: 10, Health: 20, Damage: 16, Shield: 4, EffectiveAttack: 6,
			Immunity: "spades", ImmunityCancelled: true,
			Plays: []PlayedSet{{PlayerID: "p1", Cards: []Card{{Suit: "clubs", Rank: 8}}, Value: 8, Damage: 16, Doubled: true, BlockedSuits: []string{"spades"}}},
		},
		CastleCount: 8, TavernCount: 20, DiscardCount: 3,
		Turn:            &Turn{Number: 4, ActivePlayerID: "p1", Step: "discard", DamageToCover: 6},
		SoloJestersLeft: 1,
		Result:          &Result{Outcome: "lost", Reason: "cannotCoverDamage", PlayerID: "p1", EnemiesDefeated: 3, SoloJestersUsed: 1},
		NextGameID:      "NEXT1",
	}
	got := FromView(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromView mismatch:\n got %+v\nwant %+v", got, want)
	}
	// The Jester is {suit:"", rank:0} on the wire.
	if s := mustJSON(t, got.Hand[1]); s != `{"suit":"","rank":0}` {
		t.Errorf("jester JSON = %s", s)
	}
	// Mapped slices must not alias the domain slices.
	in.Hand[0] = domain.NewCard(domain.SuitClubs, 9)
	if got.Hand[0].Suit != "hearts" {
		t.Error("Hand aliases domain slice")
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
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseLobby, HostName: "Anna", PlayerCount: 2, MaxPlayers: 4, Joinable: true},
			GameSummary{GameID: "ABCDE", Phase: "lobby", HostName: "Anna", PlayerCount: 2, MaxPlayers: 4, Joinable: true},
		},
		{
			"created and won",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseWon, PlayerCount: 1, MaxPlayers: 4, CreatedAt: created.In(time.FixedZone("x", 3600))},
			GameSummary{GameID: "ABCDE", Phase: "won", PlayerCount: 1, MaxPlayers: 4, CreatedAt: "2026-10-10T12:30:05Z"},
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

func TestCardRoundTripAndValidation(t *testing.T) {
	for _, c := range []domain.Card{domain.Jester(), domain.NewCard(domain.SuitHearts, 1), domain.NewCard(domain.SuitSpades, 13), domain.NewCard(domain.SuitClubs, 10)} {
		back, err := FromCard(c).ToDomain()
		if err != nil || back != c {
			t.Errorf("round trip %v: %v %v", c, back, err)
		}
	}
	for _, bad := range []Card{{Suit: "", Rank: 3}, {Suit: "hearts", Rank: 0}, {Suit: "hearts", Rank: 14}, {Suit: "stars", Rank: 5}, {Suit: "clubs", Rank: -1}} {
		if _, err := bad.ToDomain(); !errors.Is(err, ErrBadRequest) {
			t.Errorf("%+v: err = %v, want bad request", bad, err)
		}
	}
	if cs, err := CardsToDomain(nil); cs != nil || err != nil {
		t.Errorf("nil: %v %v", cs, err)
	}
	if cs, err := CardsToDomain([]Card{}); cs == nil || len(cs) != 0 || err != nil {
		t.Errorf("empty: %#v %v", cs, err)
	}
	if _, err := CardsToDomain([]Card{{Suit: "hearts", Rank: 2}, {Suit: "x", Rank: 2}}); !errors.Is(err, ErrBadRequest) {
		t.Errorf("mixed: %v", err)
	}
}

func TestSignalMapping(t *testing.T) {
	for in, wire := range map[domain.JesterSignal]string{domain.SignalNone: "none", domain.SignalGoodPlay: "goodPlay", domain.SignalRatherNot: "ratherNot"} {
		if got := SignalToWire(in); got != wire {
			t.Errorf("SignalToWire(%q) = %q", in, got)
		}
		if got := SignalFromWire(wire); got != in {
			t.Errorf("SignalFromWire(%q) = %q", wire, got)
		}
	}
	if got := SignalFromWire("bogus"); got != "bogus" {
		t.Errorf("unknown signal not passed through: %q", got)
	}
}

func TestFromEvents(t *testing.T) {
	k := domain.NewCard(domain.SuitClubs, 13)
	c8 := domain.NewCard(domain.SuitClubs, 8)
	evs := []domain.Event{
		domain.PlayerJoined{PlayerID: "p2", Name: "Ben"},
		domain.PlayerLeft{PlayerID: "p2", HostID: "p1"},
		domain.GameStarted{FirstPlayerID: "p1"},
		domain.EnemyRevealed{Enemy: k},
		domain.TurnStarted{PlayerID: "p1", Number: 1},
		domain.CardsPlayed{PlayerID: "p1", Cards: []domain.Card{c8}, Value: 8, Healed: 0, Drawn: 2, Damage: 16, BlockedSuits: []domain.Suit{domain.SuitSpades}},
		domain.JesterPlayed{PlayerID: "p1"},
		domain.Yielded{PlayerID: "p2"},
		domain.EnemyDefeated{Enemy: k, Exact: true},
		domain.CardsDiscarded{PlayerID: "p1", Cards: []domain.Card{c8}, Attack: 8},
		domain.AttackBlocked{PlayerID: "p1"},
		domain.NextPlayerChosen{ByPlayerID: "p1", PlayerID: "p2"},
		domain.JesterSignalSet{PlayerID: "p2", Signal: domain.SignalNone},
		domain.JesterSignalSet{PlayerID: "p2", Signal: domain.SignalRatherNot},
		domain.SoloJesterUsed{PlayerID: "p1", Drawn: 8, Left: 0},
		domain.GameWon{Grade: domain.GradeGold},
		domain.GameLost{Reason: domain.LossCannotAct, PlayerID: "p2"},
		domain.GameEnded{ByPlayerID: "p1"},
		domain.RematchCreated{GameID: "NEXT1"},
	}
	want := []string{
		`{"type":"playerJoined","playerId":"p2","name":"Ben"}`,
		`{"type":"playerLeft","playerId":"p2","hostId":"p1"}`,
		`{"type":"gameStarted","playerId":"p1"}`,
		`{"type":"enemyRevealed","card":{"suit":"clubs","rank":13}}`,
		`{"type":"turnStarted","playerId":"p1","number":1}`,
		`{"type":"cardsPlayed","playerId":"p1","cards":[{"suit":"clubs","rank":8}],"value":8,"healed":0,"drawn":2,"damage":16,"blockedSuits":["spades"]}`,
		`{"type":"jesterPlayed","playerId":"p1"}`,
		`{"type":"yielded","playerId":"p2"}`,
		`{"type":"enemyDefeated","card":{"suit":"clubs","rank":13},"exact":true}`,
		`{"type":"cardsDiscarded","playerId":"p1","cards":[{"suit":"clubs","rank":8}],"attack":8}`,
		`{"type":"attackBlocked","playerId":"p1"}`,
		`{"type":"nextPlayerChosen","playerId":"p2","byPlayerId":"p1"}`,
		`{"type":"signalSet","playerId":"p2","signal":"none"}`,
		`{"type":"signalSet","playerId":"p2","signal":"ratherNot"}`,
		`{"type":"soloJesterUsed","playerId":"p1","drawn":8,"left":0}`,
		`{"type":"gameWon","grade":"gold"}`,
		`{"type":"gameLost","playerId":"p2","reason":"cannotAct"}`,
		`{"type":"gameEnded","byPlayerId":"p1"}`,
		`{"type":"rematchCreated","gameId":"NEXT1"}`,
	}
	got := FromEvents(evs)
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
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
	if err := json.Unmarshal([]byte(`{"type":"play","requestId":"4","cards":[{"suit":"hearts","rank":5},{"suit":"","rank":0}]}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != CmdPlay || m.RequestID != "4" || len(m.Cards) != 2 || m.Cards[0] != (Card{Suit: "hearts", Rank: 5}) {
		t.Fatalf("decoded %+v", m)
	}
	m = ClientMessage{}
	if err := json.Unmarshal([]byte(`{"type":"discard","requestId":"5","cards":[]}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Cards == nil || len(m.Cards) != 0 {
		t.Fatalf("empty cards must decode to a non-nil slice: %#v", m.Cards)
	}
	m = ClientMessage{}
	if err := json.Unmarshal([]byte(`{"type":"setSignal","requestId":"6","signal":"none"}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Signal == nil || *m.Signal != "none" {
		t.Fatalf("signal = %v", m.Signal)
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
		var v CardsRequest
		typeErr = json.Unmarshal([]byte(`{"cards":"x"}`), &v)
	}
	tests := []struct {
		name       string
		err        error
		status     int
		code       string
		messageHas string
	}{
		{"bad request", BadRequest("missing %s", "cards"), 400, "bad_request", "missing cards"},
		{"json syntax", fmt.Errorf("decode: %w", syntaxErr), 400, "bad_request", ""},
		{"json type", typeErr, 400, "bad_request", ""},
		{"max bytes", &http.MaxBytesError{Limit: 4096}, 400, "bad_request", ""},
		{"rate limited", ErrRateLimited, 429, "rate_limited", ""},
		{"unauthorized", ports.ErrUnauthorized, 401, "unauthorized", ""},
		{"not host", domain.ErrNotHost, 403, "not_host", ""},
		{"game not found", fmt.Errorf("load: %w", ports.ErrGameNotFound), 404, "game_not_found", ""},
		{"wrong phase", domain.ErrWrongPhase, 409, "wrong_phase", ""},
		{"wrong step", fmt.Errorf("%w: discard", domain.ErrWrongStep), 409, "wrong_step", "discard"},
		{"game full", domain.ErrGameFull, 409, "game_full", ""},
		{"rematch", domain.ErrRematchExists, 409, "rematch_exists", ""},
		{"name taken", fmt.Errorf("%w: Anna", domain.ErrNameTaken), 409, "name_taken", "Anna"},
		{"not your turn", domain.ErrNotYourTurn, 409, "not_your_turn", ""},
		{"cannot leave", domain.ErrCannotLeave, 409, "cannot_leave", ""},
		{"conflict", ports.ErrVersionConflict, 409, "conflict", ""},
		{"invalid name", domain.ErrInvalidName, 422, "invalid_name", ""},
		{"unknown player", domain.ErrUnknownPlayer, 422, "unknown_player", ""},
		{"card not in hand", domain.ErrCardNotInHand, 422, "card_not_in_hand", ""},
		{"illegal play", fmt.Errorf("%w: mixedRanks", domain.ErrIllegalPlay), 422, "illegal_play", "mixedRanks"},
		{"illegal discard", fmt.Errorf("svc: %w: notEnough", domain.ErrIllegalDiscard), 422, "illegal_discard", "notEnough"},
		{"yield not allowed", domain.ErrYieldNotAllowed, 422, "yield_not_allowed", ""},
		{"not solo", domain.ErrNotSolo, 422, "not_solo", ""},
		{"no solo jester", domain.ErrNoSoloJester, 422, "no_solo_jester", ""},
		{"illegal signal", domain.ErrIllegalSignal, 422, "illegal_signal", ""},
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
