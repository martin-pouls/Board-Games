package domain

// Event is a domain event. Events carry only public information, never hand contents.
// Events are emitted as values (not pointers).
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

type OptionsChanged struct{ Options Options }

type GameStarted struct{}

type RematchCreated struct{ GameID GameID }

type FirstPlayerChosen struct{ PlayerID PlayerID }

type TurnStarted struct {
	PlayerID PlayerID
	Number   int
}

type CardPlayed struct {
	PlayerID PlayerID
	Card     Card
	Pile     int
}

// PlayUndone: the active player took the last card of this turn back from Pile.
type PlayUndone struct {
	PlayerID PlayerID
	Card     Card
	Pile     int
}

type TurnEnded struct {
	PlayerID   PlayerID
	CardsDrawn int
}

type PlayerSkipped struct{ PlayerID PlayerID }

// GameWon: every card was played (score 0).
type GameWon struct{}

// GameLost: PlayerID is the stuck player or the one who gave up; Score counts the cards left.
type GameLost struct {
	Reason   LossReason
	PlayerID PlayerID
	Score    int
}

// GameEnded: the host closed the lobby before the game started.
type GameEnded struct{ ByPlayerID PlayerID }

func (PlayerJoined) isEvent()      {}
func (PlayerLeft) isEvent()        {}
func (OptionsChanged) isEvent()    {}
func (GameStarted) isEvent()       {}
func (FirstPlayerChosen) isEvent() {}
func (TurnStarted) isEvent()       {}
func (CardPlayed) isEvent()        {}
func (PlayUndone) isEvent()        {}
func (TurnEnded) isEvent()         {}
func (PlayerSkipped) isEvent()     {}
func (GameWon) isEvent()           {}
func (GameLost) isEvent()          {}
func (GameEnded) isEvent()         {}
func (RematchCreated) isEvent()    {}
