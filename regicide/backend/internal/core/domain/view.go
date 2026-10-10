package domain

import (
	"fmt"
	"slices"
	"time"
)

// PlayerView is the projection of a game for one player. It never exposes another player's hand
// before the game ends, nor the order of the Tavern or Castle, nor the Discard contents.
type PlayerView struct {
	GameID          GameID
	Revision        int64
	Phase           Phase
	You, HostID     PlayerID
	Rules           RulesView
	Players         []PlayerSummary // seat order
	Hand            []Card          // only the viewer's hand, sorted; [] in the lobby
	Enemy           *EnemyView      // nil in the lobby and after a win
	CastleCount     int             // facedown Castle cards (excludes Enemy)
	TavernCount     int
	DiscardCount    int
	Turn            *TurnView // nil in the lobby
	SoloJestersLeft int       // 0 unless solo
	Result          *ResultView
	NextGameID      GameID
}

// RulesView holds the ruleset values for the current player count.
type RulesView struct {
	MinPlayers, MaxPlayers, MaxHandSize, TavernJesters, SoloJesters int
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID              PlayerID
	Name            string
	Seat            int
	HandCount       int
	IsHost, IsYou   bool
	IsActive        bool // playing && seat == ActiveSeat
	YieldedLastTurn bool
	Signal          JesterSignal
	RevealedHand    []Card // every player's sorted hand once the game is over; nil before
}

// EnemyView is the current enemy with its derived numbers.
type EnemyView struct {
	Card                            Card
	Attack, Health                  int // printed values
	Damage, Shield, EffectiveAttack int
	Immunity                        Suit // the enemy's suit
	ImmunityCancelled               bool // a Jester was played against it
	Plays                           []PlayedSetView
}

// PlayedSetView is one play against the current enemy with its derived effect.
type PlayedSetView struct {
	By           PlayerID
	Cards        []Card
	Value        int
	Damage       int
	Doubled      bool
	BlockedSuits []Suit
	Jester       bool
}

// TurnView describes the current turn.
type TurnView struct {
	Number         int
	ActivePlayerID PlayerID
	Step           Step // StepNone once the game is over
	CanYield       bool // Step == play && the active player may yield
	DamageToCover  int  // the effective attack in Step discard, else 0
}

// ResultView describes a finished game.
type ResultView struct {
	Outcome         Phase // won | lost
	Reason          LossReason
	PlayerID        PlayerID
	Grade           Grade
	EnemiesDefeated int
	SoloJestersUsed int
}

// GameSummary is the public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool // lobby && PlayerCount < MaxPlayers
	CreatedAt   time.Time
}

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	me, _, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	over := g.Phase == PhaseWon || g.Phase == PhaseLost
	r := g.Rules()
	v := PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    g.Phase,
		You:      id,
		HostID:   g.HostID,
		Rules: RulesView{
			MinPlayers:    MinPlayers,
			MaxPlayers:    MaxPlayers,
			MaxHandSize:   r.MaxHandSize,
			TavernJesters: r.TavernJesters,
			SoloJesters:   r.SoloJesters,
		},
		Players:         make([]PlayerSummary, len(g.Players)),
		Hand:            cloneCards(me.Hand),
		CastleCount:     len(g.Castle),
		TavernCount:     len(g.Tavern),
		DiscardCount:    len(g.DiscardPile),
		SoloJestersLeft: g.SoloJestersLeft(),
		NextGameID:      g.NextGameID,
	}
	for i, p := range g.Players {
		ps := PlayerSummary{
			ID:              p.ID,
			Name:            p.Name,
			Seat:            i,
			HandCount:       len(p.Hand),
			IsHost:          p.ID == g.HostID,
			IsYou:           p.ID == id,
			IsActive:        g.Phase == PhasePlaying && i == g.Turn.ActiveSeat,
			YieldedLastTurn: p.YieldedLastTurn,
			Signal:          p.Signal,
		}
		if over {
			ps.RevealedHand = cloneCards(p.Hand)
		}
		v.Players[i] = ps
	}
	if g.Enemy != nil {
		ev := &EnemyView{
			Card:              *g.Enemy,
			Attack:            EnemyAttack(g.Enemy.Rank),
			Health:            EnemyHealth(g.Enemy.Rank),
			Damage:            g.Damage(),
			Shield:            g.Shield(),
			EffectiveAttack:   g.EffectiveAttack(),
			Immunity:          g.Enemy.Suit,
			ImmunityCancelled: g.JesterPlayed(),
			Plays:             make([]PlayedSetView, len(g.Played)),
		}
		for i, ps := range g.Played {
			e := g.PlayEffectAt(i)
			ev.Plays[i] = PlayedSetView{
				By:           ps.By,
				Cards:        cloneCards(ps.Cards),
				Value:        e.Value,
				Damage:       e.Damage,
				Doubled:      e.Doubled,
				BlockedSuits: slices.Clone(e.BlockedSuits),
				Jester:       e.Jester,
			}
		}
		v.Enemy = ev
	}
	if g.Phase != PhaseLobby {
		tv := &TurnView{
			Number:         g.Turn.Number,
			ActivePlayerID: g.Players[g.Turn.ActiveSeat].ID,
			Step:           g.Step,
			CanYield:       g.Step == StepPlay && g.CanYield(g.Turn.ActiveSeat),
		}
		if g.Step == StepDiscard {
			tv.DamageToCover = g.EffectiveAttack()
		}
		v.Turn = tv
	}
	if over && g.Result != nil {
		v.Result = &ResultView{
			Outcome:         g.Phase,
			Reason:          g.Result.Reason,
			PlayerID:        g.Result.PlayerID,
			Grade:           g.Grade(),
			EnemiesDefeated: g.EnemiesDefeated(),
			SoloJestersUsed: g.SoloJestersUsed,
		}
	}
	return v, nil
}

// cloneCards copies cs; the result is never nil.
func cloneCards(cs []Card) []Card {
	out := make([]Card, len(cs))
	copy(out, cs)
	return out
}
