package domain

import (
	"fmt"
	"time"
)

// PlayerView is the projection of a game for one player.
type PlayerView struct {
	GameID      GameID
	Revision    int64
	Phase       Phase
	You, HostID PlayerID
	Players     []PlayerSummary // seat order
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID            PlayerID
	Name          string
	Seat          int
	IsHost, IsYou bool
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
	if _, _, ok := g.PlayerByID(id); !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	v := PlayerView{
		GameID:   g.ID,
		Revision: g.Revision,
		Phase:    g.Phase,
		You:      id,
		HostID:   g.HostID,
		Players:  make([]PlayerSummary, len(g.Players)),
	}
	for i, p := range g.Players {
		v.Players[i] = PlayerSummary{
			ID:     p.ID,
			Name:   p.Name,
			Seat:   i,
			IsHost: p.ID == g.HostID,
			IsYou:  p.ID == id,
		}
	}
	return v, nil
}
