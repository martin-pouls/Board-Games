package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// maxNameRunes is the maximum length of a player name in runes.
const maxNameRunes = 20

// NewGame creates a game in the lobby with the host in seat 0.
// Errors: ErrInvalidName. Revision = 1; no events.
func NewGame(id GameID, hostID PlayerID, hostName string) (*Game, error) {
	name, err := normalizeName(hostName)
	if err != nil {
		return nil, err
	}
	return newLobby(id, hostID, []Player{{ID: hostID, Name: name, Hand: []Card{}}}), nil
}

// newLobby builds a fresh lobby game with the given seated players.
func newLobby(id GameID, hostID PlayerID, players []Player) *Game {
	return &Game{
		ID:          id,
		Phase:       PhaseLobby,
		HostID:      hostID,
		Players:     players,
		Castle:      []Card{},
		Tavern:      []Card{},
		DiscardPile: []Card{},
		Revision:    1,
		CreatedAt:   time.Now().UTC().Truncate(time.Second),
	}
}

// normalizeName trims name and validates it: 1..20 runes, valid UTF-8, no control characters.
func normalizeName(name string) (string, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return "", fmt.Errorf("%w: name is empty", ErrInvalidName)
	}
	if !utf8.ValidString(n) {
		return "", fmt.Errorf("%w: name is not valid UTF-8", ErrInvalidName)
	}
	if utf8.RuneCountInString(n) > maxNameRunes {
		return "", fmt.Errorf("%w: name is longer than %d characters", ErrInvalidName, maxNameRunes)
	}
	if strings.IndexFunc(n, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("%w: name contains control characters", ErrInvalidName)
	}
	return n, nil
}

// AddPlayer seats a new player in the lobby.
// Errors: ErrWrongPhase, ErrDuplicatePlayer, ErrGameFull, ErrInvalidName, ErrNameTaken. Emits PlayerJoined.
func (g *Game) AddPlayer(id PlayerID, name string) error {
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: players can only join in the lobby", ErrWrongPhase)
	}
	if _, _, ok := g.PlayerByID(id); ok {
		return fmt.Errorf("%w: %s", ErrDuplicatePlayer, id)
	}
	if len(g.Players) >= MaxPlayers {
		return fmt.Errorf("%w: at most %d players", ErrGameFull, MaxPlayers)
	}
	n, err := normalizeName(name)
	if err != nil {
		return err
	}
	for _, p := range g.Players {
		if strings.EqualFold(p.Name, n) {
			return fmt.Errorf("%w: %q", ErrNameTaken, n)
		}
	}
	g.Players = append(g.Players, Player{ID: id, Name: n, Hand: []Card{}})
	g.emit(PlayerJoined{PlayerID: id, Name: n})
	g.commit()
	return nil
}

// requireSeated returns the seat of by. Errors: ErrUnknownPlayer.
func (g *Game) requireSeated(by PlayerID) (int, error) {
	_, seat, ok := g.PlayerByID(by)
	if !ok {
		return -1, fmt.Errorf("%w: %s", ErrUnknownPlayer, by)
	}
	return seat, nil
}

// requireHost checks that by is seated and is the host.
func (g *Game) requireHost(by PlayerID) error {
	if _, err := g.requireSeated(by); err != nil {
		return err
	}
	if !g.IsHost(by) {
		return fmt.Errorf("%w: only the host may do this", ErrNotHost)
	}
	return nil
}

// Leave removes a player from the lobby. If the host leaves, the player now in seat 0 becomes host.
// Errors: ErrUnknownPlayer, ErrWrongPhase (not in the lobby), ErrCannotLeave (sole player).
// Emits PlayerLeft.
func (g *Game) Leave(by PlayerID) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: players can only leave in the lobby", ErrWrongPhase)
	}
	if len(g.Players) == 1 {
		return fmt.Errorf("%w: the last player cannot leave", ErrCannotLeave)
	}
	players := make([]Player, 0, len(g.Players)-1)
	players = append(players, g.Players[:seat]...)
	g.Players = append(players, g.Players[seat+1:]...)
	if g.HostID == by {
		g.HostID = g.Players[0].ID
	}
	g.emit(PlayerLeft{PlayerID: by, HostID: g.HostID})
	g.commit()
	return nil
}

// Start sets up the decks and deals (host only). first is the first player; "" picks one at random.
// RNG use order: NewCastle (Jacks, Queens, Kings), the Tavern, then (random first only) the seats.
// Errors: ErrUnknownPlayer (by or first), ErrNotHost, ErrWrongPhase.
// Emits GameStarted, EnemyRevealed, TurnStarted.
func (g *Game) Start(by, first PlayerID, rng RNG) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: game already started", ErrWrongPhase)
	}
	firstSeat := -1
	if first != "" {
		_, s, ok := g.PlayerByID(first)
		if !ok {
			return fmt.Errorf("%w: first player %s", ErrUnknownPlayer, first)
		}
		firstSeat = s
	}

	// All validation done; mutate.
	rules := g.Rules()
	n := len(g.Players)
	castle := NewCastle(rng)
	tavern := NewTavern(rules.TavernJesters)
	rng.Shuffle(len(tavern), func(i, j int) { tavern[i], tavern[j] = tavern[j], tavern[i] })
	if firstSeat < 0 {
		seats := make([]int, n)
		for i := range seats {
			seats[i] = i
		}
		rng.Shuffle(n, func(i, j int) { seats[i], seats[j] = seats[j], seats[i] })
		firstSeat = seats[0]
	}

	hands := make([][]Card, n)
	for i := range hands {
		hands[i] = make([]Card, 0, rules.MaxHandSize)
	}
	for range rules.MaxHandSize {
		for seat := range n {
			top := len(tavern) - 1
			hands[seat] = append(hands[seat], tavern[top])
			tavern = tavern[:top]
		}
	}
	for seat := range g.Players {
		SortCards(hands[seat])
		p := &g.Players[seat]
		p.Hand = hands[seat]
		p.YieldedLastTurn = false
		p.Signal = SignalNone
	}
	top := len(castle) - 1
	enemy := castle[top]
	g.Castle = castle[:top]
	g.Enemy = &enemy
	g.Tavern = tavern
	g.DiscardPile = []Card{}
	g.Played = nil
	g.SoloJestersUsed = 0
	g.Result = nil
	g.Turn = Turn{}
	g.Phase = PhasePlaying
	g.emit(GameStarted{FirstPlayerID: g.Players[firstSeat].ID})
	g.emit(EnemyRevealed{Enemy: enemy})
	g.beginTurn(firstSeat)
	g.commit()
	return nil
}

// GiveUp ends the game as a loss for everyone (any player, any step).
// Errors: ErrUnknownPlayer, ErrWrongPhase. Emits GameLost.
func (g *Game) GiveUp(by PlayerID) error {
	if _, err := g.requireSeated(by); err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: only a running game can be given up", ErrWrongPhase)
	}
	g.lose(LossGaveUp, by)
	g.commit()
	return nil
}

// Rematch creates the lobby for a new game with the same players (same IDs, names, seat order)
// and an empty table; by becomes host. Any player may call it once the game is over.
// Errors: ErrUnknownPlayer, ErrWrongPhase (not won/lost), ErrRematchExists (NextGameID already set).
// On success sets g.NextGameID = newID, emits RematchCreated, commits and returns the new game.
func (g *Game) Rematch(by PlayerID, newID GameID) (*Game, error) {
	if _, err := g.requireSeated(by); err != nil {
		return nil, err
	}
	if g.Phase != PhaseWon && g.Phase != PhaseLost {
		return nil, fmt.Errorf("%w: rematch needs a finished game", ErrWrongPhase)
	}
	if g.NextGameID != "" {
		return nil, fmt.Errorf("%w: %s", ErrRematchExists, g.NextGameID)
	}
	players := make([]Player, len(g.Players))
	for i, p := range g.Players {
		players[i] = Player{ID: p.ID, Name: p.Name, Hand: []Card{}}
	}
	next := newLobby(newID, by, players)

	g.NextGameID = newID
	g.emit(RematchCreated{GameID: newID})
	g.commit()
	return next, nil
}
