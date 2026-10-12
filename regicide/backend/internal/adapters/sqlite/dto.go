package sqlite

import (
	"fmt"
	"time"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
)

// schemaVersion is the version of gameDTO. Bump it (and migrate in fromDTO) on breaking changes.
const schemaVersion = 1

type cardDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type playerDTO struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Hand    []cardDTO `json:"hand"`
	Yielded bool      `json:"yielded"`
	Signal  string    `json:"signal"`
}

type playedDTO struct {
	By    string    `json:"by"`
	Cards []cardDTO `json:"cards"`
}

type lastTurnDTO struct {
	By        string    `json:"by"`
	Played    []cardDTO `json:"played"`
	Defeated  *cardDTO  `json:"defeated,omitempty"`
	Exact     bool      `json:"exact,omitempty"`
	Discarded []cardDTO `json:"discarded"`
}

type turnDTO struct {
	Number     int `json:"number"`
	ActiveSeat int `json:"activeSeat"`
}

type resultDTO struct {
	Reason   string `json:"reason"`
	PlayerID string `json:"playerId"`
}

// gameDTO is the persisted JSON snapshot of a domain.Game (pending events excluded).
type gameDTO struct {
	SchemaVersion   int          `json:"schemaVersion"`
	ID              string       `json:"id"`
	Phase           string       `json:"phase"`
	Step            string       `json:"step"`
	HostID          string       `json:"hostId"`
	Players         []playerDTO  `json:"players"`
	Castle          []cardDTO    `json:"castle"`
	Enemy           *cardDTO     `json:"enemy"`
	Tavern          []cardDTO    `json:"tavern"`
	Discard         []cardDTO    `json:"discard"`
	Played          []playedDTO  `json:"played"`
	LastTurn        *lastTurnDTO `json:"lastTurn,omitempty"` // optional: older snapshots have none
	Turn            turnDTO      `json:"turn"`
	SoloJestersUsed int          `json:"soloJestersUsed"`
	Result          *resultDTO   `json:"result"`
	NextGameID      string       `json:"nextGameId"`
	Revision        int64        `json:"revision"`
	CreatedAt       time.Time    `json:"createdAt"`
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

func toDTO(g *domain.Game) gameDTO {
	d := gameDTO{
		SchemaVersion:   schemaVersion,
		ID:              string(g.ID),
		Phase:           string(g.Phase),
		Step:            string(g.Step),
		HostID:          string(g.HostID),
		Castle:          cardsToDTO(g.Castle),
		Tavern:          cardsToDTO(g.Tavern),
		Discard:         cardsToDTO(g.DiscardPile),
		Turn:            turnDTO{Number: g.Turn.Number, ActiveSeat: g.Turn.ActiveSeat},
		SoloJestersUsed: g.SoloJestersUsed,
		NextGameID:      string(g.NextGameID),
		Revision:        g.Revision,
		CreatedAt:       g.CreatedAt,
	}
	if g.Players != nil {
		d.Players = make([]playerDTO, len(g.Players))
		for i, p := range g.Players {
			d.Players[i] = playerDTO{
				ID: string(p.ID), Name: p.Name, Hand: cardsToDTO(p.Hand),
				Yielded: p.YieldedLastTurn, Signal: string(p.Signal),
			}
		}
	}
	if g.Enemy != nil {
		e := cardToDTO(*g.Enemy)
		d.Enemy = &e
	}
	if g.Played != nil {
		d.Played = make([]playedDTO, len(g.Played))
		for i, ps := range g.Played {
			d.Played[i] = playedDTO{By: string(ps.By), Cards: cardsToDTO(ps.Cards)}
		}
	}
	if lt := g.LastTurn; lt != nil {
		d.LastTurn = &lastTurnDTO{By: string(lt.By), Played: cardsToDTO(lt.Played), Discarded: cardsToDTO(lt.Discarded)}
		if lt.Defeated != nil {
			e := cardToDTO(*lt.Defeated)
			d.LastTurn.Defeated = &e
			d.LastTurn.Exact = lt.Exact
		}
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
		ID:              domain.GameID(d.ID),
		Phase:           domain.Phase(d.Phase),
		Step:            domain.Step(d.Step),
		HostID:          domain.PlayerID(d.HostID),
		Castle:          cardsFromDTO(d.Castle),
		Tavern:          cardsFromDTO(d.Tavern),
		DiscardPile:     cardsFromDTO(d.Discard),
		Turn:            domain.Turn{Number: d.Turn.Number, ActiveSeat: d.Turn.ActiveSeat},
		SoloJestersUsed: d.SoloJestersUsed,
		NextGameID:      domain.GameID(d.NextGameID),
		Revision:        d.Revision,
		CreatedAt:       d.CreatedAt,
	}
	if d.Players != nil {
		g.Players = make([]domain.Player, len(d.Players))
		for i, p := range d.Players {
			g.Players[i] = domain.Player{
				ID: domain.PlayerID(p.ID), Name: p.Name, Hand: cardsFromDTO(p.Hand),
				YieldedLastTurn: p.Yielded, Signal: domain.JesterSignal(p.Signal),
			}
		}
	}
	if d.Enemy != nil {
		e := cardFromDTO(*d.Enemy)
		g.Enemy = &e
	}
	if d.Played != nil {
		g.Played = make([]domain.PlayedSet, len(d.Played))
		for i, ps := range d.Played {
			g.Played[i] = domain.PlayedSet{By: domain.PlayerID(ps.By), Cards: cardsFromDTO(ps.Cards)}
		}
	}
	if lt := d.LastTurn; lt != nil {
		g.LastTurn = &domain.LastTurn{By: domain.PlayerID(lt.By), Played: cardsFromDTO(lt.Played), Discarded: cardsFromDTO(lt.Discarded)}
		if lt.Defeated != nil {
			e := cardFromDTO(*lt.Defeated)
			g.LastTurn.Defeated = &e
			g.LastTurn.Exact = lt.Exact
		}
	}
	if d.Result != nil {
		g.Result = &domain.Result{Reason: domain.LossReason(d.Result.Reason), PlayerID: domain.PlayerID(d.Result.PlayerID)}
	}
	return g, nil
}
