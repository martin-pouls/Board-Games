package domain

// NewGame creates a game in the lobby with the host in seat 0.
// Errors: ErrInvalidName, ErrInvalidOptions. Piles = NewPiles; Revision = 1; no events.
func NewGame(id GameID, hostID PlayerID, hostName string, opts Options) (*Game, error) {
	panic("not implemented: WP3")
}

// AddPlayer seats a new player in the lobby.
// Errors: ErrWrongPhase, ErrGameFull, ErrInvalidName, ErrNameTaken, ErrDuplicatePlayer. Emits PlayerJoined.
func (g *Game) AddPlayer(id PlayerID, name string) error {
	panic("not implemented: WP3")
}

// SetOptions changes the options in the lobby (host only).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidOptions. Emits OptionsChanged.
func (g *Game) SetOptions(by PlayerID, opts Options) error {
	panic("not implemented: WP3")
}

// Start deals the cards (host only). Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase.
// Emits GameStarted (+ solo: FirstPlayerChosen, TurnStarted).
func (g *Game) Start(by PlayerID, rng RNG) error {
	panic("not implemented: WP3")
}

// ChooseFirstPlayer picks the starting player (host only).
// Errors: ErrUnknownPlayer (by or first), ErrNotHost, ErrWrongPhase. Emits FirstPlayerChosen, TurnStarted.
func (g *Game) ChooseFirstPlayer(by, first PlayerID) error {
	panic("not implemented: WP3")
}

// Rematch creates the lobby for a new game with the same players (same IDs, names, seat order),
// Options and an empty table; by becomes host. Any player may call it once the game is finished.
// Errors: ErrUnknownPlayer, ErrWrongPhase (not finished), ErrRematchExists (NextGameID already set).
// On success sets g.NextGameID = newID, emits RematchCreated, commits and returns the new game.
func (g *Game) Rematch(by PlayerID, newID GameID) (*Game, error) {
	panic("not implemented: WP3")
}
