package apidto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// rng is a deterministic RNG that leaves every deck in order.
type rng struct{}

func (rng) Shuffle(int, func(i, j int)) {}

func startedGame(t *testing.T) *domain.Game {
	t.Helper()
	g, err := domain.NewGame("K7MQ2", "p1", "Anna")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Ben"); err != nil {
		t.Fatal(err)
	}
	if err := g.Start("p1", 1, rng{}); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestFromViewEmptySlicesMarshalAsArrays(t *testing.T) {
	s := mustJSON(t, FromView(domain.PlayerView{}))
	for _, want := range []string{`"players":[]`, `"hand":[]`, `"playable":[]`, `"board":null`, `"decree":null`, `"trick":null`, `"lastTrick":null`, `"pending":null`, `"lastGift":null`, `"result":null`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if strings.Contains(s, "revealedHand") || strings.Contains(s, "nextGameId") {
		t.Errorf("optional keys present in %s", s)
	}
	s = mustJSON(t, FromView(domain.PlayerView{
		Board:   &domain.BoardView{},
		Pending: &domain.PendingView{Step: domain.StepLead},
	}))
	for _, want := range []string{`"spaces":[]`, `"playerIds":[]`, `"moveOptions":[]`, `"ends":[]`, `"yourGift":null`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in %s", want, s)
		}
	}
	if got := mustJSON(t, FromEvents(nil)); got != "[]" {
		t.Errorf("FromEvents(nil) = %s", got)
	}
}

func TestFromViewStartedGame(t *testing.T) {
	g := startedGame(t)
	for seat, p := range g.Players {
		v, err := g.ViewFor(p.ID)
		if err != nil {
			t.Fatal(err)
		}
		got := FromView(v)
		other := g.Players[1-seat]
		if len(got.Hand) != domain.HandSize || len(got.Playable) > len(got.Hand) {
			t.Fatalf("hand %d, playable %d", len(got.Hand), len(got.Playable))
		}
		for _, c := range got.Hand {
			dc := domain.NewCard(domain.Suit(c.Suit), domain.Rank(c.Rank))
			if c.Movement != dc.Movement() || c.Ability != string(dc.Ability()) {
				t.Errorf("card %+v: wrong movement/ability", c)
			}
		}
		// The hidden hand of the teammate never reaches the view.
		s := mustJSON(t, got)
		if strings.Contains(s, "revealedHand") {
			t.Errorf("hand revealed in a running game: %s", s)
		}
		for _, c := range other.Hand {
			if slices.Contains(got.Hand, FromCard(c)) {
				continue
			}
			if strings.Contains(s, mustJSON(t, FromCard(c))) {
				t.Errorf("teammate card %v leaked into %s", c, s)
			}
		}
		if got.Players[1-seat].HandCount != domain.HandSize || got.Players[seat].HandCount != domain.HandSize {
			t.Errorf("hand counts = %+v", got.Players)
		}
		if got.Board == nil || len(got.Board.Spaces) == 0 || got.Decree == nil || got.Pending == nil || got.Pending.Step != "lead" {
			t.Errorf("view = %+v", got)
		}
		if got.Players[0].End != "left" || got.Players[1].End != "right" {
			t.Errorf("ends = %q %q", got.Players[0].End, got.Players[1].End)
		}
		if len(got.Pending.PlayerIDs) != 1 || got.Pending.MoveOptions == nil || got.Pending.Ends == nil {
			t.Errorf("pending = %+v", got.Pending)
		}
	}
}

func TestFromViewMapsEveryField(t *testing.T) {
	d := domain.NewCard(domain.SuitDoves, 3)
	r := domain.NewCard(domain.SuitRoses, 10)
	s7 := domain.NewCard(domain.SuitStars, 7)
	in := domain.PlayerView{
		GameID: "K7MQ2", Revision: 9, Phase: domain.PhasePlaying, You: "p2", HostID: "p1", Level: 2,
		Players: []domain.PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, End: domain.EndLeft, HandCount: 5, IsHost: true, IsDealer: true, RevealedHand: []domain.Card{d}},
			{ID: "p2", Name: "Ben", Seat: 1, End: domain.EndRight, HandCount: 4, IsYou: true, MustAct: true},
		},
		Hand:     []domain.Card{d, r},
		Playable: []domain.Card{r},
		Board: &domain.BoardView{
			Side: domain.SideB, Start: 4, Traveller: 2, ForestTokens: 1, ForestTokensTotal: 3, GemsLeft: 5,
			Spaces: []domain.SpaceView{{Gems: 2, Plus: true, Covered: true}, {Gems: 0}},
		},
		Round:  2,
		Decree: &s7,
		Trick:  &domain.TrickView{Number: 4, LeaderID: "p1", Lead: &d, Follow: &r, WinnerID: "p2"},
		LastTrick: &domain.TrickRecordView{
			Number: 3, LeaderID: "p2", Lead: d, Follow: r, WinnerID: "p2", From: 4, To: 5, Delta: 1, Beyond: domain.EndRight, Collected: true,
		},
		Pending: &domain.PendingView{
			Step: domain.StepMove, PlayerIDs: []domain.PlayerID{"p2"}, ChosenID: "p1",
			MoveOptions: []domain.MoveOption{{Delta: -1, To: 3, Beyond: domain.EndNone, Gem: true}, {Delta: 5, To: 4, Beyond: domain.EndRight, Lost: true}},
			Ends:        []domain.End{domain.EndLeft, domain.EndRight},
			YourGift:    &r,
		},
		LastGift:   &domain.GiftView{Gave: d, Received: r},
		Result:     &domain.ResultView{Outcome: domain.PhaseWon, Reason: "", PlayerID: "p1", Round: 3, GemsLeft: 0, Score: &domain.Score{Level: 20, Cards: 4, EarlyWin: 10, ForestTokens: 3, Total: 37}},
		NextGameID: "NEXT1",
	}
	dw := Card{Suit: "doves", Rank: 3, Movement: 2, Ability: "foxes"}
	rw := Card{Suit: "roses", Rank: 10, Movement: 3}
	want := PlayerView{
		GameID: "K7MQ2", Revision: 9, Phase: "playing", You: "p2", HostID: "p1", Level: 2,
		Players: []PlayerSummary{
			{ID: "p1", Name: "Anna", Seat: 0, End: "left", HandCount: 5, IsHost: true, IsDealer: true, RevealedHand: []Card{dw}},
			{ID: "p2", Name: "Ben", Seat: 1, End: "right", HandCount: 4, IsYou: true, MustAct: true},
		},
		Hand:     []Card{dw, rw},
		Playable: []Card{rw},
		Board: &Board{
			Side: "B", Start: 4, Traveller: 2, ForestTokens: 1, ForestTokensTotal: 3, GemsLeft: 5,
			Spaces: []Space{{Gems: 2, Plus: true, Covered: true}, {}},
		},
		Round:  2,
		Decree: &Card{Suit: "stars", Rank: 7, Movement: 0, Ability: "gift"},
		Trick:  &Trick{Number: 4, LeaderID: "p1", Lead: &dw, Follow: &rw, WinnerID: "p2"},
		LastTrick: &LastTrick{
			Number: 3, LeaderID: "p2", Lead: dw, Follow: rw, WinnerID: "p2", From: 4, To: 5, Delta: 1, Beyond: "right", Collected: true,
		},
		Pending: &Pending{
			Step: "move", PlayerIDs: []string{"p2"}, ChosenID: "p1",
			MoveOptions: []MoveOption{{Delta: -1, To: 3, Beyond: "", Gem: true}, {Delta: 5, To: 4, Beyond: "right", Lost: true}},
			Ends:        []string{"left", "right"},
			YourGift:    &rw,
		},
		LastGift:   &Gift{Gave: dw, Received: rw},
		Result:     &Result{Outcome: "won", PlayerID: "p1", Round: 3, Score: &Score{Level: 20, Cards: 4, EarlyWin: 10, ForestTokens: 3, Total: 37}},
		NextGameID: "NEXT1",
	}
	got := FromView(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromView mismatch:\n got %+v\nwant %+v", got, want)
	}
	// Mapped slices must not alias the domain slices.
	in.Hand[0] = r
	if got.Hand[0].Suit != "doves" {
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
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseLobby, HostName: "Anna", PlayerCount: 1, MaxPlayers: 2, Joinable: true},
			GameSummary{GameID: "ABCDE", Phase: "lobby", HostName: "Anna", PlayerCount: 1, MaxPlayers: 2, Joinable: true},
		},
		{
			"created and won",
			domain.GameSummary{GameID: "ABCDE", Phase: domain.PhaseWon, PlayerCount: 2, MaxPlayers: 2, CreatedAt: created.In(time.FixedZone("x", 3600))},
			GameSummary{GameID: "ABCDE", Phase: "won", PlayerCount: 2, MaxPlayers: 2, CreatedAt: "2026-10-10T12:30:05Z"},
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

func TestCardMappingAndValidation(t *testing.T) {
	for _, c := range domain.NewDeck() {
		w := FromCard(c)
		if w.Movement != c.Movement() || w.Ability != string(c.Ability()) {
			t.Errorf("%v: %+v", c, w)
		}
		back, err := w.ToDomain()
		if err != nil || back != c {
			t.Errorf("round trip %v: %v %v", c, back, err)
		}
	}
	// Movement and ability are ignored on input.
	if c, err := (Card{Suit: "roses", Rank: 4, Movement: 3, Ability: "gift"}).ToDomain(); err != nil || c != domain.NewCard(domain.SuitRoses, 4) {
		t.Errorf("movement/ability not ignored: %v %v", c, err)
	}
	for _, bad := range []Card{{Suit: "", Rank: 3}, {Suit: "doves", Rank: 0}, {Suit: "doves", Rank: 11}, {Suit: "hearts", Rank: 5}, {Suit: "stars", Rank: -1}} {
		if _, err := bad.ToDomain(); !errors.Is(err, ErrBadRequest) {
			t.Errorf("%+v: err = %v, want bad request", bad, err)
		}
	}
}

func TestLevelAndEndMapping(t *testing.T) {
	if LevelOrDefault(0) != 1 || LevelOrDefault(3) != 3 || LevelOrDefault(9) != 9 {
		t.Error("LevelOrDefault")
	}
	if EndFromWire("left") != domain.EndLeft || EndFromWire("right") != domain.EndRight || EndFromWire("up") != "up" {
		t.Error("EndFromWire")
	}
}

func TestFromEvents(t *testing.T) {
	d3 := domain.NewCard(domain.SuitDoves, 3)
	s7 := domain.NewCard(domain.SuitStars, 7)
	evs := []domain.Event{
		domain.PlayerJoined{PlayerID: "p2", Name: "Ben"},
		domain.PlayerLeft{PlayerID: "p2", HostID: "p1"},
		domain.GameStarted{Level: 2},
		domain.RoundStarted{Round: 1, DealerID: "p1", Decree: s7},
		domain.CardPlayed{PlayerID: "p1", Card: d3, Lead: true},
		domain.FoxesChosen{ByPlayerID: "p1", PlayerID: "p2", Skipped: false},
		domain.DecreeSwapped{PlayerID: "p2", Old: s7, New: d3},
		domain.DecreeKept{PlayerID: "p2"},
		domain.GiftChosen{PlayerID: "p1", Auto: true},
		domain.GiftExchanged{},
		domain.TrickWon{PlayerID: "p2", Number: 4},
		domain.TravellerMoved{From: 5, To: 3, Delta: -2},
		domain.TravellerMoved{From: 1, To: 5, Delta: -3, Beyond: domain.EndLeft},
		domain.ForestTokenPlaced{End: domain.EndRight, Space: 10, GemsMoved: 1, RoundEnd: false, Left: 0},
		domain.GemCollected{Space: 3, Left: 0},
		domain.GemsAdded{Count: 4},
		domain.GameWon{Score: 37},
		domain.GameLost{Reason: domain.LossGaveUp, PlayerID: "p2"},
		domain.GameEnded{ByPlayerID: "p1"},
		domain.RematchCreated{GameID: "NEXT1"},
	}
	card3 := `{"suit":"doves","rank":3,"movement":2,"ability":"foxes"}`
	card7 := `{"suit":"stars","rank":7,"movement":0,"ability":"gift"}`
	want := []string{
		`{"type":"playerJoined","playerId":"p2","name":"Ben"}`,
		`{"type":"playerLeft","playerId":"p2","hostId":"p1"}`,
		`{"type":"gameStarted","level":2}`,
		`{"type":"roundStarted","dealerId":"p1","round":1,"card":` + card7 + `}`,
		`{"type":"cardPlayed","playerId":"p1","card":` + card3 + `,"lead":true}`,
		`{"type":"foxesChosen","playerId":"p2","byPlayerId":"p1","skipped":false}`,
		`{"type":"decreeSwapped","playerId":"p2","old":` + card7 + `,"new":` + card3 + `}`,
		`{"type":"decreeKept","playerId":"p2"}`,
		`{"type":"giftChosen","playerId":"p1","auto":true}`,
		`{"type":"giftExchanged"}`,
		`{"type":"trickWon","playerId":"p2","number":4}`,
		`{"type":"travellerMoved","from":5,"to":3,"delta":-2}`,
		`{"type":"travellerMoved","end":"left","from":1,"to":5,"delta":-3}`,
		`{"type":"forestTokenPlaced","end":"right","space":10,"gemsMoved":1,"roundEnd":false,"left":0}`,
		`{"type":"gemCollected","space":3,"left":0}`,
		`{"type":"gemsAdded","count":4}`,
		`{"type":"gameWon","score":37}`,
		`{"type":"gameLost","reason":"gaveUp","playerId":"p2"}`,
		`{"type":"gameEnded","byPlayerId":"p1"}`,
		`{"type":"rematchCreated","gameId":"NEXT1"}`,
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
	if err := json.Unmarshal([]byte(`{"type":"playCard","requestId":"4","card":{"suit":"doves","rank":5}}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Type != CmdPlayCard || m.RequestID != "4" || m.Card == nil || *m.Card != (Card{Suit: "doves", Rank: 5}) {
		t.Fatalf("decoded %+v", m)
	}
	m = ClientMessage{}
	if err := json.Unmarshal([]byte(`{"type":"chooseMove","requestId":"5","delta":0}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Delta == nil || *m.Delta != 0 {
		t.Fatalf("a delta of 0 must decode to a non-nil pointer: %v", m.Delta)
	}
	m = ClientMessage{}
	if err := json.Unmarshal([]byte(`{"type":"keepDecree","requestId":"6"}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Delta != nil || m.Card != nil {
		t.Fatalf("absent fields must stay nil: %+v", m)
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
		var v StartGameRequest
		typeErr = json.Unmarshal([]byte(`{"level":"x"}`), &v)
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
		{"rate limited", ErrRateLimited, 429, "rate_limited", ""},
		{"unauthorized", ports.ErrUnauthorized, 401, "unauthorized", ""},
		{"not host", domain.ErrNotHost, 403, "not_host", ""},
		{"game not found", fmt.Errorf("load: %w", ports.ErrGameNotFound), 404, "game_not_found", ""},
		{"wrong phase", domain.ErrWrongPhase, 409, "wrong_phase", ""},
		{"wrong step", fmt.Errorf("%w: gift", domain.ErrWrongStep), 409, "wrong_step", "gift"},
		{"game full", domain.ErrGameFull, 409, "game_full", ""},
		{"rematch", domain.ErrRematchExists, 409, "rematch_exists", ""},
		{"name taken", fmt.Errorf("%w: Anna", domain.ErrNameTaken), 409, "name_taken", "Anna"},
		{"not your turn", domain.ErrNotYourTurn, 409, "not_your_turn", ""},
		{"cannot leave", domain.ErrCannotLeave, 409, "cannot_leave", ""},
		{"conflict", ports.ErrVersionConflict, 409, "conflict", ""},
		{"invalid name", domain.ErrInvalidName, 422, "invalid_name", ""},
		{"unknown player", domain.ErrUnknownPlayer, 422, "unknown_player", ""},
		{"not enough players", domain.ErrNotEnoughPlayers, 409, "not_enough_players", ""},
		{"gift already chosen", domain.ErrGiftAlreadyChosen, 409, "gift_already_chosen", ""},
		{"invalid level", domain.ErrInvalidLevel, 422, "invalid_level", ""},
		{"card not in hand", domain.ErrCardNotInHand, 422, "card_not_in_hand", ""},
		{"must follow suit", fmt.Errorf("%w: doves", domain.ErrMustFollowSuit), 422, "must_follow_suit", "doves"},
		{"illegal move", fmt.Errorf("svc: %w: 7", domain.ErrIllegalMove), 422, "illegal_move", "7"},
		{"illegal end", domain.ErrIllegalEnd, 422, "illegal_end", ""},
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
