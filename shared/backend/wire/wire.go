// Package wire is the JSON envelope every game speaks over REST and WebSocket. The game's own
// view, event and command shapes go inside it. The frontend mirror is shared/frontend/src/wire.ts.
package wire

import "encoding/json"

// Error codes every game uses. Games add their own (domain errors) next to these.
const (
	CodeBadRequest   = "bad_request"
	CodeUnauthorized = "unauthorized"
	CodeGameNotFound = "game_not_found"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeRateLimited  = "rate_limited"
	CodeInternal     = "internal"
)

// InternalMessage is the message sent for unmapped errors (log the real error instead).
const InternalMessage = "internal error"

// ErrorBody is the error payload of a REST error response and of a WS "error" message.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the body of every REST error response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// Server -> client WebSocket message types.
const (
	MsgView     = "view"
	MsgPresence = "presence"
	MsgAck      = "ack"
	MsgError    = "error"
)

// ServerMessage is a server -> client WebSocket message carrying the game's view V and events E.
// For Type "view", MarshalJSON always emits "events" (as [] when empty).
// For Type "presence", MarshalJSON always emits "online" (as [] when empty).
type ServerMessage[V, E any] struct {
	Type      string     `json:"type"` // view|presence|ack|error
	RequestID string     `json:"requestId,omitempty"`
	View      *V         `json:"view,omitempty"`
	Events    []E        `json:"events,omitempty"`
	Online    []string   `json:"online,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
}

// MarshalJSON guarantees that view messages carry "events" and presence
// messages carry "online", even when empty.
func (m ServerMessage[V, E]) MarshalJSON() ([]byte, error) {
	type plain ServerMessage[V, E] // no methods: avoids recursion
	switch m.Type {
	case MsgView:
		evs := m.Events
		if evs == nil {
			evs = []E{}
		}
		return json.Marshal(struct {
			plain
			Events []E `json:"events"`
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
func ViewMessage[V, E any](v V, events []E) ServerMessage[V, E] {
	if events == nil {
		events = []E{}
	}
	return ServerMessage[V, E]{Type: MsgView, View: &v, Events: events}
}

// The messages below carry no view or events; they marshal the same for any V and E.
type bare = ServerMessage[struct{}, struct{}]

// PresenceMessage returns the marshalled "presence" message. A nil slice is sent as [].
func PresenceMessage(online []string) ([]byte, error) {
	if online == nil {
		online = []string{}
	}
	return json.Marshal(bare{Type: MsgPresence, Online: online})
}

// AckMessage returns the marshalled "ack" message.
func AckMessage(requestID string) ([]byte, error) {
	return json.Marshal(bare{Type: MsgAck, RequestID: requestID})
}

// ErrorMessage returns the marshalled "error" message; requestID may be empty for malformed input.
func ErrorMessage(requestID string, body ErrorBody) ([]byte, error) {
	return json.Marshal(bare{Type: MsgError, RequestID: requestID, Error: &body})
}

// Command is the part of every client -> server WebSocket message the transport reads: the
// command type and the request ID echoed in the ack or error. Games decode the rest.
type Command struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
}
