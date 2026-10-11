package domain

// Event is a domain event. Events carry only public information, never hand contents or other
// private choices. Events are emitted as values.
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

type GameEnded struct{ ByPlayerID PlayerID }

func (PlayerJoined) isEvent() {}
func (PlayerLeft) isEvent()   {}
func (GameEnded) isEvent()    {}
