package sqlite

import (
	"fmt"
	"slices"
	"time"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
)

// schemaVersion is the version of gameDTO. Bump it (and migrate in fromDTO) on breaking changes.
const schemaVersion = 1

type cardDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type playerDTO struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
	Hand []cardDTO `json:"hand"`
}

type boardDTO struct {
	Gems         []int `json:"gems"`
	CoveredLeft  int   `json:"coveredLeft"`
	CoveredRight int   `json:"coveredRight"`
	Traveller    int   `json:"traveller"`
	ForestTokens int   `json:"forestTokens"`
}

type trickDTO struct {
	Number int         `json:"number"`
	Leader int         `json:"leader"`
	Lead   *cardDTO    `json:"lead"`
	Follow *cardDTO    `json:"follow"`
	Chosen int         `json:"chosen"`
	Gift   [2]*cardDTO `json:"gift"`
}

type trickRecordDTO struct {
	Number    int     `json:"number"`
	Leader    int     `json:"leader"`
	Lead      cardDTO `json:"lead"`
	Follow    cardDTO `json:"follow"`
	Winner    int     `json:"winner"`
	From      int     `json:"from"`
	Delta     int     `json:"delta"`
	To        int     `json:"to"`
	Beyond    string  `json:"beyond"`
	Collected bool    `json:"collected"`
}

type giftRecordDTO struct {
	Given [2]cardDTO `json:"given"`
}

type resultDTO struct {
	Reason   string `json:"reason"`
	PlayerID string `json:"playerId"`
}

// gameDTO is the persisted JSON snapshot of a domain.Game (pending events excluded).
type gameDTO struct {
	SchemaVersion int             `json:"schemaVersion"`
	ID            string          `json:"id"`
	Phase         string          `json:"phase"`
	Step          string          `json:"step"`
	HostID        string          `json:"hostId"`
	Players       []playerDTO     `json:"players"`
	Level         int             `json:"level"`
	Board         boardDTO        `json:"board"`
	Round         int             `json:"round"`
	Dealer        int             `json:"dealer"`
	Deck          []cardDTO       `json:"deck"`
	Decree        *cardDTO        `json:"decree"`
	Trick         trickDTO        `json:"trick"`
	LastTrick     *trickRecordDTO `json:"lastTrick"`
	LastGift      *giftRecordDTO  `json:"lastGift"`
	Result        *resultDTO      `json:"result"`
	NextGameID    string          `json:"nextGameId"`
	Revision      int64           `json:"revision"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func cardToDTO(c domain.Card) cardDTO {
	return cardDTO{Suit: string(c.Suit), Rank: int(c.Rank)}
}

func cardFromDTO(c cardDTO) domain.Card {
	return domain.Card{Suit: domain.Suit(c.Suit), Rank: domain.Rank(c.Rank)}
}

// cardsToDTO keeps nil slices nil and empty slices empty, so snapshots round-trip exactly.
func cardsToDTO(cs []domain.Card) []cardDTO {
	if cs == nil {
		return nil
	}
	out := make([]cardDTO, len(cs))
	for i, c := range cs {
		out[i] = cardToDTO(c)
	}
	return out
}

func cardsFromDTO(cs []cardDTO) []domain.Card {
	if cs == nil {
		return nil
	}
	out := make([]domain.Card, len(cs))
	for i, c := range cs {
		out[i] = cardFromDTO(c)
	}
	return out
}

func cardPtrToDTO(c *domain.Card) *cardDTO {
	if c == nil {
		return nil
	}
	d := cardToDTO(*c)
	return &d
}

func cardPtrFromDTO(c *cardDTO) *domain.Card {
	if c == nil {
		return nil
	}
	d := cardFromDTO(*c)
	return &d
}

func toDTO(g *domain.Game) gameDTO {
	d := gameDTO{
		SchemaVersion: schemaVersion,
		ID:            string(g.ID),
		Phase:         string(g.Phase),
		Step:          string(g.Step),
		HostID:        string(g.HostID),
		Level:         g.Level,
		Board: boardDTO{
			Gems:         slices.Clone(g.Board.Gems),
			CoveredLeft:  g.Board.CoveredLeft,
			CoveredRight: g.Board.CoveredRight,
			Traveller:    g.Board.Traveller,
			ForestTokens: g.Board.ForestTokens,
		},
		Round:  g.Round,
		Dealer: g.Dealer,
		Deck:   cardsToDTO(g.Deck),
		Decree: cardPtrToDTO(g.Decree),
		Trick: trickDTO{
			Number: g.Trick.Number,
			Leader: g.Trick.Leader,
			Lead:   cardPtrToDTO(g.Trick.Lead),
			Follow: cardPtrToDTO(g.Trick.Follow),
			Chosen: g.Trick.Chosen,
			Gift:   [2]*cardDTO{cardPtrToDTO(g.Trick.Gift[0]), cardPtrToDTO(g.Trick.Gift[1])},
		},
		NextGameID: string(g.NextGameID),
		Revision:   g.Revision,
		CreatedAt:  g.CreatedAt,
	}
	if g.Players != nil {
		d.Players = make([]playerDTO, len(g.Players))
		for i, p := range g.Players {
			d.Players[i] = playerDTO{ID: string(p.ID), Name: p.Name, Hand: cardsToDTO(p.Hand)}
		}
	}
	if t := g.LastTrick; t != nil {
		d.LastTrick = &trickRecordDTO{
			Number: t.Number, Leader: t.Leader, Lead: cardToDTO(t.Lead), Follow: cardToDTO(t.Follow),
			Winner: t.Winner, From: t.From, Delta: t.Delta, To: t.To, Beyond: string(t.Beyond), Collected: t.Collected,
		}
	}
	if g.LastGift != nil {
		d.LastGift = &giftRecordDTO{Given: [2]cardDTO{cardToDTO(g.LastGift.Given[0]), cardToDTO(g.LastGift.Given[1])}}
	}
	if g.Result != nil {
		d.Result = &resultDTO{Reason: string(g.Result.Reason), PlayerID: string(g.Result.PlayerID)}
	}
	return d
}

func fromDTO(d gameDTO) (*domain.Game, error) {
	if d.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf("unsupported game schema version %d (want %d)", d.SchemaVersion, schemaVersion)
	}
	g := &domain.Game{
		ID:     domain.GameID(d.ID),
		Phase:  domain.Phase(d.Phase),
		Step:   domain.Step(d.Step),
		HostID: domain.PlayerID(d.HostID),
		Level:  d.Level,
		Board: domain.Board{
			Gems:         slices.Clone(d.Board.Gems),
			CoveredLeft:  d.Board.CoveredLeft,
			CoveredRight: d.Board.CoveredRight,
			Traveller:    d.Board.Traveller,
			ForestTokens: d.Board.ForestTokens,
		},
		Round:  d.Round,
		Dealer: d.Dealer,
		Deck:   cardsFromDTO(d.Deck),
		Decree: cardPtrFromDTO(d.Decree),
		Trick: domain.Trick{
			Number: d.Trick.Number,
			Leader: d.Trick.Leader,
			Lead:   cardPtrFromDTO(d.Trick.Lead),
			Follow: cardPtrFromDTO(d.Trick.Follow),
			Chosen: d.Trick.Chosen,
			Gift:   [2]*domain.Card{cardPtrFromDTO(d.Trick.Gift[0]), cardPtrFromDTO(d.Trick.Gift[1])},
		},
		NextGameID: domain.GameID(d.NextGameID),
		Revision:   d.Revision,
		CreatedAt:  d.CreatedAt,
	}
	if d.Players != nil {
		g.Players = make([]domain.Player, len(d.Players))
		for i, p := range d.Players {
			g.Players[i] = domain.Player{ID: domain.PlayerID(p.ID), Name: p.Name, Hand: cardsFromDTO(p.Hand)}
		}
	}
	if t := d.LastTrick; t != nil {
		g.LastTrick = &domain.TrickRecord{
			Number: t.Number, Leader: t.Leader, Lead: cardFromDTO(t.Lead), Follow: cardFromDTO(t.Follow),
			Winner: t.Winner, From: t.From, Delta: t.Delta, To: t.To, Beyond: domain.End(t.Beyond), Collected: t.Collected,
		}
	}
	if d.LastGift != nil {
		g.LastGift = &domain.GiftRecord{Given: [2]domain.Card{cardFromDTO(d.LastGift.Given[0]), cardFromDTO(d.LastGift.Given[1])}}
	}
	if d.Result != nil {
		g.Result = &domain.Result{Reason: domain.LossReason(d.Result.Reason), PlayerID: domain.PlayerID(d.Result.PlayerID)}
	}
	return g, nil
}
