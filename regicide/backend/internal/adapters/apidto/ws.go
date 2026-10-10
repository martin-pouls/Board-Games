package apidto

import "encoding/json"

// Server -> client message types.
const (
	MsgView     = "view"
	MsgPresence = "presence"
	MsgAck      = "ack"
	MsgError    = "error"
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
	CmdRematch       = "rematch"
)

// ServerMessage is a server -> client WebSocket message.
// For Type "view", MarshalJSON always emits "events" (as [] when empty).
// For Type "presence", MarshalJSON always emits "online" (as [] when empty).
type ServerMessage struct {
	Type      string      `json:"type"` // view|presence|ack|error
	RequestID string      `json:"requestId,omitempty"`
	View      *PlayerView `json:"view,omitempty"`
	Events    []Event     `json:"events,omitempty"`
	Online    []string    `json:"online,omitempty"`
	Error     *ErrorBody  `json:"error,omitempty"`
}

// MarshalJSON guarantees that view messages carry "events" and presence
// messages carry "online", even when empty.
func (m ServerMessage) MarshalJSON() ([]byte, error) {
	type plain ServerMessage // no methods: avoids recursion
	switch m.Type {
	case MsgView:
		evs := m.Events
		if evs == nil {
			evs = []Event{}
		}
		return json.Marshal(struct {
			plain
			Events []Event `json:"events"`
		}{plain(m), evs})
	case MsgPresence:
		online := m.Online
		if online == nil {
			online = []string{}
		}
		return json.Marshal(struct {
			plain
			Online []string `json:"online"`
		}{plain(m), online})
	default:
		return json.Marshal(plain(m))
	}
}

// ViewMessage builds a "view" message. A nil events slice is sent as [].
func ViewMessage(v PlayerView, events []Event) ServerMessage {
	if events == nil {
		events = []Event{}
	}
	return ServerMessage{Type: MsgView, View: &v, Events: events}
}

// PresenceMessage builds a "presence" message. A nil slice is sent as [].
func PresenceMessage(online []string) ServerMessage {
	if online == nil {
		online = []string{}
	}
	return ServerMessage{Type: MsgPresence, Online: online}
}

// AckMessage builds an "ack" message.
func AckMessage(requestID string) ServerMessage {
	return ServerMessage{Type: MsgAck, RequestID: requestID}
}

// ErrorMessage builds an "error" message; requestID may be empty for malformed input.
func ErrorMessage(requestID string, body ErrorBody) ServerMessage {
	return ServerMessage{Type: MsgError, RequestID: requestID, Error: &body}
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
