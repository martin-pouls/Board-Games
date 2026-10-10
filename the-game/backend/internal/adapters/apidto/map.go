package apidto

import (
	"time"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
)

// FromCard maps a domain card.
func FromCard(c domain.Card) Card {
	return Card{Value: c.Value, Kind: string(c.Kind), Command: string(c.Command)}
}

// FromOptions maps domain options.
func FromOptions(o domain.Options) GameOptions {
	return GameOptions{Variant: string(o.Variant), Pro: o.Pro, ReducedHandSize: o.ReducedHandSize}
}

// ToDomain maps wire options to domain options. An empty variant means the default variant.
func (o GameOptions) ToDomain() domain.Options {
	v := domain.Variant(o.Variant)
	if v == "" {
		v = domain.DefaultOptions().Variant
	}
	return domain.Options{Variant: v, Pro: o.Pro, ReducedHandSize: o.ReducedHandSize}
}

// FromView maps a player view. Slices are never nil.
func FromView(v domain.PlayerView) PlayerView {
	out := PlayerView{
		GameID:   string(v.GameID),
		Revision: v.Revision,
		Phase:    string(v.Phase),
		You:      string(v.You),
		HostID:   string(v.HostID),
		Options:  FromOptions(v.Options),
		Rules: Rules{
			MinCardsPerTurn:       v.Rules.MinCardsPerTurn,
			MinCardsDrawPileEmpty: v.Rules.MinCardsDrawPileEmpty,
			HandSize:              v.Rules.HandSize,
			MinPlayers:            v.Rules.MinPlayers,
			MaxPlayers:            v.Rules.MaxPlayers,
		},
		Players:                 make([]PlayerSummary, 0, len(v.Players)),
		Hand:                    make([]HandCard, 0, len(v.Hand)),
		Piles:                   make([]Pile, 0, len(v.Piles)),
		DrawPileCount:           v.DrawPileCount,
		CommunicationRestricted: v.CommunicationRestricted,
		NextGameID:              string(v.NextGameID),
		ActiveCommands:          make([]ActiveCommand, 0, len(v.ActiveCommands)),
	}
	for _, a := range v.ActiveCommands {
		out.ActiveCommands = append(out.ActiveCommands, FromActiveCommand(a))
	}
	for _, p := range v.Players {
		var revealed []Card
		for _, c := range p.RevealedHand {
			revealed = append(revealed, FromCard(c))
		}
		out.Players = append(out.Players, PlayerSummary{
			RevealedHand: revealed,
			ID:           string(p.ID),
			Name:         p.Name,
			Seat:         p.Seat,
			HandCount:    p.HandCount,
			IsHost:       p.IsHost,
			IsYou:        p.IsYou,
			IsActive:     p.IsActive,
		})
	}
	for _, h := range v.Hand {
		hc := HandCard{
			Card:       FromCard(h.Card),
			LegalPiles: intsOrEmpty(h.LegalPiles),
			Blocked:    blockedCode(h.Blocked),
		}
		for _, veto := range h.Vetoes {
			if code := blockedCode(veto.Err); code != "" {
				hc.Vetoes = append(hc.Vetoes, PileVeto{Pile: veto.Pile, Code: code})
			}
		}
		out.Hand = append(out.Hand, hc)
	}
	for _, p := range v.Piles {
		pile := Pile{
			Index:         p.Index,
			Direction:     string(p.Direction),
			StartValue:    p.StartValue,
			TopValue:      p.TopValue,
			CardCount:     p.CardCount,
			Cards:         make([]Card, 0, len(p.Cards)),
			CommandActive: p.CommandActive,
		}
		for _, c := range p.Cards {
			pile.Cards = append(pile.Cards, FromCard(c))
		}
		if p.TopCard != nil {
			c := FromCard(*p.TopCard)
			pile.TopCard = &c
		}
		out.Piles = append(out.Piles, pile)
	}
	if t := v.Turn; t != nil {
		turn := &Turn{
			Number:          t.Number,
			ActivePlayerID:  string(t.ActivePlayerID),
			CardsPlayed:     t.CardsPlayed,
			RequiredMinimum: t.RequiredMinimum,
			CanEndTurn:      t.CanEndTurn,
			Stuck:           t.Stuck,
			Plays:           make([]Play, 0, len(t.Plays)),
			Constraints:     fromConstraints(t.Constraints),
		}
		for _, p := range t.Plays {
			turn.Plays = append(turn.Plays, Play{Card: FromCard(p.Card), Pile: p.Pile})
		}
		out.Turn = turn
	}
	if r := v.Result; r != nil {
		out.Result = &Result{
			Outcome:             string(r.Outcome),
			Score:               r.Score,
			Excellent:           r.Excellent,
			CardsLeftInHands:    r.CardsLeftInHands,
			CardsLeftInDrawPile: r.CardsLeftInDrawPile,
			StuckPlayerID:       string(r.StuckPlayerID),
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
		Outcome:     string(s.Outcome),
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

// FromEvent maps one domain event; false for unknown types.
func FromEvent(e domain.Event) (Event, bool) {
	switch e := e.(type) {
	case domain.PlayerJoined:
		return Event{Type: EventPlayerJoined, PlayerID: string(e.PlayerID), Name: e.Name}, true
	case domain.OptionsChanged:
		return Event{Type: EventOptionsChanged}, true
	case domain.GameStarted:
		return Event{Type: EventGameStarted}, true
	case domain.FirstPlayerChosen:
		return Event{Type: EventFirstPlayerChosen, PlayerID: string(e.PlayerID)}, true
	case domain.TurnStarted:
		return Event{Type: EventTurnStarted, PlayerID: string(e.PlayerID), Number: ptr(e.Number)}, true
	case domain.CardPlayed:
		c := FromCard(e.Card)
		return Event{Type: EventCardPlayed, PlayerID: string(e.PlayerID), Card: &c, Pile: ptr(e.Pile)}, true
	case domain.PlayUndone:
		c := FromCard(e.Card)
		return Event{Type: EventPlayUndone, PlayerID: string(e.PlayerID), Card: &c, Pile: ptr(e.Pile)}, true
	case domain.TurnEnded:
		return Event{Type: EventTurnEnded, PlayerID: string(e.PlayerID), CardsDrawn: ptr(e.CardsDrawn)}, true
	case domain.PlayerSkipped:
		return Event{Type: EventPlayerSkipped, PlayerID: string(e.PlayerID)}, true
	case domain.GameFinished:
		return Event{
			Type:          EventGameFinished,
			Outcome:       string(e.Outcome),
			Score:         ptr(e.Score),
			StuckPlayerID: string(e.StuckPlayerID),
		}, true
	case domain.RematchCreated:
		return Event{Type: EventRematchCreated, GameID: string(e.GameID)}, true
	default:
		return Event{}, false
	}
}

// FromActiveCommand maps a visible Infinity command.
func FromActiveCommand(a domain.ActiveCommand) ActiveCommand {
	return ActiveCommand{Command: string(a.Command), Pile: a.Pile}
}

// fromConstraints maps turn constraints: ExactCount 0 is null, SkullPiles is never null,
// BoundPile is copied into a fresh pointer.
func fromConstraints(c domain.TurnConstraints) TurnConstraints {
	out := TurnConstraints{
		SinglePile: c.SinglePile,
		SkullPiles: intsOrEmpty(c.SkullPiles),
		StopPlayed: c.StopPlayed,
	}
	if c.ExactCount != 0 {
		out.ExactCount = ptr(c.ExactCount)
	}
	if c.BoundPile != nil {
		out.BoundPile = ptr(*c.BoundPile)
	}
	return out
}

// blockedCode maps a hand card's command veto to its wire error code:
// "" for nil and for errors without a specific code (internal).
func blockedCode(err error) string {
	if err == nil {
		return ""
	}
	if _, body := ErrorFor(err); body.Code != CodeInternal {
		return body.Code
	}
	return ""
}

func ptr[T any](v T) *T { return &v }

func intsOrEmpty(s []int) []int {
	if s == nil {
		return []int{}
	}
	return append([]int{}, s...)
}
