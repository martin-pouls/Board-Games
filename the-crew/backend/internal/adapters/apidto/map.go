package apidto

import (
	"slices"
	"time"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// FromCard maps a domain card.
func FromCard(c domain.Card) Card { return Card{Suit: string(c.Suit), Rank: int(c.Rank)} }

// ToDomain maps a wire card to a domain card; the result must be a valid card.
func (c Card) ToDomain() (domain.Card, error) {
	dc := domain.NewCard(domain.Suit(c.Suit), domain.Rank(c.Rank))
	if !dc.Valid() {
		return domain.Card{}, BadRequest("invalid card (suit %q, rank %d)", c.Suit, c.Rank)
	}
	return dc, nil
}

// fromCards maps cards; the result is never nil.
func fromCards(cs []domain.Card) []Card {
	out := make([]Card, len(cs))
	for i, c := range cs {
		out[i] = FromCard(c)
	}
	return out
}

func fromCardPtr(c *domain.Card) *Card {
	if c == nil {
		return nil
	}
	v := FromCard(*c)
	return &v
}

func ptr[T any](v T) *T { return &v }

func ints(s []int) []int { return append([]int{}, s...) }

// FromOptions maps domain options.
func FromOptions(o domain.Options) GameOptions { return GameOptions{Distress: o.Distress} }

// ToDomain maps wire options to domain options.
func (o GameOptions) ToDomain() domain.Options { return domain.Options{Distress: o.Distress} }

// FromView maps a player view. Slices are never nil, except Member.Cards (null until the debrief).
func FromView(v domain.PlayerView) PlayerView {
	out := PlayerView{
		GameID:   string(v.GameID),
		Revision: v.Revision,
		Phase:    string(v.Phase),
		You:      string(v.You),
		HostID:   string(v.HostID),
		Players:  make([]PlayerSummary, 0, len(v.Players)),
		Edition:  string(v.Edition),
		Options:  FromOptions(v.Options),
		Mission:  v.Mission,
		Logbook:  fromLogbook(v.Logbook),
	}
	for _, p := range v.Players {
		out.Players = append(out.Players, PlayerSummary{
			ID:     string(p.ID),
			Name:   p.Name,
			Seat:   p.Seat,
			IsHost: p.IsHost,
			IsYou:  p.IsYou,
		})
	}
	if m, ok := domain.MissionFor(v.Edition, v.Mission); ok {
		def := FromMission(m)
		out.MissionDef = &def
	}
	if a := v.Attempt; a != nil {
		out.Attempt = fromAttempt(a)
	}
	return out
}

// fromLogbook maps the logbook to a list sorted by mission number.
func fromLogbook(l map[int]domain.LogEntry) []LogEntry {
	out := make([]LogEntry, 0, len(l))
	for n, e := range l {
		out = append(out, LogEntry{Mission: n, Attempts: e.Attempts, Distress: e.Distress, Completed: e.Completed, Total: e.Total()})
	}
	slices.SortFunc(out, func(a, b LogEntry) int { return a.Mission - b.Mission })
	return out
}

func fromObjective(o domain.Objective) Objective {
	return Objective{
		Kind:         string(o.Kind),
		Owner:        o.Owner,
		Cards:        fromCards(o.Cards),
		Count:        o.Count,
		Tricks:       ints(o.Tricks),
		Suit:         string(o.Suit),
		Rank:         int(o.Rank),
		NotWithTrump: o.NotWithTrump,
		Ordered:      o.Ordered,
	}
}

func fromPlays(ps []domain.Play) []Play {
	out := make([]Play, len(ps))
	for i, p := range ps {
		out[i] = Play{Seat: p.Seat, Card: FromCard(p.Card)}
	}
	return out
}

func fromAttempt(a *domain.AttemptView) *Attempt {
	out := &Attempt{
		Number:      a.Number,
		Seat:        a.Seat,
		Commander:   a.Commander,
		Step:        string(a.Step),
		Pick:        string(a.Pick),
		Actors:      ints(a.Actors),
		Turn:        a.Turn,
		TrickNumber: a.TrickNumber,
		TotalTricks: a.TotalTricks,
		Members:     make([]Member, 0, len(a.Members)),
		Hand:        fromCards(a.Hand),
		Playable:    fromCards(a.Playable),
		Tasks:       make([]Task, 0, len(a.Tasks)),
		Objectives:  make([]ObjectiveView, 0, len(a.Objectives)),
		Silenced:    a.Silenced,
		Pickable:    ints(a.Pickable),
		Trick:       fromPlays(a.Trick),
		TrickLeader: a.TrickLeader,
		Distress:    a.Distress,
		Actions: Actions{
			Communicate:  make([]CommOption, 0, len(a.Actions.Communicate)),
			SendDistress: a.Actions.SendDistress,
			Handover:     a.Actions.Handover,
		},
	}
	for _, m := range a.Members {
		mv := Member{
			Seat:         m.Seat,
			PlayerID:     string(m.PlayerID),
			Name:         m.Name,
			Dummy:        m.Dummy,
			CardCount:    m.CardCount,
			TricksWon:    m.TricksWon,
			Communicated: m.Communicated,
			Stacks:       make([]Stack, 0, len(m.Stacks)),
			MustAct:      m.MustAct,
		}
		if c := m.Communication; c != nil {
			mv.Communication = &Communication{Card: FromCard(c.Card), Position: string(c.Position), Trick: c.Trick}
		}
		for _, s := range m.Stacks {
			mv.Stacks = append(mv.Stacks, Stack{Up: fromCardPtr(s.Up), Hidden: s.Hidden})
		}
		if m.Cards != nil {
			mv.Cards = fromCards(m.Cards)
		}
		out.Members = append(out.Members, mv)
	}
	for _, t := range a.Tasks {
		out.Tasks = append(out.Tasks, Task{
			ID:       t.ID,
			Card:     fromCardPtr(t.Card),
			Token:    string(t.Token),
			Owner:    t.Owner,
			Status:   string(t.Status),
			FaceDown: t.FaceDown,
		})
	}
	for _, o := range a.Objectives {
		out.Objectives = append(out.Objectives, ObjectiveView{Objective: fromObjective(o.Objective), Status: string(o.Status)})
	}
	if p := a.Poll; p != nil {
		answers := make([]*bool, len(p.Answers))
		for i, y := range p.Answers {
			if y != nil {
				answers[i] = ptr(*y)
			}
		}
		out.Poll = &Poll{Kind: string(p.Kind), Answers: answers, Objective: p.Objective, Task: p.Task}
	}
	if p := a.Pass; p != nil {
		out.Pass = &Pass{Direction: string(p.Direction), Chosen: fromCardPtr(p.Chosen), Waiting: ints(p.Waiting)}
	}
	if e := a.Exchange; e != nil {
		out.Exchange = &Exchange{Kind: string(e.Kind), Gave: fromCardPtr(e.Gave), Got: fromCardPtr(e.Got)}
	}
	if t := a.LastTrick; t != nil {
		out.LastTrick = &CompletedTrick{Leader: t.Leader, Plays: fromPlays(t.Plays), Winner: t.Winner}
	}
	for _, c := range a.Actions.Communicate {
		out.Actions.Communicate = append(out.Actions.Communicate, CommOption{Card: FromCard(c.Card), Position: string(c.Position)})
	}
	if r := a.Result; r != nil {
		out.Result = &AttemptResult{
			Success: r.Success, Reason: string(r.Reason), Task: r.Task, Objective: r.Objective, Trick: r.Trick, Seat: r.Seat,
		}
	}
	return out
}

// FromMission maps a mission definition.
func FromMission(m domain.MissionDef) MissionDef {
	out := MissionDef{
		Number:       m.Number,
		Tasks:        m.Tasks,
		Tokens:       make([]string, 0, len(m.Tokens)),
		Distribution: string(m.Distribution),
		DeadZone:     m.Comm.DeadZone,
		FromTrick:    m.Comm.FromTrick,
		Golden:       m.Golden,
		Setup:        make([]string, 0, len(m.Setup)),
		Poll:         string(m.Poll),
		Objectives:   make([]ObjectiveSpec, 0, len(m.Objectives)),
	}
	for _, t := range m.Tokens {
		out.Tokens = append(out.Tokens, string(t))
	}
	for _, s := range m.Setup {
		out.Setup = append(out.Setup, string(s))
	}
	for _, s := range m.Objectives {
		o := fromObjective(s.Objective)
		o.Owner = domain.OwnerUnassigned // the template's owner is ignored; the owner rule decides
		spec := ObjectiveSpec{Objective: o, OwnerRule: string(s.Owner), TaskToken: string(s.TaskToken)}
		if s.Owner == domain.OwnerRuleLeftOfHolder {
			spec.Holder = ptr(FromCard(s.Holder))
		}
		out.Objectives = append(out.Objectives, spec)
	}
	return out
}

// Catalogue returns the mission table of the edition, in mission order.
func Catalogue(e domain.Edition) MissionCatalogue {
	n := domain.Missions(e)
	out := MissionCatalogue{Edition: string(e), Missions: make([]MissionDef, 0, n)}
	for i := 1; i <= n; i++ {
		if m, ok := domain.MissionFor(e, i); ok {
			out.Missions = append(out.Missions, FromMission(m))
		}
	}
	return out
}

// FromSummary maps a game summary.
func FromSummary(s domain.GameSummary) GameSummary {
	var createdAt string
	if !s.CreatedAt.IsZero() {
		createdAt = s.CreatedAt.UTC().Format(time.RFC3339)
	}
	return GameSummary{
		GameID:      string(s.GameID),
		Phase:       string(s.Phase),
		HostName:    s.HostName,
		PlayerCount: s.PlayerCount,
		MaxPlayers:  s.MaxPlayers,
		Joinable:    s.Joinable,
		CreatedAt:   createdAt,
	}
}

// FromEvents maps domain events. The result is never nil. Unknown event types are skipped.
func FromEvents(evs []domain.Event) []Event {
	out := make([]Event, 0, len(evs))
	for _, e := range evs {
		if ev, ok := FromEvent(e); ok {
			out = append(out, ev)
		}
	}
	return out
}

// FromEvent maps one domain event; false for unknown types. Events are public and carry no
// hidden card: the pass and draw events name no cards.
func FromEvent(e domain.Event) (Event, bool) {
	switch e := e.(type) {
	case domain.PlayerJoined:
		return Event{Type: EventPlayerJoined, PlayerID: string(e.PlayerID), Name: e.Name}, true
	case domain.PlayerLeft:
		return Event{Type: EventPlayerLeft, PlayerID: string(e.PlayerID), HostID: string(e.HostID)}, true
	case domain.MissionSelected:
		return Event{Type: EventMissionSelected, Mission: ptr(e.Mission)}, true
	case domain.OptionsChanged:
		return Event{Type: EventOptionsChanged}, true
	case domain.AttemptStarted:
		return Event{Type: EventAttemptStarted, Mission: ptr(e.Mission), Number: ptr(e.Number), Commander: ptr(e.Commander)}, true
	case domain.TokensAdjusted:
		return Event{Type: EventTokensAdjusted, From: ptr(e.From), To: ptr(e.To)}, true
	case domain.PollAnswered:
		return Event{Type: EventPollAnswered, Seat: ptr(e.Seat), Yes: ptr(e.Yes)}, true
	case domain.Picked:
		return Event{Type: EventPicked, Seats: ints(e.Seats)}, true
	case domain.TaskPicked:
		return Event{Type: EventTaskPicked, TaskID: ptr(e.TaskID), Seat: ptr(e.Seat)}, true
	case domain.DistressSignal:
		return Event{Type: EventDistressSignal, Direction: string(e.Direction), Skipped: ptr(e.Skipped)}, true
	case domain.PassCardChosen:
		return Event{Type: EventPassCardChosen, Seat: ptr(e.Seat)}, true
	case domain.CardsPassed:
		return Event{Type: EventCardsPassed, Direction: string(e.Direction)}, true
	case domain.TaskHandedOver:
		return Event{Type: EventTaskHandedOver, TaskID: ptr(e.TaskID), From: ptr(e.From), To: ptr(e.To)}, true
	case domain.Communicated:
		return Event{Type: EventCommunicated, Seat: ptr(e.Seat), Card: fromCardPtr(&e.Card), Position: string(e.Position)}, true
	case domain.CardPlayed:
		return Event{Type: EventCardPlayed, Seat: ptr(e.Seat), Card: fromCardPtr(&e.Card)}, true
	case domain.TrickWon:
		return Event{Type: EventTrickWon, Number: ptr(e.Number), Seat: ptr(e.Seat)}, true
	case domain.CardsDrawn:
		return Event{Type: EventCardsDrawn}, true
	case domain.AttemptEnded:
		return Event{Type: EventAttemptEnded, Success: ptr(e.Success), Reason: string(e.Reason)}, true
	case domain.BackInLobby:
		return Event{Type: EventBackInLobby}, true
	case domain.GameEnded:
		return Event{Type: EventGameEnded, ByPlayerID: string(e.ByPlayerID)}, true
	default:
		return Event{}, false
	}
}
