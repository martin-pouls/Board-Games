package domain

// Event is a domain event. Events are public: every player receives them. They never carry a
// hidden card: no hand contents, no hidden dummy card, no pass or draw card, no face-down task.
// Private facts reach a player only through their own view. Events are emitted as values.
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

type MissionSelected struct{ Mission int }

// OptionsChanged: the host changed the lobby options.
type OptionsChanged struct{ Options Options }

// AttemptStarted: the cards were dealt and the tasks drawn.
type AttemptStarted struct {
	Mission, Number int
	Commander       int // seat
}

// TokensAdjusted: the commander swapped or moved tokens between two tasks, or skipped (From and To
// are -1).
type TokensAdjusted struct{ From, To int }

// PollAnswered: answers are public as they come in.
type PollAnswered struct {
	Seat int
	Yes  bool
}

// Picked: the commander picked seats (silenced member, roles, chosen member, receiver of tasks).
type Picked struct{ Seats []int }

// TaskPicked: Seat took (or was given) a task.
type TaskPicked struct{ TaskID, Seat int }

// DistressSignal: the commander sent the signal (Skipped false) or let it pass.
type DistressSignal struct {
	Direction Direction
	Skipped   bool
}

// PassCardChosen: Seat has chosen a card to pass (which one stays secret).
type PassCardChosen struct{ Seat int }

// CardsPassed: every human passed a card in Direction.
type CardsPassed struct{ Direction Direction }

// TaskHandedOver: the five-player handover.
type TaskHandedOver struct{ TaskID, From, To int }

// Communicated: Position is empty in a dead zone.
type Communicated struct {
	Seat     int
	Card     Card
	Position Position
}

type CardPlayed struct {
	Seat int
	Card Card
}

type TrickWon struct{ Number, Seat int }

// CardsDrawn: mission 12's draw after the first trick happened (which cards stays secret).
type CardsDrawn struct{}

type AttemptEnded struct {
	Success bool
	Reason  Reason
}

// Redealt: Deep Sea: the playing cards were dealt again Times times, because an alignment made a
// task impossible (not an attempt). Commander is the captain after the re-deal.
type Redealt struct{ Times, Commander int }

// TaskReplaced: Deep Sea: the captain replaced task TaskID by another card of the same difficulty.
type TaskReplaced struct{ TaskID int }

// TaskPassed: Deep Sea: Seat passed in the draft.
type TaskPassed struct{ Seat int }

// Predicted: Deep Sea: Seat predicted Tricks tricks for task TaskID (-1: a hidden prediction).
type Predicted struct{ TaskID, Seat, Tricks int }

// BackInLobby: the host returned the crew to the lobby.
type BackInLobby struct{}

type GameEnded struct{ ByPlayerID PlayerID }

func (PlayerJoined) isEvent()    {}
func (PlayerLeft) isEvent()      {}
func (MissionSelected) isEvent() {}
func (OptionsChanged) isEvent()  {}
func (AttemptStarted) isEvent()  {}
func (TokensAdjusted) isEvent()  {}
func (PollAnswered) isEvent()    {}
func (Picked) isEvent()          {}
func (TaskPicked) isEvent()      {}
func (DistressSignal) isEvent()  {}
func (PassCardChosen) isEvent()  {}
func (CardsPassed) isEvent()     {}
func (TaskHandedOver) isEvent()  {}
func (Communicated) isEvent()    {}
func (CardPlayed) isEvent()      {}
func (TrickWon) isEvent()        {}
func (CardsDrawn) isEvent()      {}
func (AttemptEnded) isEvent()    {}
func (Redealt) isEvent()         {}
func (TaskReplaced) isEvent()    {}
func (TaskPassed) isEvent()      {}
func (Predicted) isEvent()       {}
func (BackInLobby) isEvent()     {}
func (GameEnded) isEvent()       {}
