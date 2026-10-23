package apidto

import (
	"time"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
)

// FromView maps a player view. Slices are never nil.
func FromView(v domain.PlayerView) PlayerView {
	out := PlayerView{
		GameID:   string(v.GameID),
		Revision: v.Revision,
		Phase:    string(v.Phase),
		You:      string(v.You),
		HostID:   string(v.HostID),
		Players:  make([]PlayerSummary, 0, len(v.Players)),
	}
	for _, p := range v.Players {
		out.Players = append(out.Players, PlayerSummary{
			ID:     string(p.ID),
			Name:   p.Name,
			Seat:   p.Seat,
			IsHost: p.IsHost,
			IsYou:  p.IsYou,
		})
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
	case domain.GameEnded:
		return Event{Type: EventGameEnded, ByPlayerID: string(e.ByPlayerID)}, true
	default:
		return Event{}, false
	}
}
