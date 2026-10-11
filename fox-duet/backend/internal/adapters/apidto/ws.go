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
	CmdStartGame         = "startGame"         // level? (absent / 0 = 1)
	CmdPlayCard          = "playCard"          // card (required)
	CmdChooseFoxesPlayer = "chooseFoxesPlayer" // playerId (required)
	CmdSwapDecree        = "swapDecree"        // card (required)
	CmdKeepDecree        = "keepDecree"
	CmdGiveGift          = "giveGift"        // card (required)
	CmdChooseMove        = "chooseMove"      // delta (required, may be 0)
	CmdChooseForestEnd   = "chooseForestEnd" // end: "left" | "right" (required)
	CmdGiveUp            = "giveUp"
	CmdEndGame           = "endGame"
	CmdRematch           = "rematch"
)

// ServerMessage is a server -> client WebSocket message: the shared envelope around
// Fox Duet's view and events. View messages always carry "events", presence messages "online".
type ServerMessage = wire.ServerMessage[PlayerView, Event]

// ViewMessage builds a "view" message. A nil events slice is sent as [].
func ViewMessage(v PlayerView, events []Event) ServerMessage {
	return wire.ViewMessage(v, events)
}

// ClientMessage is a client -> server WebSocket message.
type ClientMessage struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Level     int    `json:"level,omitempty"`
	Card      *Card  `json:"card,omitempty"`
	PlayerID  string `json:"playerId,omitempty"`
	Delta     *int   `json:"delta,omitempty"`
	End       string `json:"end,omitempty"`
}
