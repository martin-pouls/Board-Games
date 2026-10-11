// Package apidto is the JSON wire contract shared by the REST (httpapi) and
// WebSocket (ws) adapters: tagged DTOs, mappers from domain types, and the
// error mapping. All JSON is camelCase and slices are never null.
// Keep in sync with frontend/src/lib/types.ts.
package apidto

import "github.com/martin-pouls/shared/backend/wire"

// PlayerSummary is public information about a seated player.
type PlayerSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Seat   int    `json:"seat"`
	IsHost bool   `json:"isHost"`
	IsYou  bool   `json:"isYou"`
}

// PlayerView is the per-player projection.
type PlayerView struct {
	GameID   string          `json:"gameId"`
	Revision int64           `json:"revision"`
	Phase    string          `json:"phase"` // "lobby" | "playing" | "debrief" | "ended"
	You      string          `json:"you"`
	HostID   string          `json:"hostId"`
	Players  []PlayerSummary `json:"players"` // never null
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      string `json:"gameId"`
	Phase       string `json:"phase"`
	HostName    string `json:"hostName"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	Joinable    bool   `json:"joinable"`
	CreatedAt   string `json:"createdAt,omitempty"` // RFC3339 UTC; omitted when unknown
}

// Event type strings.
const (
	EventPlayerJoined = "playerJoined"
	EventPlayerLeft   = "playerLeft"
	EventGameEnded    = "gameEnded"
)

// Event is a flattened union of domain events; Type is one of the Event* constants.
// Pointer fields are set exactly when the event carries them, so 0 and false survive omitempty.
type Event struct {
	Type       string `json:"type"`
	PlayerID   string `json:"playerId,omitempty"`
	ByPlayerID string `json:"byPlayerId,omitempty"`
	HostID     string `json:"hostId,omitempty"`
	Name       string `json:"name,omitempty"`
}

// CreateGameRequest is the body of POST /api/games.
type CreateGameRequest struct {
	Name string `json:"name"`
}

// JoinGameRequest is the body of POST /api/games/{code}/players.
type JoinGameRequest struct {
	Name string `json:"name"`
}

// SessionResponse is returned when creating or joining a game.
type SessionResponse struct {
	GameID   string     `json:"gameId"`
	PlayerID string     `json:"playerId"`
	Token    string     `json:"token"`
	View     PlayerView `json:"view"`
}

// ErrorBody is the error payload used on REST and WS (shared envelope).
type ErrorBody = wire.ErrorBody

// ErrorResponse is the body of every non-2xx REST response (shared envelope).
type ErrorResponse = wire.ErrorResponse
