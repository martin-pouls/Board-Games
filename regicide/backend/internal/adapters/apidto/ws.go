package apidto

import "github.com/martin-pouls/shared/backend/wire"

// Server -> client message types (shared envelope).
const (
	MsgView     = wire.MsgView
	MsgPresence = wire.MsgPresence
	MsgAck      = wire.MsgAck
	MsgError    = wire.MsgError
)

// Client -> server message types.
const (
	CmdStartGame     = "startGame" // firstPlayerId? ("" / absent = random)
	CmdPlay          = "play"      // cards (required, non-null)
	CmdYield         = "yield"
	CmdDiscard       = "discard"    // cards (required, non-null)
	CmdChooseNext    = "chooseNext" // playerId (required)
	CmdUseSoloJester = "useSoloJester"
	CmdSetSignal     = "setSignal" // signal: "none" | "goodPlay" | "ratherNot" (required)
	CmdGiveUp        = "giveUp"
	CmdEndGame       = "endGame"
	CmdRematch       = "rematch"
)

// ServerMessage is a server -> client WebSocket message: the shared envelope around
// Regicide's view and events. View messages always carry "events", presence messages "online".
type ServerMessage = wire.ServerMessage[PlayerView, Event]

// ViewMessage builds a "view" message. A nil events slice is sent as [].
func ViewMessage(v PlayerView, events []Event) ServerMessage {
	return wire.ViewMessage(v, events)
}

// ClientMessage is a client -> server WebSocket message.
type ClientMessage struct {
	Type          string  `json:"type"`
	RequestID     string  `json:"requestId"`
	FirstPlayerID string  `json:"firstPlayerId,omitempty"`
	Cards         []Card  `json:"cards,omitempty"`
	PlayerID      string  `json:"playerId,omitempty"`
	Signal        *string `json:"signal,omitempty"`
}
