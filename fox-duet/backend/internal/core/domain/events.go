package domain

// Event is a domain event. Events carry only public information, never hand contents or gift
// choices (played cards and the decree are public). Events are emitted as values.
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

type GameStarted struct{ Level int }

// RoundStarted: the cards were dealt and the decree revealed.
type RoundStarted struct {
	Round    int
	DealerID PlayerID
	Decree   Card
}

type CardPlayed struct {
	PlayerID PlayerID
	Card     Card
	Lead     bool
}

// FoxesChosen: ByPlayerID played the Foxes and chose PlayerID. Skipped is true when the chosen
// player has no card to swap, so nothing happens.
type FoxesChosen struct {
	ByPlayerID, PlayerID PlayerID
	Skipped              bool
}

// DecreeSwapped: PlayerID put New from their hand as decree and took Old.
type DecreeSwapped struct {
	PlayerID PlayerID
	Old, New Card
}

type DecreeKept struct{ PlayerID PlayerID }

// GiftChosen: PlayerID has chosen their gift card (which one stays secret). Auto is true when it
// was their only card.
type GiftChosen struct {
	PlayerID PlayerID
	Auto     bool
}

// GiftExchanged: both gift cards changed hands.
type GiftExchanged struct{}

type TrickWon struct {
	PlayerID PlayerID
	Number   int
}

// TravellerMoved: the traveller moved by Delta from From to To. With Beyond set it left the path
// at that end and To is the start space.
type TravellerMoved struct {
	From, To, Delta int
	Beyond          End
}

// ForestTokenPlaced: a token now covers Space at End; GemsMoved gems moved toward the centre.
// RoundEnd is false when the traveller left the path.
type ForestTokenPlaced struct {
	End       End
	Space     int
	GemsMoved int
	RoundEnd  bool
	Left      int // forest tokens left
}

type GemCollected struct {
	Space int
	Left  int // gems left on the board
}

// GemsAdded: the round-end gems were added.
type GemsAdded struct{ Count int }

type GameWon struct{ Score int }

type GameLost struct {
	Reason   LossReason
	PlayerID PlayerID
}

type GameEnded struct{ ByPlayerID PlayerID }

type RematchCreated struct{ GameID GameID }

func (PlayerJoined) isEvent()      {}
func (PlayerLeft) isEvent()        {}
func (GameStarted) isEvent()       {}
func (RoundStarted) isEvent()      {}
func (CardPlayed) isEvent()        {}
func (FoxesChosen) isEvent()       {}
func (DecreeSwapped) isEvent()     {}
func (DecreeKept) isEvent()        {}
func (GiftChosen) isEvent()        {}
func (GiftExchanged) isEvent()     {}
func (TrickWon) isEvent()          {}
func (TravellerMoved) isEvent()    {}
func (ForestTokenPlaced) isEvent() {}
func (GemCollected) isEvent()      {}
func (GemsAdded) isEvent()         {}
func (GameWon) isEvent()           {}
func (GameLost) isEvent()          {}
func (GameEnded) isEvent()         {}
func (RematchCreated) isEvent()    {}
