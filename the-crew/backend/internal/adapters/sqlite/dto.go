package sqlite

import (
	"fmt"
	"time"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// schemaVersion is the version of gameDTO. Bump it (and migrate in fromDTO) on breaking changes.
const schemaVersion = 1

type playerDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// gameDTO is the persisted JSON snapshot of a domain.Game (pending events excluded).
type gameDTO struct {
	SchemaVersion int         `json:"schemaVersion"`
	ID            string      `json:"id"`
	Phase         string      `json:"phase"`
	HostID        string      `json:"hostId"`
	Players       []playerDTO `json:"players"`
	Revision      int64       `json:"revision"`
	CreatedAt     time.Time   `json:"createdAt"`
}

func toDTO(g *domain.Game) gameDTO {
	d := gameDTO{
		SchemaVersion: schemaVersion,
		ID:            string(g.ID),
		Phase:         string(g.Phase),
		HostID:        string(g.HostID),
		Revision:      g.Revision,
		CreatedAt:     g.CreatedAt,
	}
	if g.Players != nil {
		d.Players = make([]playerDTO, len(g.Players))
		for i, p := range g.Players {
			d.Players[i] = playerDTO{ID: string(p.ID), Name: p.Name}
		}
	}
	return d
}

func fromDTO(d gameDTO) (*domain.Game, error) {
	if d.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf("unsupported game schema version %d (want %d)", d.SchemaVersion, schemaVersion)
	}
	g := &domain.Game{
		ID:        domain.GameID(d.ID),
		Phase:     domain.Phase(d.Phase),
		HostID:    domain.PlayerID(d.HostID),
		Revision:  d.Revision,
		CreatedAt: d.CreatedAt,
	}
	if d.Players != nil {
		g.Players = make([]domain.Player, len(d.Players))
		for i, p := range d.Players {
			g.Players[i] = domain.Player{ID: domain.PlayerID(p.ID), Name: p.Name}
		}
	}
	return g, nil
}
