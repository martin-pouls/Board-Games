package sqlite

import (
	"fmt"
	"time"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// schemaVersion is the version of gameDTO. Bump it (and migrate in fromDTO) on breaking changes.
const schemaVersion = 1

// The snapshot is explicit: no struct tags on domain types. Nil and empty slices (and maps) are
// kept apart, so a loaded game is reflect.DeepEqual to the saved one. Missions are not stored;
// they are looked up by edition and number.

type cardDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type playerDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type logEntryDTO struct {
	Attempts  int  `json:"attempts"`
	Distress  bool `json:"distress"`
	Completed bool `json:"completed"`
}

// stackDTO is one of the dummy's stacks; a nil card is a played or turned-up position.
type stackDTO struct {
	Down *cardDTO `json:"down"`
	Up   *cardDTO `json:"up"`
}

type memberDTO struct {
	PlayerID string     `json:"playerId"`
	Dummy    bool       `json:"dummy"`
	Hand     []cardDTO  `json:"hand"`
	Stacks   []stackDTO `json:"stacks"`
}

type objectiveDTO struct {
	Kind         string    `json:"kind"`
	Owner        int       `json:"owner"`
	Cards        []cardDTO `json:"cards"`
	Count        int       `json:"count"`
	Tricks       []int     `json:"tricks"`
	Suit         string    `json:"suit"`
	Rank         int       `json:"rank"`
	NotWithTrump bool      `json:"notWithTrump"`
	Ordered      bool      `json:"ordered"`

	// Deep Sea. Missing in older snapshots: zero values.
	Cmp        string         `json:"cmp,omitempty"`
	Suits      []string       `json:"suits"`
	Ranks      []int          `json:"ranks"`
	With       *cardMatchDTO  `json:"with,omitempty"`
	Target     *cardMatchDTO  `json:"target,omitempty"`
	Where      string         `json:"where,omitempty"`
	Max        int            `json:"max,omitempty"`
	Exact      bool           `json:"exact,omitempty"`
	InOneTrick bool           `json:"inOneTrick,omitempty"`
	Hidden     bool           `json:"hidden,omitempty"`
	Parts      []objectiveDTO `json:"parts"`
}

type cardMatchDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type taskDTO struct {
	ID        int          `json:"id"`
	Objective objectiveDTO `json:"objective"`
	Token     string       `json:"token"`
	FaceDown  bool         `json:"faceDown"`

	CardID     int `json:"cardId,omitempty"` // Deep Sea
	Difficulty int `json:"difficulty,omitempty"`
}

type pollDTO struct {
	Answers []*bool `json:"answers"`
}

type passDTO struct {
	Direction string     `json:"direction"`
	Chosen    []*cardDTO `json:"chosen"`
}

type exchangeDTO struct {
	Kind string     `json:"kind"`
	Gave []*cardDTO `json:"gave"`
	Got  []*cardDTO `json:"got"`
}

type commDTO struct {
	Card     cardDTO `json:"card"`
	Position string  `json:"position"`
	Trick    int     `json:"trick"`
}

type playDTO struct {
	Seat int     `json:"seat"`
	Card cardDTO `json:"card"`
}

type trickDTO struct {
	Leader int       `json:"leader"`
	Plays  []playDTO `json:"plays"`
}

type trickRecordDTO struct {
	Leader int       `json:"leader"`
	Plays  []playDTO `json:"plays"`
	Winner int       `json:"winner"`
}

type resultDTO struct {
	Success   bool   `json:"success"`
	Reason    string `json:"reason"`
	Task      int    `json:"task"`
	Objective int    `json:"objective"`
	Trick     int    `json:"trick"`
	Seat      int    `json:"seat"`
}

type attemptDTO struct {
	Number          int              `json:"number"`
	Members         []memberDTO      `json:"members"`
	Commander       int              `json:"commander"`
	Step            string           `json:"step"`
	Tasks           []taskDTO        `json:"tasks"`
	Objectives      []objectiveDTO   `json:"objectives"`
	Poll            *pollDTO         `json:"poll"`
	Silenced        int              `json:"silenced"`
	Comms           []*commDTO       `json:"comms"`               // per seat: the first communication
	MoreComms       [][]commDTO      `json:"moreComms,omitempty"` // per seat: the further ones (rapture of the deep)
	DistressDecided bool             `json:"distressDecided"`
	Pass            *passDTO         `json:"pass"`
	Exchange        *exchangeDTO     `json:"exchange"`
	HandoverUsed    bool             `json:"handoverUsed"`
	TokensAdjusted  bool             `json:"tokensAdjusted"`
	Trick           trickDTO         `json:"trick"`
	Tricks          []trickRecordDTO `json:"tricks"`
	Result          *resultDTO       `json:"result"`

	// Deep Sea. Missing in older snapshots: zero values.
	Comm     string   `json:"comm,omitempty"`
	Terrain  *cardDTO `json:"terrain,omitempty"`
	TaskDeck []int    `json:"taskDeck"`
	Passes   int      `json:"passes,omitempty"`
	Chosen   []int    `json:"chosen"`
	Redeals  int      `json:"redeals,omitempty"`
}

// optionsDTO holds the lobby options. A snapshot without it loads with every option off.
type optionsDTO struct {
	Distress bool `json:"distress"`
}

// gameDTO is the persisted JSON snapshot of a domain.Game (pending events excluded).
type gameDTO struct {
	SchemaVersion int                 `json:"schemaVersion"`
	ID            string              `json:"id"`
	Phase         string              `json:"phase"`
	HostID        string              `json:"hostId"`
	Players       []playerDTO         `json:"players"`
	Edition       string              `json:"edition"`
	Options       optionsDTO          `json:"options"`
	Mission       int                 `json:"mission"`
	Logbook       map[int]logEntryDTO `json:"logbook"`
	Attempt       *attemptDTO         `json:"attempt"`
	UsedTasks     []int               `json:"usedTasks"` // Deep Sea
	Revision      int64               `json:"revision"`
	CreatedAt     time.Time           `json:"createdAt"`
}

// mapSlice maps in to a new slice, keeping nil as nil and empty as empty.
func mapSlice[T, U any](in []T, f func(T) U) []U {
	if in == nil {
		return nil
	}
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

// mapPtr maps what p points to; a nil pointer stays nil.
func mapPtr[T, U any](p *T, f func(T) U) *U {
	if p == nil {
		return nil
	}
	v := f(*p)
	return &v
}

func cardToDTO(c domain.Card) cardDTO { return cardDTO{Suit: string(c.Suit), Rank: int(c.Rank)} }
func cardFromDTO(c cardDTO) domain.Card {
	return domain.Card{Suit: domain.Suit(c.Suit), Rank: domain.Rank(c.Rank)}
}

func copyBool(b *bool) *bool { return mapPtr(b, func(v bool) bool { return v }) }

func toDTO(g *domain.Game) gameDTO {
	d := gameDTO{
		SchemaVersion: schemaVersion,
		ID:            string(g.ID),
		Phase:         string(g.Phase),
		HostID:        string(g.HostID),
		Players:       mapSlice(g.Players, func(p domain.Player) playerDTO { return playerDTO{ID: string(p.ID), Name: p.Name} }),
		Edition:       string(g.Edition),
		Options:       optionsDTO{Distress: g.Options.Distress},
		Mission:       g.Mission,
		Attempt:       attemptToDTO(g.Attempt),
		UsedTasks:     mapSlice(g.UsedTasks, intCopy),
		Revision:      g.Revision,
		CreatedAt:     g.CreatedAt,
	}
	if g.Logbook != nil {
		d.Logbook = make(map[int]logEntryDTO, len(g.Logbook))
		for n, e := range g.Logbook {
			d.Logbook[n] = logEntryDTO{Attempts: e.Attempts, Distress: e.Distress, Completed: e.Completed}
		}
	}
	return d
}

func attemptToDTO(a *domain.Attempt) *attemptDTO {
	if a == nil {
		return nil
	}
	d := &attemptDTO{
		Number:          a.Number,
		Commander:       a.Commander,
		Step:            string(a.Step),
		Silenced:        a.Silenced,
		DistressDecided: a.DistressDecided,
		HandoverUsed:    a.HandoverUsed,
		TokensAdjusted:  a.TokensAdjusted,
		Trick:           trickDTO{Leader: a.Trick.Leader, Plays: mapSlice(a.Trick.Plays, playToDTO)},
		Members:         mapSlice(a.Members, memberToDTO),
		Tasks:           mapSlice(a.Tasks, taskToDTO),
		Objectives:      mapSlice(a.Objectives, objectiveToDTO),
		Comms: mapSlice(a.Comms, func(cs []domain.Communication) *commDTO {
			if len(cs) == 0 {
				return nil
			}
			c := commToDTO(cs[0])
			return &c
		}),
		MoreComms: moreCommsToDTO(a.Comms),
		Tricks: mapSlice(a.Tricks, func(t domain.TrickRecord) trickRecordDTO {
			return trickRecordDTO{Leader: t.Leader, Plays: mapSlice(t.Plays, playToDTO), Winner: t.Winner}
		}),
		Poll: mapPtr(a.Poll, func(p domain.Poll) pollDTO { return pollDTO{Answers: mapSlice(p.Answers, copyBool)} }),
		Pass: mapPtr(a.Pass, func(p domain.Pass) passDTO {
			return passDTO{Direction: string(p.Direction), Chosen: mapSlice(p.Chosen, cardPtrToDTO)}
		}),
		Exchange: mapPtr(a.Exchange, func(e domain.Exchange) exchangeDTO {
			return exchangeDTO{Kind: string(e.Kind), Gave: mapSlice(e.Gave, cardPtrToDTO), Got: mapSlice(e.Got, cardPtrToDTO)}
		}),
		Result: mapPtr(a.Result, func(r domain.AttemptResult) resultDTO {
			return resultDTO{Success: r.Success, Reason: string(r.Reason), Task: r.Task, Objective: r.Objective, Trick: r.Trick, Seat: r.Seat}
		}),
		Comm:     string(a.Comm),
		Terrain:  cardPtrToDTO(a.Terrain),
		TaskDeck: mapSlice(a.TaskDeck, intCopy),
		Passes:   a.Passes,
		Chosen:   mapSlice(a.Chosen, intCopy),
		Redeals:  a.Redeals,
	}
	return d
}

func intCopy(n int) int { return n }

func cardPtrToDTO(c *domain.Card) *cardDTO   { return mapPtr(c, cardToDTO) }
func cardPtrFromDTO(c *cardDTO) *domain.Card { return mapPtr(c, cardFromDTO) }

func playToDTO(p domain.Play) playDTO   { return playDTO{Seat: p.Seat, Card: cardToDTO(p.Card)} }
func playFromDTO(p playDTO) domain.Play { return domain.Play{Seat: p.Seat, Card: cardFromDTO(p.Card)} }

func memberToDTO(m domain.Member) memberDTO {
	return memberDTO{
		PlayerID: string(m.PlayerID),
		Dummy:    m.Dummy,
		Hand:     mapSlice(m.Hand, cardToDTO),
		Stacks: mapSlice(m.Stacks, func(s domain.DummyStack) stackDTO {
			return stackDTO{Down: cardPtrToDTO(s.Down), Up: cardPtrToDTO(s.Up)}
		}),
	}
}

func memberFromDTO(d memberDTO) domain.Member {
	return domain.Member{
		PlayerID: domain.PlayerID(d.PlayerID),
		Dummy:    d.Dummy,
		Hand:     mapSlice(d.Hand, cardFromDTO),
		Stacks: mapSlice(d.Stacks, func(s stackDTO) domain.DummyStack {
			return domain.DummyStack{Down: cardPtrFromDTO(s.Down), Up: cardPtrFromDTO(s.Up)}
		}),
	}
}

func cardMatchToDTO(m domain.CardMatch) *cardMatchDTO {
	if m.IsZero() {
		return nil
	}
	return &cardMatchDTO{Suit: string(m.Suit), Rank: int(m.Rank)}
}

func cardMatchFromDTO(d *cardMatchDTO) domain.CardMatch {
	if d == nil {
		return domain.CardMatch{}
	}
	return domain.CardMatch{Suit: domain.Suit(d.Suit), Rank: domain.Rank(d.Rank)}
}

func objectiveToDTO(o domain.Objective) objectiveDTO {
	return objectiveDTO{
		Kind: string(o.Kind), Owner: o.Owner, Cards: mapSlice(o.Cards, cardToDTO), Count: o.Count,
		Tricks: mapSlice(o.Tricks, intCopy), Suit: string(o.Suit), Rank: int(o.Rank),
		NotWithTrump: o.NotWithTrump, Ordered: o.Ordered,
		Cmp:   string(o.Cmp),
		Suits: mapSlice(o.Suits, func(s domain.Suit) string { return string(s) }),
		Ranks: mapSlice(o.Ranks, func(r domain.Rank) int { return int(r) }),
		With:  cardMatchToDTO(o.With), Target: cardMatchToDTO(o.Target),
		Where: string(o.Where), Max: o.Max, Exact: o.Exact, InOneTrick: o.InOneTrick, Hidden: o.Hidden,
		Parts: mapSlice(o.Parts, objectiveToDTO),
	}
}

func objectiveFromDTO(d objectiveDTO) domain.Objective {
	return domain.Objective{
		Kind: domain.ObjectiveKind(d.Kind), Owner: d.Owner, Cards: mapSlice(d.Cards, cardFromDTO), Count: d.Count,
		Tricks: mapSlice(d.Tricks, intCopy), Suit: domain.Suit(d.Suit), Rank: domain.Rank(d.Rank),
		NotWithTrump: d.NotWithTrump, Ordered: d.Ordered,
		Cmp:   domain.Comparison(d.Cmp),
		Suits: mapSlice(d.Suits, func(s string) domain.Suit { return domain.Suit(s) }),
		Ranks: mapSlice(d.Ranks, func(r int) domain.Rank { return domain.Rank(r) }),
		With:  cardMatchFromDTO(d.With), Target: cardMatchFromDTO(d.Target),
		Where: domain.TrickCondition(d.Where), Max: d.Max, Exact: d.Exact, InOneTrick: d.InOneTrick, Hidden: d.Hidden,
		Parts: mapSlice(d.Parts, objectiveFromDTO),
	}
}

func taskToDTO(t domain.Task) taskDTO {
	return taskDTO{ID: t.ID, Objective: objectiveToDTO(t.Objective), Token: string(t.Token), FaceDown: t.FaceDown,
		CardID: t.CardID, Difficulty: t.Difficulty}
}

func taskFromDTO(d taskDTO) domain.Task {
	return domain.Task{ID: d.ID, Objective: objectiveFromDTO(d.Objective), Token: domain.Token(d.Token), FaceDown: d.FaceDown,
		CardID: d.CardID, Difficulty: d.Difficulty}
}

func fromDTO(d gameDTO) (*domain.Game, error) {
	if d.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf("unsupported game schema version %d (want %d)", d.SchemaVersion, schemaVersion)
	}
	g := &domain.Game{
		ID:        domain.GameID(d.ID),
		Phase:     domain.Phase(d.Phase),
		HostID:    domain.PlayerID(d.HostID),
		Players:   mapSlice(d.Players, func(p playerDTO) domain.Player { return domain.Player{ID: domain.PlayerID(p.ID), Name: p.Name} }),
		Edition:   domain.Edition(d.Edition),
		Options:   domain.Options{Distress: d.Options.Distress},
		Mission:   d.Mission,
		Attempt:   attemptFromDTO(d.Attempt),
		UsedTasks: mapSlice(d.UsedTasks, intCopy),
		Revision:  d.Revision,
		CreatedAt: d.CreatedAt,
	}
	if d.Logbook != nil {
		g.Logbook = make(map[int]domain.LogEntry, len(d.Logbook))
		for n, e := range d.Logbook {
			g.Logbook[n] = domain.LogEntry{Attempts: e.Attempts, Distress: e.Distress, Completed: e.Completed}
		}
	}
	return g, nil
}

func attemptFromDTO(d *attemptDTO) *domain.Attempt {
	if d == nil {
		return nil
	}
	return &domain.Attempt{
		Number:          d.Number,
		Members:         mapSlice(d.Members, memberFromDTO),
		Commander:       d.Commander,
		Step:            domain.Step(d.Step),
		Tasks:           mapSlice(d.Tasks, taskFromDTO),
		Objectives:      mapSlice(d.Objectives, objectiveFromDTO),
		Poll:            mapPtr(d.Poll, func(p pollDTO) domain.Poll { return domain.Poll{Answers: mapSlice(p.Answers, copyBool)} }),
		Silenced:        d.Silenced,
		Comms:           commsFromDTO(d.Comms, d.MoreComms),
		DistressDecided: d.DistressDecided,
		Pass: mapPtr(d.Pass, func(p passDTO) domain.Pass {
			return domain.Pass{Direction: domain.Direction(p.Direction), Chosen: mapSlice(p.Chosen, cardPtrFromDTO)}
		}),
		Exchange: mapPtr(d.Exchange, func(e exchangeDTO) domain.Exchange {
			return domain.Exchange{Kind: domain.ExchangeKind(e.Kind), Gave: mapSlice(e.Gave, cardPtrFromDTO), Got: mapSlice(e.Got, cardPtrFromDTO)}
		}),
		HandoverUsed:   d.HandoverUsed,
		TokensAdjusted: d.TokensAdjusted,
		Trick:          domain.Trick{Leader: d.Trick.Leader, Plays: mapSlice(d.Trick.Plays, playFromDTO)},
		Tricks: mapSlice(d.Tricks, func(t trickRecordDTO) domain.TrickRecord {
			return domain.TrickRecord{Leader: t.Leader, Plays: mapSlice(t.Plays, playFromDTO), Winner: t.Winner}
		}),
		Result: mapPtr(d.Result, func(r resultDTO) domain.AttemptResult {
			return domain.AttemptResult{Success: r.Success, Reason: domain.Reason(r.Reason), Task: r.Task, Objective: r.Objective, Trick: r.Trick, Seat: r.Seat}
		}),
		Comm:     domain.CommMode(d.Comm),
		Terrain:  cardPtrFromDTO(d.Terrain),
		TaskDeck: mapSlice(d.TaskDeck, intCopy),
		Passes:   d.Passes,
		Chosen:   mapSlice(d.Chosen, intCopy),
		Redeals:  d.Redeals,
	}
}

func commToDTO(c domain.Communication) commDTO {
	return commDTO{Card: cardToDTO(c.Card), Position: string(c.Position), Trick: c.Trick}
}

func commFromDTO(c commDTO) domain.Communication {
	return domain.Communication{Card: cardFromDTO(c.Card), Position: domain.Position(c.Position), Trick: c.Trick}
}

// moreCommsToDTO returns each seat's communications after the first, or nil if there are none.
func moreCommsToDTO(comms [][]domain.Communication) [][]commDTO {
	var out [][]commDTO
	for i, cs := range comms {
		if len(cs) < 2 {
			continue
		}
		if out == nil {
			out = make([][]commDTO, len(comms))
		}
		out[i] = mapSlice(cs[1:], commToDTO)
	}
	return out
}

// commsFromDTO joins the first and the further communications per seat. Snapshots before Deep
// Sea only have the first.
func commsFromDTO(first []*commDTO, more [][]commDTO) [][]domain.Communication {
	out := make([][]domain.Communication, len(first))
	for i, c := range first {
		if c == nil {
			continue
		}
		out[i] = []domain.Communication{commFromDTO(*c)}
		if i < len(more) {
			out[i] = append(out[i], mapSlice(more[i], commFromDTO)...)
		}
	}
	return out
}
