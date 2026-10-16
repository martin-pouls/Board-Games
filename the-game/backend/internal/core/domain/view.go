package domain

import (
	"fmt"
	"slices"
	"time"
)

// PlayerView is the per-player projection of a game. It never contains other players' hands.
type PlayerView struct {
	GameID                  GameID
	Revision                int64
	Phase                   Phase
	You                     PlayerID
	HostID                  PlayerID
	Options                 Options
	Rules                   RulesView
	Players                 []PlayerSummary // seat order
	Hand                    []HandCard      // ONLY the viewer's hand, sorted ascending; empty in lobby
	Piles                   []PileView
	DrawPileCount           int
	Turn                    *TurnView // nil in lobby / choosingFirstPlayer / ended, and when given up before the first turn
	CommunicationRestricted bool
	Result                  *ResultView     // nil unless won or lost
	NextGameID              GameID          // empty until a rematch exists
	ActiveCommands          []ActiveCommand // visible Infinity commands, pile order; nil if none
}

// RulesView is the public subset of the ruleset.
type RulesView struct {
	MinCardsPerTurn, MinCardsDrawPileEmpty int
	HandSize                               int // for the current player count (preview in lobby)
	MinPlayers, MaxPlayers                 int
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID                      PlayerID
	Name                    string
	Seat                    int
	HandCount               int
	IsHost, IsYou, IsActive bool
	RevealedHand            []Card // sorted; filled for every player only once won or lost, else nil
}

// HandCard is a card in the viewer's hand plus where it may legally go now.
type HandCard struct {
	Card       Card
	LegalPiles []int // empty unless viewer is active in playing
	// Blocked is the first command veto (a typed sentinel from errors.go, wrapped) for a card
	// with no legal pile that the base rules alone would allow somewhere; nil otherwise.
	// Only set for the active viewer in playing.
	Blocked error
	// Vetoes are the piles the base rules allow but a command forbids, with the reason,
	// in pile order. Only set for the active viewer in playing.
	Vetoes []PileVeto
}

// PileView is the public state of a pile.
type PileView struct {
	Index      int
	Direction  Direction
	StartValue int
	TopValue   int
	TopCard    *Card // nil when only the start card shows
	CardCount  int   // played cards on the pile
	// CommandActive says the top card's command is in effect: an Infinity top, or a
	// lightning top played this turn.
	CommandActive bool
}

// TurnView is the public state of the current turn.
type TurnView struct {
	Number          int
	ActivePlayerID  PlayerID
	CardsPlayed     int
	RequiredMinimum int
	CanEndTurn      bool // CanEndTurn() == nil
	// Stuck: the active player can neither play nor end the turn; they must undo or give up.
	Stuck       bool
	Plays       []Play
	Constraints TurnConstraints // zero value = no command constrains this turn
}

// TurnConstraints are the command constraints on the current turn (server facts the
// client renders without re-deriving rules).
type TurnConstraints struct {
	ExactCount int   // 3 once a 3! was played this turn; 0 = none
	SinglePile bool  // a visible Single pile card restricts this turn
	BoundPile  *int  // pile all further cards must go on; nil = unbound (the next card picks) or no Single pile
	SkullPiles []int // piles showing a Skull played this turn; nil = none
	StopPlayed bool  // a STOP was played this turn: no further card, only End turn (or undo)
}

// ResultView is the public final result (the phase says won or lost).
type ResultView struct {
	Score               int
	Excellent           bool // Score < 10
	CardsLeftInHands    int
	CardsLeftInDrawPile int
	Reason              LossReason // "" when won
	PlayerID            PlayerID   // the stuck player or the one who gave up
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool      // Phase == lobby && PlayerCount < MaxPlayers
	CreatedAt   time.Time // zero when unknown (old snapshot)
}

// excellentScoreBelow: a score below this is "excellent" (rules.md, End of game).
const excellentScoreBelow = 10

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	me, mySeat, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	playing := g.Phase == PhasePlaying
	over := g.IsOver()
	cs := g.commands()

	v := PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    g.Phase,
		You:      id,
		HostID:   g.HostID,
		Options:  g.Options,
		Rules: RulesView{
			MinCardsPerTurn:       g.Rules.MinCardsPerTurn,
			MinCardsDrawPileEmpty: g.Rules.MinCardsDrawPileEmpty,
			HandSize:              g.Rules.HandSize(len(g.Players)),
			MinPlayers:            g.Rules.MinPlayers,
			MaxPlayers:            g.Rules.MaxPlayers,
		},
		Players:                 make([]PlayerSummary, len(g.Players)),
		Hand:                    make([]HandCard, 0, len(me.Hand)),
		Piles:                   make([]PileView, len(g.Piles)),
		DrawPileCount:           len(g.DrawPile),
		CommunicationRestricted: g.CommunicationRestricted(),
		NextGameID:              g.NextGameID,
		ActiveCommands:          slices.Clone(cs.active),
	}

	for seat, p := range g.Players {
		s := PlayerSummary{
			ID:        p.ID,
			Name:      p.Name,
			Seat:      seat,
			HandCount: len(p.Hand),
			IsHost:    g.IsHost(p.ID),
			IsYou:     p.ID == id,
			IsActive:  playing && seat == g.Turn.ActiveSeat,
		}
		if over {
			s.RevealedHand = make([]Card, len(p.Hand))
			copy(s.RevealedHand, p.Hand)
			sortCards(s.RevealedHand)
		}
		v.Players[seat] = s
	}

	active := playing && mySeat == g.Turn.ActiveSeat
	var cons []PlacementConstraint
	if active {
		cons = g.placementConstraints()
	}
	for _, c := range me.Hand {
		hc := HandCard{Card: c}
		if active {
			hc.LegalPiles, hc.Vetoes = g.cardLegality(cons, c)
			if len(hc.LegalPiles) == 0 && len(hc.Vetoes) > 0 {
				hc.Blocked = hc.Vetoes[0].Err
			}
		}
		v.Hand = append(v.Hand, hc)
	}

	for i, p := range g.Piles {
		pv := PileView{
			Index:      i,
			Direction:  p.Direction,
			StartValue: p.StartValue,
			TopValue:   p.TopValue(),
			CardCount:  len(p.Cards),
		}
		if top, ok := p.TopCard(); ok {
			pv.TopCard = &top
			pv.CommandActive = top.Command.IsInfinity() ||
				(top.Command.IsLightning() && playedThisTurn(g.Turn.Plays, i))
		}
		v.Piles[i] = pv
	}

	// A game given up while choosing the first player never had a turn.
	if playing || (over && g.Turn.Number > 0) {
		v.Turn = &TurnView{
			Number:          g.Turn.Number,
			ActivePlayerID:  g.Players[g.Turn.ActiveSeat].ID,
			CardsPlayed:     len(g.Turn.Plays),
			RequiredMinimum: g.RequiredMinimum(),
			CanEndTurn:      g.CanEndTurn() == nil,
			Stuck:           g.ActivePlayerStuck(),
			Plays:           slices.Clone(g.Turn.Plays),
			Constraints: TurnConstraints{
				ExactCount: cs.exactCount,
				SinglePile: cs.singlePile,
				SkullPiles: slices.Clone(cs.skullPiles),
				StopPlayed: cs.stopPlayed,
			},
		}
		if cs.singlePile && cs.boundPile >= 0 {
			bound := cs.boundPile
			v.Turn.Constraints.BoundPile = &bound
		}
	}

	if over && g.Result != nil {
		r := g.Result
		v.Result = &ResultView{
			Score:               r.Score,
			Excellent:           r.Score < excellentScoreBelow,
			CardsLeftInHands:    r.CardsLeftInHands,
			CardsLeftInDrawPile: r.CardsLeftInDrawPile,
			Reason:              r.Reason,
			PlayerID:            r.PlayerID,
		}
	}
	return v, nil
}
