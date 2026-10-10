package apidto

import "github.com/martin-pouls/the-game/backend/internal/core/domain"

// FromCard maps a domain card.
func FromCard(c domain.Card) Card {
	return Card{Value: c.Value, Kind: string(c.Kind), Command: string(c.Command)}
}

// FromOptions maps domain options.
func FromOptions(o domain.Options) GameOptions {
	return GameOptions{Variant: string(o.Variant), Pro: o.Pro, ReducedHandSize: o.ReducedHandSize}
}

// ToDomain maps wire options to domain options. An empty variant means classic.
func (o GameOptions) ToDomain() domain.Options {
	v := domain.Variant(o.Variant)
	if v == "" {
		v = domain.VariantClassic
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
		out.Hand = append(out.Hand, HandCard{Card: FromCard(h.Card), LegalPiles: intsOrEmpty(h.LegalPiles)})
	}
	for _, p := range v.Piles {
		pile := Pile{
			Index:      p.Index,
			Direction:  string(p.Direction),
			StartValue: p.StartValue,
			TopValue:   p.TopValue,
			CardCount:  p.CardCount,
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
			Plays:           make([]Play, 0, len(t.Plays)),
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
	return GameSummary{
		GameID:      string(s.GameID),
		Phase:       string(s.Phase),
		HostName:    s.HostName,
		PlayerCount: s.PlayerCount,
		MaxPlayers:  s.MaxPlayers,
		Joinable:    s.Joinable,
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
	case domain.TurnEnded:
		return Event{Type: EventTurnEnded, PlayerID: string(e.PlayerID), CardsDrawn: ptr(e.CardsDrawn), Auto: e.Auto}, true
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

func ptr[T any](v T) *T { return &v }

func intsOrEmpty(s []int) []int {
	if s == nil {
		return []int{}
	}
	return append([]int{}, s...)
}
