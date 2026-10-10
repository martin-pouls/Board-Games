package domain

import (
	"fmt"
	"slices"
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
	Turn                    *TurnView // nil in lobby / choosingFirstPlayer
	CommunicationRestricted bool
	Result                  *ResultView // nil unless finished
	NextGameID              GameID      // empty until a rematch exists
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
	RevealedHand            []Card // sorted; filled for every player only when Phase == finished, else nil
}

// HandCard is a card in the viewer's hand plus where it may legally go now.
type HandCard struct {
	Card       Card
	LegalPiles []int // empty unless viewer is active in playing
}

// PileView is the public state of a pile.
type PileView struct {
	Index      int
	Direction  Direction
	StartValue int
	TopValue   int
	TopCard    *Card // nil when only the start card shows
	CardCount  int   // played cards on the pile
}

// TurnView is the public state of the current turn.
type TurnView struct {
	Number          int
	ActivePlayerID  PlayerID
	CardsPlayed     int
	RequiredMinimum int
	CanEndTurn      bool // CanEndTurn() == nil
	Plays           []Play
}

// ResultView is the public final result.
type ResultView struct {
	Outcome             Outcome
	Score               int
	Excellent           bool // Score < 10
	CardsLeftInHands    int
	CardsLeftInDrawPile int
	StuckPlayerID       PlayerID
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool // Phase == lobby && PlayerCount < MaxPlayers
}

// excellentScoreBelow: a score below this is "excellent" (rules.md, End of game).
const excellentScoreBelow = 10

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	me, _, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	playing := g.Phase == PhasePlaying
	finished := g.Phase == PhaseFinished

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
		if finished {
			s.RevealedHand = make([]Card, len(p.Hand))
			copy(s.RevealedHand, p.Hand)
			sortCards(s.RevealedHand)
		}
		v.Players[seat] = s
	}

	legal := map[int][]int{}
	for _, m := range g.LegalMoves(id) { // nil unless id is the active player in playing
		legal[m.Card] = append(legal[m.Card], m.Pile)
	}
	for _, c := range me.Hand {
		v.Hand = append(v.Hand, HandCard{Card: c, LegalPiles: legal[c.Value]})
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
		}
		v.Piles[i] = pv
	}

	if playing || finished {
		v.Turn = &TurnView{
			Number:          g.Turn.Number,
			ActivePlayerID:  g.Players[g.Turn.ActiveSeat].ID,
			CardsPlayed:     len(g.Turn.Plays),
			RequiredMinimum: g.RequiredMinimum(),
			CanEndTurn:      g.CanEndTurn() == nil,
			Plays:           slices.Clone(g.Turn.Plays),
		}
	}

	if finished && g.Result != nil {
		r := g.Result
		v.Result = &ResultView{
			Outcome:             r.Outcome,
			Score:               r.Score,
			Excellent:           r.Score < excellentScoreBelow,
			CardsLeftInHands:    r.CardsLeftInHands,
			CardsLeftInDrawPile: r.CardsLeftInDrawPile,
			StuckPlayerID:       r.StuckPlayerID,
		}
	}
	return v, nil
}
