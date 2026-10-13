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
// The comments list the required fields of each command.
const (
	CmdSetOptions     = "setOptions"     // options
	CmdSelectMission  = "selectMission"  // mission
	CmdStartMission   = "startMission"   //
	CmdAdjustTokens   = "adjustTokens"   // from, to
	CmdSkipTokens     = "skipTokens"     //
	CmdAnswerPoll     = "answerPoll"     // yes
	CmdPick           = "pick"           // seats (one seat, two for the roles of mission 50)
	CmdPickTask       = "pickTask"       // taskId
	CmdSendDistress   = "sendDistress"   // direction: "left" | "right"
	CmdSkipDistress   = "skipDistress"   //
	CmdChoosePassCard = "choosePassCard" // card
	CmdHandover       = "handover"       // taskId, seat
	CmdCommunicate    = "communicate"    // card, position (may be empty in a dead zone)
	CmdPlayCard       = "playCard"       // card
	CmdGiveUp         = "giveUp"         //
	CmdNextAttempt    = "nextAttempt"    //
	CmdNextMission    = "nextMission"    //
	CmdBackToLobby    = "backToLobby"    //
	CmdEndGame        = "endGame"        //
)

// ServerMessage is a server -> client WebSocket message: the shared envelope around
// The Crew's view and events. View messages always carry "events", presence messages "online".
type ServerMessage = wire.ServerMessage[PlayerView, Event]

// ViewMessage builds a "view" message. A nil events slice is sent as [].
func ViewMessage(v PlayerView, events []Event) ServerMessage {
	return wire.ViewMessage(v, events)
}

// ClientMessage is a client -> server WebSocket message. Pointer fields tell "absent" from 0 and
// false; a missing required field is bad_request.
type ClientMessage struct {
	Type      string       `json:"type"`
	RequestID string       `json:"requestId"`
	Options   *GameOptions `json:"options,omitempty"`
	Mission   *int         `json:"mission,omitempty"`
	Seats     []int        `json:"seats,omitempty"`
	TaskID    *int         `json:"taskId,omitempty"`
	Card      *Card        `json:"card,omitempty"`
	Position  string       `json:"position,omitempty"`
	Direction string       `json:"direction,omitempty"`
	Yes       *bool        `json:"yes,omitempty"`
	From      *int         `json:"from,omitempty"`
	To        *int         `json:"to,omitempty"`
	Seat      *int         `json:"seat,omitempty"`
}
