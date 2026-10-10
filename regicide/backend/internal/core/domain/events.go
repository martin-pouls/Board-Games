package domain

// Event is a domain event. Events carry only public information, never hand contents
// (played and discarded cards are public). Events are emitted as values; their slices are
// freshly cloned when emitted and never mutated afterwards.
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

type GameStarted struct{ FirstPlayerID PlayerID }

type EnemyRevealed struct{ Enemy Card }

type TurnStarted struct {
	PlayerID PlayerID
	Number   int
}

// CardsPlayed is a non-Jester play: Healed/Drawn are the cards moved by ♥/♦, Damage the damage
// dealt by this play, BlockedSuits the suits whose power was not in effect (immunity).
type CardsPlayed struct {
	PlayerID     PlayerID
	Cards        []Card
	Value        int
	Healed       int
	Drawn        int
	Damage       int
	BlockedSuits []Suit
}

type JesterPlayed struct{ PlayerID PlayerID }

type Yielded struct{ PlayerID PlayerID }

type EnemyDefeated struct {
	Enemy Card
	Exact bool // the enemy card went on top of the Tavern
}

type CardsDiscarded struct {
	PlayerID PlayerID
	Cards    []Card
	Attack   int // the effective attack that was covered
}

// AttackBlocked: the effective attack was 0, so Step 4 was skipped.
type AttackBlocked struct{ PlayerID PlayerID }

type NextPlayerChosen struct{ ByPlayerID, PlayerID PlayerID }

type JesterSignalSet struct {
	PlayerID PlayerID
	Signal   JesterSignal
}

type SoloJesterUsed struct {
	PlayerID PlayerID
	Drawn    int
	Left     int
}

type GameWon struct{ Grade Grade }

type GameLost struct {
	Reason   LossReason
	PlayerID PlayerID
}

type GameEnded struct{ ByPlayerID PlayerID }

type RematchCreated struct{ GameID GameID }

func (PlayerJoined) isEvent()     {}
func (PlayerLeft) isEvent()       {}
func (GameStarted) isEvent()      {}
func (EnemyRevealed) isEvent()    {}
func (TurnStarted) isEvent()      {}
func (CardsPlayed) isEvent()      {}
func (JesterPlayed) isEvent()     {}
func (Yielded) isEvent()          {}
func (EnemyDefeated) isEvent()    {}
func (CardsDiscarded) isEvent()   {}
func (AttackBlocked) isEvent()    {}
func (NextPlayerChosen) isEvent() {}
func (JesterSignalSet) isEvent()  {}
func (SoloJesterUsed) isEvent()   {}
func (GameWon) isEvent()          {}
func (GameLost) isEvent()         {}
func (GameEnded) isEvent()        {}
func (RematchCreated) isEvent()   {}
