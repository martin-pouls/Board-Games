package apidto

import (
	"time"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
)

// FromCard maps a domain card.
func FromCard(c domain.Card) Card {
	return Card{Suit: string(c.Suit), Rank: int(c.Rank)}
}

// ToDomain maps a wire card to a domain card; the result must be a valid card.
func (c Card) ToDomain() (domain.Card, error) {
	dc := domain.NewCard(domain.Suit(c.Suit), domain.Rank(c.Rank))
	if !dc.Valid() {
		return domain.Card{}, BadRequest("invalid card (suit %q, rank %d)", c.Suit, c.Rank)
	}
	return dc, nil
}

// CardsToDomain maps wire cards to domain cards. A nil input yields nil, an empty one an empty slice.
func CardsToDomain(cs []Card) ([]domain.Card, error) {
	if cs == nil {
		return nil, nil
	}
	out := make([]domain.Card, len(cs))
	for i, c := range cs {
		dc, err := c.ToDomain()
		if err != nil {
			return nil, err
		}
		out[i] = dc
	}
	return out, nil
}

// SignalToWire maps a Jester signal to its wire value; the empty signal is "none".
func SignalToWire(s domain.JesterSignal) string {
	if s == domain.SignalNone {
		return "none"
	}
	return string(s)
}

// SignalFromWire maps a wire signal to the domain; "none" is the empty signal and
// anything else is passed through for the domain to validate.
func SignalFromWire(s string) domain.JesterSignal {
	if s == "none" {
		return domain.SignalNone
	}
	return domain.JesterSignal(s)
}

func fromCards(cs []domain.Card) []Card {
	out := make([]Card, len(cs))
	for i, c := range cs {
		out[i] = FromCard(c)
	}
	return out
}

func suitsToWire(ss []domain.Suit) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = string(s)
	}
	return out
}

// FromView maps a player view. Slices are never nil, except RevealedHand (nil until the game ends).
func FromView(v domain.PlayerView) PlayerView {
	out := PlayerView{
		GameID:   string(v.GameID),
		Revision: v.Revision,
		Phase:    string(v.Phase),
		You:      string(v.You),
		HostID:   string(v.HostID),
		Rules: Rules{
			MinPlayers:    v.Rules.MinPlayers,
			MaxPlayers:    v.Rules.MaxPlayers,
			MaxHandSize:   v.Rules.MaxHandSize,
			TavernJesters: v.Rules.TavernJesters,
			SoloJesters:   v.Rules.SoloJesters,
		},
		Players:         make([]PlayerSummary, 0, len(v.Players)),
		Hand:            fromCards(v.Hand),
		CastleCount:     v.CastleCount,
		TavernCount:     v.TavernCount,
		DiscardCount:    v.DiscardCount,
		SoloJestersLeft: v.SoloJestersLeft,
		NextGameID:      string(v.NextGameID),
	}
	for _, p := range v.Players {
		ps := PlayerSummary{
			ID:              string(p.ID),
			Name:            p.Name,
			Seat:            p.Seat,
			HandCount:       p.HandCount,
			IsHost:          p.IsHost,
			IsYou:           p.IsYou,
			IsActive:        p.IsActive,
			YieldedLastTurn: p.YieldedLastTurn,
			Signal:          SignalToWire(p.Signal),
		}
		if p.RevealedHand != nil {
			ps.RevealedHand = fromCards(p.RevealedHand)
		}
		out.Players = append(out.Players, ps)
	}
	if e := v.Enemy; e != nil {
		enemy := &Enemy{
			Card:              FromCard(e.Card),
			Attack:            e.Attack,
			Health:            e.Health,
			Damage:            e.Damage,
			Shield:            e.Shield,
			EffectiveAttack:   e.EffectiveAttack,
			Immunity:          string(e.Immunity),
			ImmunityCancelled: e.ImmunityCancelled,
			Plays:             make([]PlayedSet, 0, len(e.Plays)),
		}
		for _, p := range e.Plays {
			enemy.Plays = append(enemy.Plays, PlayedSet{
				PlayerID:     string(p.By),
				Cards:        fromCards(p.Cards),
				Value:        p.Value,
				Damage:       p.Damage,
				Doubled:      p.Doubled,
				BlockedSuits: suitsToWire(p.BlockedSuits),
				Jester:       p.Jester,
			})
		}
		out.Enemy = enemy
	}
	if lt := v.LastTurn; lt != nil {
		out.LastTurn = &LastTurn{PlayerID: string(lt.By)}
		if lt.Played != nil {
			out.LastTurn.Played = fromCards(lt.Played)
		}
		if lt.Defeated != nil {
			out.LastTurn.Defeated = ptr(FromCard(*lt.Defeated))
			out.LastTurn.Exact = lt.Exact
		}
		if lt.Discarded != nil {
			out.LastTurn.Discarded = fromCards(lt.Discarded)
		}
	}
	if t := v.Turn; t != nil {
		out.Turn = &Turn{
			Number:         t.Number,
			ActivePlayerID: string(t.ActivePlayerID),
			Step:           string(t.Step),
			CanYield:       t.CanYield,
			DamageToCover:  t.DamageToCover,
		}
	}
	if r := v.Result; r != nil {
		out.Result = &Result{
			Outcome:         string(r.Outcome),
			Reason:          string(r.Reason),
			PlayerID:        string(r.PlayerID),
			Grade:           string(r.Grade),
			EnemiesDefeated: r.EnemiesDefeated,
			SoloJestersUsed: r.SoloJestersUsed,
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

// FromEvent maps one domain event; false for unknown types.
func FromEvent(e domain.Event) (Event, bool) {
	switch e := e.(type) {
	case domain.PlayerJoined:
		return Event{Type: EventPlayerJoined, PlayerID: string(e.PlayerID), Name: e.Name}, true
	case domain.PlayerLeft:
		return Event{Type: EventPlayerLeft, PlayerID: string(e.PlayerID), HostID: string(e.HostID)}, true
	case domain.GameStarted:
		return Event{Type: EventGameStarted, PlayerID: string(e.FirstPlayerID)}, true
	case domain.EnemyRevealed:
		c := FromCard(e.Enemy)
		return Event{Type: EventEnemyRevealed, Card: &c}, true
	case domain.TurnStarted:
		return Event{Type: EventTurnStarted, PlayerID: string(e.PlayerID), Number: ptr(e.Number)}, true
	case domain.CardsPlayed:
		return Event{
			Type:         EventCardsPlayed,
			PlayerID:     string(e.PlayerID),
			Cards:        fromCards(e.Cards),
			Value:        ptr(e.Value),
			Healed:       ptr(e.Healed),
			Drawn:        ptr(e.Drawn),
			Damage:       ptr(e.Damage),
			BlockedSuits: suitsToWire(e.BlockedSuits),
		}, true
	case domain.JesterPlayed:
		return Event{Type: EventJesterPlayed, PlayerID: string(e.PlayerID)}, true
	case domain.Yielded:
		return Event{Type: EventYielded, PlayerID: string(e.PlayerID)}, true
	case domain.EnemyDefeated:
		c := FromCard(e.Enemy)
		return Event{Type: EventEnemyDefeated, Card: &c, Exact: ptr(e.Exact)}, true
	case domain.CardsDiscarded:
		return Event{Type: EventCardsDiscarded, PlayerID: string(e.PlayerID), Cards: fromCards(e.Cards), Attack: ptr(e.Attack)}, true
	case domain.AttackBlocked:
		return Event{Type: EventAttackBlocked, PlayerID: string(e.PlayerID)}, true
	case domain.NextPlayerChosen:
		return Event{Type: EventNextPlayerChosen, ByPlayerID: string(e.ByPlayerID), PlayerID: string(e.PlayerID)}, true
	case domain.JesterSignalSet:
		return Event{Type: EventSignalSet, PlayerID: string(e.PlayerID), Signal: SignalToWire(e.Signal)}, true
	case domain.SoloJesterUsed:
		return Event{Type: EventSoloJesterUsed, PlayerID: string(e.PlayerID), Drawn: ptr(e.Drawn), Left: ptr(e.Left)}, true
	case domain.GameWon:
		return Event{Type: EventGameWon, Grade: string(e.Grade)}, true
	case domain.GameLost:
		return Event{Type: EventGameLost, Reason: string(e.Reason), PlayerID: string(e.PlayerID)}, true
	case domain.GameEnded:
		return Event{Type: EventGameEnded, ByPlayerID: string(e.ByPlayerID)}, true
	case domain.RematchCreated:
		return Event{Type: EventRematchCreated, GameID: string(e.GameID)}, true
	default:
		return Event{}, false
	}
}

func ptr[T any](v T) *T { return &v }
