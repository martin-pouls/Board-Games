package sqlite

import (
	"fmt"
	"maps"
	"time"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
)

// schemaVersion is the version of gameDTO. Bump it (and migrate in fromDTO) on breaking changes.
const schemaVersion = 1

type cardDTO struct {
	Value   int    `json:"value"`
	Kind    string `json:"kind"`
	Command string `json:"command"`
}

type optionsDTO struct {
	Variant         string `json:"variant"`
	Pro             bool   `json:"pro"`
	ReducedHandSize bool   `json:"reducedHandSize"`
}

type rulesDTO struct {
	Variant               string      `json:"variant"`
	MinPlayers            int         `json:"minPlayers"`
	MaxPlayers            int         `json:"maxPlayers"`
	MinCardsPerTurn       int         `json:"minCardsPerTurn"`
	MinCardsDrawPileEmpty int         `json:"minCardsDrawPileEmpty"`
	HandSizes             map[int]int `json:"handSizes"`
	HandSizeReduction     int         `json:"handSizeReduction"`
	BackwardsTrickStep    int         `json:"backwardsTrickStep"`
	AscendingStart        int         `json:"ascendingStart"`
	DescendingStart       int         `json:"descendingStart"`
	LowestCard            int         `json:"lowestCard"`
	HighestCard           int         `json:"highestCard"`
	AscendingPiles        int         `json:"ascendingPiles"`
	DescendingPiles       int         `json:"descendingPiles"`
}

type playerDTO struct {
	ID   string    `json:"id"`
	Name string    `json:"name"`
	Hand []cardDTO `json:"hand"`
}

type pileDTO struct {
	Direction  string    `json:"direction"`
	StartValue int       `json:"startValue"`
	Cards      []cardDTO `json:"cards"`
}

type playDTO struct {
	Card cardDTO `json:"card"`
	Pile int     `json:"pile"`
}

type turnDTO struct {
	Number        int       `json:"number"`
	ActiveSeat    int       `json:"activeSeat"`
	Plays         []playDTO `json:"plays"`
	StartHandSize int       `json:"startHandSize"`
}

type resultDTO struct {
	Score               int    `json:"score"`
	CardsLeftInHands    int    `json:"cardsLeftInHands"`
	CardsLeftInDrawPile int    `json:"cardsLeftInDrawPile"`
	Reason              string `json:"reason,omitempty"`
	PlayerID            string `json:"playerId,omitempty"`

	// Read only, from snapshots written while a game over had phase "finished" plus an outcome.
	LegacyOutcome       string `json:"outcome,omitempty"`
	LegacyStuckPlayerID string `json:"stuckPlayerId,omitempty"`
}

// legacyPhaseFinished is the single phase old snapshots used for every game over.
const legacyPhaseFinished = "finished"

// migrateFinished maps an old "finished" snapshot onto the won / lost phases: outcome won → won,
// lost → lost (stuck), aborted (ended early by the host) → lost (given up, by nobody in particular).
// Other snapshots are returned unchanged.
func migrateFinished(d gameDTO) gameDTO {
	if d.Phase != legacyPhaseFinished {
		return d
	}
	r := resultDTO{}
	if d.Result != nil {
		r = *d.Result
	}
	switch r.LegacyOutcome {
	case "won":
		d.Phase = string(domain.PhaseWon)
		r.Reason, r.PlayerID = "", ""
	case "aborted":
		d.Phase = string(domain.PhaseLost)
		r.Reason, r.PlayerID = string(domain.LossGaveUp), ""
	default: // "lost"
		d.Phase = string(domain.PhaseLost)
		r.Reason, r.PlayerID = string(domain.LossStuck), r.LegacyStuckPlayerID
	}
	r.LegacyOutcome, r.LegacyStuckPlayerID = "", ""
	d.Result = &r
	return d
}

// gameDTO is the persisted JSON snapshot of a domain.Game (pending events excluded).
type gameDTO struct {
	SchemaVersion int         `json:"schemaVersion"`
	ID            string      `json:"id"`
	Options       optionsDTO  `json:"options"`
	Rules         rulesDTO    `json:"rules"`
	Phase         string      `json:"phase"`
	HostID        string      `json:"hostId"`
	Players       []playerDTO `json:"players"`
	DrawPile      []cardDTO   `json:"drawPile"`
	Piles         []pileDTO   `json:"piles"`
	Turn          turnDTO     `json:"turn"`
	Result        *resultDTO  `json:"result"`
	NextGameID    string      `json:"nextGameId"`
	Revision      int64       `json:"revision"`
	CreatedAt     time.Time   `json:"createdAt"` // absent in old snapshots: zero
}

func cardToDTO(c domain.Card) cardDTO {
	return cardDTO{Value: c.Value, Kind: string(c.Kind), Command: string(c.Command)}
}

func cardFromDTO(c cardDTO) domain.Card {
	return domain.Card{Value: c.Value, Kind: domain.CardKind(c.Kind), Command: domain.Command(c.Command)}
}

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
	r := g.Rules
	d := gameDTO{
		SchemaVersion: schemaVersion,
		ID:            string(g.ID),
		Options:       optionsDTO{Variant: string(g.Options.Variant), Pro: g.Options.Pro, ReducedHandSize: g.Options.ReducedHandSize},
		Rules: rulesDTO{
			Variant: string(r.Variant), MinPlayers: r.MinPlayers, MaxPlayers: r.MaxPlayers,
			MinCardsPerTurn: r.MinCardsPerTurn, MinCardsDrawPileEmpty: r.MinCardsDrawPileEmpty,
			HandSizes:         maps.Clone(r.HandSizes),
			HandSizeReduction: r.HandSizeReduction, BackwardsTrickStep: r.BackwardsTrickStep,
			AscendingStart: r.AscendingStart, DescendingStart: r.DescendingStart,
			LowestCard: r.LowestCard, HighestCard: r.HighestCard,
			AscendingPiles: r.AscendingPiles, DescendingPiles: r.DescendingPiles,
		},
		Phase:      string(g.Phase),
		HostID:     string(g.HostID),
		DrawPile:   cardsToDTO(g.DrawPile),
		NextGameID: string(g.NextGameID),
		Revision:   g.Revision,
		CreatedAt:  g.CreatedAt,
		Turn:       turnDTO{Number: g.Turn.Number, ActiveSeat: g.Turn.ActiveSeat, StartHandSize: g.Turn.StartHandSize},
	}
	if g.Players != nil {
		d.Players = make([]playerDTO, len(g.Players))
		for i, p := range g.Players {
			d.Players[i] = playerDTO{ID: string(p.ID), Name: p.Name, Hand: cardsToDTO(p.Hand)}
		}
	}
	if g.Piles != nil {
		d.Piles = make([]pileDTO, len(g.Piles))
		for i, p := range g.Piles {
			d.Piles[i] = pileDTO{Direction: string(p.Direction), StartValue: p.StartValue, Cards: cardsToDTO(p.Cards)}
		}
	}
	if g.Turn.Plays != nil {
		d.Turn.Plays = make([]playDTO, len(g.Turn.Plays))
		for i, p := range g.Turn.Plays {
			d.Turn.Plays[i] = playDTO{Card: cardToDTO(p.Card), Pile: p.Pile}
		}
	}
	if g.Result != nil {
		d.Result = &resultDTO{
			Score:            g.Result.Score,
			CardsLeftInHands: g.Result.CardsLeftInHands, CardsLeftInDrawPile: g.Result.CardsLeftInDrawPile,
			Reason: string(g.Result.Reason), PlayerID: string(g.Result.PlayerID),
		}
	}
	return d
}

func fromDTO(d gameDTO) (*domain.Game, error) {
	if d.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf("unsupported game schema version %d (want %d)", d.SchemaVersion, schemaVersion)
	}
	d = migrateFinished(d)
	r := d.Rules
	g := &domain.Game{
		ID:      domain.GameID(d.ID),
		Options: domain.Options{Variant: domain.Variant(d.Options.Variant), Pro: d.Options.Pro, ReducedHandSize: d.Options.ReducedHandSize},
		Rules: domain.Ruleset{
			Variant: domain.Variant(r.Variant), MinPlayers: r.MinPlayers, MaxPlayers: r.MaxPlayers,
			MinCardsPerTurn: r.MinCardsPerTurn, MinCardsDrawPileEmpty: r.MinCardsDrawPileEmpty,
			HandSizes:         maps.Clone(r.HandSizes),
			HandSizeReduction: r.HandSizeReduction, BackwardsTrickStep: r.BackwardsTrickStep,
			AscendingStart: r.AscendingStart, DescendingStart: r.DescendingStart,
			LowestCard: r.LowestCard, HighestCard: r.HighestCard,
			AscendingPiles: r.AscendingPiles, DescendingPiles: r.DescendingPiles,
		},
		Phase:      domain.Phase(d.Phase),
		HostID:     domain.PlayerID(d.HostID),
		DrawPile:   cardsFromDTO(d.DrawPile),
		NextGameID: domain.GameID(d.NextGameID),
		Revision:   d.Revision,
		CreatedAt:  d.CreatedAt,
		Turn:       domain.Turn{Number: d.Turn.Number, ActiveSeat: d.Turn.ActiveSeat, StartHandSize: d.Turn.StartHandSize},
	}
	if d.Players != nil {
		g.Players = make([]domain.Player, len(d.Players))
		for i, p := range d.Players {
			g.Players[i] = domain.Player{ID: domain.PlayerID(p.ID), Name: p.Name, Hand: cardsFromDTO(p.Hand)}
		}
	}
	if d.Piles != nil {
		g.Piles = make([]domain.Pile, len(d.Piles))
		for i, p := range d.Piles {
			g.Piles[i] = domain.Pile{Direction: domain.Direction(p.Direction), StartValue: p.StartValue, Cards: cardsFromDTO(p.Cards)}
		}
	}
	if d.Turn.Plays != nil {
		g.Turn.Plays = make([]domain.Play, len(d.Turn.Plays))
		for i, p := range d.Turn.Plays {
			g.Turn.Plays[i] = domain.Play{Card: cardFromDTO(p.Card), Pile: p.Pile}
		}
	}
	if d.Result != nil {
		g.Result = &domain.Result{
			Score:            d.Result.Score,
			CardsLeftInHands: d.Result.CardsLeftInHands, CardsLeftInDrawPile: d.Result.CardsLeftInDrawPile,
			Reason: domain.LossReason(d.Result.Reason), PlayerID: domain.PlayerID(d.Result.PlayerID),
		}
	}
	return g, nil
}
