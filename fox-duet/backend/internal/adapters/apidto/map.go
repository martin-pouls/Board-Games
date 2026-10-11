package apidto

import (
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
)

// FromCard maps a domain card; movement and ability are derived from the rank.
func FromCard(c domain.Card) Card {
	return Card{
		Suit:     string(c.Suit),
		Rank:     int(c.Rank),
		Movement: c.Movement(),
		Ability:  string(c.Ability()),
	}
}

// ToDomain maps a wire card to a domain card; the result must be a valid card.
// Movement and ability are ignored on input.
func (c Card) ToDomain() (domain.Card, error) {
	dc := domain.NewCard(domain.Suit(c.Suit), domain.Rank(c.Rank))
	if !dc.Valid() {
		return domain.Card{}, BadRequest("invalid card (suit %q, rank %d)", c.Suit, c.Rank)
	}
	return dc, nil
}

// LevelOrDefault maps a missing level (0) to level 1; the domain validates the rest.
func LevelOrDefault(level int) int {
	if level == 0 {
		return domain.MinLevel
	}
	return level
}

// EndFromWire maps a wire end to the domain; the domain validates it.
func EndFromWire(s string) domain.End { return domain.End(s) }

func endToWire(e domain.End) string { return string(e) }

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

// FromView maps a player view. Slices are never nil, except RevealedHand (nil until the game ends).
func FromView(v domain.PlayerView) PlayerView {
	out := PlayerView{
		GameID:     string(v.GameID),
		Revision:   v.Revision,
		Phase:      string(v.Phase),
		You:        string(v.You),
		HostID:     string(v.HostID),
		Level:      v.Level,
		Players:    make([]PlayerSummary, 0, len(v.Players)),
		Hand:       fromCards(v.Hand),
		Playable:   fromCards(v.Playable),
		Round:      v.Round,
		Decree:     fromCardPtr(v.Decree),
		NextGameID: string(v.NextGameID),
	}
	for _, p := range v.Players {
		ps := PlayerSummary{
			ID:        string(p.ID),
			Name:      p.Name,
			Seat:      p.Seat,
			End:       endToWire(p.End),
			HandCount: p.HandCount,
			IsHost:    p.IsHost,
			IsYou:     p.IsYou,
			IsDealer:  p.IsDealer,
			MustAct:   p.MustAct,
		}
		if p.RevealedHand != nil {
			ps.RevealedHand = fromCards(p.RevealedHand)
		}
		out.Players = append(out.Players, ps)
	}
	if b := v.Board; b != nil {
		board := &Board{
			Side:              string(b.Side),
			Start:             b.Start,
			Spaces:            make([]Space, 0, len(b.Spaces)),
			Traveller:         b.Traveller,
			ForestTokens:      b.ForestTokens,
			ForestTokensTotal: b.ForestTokensTotal,
			GemsLeft:          b.GemsLeft,
		}
		for _, s := range b.Spaces {
			board.Spaces = append(board.Spaces, Space{Gems: s.Gems, Plus: s.Plus, Covered: s.Covered})
		}
		out.Board = board
	}
	if t := v.Trick; t != nil {
		out.Trick = &Trick{
			Number:   t.Number,
			LeaderID: string(t.LeaderID),
			Lead:     fromCardPtr(t.Lead),
			Follow:   fromCardPtr(t.Follow),
			WinnerID: string(t.WinnerID),
		}
	}
	if t := v.LastTrick; t != nil {
		out.LastTrick = &LastTrick{
			Number:    t.Number,
			LeaderID:  string(t.LeaderID),
			Lead:      FromCard(t.Lead),
			Follow:    FromCard(t.Follow),
			WinnerID:  string(t.WinnerID),
			From:      t.From,
			To:        t.To,
			Delta:     t.Delta,
			Beyond:    endToWire(t.Beyond),
			Collected: t.Collected,
		}
	}
	if p := v.Pending; p != nil {
		pending := &Pending{
			Step:        string(p.Step),
			PlayerIDs:   make([]string, 0, len(p.PlayerIDs)),
			ChosenID:    string(p.ChosenID),
			MoveOptions: make([]MoveOption, 0, len(p.MoveOptions)),
			Ends:        make([]string, 0, len(p.Ends)),
			YourGift:    fromCardPtr(p.YourGift),
		}
		for _, id := range p.PlayerIDs {
			pending.PlayerIDs = append(pending.PlayerIDs, string(id))
		}
		for _, o := range p.MoveOptions {
			pending.MoveOptions = append(pending.MoveOptions, MoveOption{
				Delta: o.Delta, To: o.To, Beyond: endToWire(o.Beyond), Gem: o.Gem, Lost: o.Lost,
			})
		}
		for _, e := range p.Ends {
			pending.Ends = append(pending.Ends, endToWire(e))
		}
		out.Pending = pending
	}
	if g := v.LastGift; g != nil {
		out.LastGift = &Gift{Gave: FromCard(g.Gave), Received: FromCard(g.Received)}
	}
	if r := v.Result; r != nil {
		res := &Result{
			Outcome:  string(r.Outcome),
			Reason:   string(r.Reason),
			PlayerID: string(r.PlayerID),
			Round:    r.Round,
			GemsLeft: r.GemsLeft,
		}
		if s := r.Score; s != nil {
			res.Score = &Score{
				Level: s.Level, Cards: s.Cards, EarlyWin: s.EarlyWin,
				ForestTokens: s.ForestTokens, Total: s.Total,
			}
		}
		out.Result = res
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
		return Event{Type: EventGameStarted, Level: ptr(e.Level)}, true
	case domain.RoundStarted:
		return Event{
			Type: EventRoundStarted, Round: ptr(e.Round), DealerID: string(e.DealerID),
			Card: fromCardPtr(&e.Decree),
		}, true
	case domain.CardPlayed:
		return Event{
			Type: EventCardPlayed, PlayerID: string(e.PlayerID), Card: fromCardPtr(&e.Card), Lead: ptr(e.Lead),
		}, true
	case domain.FoxesChosen:
		return Event{
			Type: EventFoxesChosen, ByPlayerID: string(e.ByPlayerID), PlayerID: string(e.PlayerID),
			Skipped: ptr(e.Skipped),
		}, true
	case domain.DecreeSwapped:
		return Event{
			Type: EventDecreeSwapped, PlayerID: string(e.PlayerID), Old: fromCardPtr(&e.Old), New: fromCardPtr(&e.New),
		}, true
	case domain.DecreeKept:
		return Event{Type: EventDecreeKept, PlayerID: string(e.PlayerID)}, true
	case domain.GiftChosen:
		return Event{Type: EventGiftChosen, PlayerID: string(e.PlayerID), Auto: ptr(e.Auto)}, true
	case domain.GiftExchanged:
		return Event{Type: EventGiftExchanged}, true
	case domain.TrickWon:
		return Event{Type: EventTrickWon, PlayerID: string(e.PlayerID), Number: ptr(e.Number)}, true
	case domain.TravellerMoved:
		return Event{
			Type: EventTravellerMoved, From: ptr(e.From), To: ptr(e.To), Delta: ptr(e.Delta), End: endToWire(e.Beyond),
		}, true
	case domain.ForestTokenPlaced:
		return Event{
			Type: EventForestTokenPlaced, End: endToWire(e.End), Space: ptr(e.Space), GemsMoved: ptr(e.GemsMoved),
			RoundEnd: ptr(e.RoundEnd), Left: ptr(e.Left),
		}, true
	case domain.GemCollected:
		return Event{Type: EventGemCollected, Space: ptr(e.Space), Left: ptr(e.Left)}, true
	case domain.GemsAdded:
		return Event{Type: EventGemsAdded, Count: ptr(e.Count)}, true
	case domain.GameWon:
		return Event{Type: EventGameWon, Score: ptr(e.Score)}, true
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
