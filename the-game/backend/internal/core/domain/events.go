package domain

// Event is a domain event. Events carry only public information, never hand contents.
// Events are emitted as values (not pointers).
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

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

type GameFinished struct {
	Outcome       Outcome
	Score         int
	StuckPlayerID PlayerID
}

func (PlayerJoined) isEvent()      {}
func (OptionsChanged) isEvent()    {}
func (GameStarted) isEvent()       {}
func (FirstPlayerChosen) isEvent() {}
func (TurnStarted) isEvent()       {}
func (CardPlayed) isEvent()        {}
func (PlayUndone) isEvent()        {}
func (TurnEnded) isEvent()         {}
func (PlayerSkipped) isEvent()     {}
func (GameFinished) isEvent()      {}
func (RematchCreated) isEvent()    {}
