package domain

import (
	"slices"
	"time"
)

// Phase is the lifecycle phase of a game.
type Phase string

const (
	PhaseLobby   Phase = "lobby"
	PhasePlaying Phase = "playing"
	PhaseWon     Phase = "won"
	PhaseLost    Phase = "lost"
	PhaseEnded   Phase = "ended" // closed by the host in the lobby, before it started
)

// Step is the pending decision while playing.
type Step string

const (
	StepNone        Step = ""            // not playing
	StepLead        Step = "lead"        // the leader plays a card
	StepFollow      Step = "follow"      // the follower plays a card
	StepFoxesChoose Step = "foxesChoose" // whoever played the Foxes chooses a player
	StepFoxesSwap   Step = "foxesSwap"   // the chosen player swaps the decree or keeps it
	StepGift        Step = "gift"        // both players choose a card to give
	StepMove        Step = "move"        // the trick winner chooses how the traveller moves
	StepForestEnd   Step = "forestEnd"   // round end: the host chooses the end for the forest token
)

// LossReason says why a game was lost.
type LossReason string

const (
	LossLostInForest LossReason = "lostInForest" // beyond the path with no forest token left
	LossOutOfTime    LossReason = "outOfTime"    // gems left after round 3
	LossGaveUp       LossReason = "gaveUp"
)

// Game sizes (rules §5).
const (
	Players          = 2
	HandSize         = 11
	TricksPerRound   = HandSize
	Rounds           = 3
	EarlyWinBonus    = 10
	ForestTokenWorth = 3
)

// Player is a seated player. Seat 0 faces the left end, seat 1 the right end.
type Player struct {
	ID   PlayerID
	Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively
	Hand []Card // always sorted (SortCards); never nil once seated ([]Card{})
}

// Trick is the trick in progress.
type Trick struct {
	Number int // 1..TricksPerRound within the round
	Leader int // seat
	Lead   *Card
	Follow *Card
	Chosen int      // StepFoxesSwap: the seat that may swap the decree
	Gift   [2]*Card // StepGift: the card each seat has chosen to give, nil while undecided
}

// played returns the seat that played the last card of the trick and that card, or -1 if no
// card has been played. A pending ability belongs to this card.
func (t *Trick) played() (int, *Card) {
	switch {
	case t.Follow != nil:
		return 1 - t.Leader, t.Follow
	case t.Lead != nil:
		return t.Leader, t.Lead
	}
	return -1, nil
}

// TrickRecord is a completed trick, shown until the next card is led.
type TrickRecord struct {
	Number    int
	Leader    int
	Lead      Card
	Follow    Card
	Winner    int  // seat
	From      int  // traveller space before the move
	Delta     int  // chosen move, negative = toward the left end
	To        int  // traveller space after the move (the start space after a reset)
	Beyond    End  // the end the traveller left the path at, EndNone if it stayed on it
	Collected bool // a gem was collected at To
}

// GiftRecord is the last completed gift exchange, shown until the next card is led.
// Given[seat] is the card that seat gave to the other.
type GiftRecord struct {
	Given [2]Card
}

// Result is set iff Phase is won or lost.
type Result struct {
	Reason   LossReason // empty on a win
	PlayerID PlayerID   // who gave up; empty otherwise
}

// Game is the aggregate root. Fields are exported for the persistence mapper and
// tests; code outside this package must mutate a Game only through its methods.
// The trick winner, legal cards and move options are derived, never stored.
type Game struct {
	ID         GameID
	Phase      Phase
	Step       Step // StepNone unless Phase == playing
	HostID     PlayerID
	Players    []Player // seat order = join order
	Level      int      // 0 in the lobby
	Board      Board    // zero in the lobby
	Round      int      // 1..Rounds; 0 in the lobby
	Dealer     int      // seat of this round's dealer
	Deck       []Card   // the hidden rest of the deck after the decree; order irrelevant
	Decree     *Card    // nil in the lobby
	Trick      Trick
	LastTrick  *TrickRecord
	LastGift   *GiftRecord
	Result     *Result
	NextGameID GameID    // set when a rematch was created; empty otherwise
	Revision   int64     // 1 after NewGame; +1 per successful mutator
	CreatedAt  time.Time // UTC, second precision
	events     []Event   // pending, not persisted
}

// Clone returns a deep copy, including pending events. Nil and empty slices are preserved.
func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	c := *g
	if g.Players != nil {
		c.Players = make([]Player, len(g.Players))
		for i, p := range g.Players {
			p.Hand = slices.Clone(p.Hand)
			c.Players[i] = p
		}
	}
	c.Board = g.Board.clone()
	c.Deck = slices.Clone(g.Deck)
	c.Decree = cloneCard(g.Decree)
	c.Trick.Lead = cloneCard(g.Trick.Lead)
	c.Trick.Follow = cloneCard(g.Trick.Follow)
	for i := range c.Trick.Gift {
		c.Trick.Gift[i] = cloneCard(g.Trick.Gift[i])
	}
	if g.LastTrick != nil {
		t := *g.LastTrick
		c.LastTrick = &t
	}
	if g.LastGift != nil {
		gr := *g.LastGift
		c.LastGift = &gr
	}
	if g.Result != nil {
		r := *g.Result
		c.Result = &r
	}
	// Events are immutable values whose slices are never shared with state,
	// so a shallow slice copy is enough.
	c.events = slices.Clone(g.events)
	return &c
}

// PullEvents returns the pending events and clears them.
func (g *Game) PullEvents() []Event {
	evs := g.events
	g.events = nil
	return evs
}

// PlayerByID returns the player with the given ID, their seat, and whether they were found.
func (g *Game) PlayerByID(id PlayerID) (Player, int, bool) {
	for i, p := range g.Players {
		if p.ID == id {
			return p, i, true
		}
	}
	return Player{}, -1, false
}

// IsHost reports whether id is the host.
func (g *Game) IsHost(id PlayerID) bool {
	return id != "" && id == g.HostID
}

// LevelDef returns the setup of the game's level (zero LevelDef in the lobby).
func (g *Game) LevelDef() LevelDef {
	l, _ := LevelFor(g.Level)
	return l
}

// Summary returns the public, pre-join summary of the game.
func (g *Game) Summary() GameSummary {
	var hostName string
	if host, _, ok := g.PlayerByID(g.HostID); ok {
		hostName = host.Name
	}
	n := len(g.Players)
	return GameSummary{
		GameID:      g.ID,
		Phase:       g.Phase,
		HostName:    hostName,
		PlayerCount: n,
		MaxPlayers:  Players,
		Joinable:    g.Phase == PhaseLobby && n < Players,
		CreatedAt:   g.CreatedAt,
	}
}

// emit queues a pending domain event.
func (g *Game) emit(e Event) {
	g.events = append(g.events, e)
}

// commit marks a successful mutation: Revision++ exactly once per exported mutator.
func (g *Game) commit() {
	g.Revision++
}
