package apidto

import "github.com/martin-pouls/shared/backend/wire"

// Server -> client message types (shared envelope).
const (
	MsgView     = wire.MsgView
	MsgPresence = wire.MsgPresence
	MsgAck      = wire.MsgAck
	MsgError    = wire.MsgError
)

// Client -> server message types. Gameplay runs over the WebSocket only; REST covers the lobby.
const (
	CmdEndGame = "endGame"
)

// ServerMessage is a server -> client WebSocket message: the shared envelope around
// The Crew's view and events. View messages always carry "events", presence messages "online".
type ServerMessage = wire.ServerMessage[PlayerView, Event]

// ViewMessage builds a "view" message. A nil events slice is sent as [].
func ViewMessage(v PlayerView, events []Event) ServerMessage {
	return wire.ViewMessage(v, events)
}

// ClientMessage is a client -> server WebSocket message. Commands added later (mission, card,
// seats, ...) extend it with their fields.
type ClientMessage struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
}
