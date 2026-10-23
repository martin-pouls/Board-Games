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
	return newLobby(id, hostID, []Player{{ID: hostID, Name: name}}), nil
}

// newLobby builds a fresh lobby game with the given seated players.
func newLobby(id GameID, hostID PlayerID, players []Player) *Game {
	return &Game{
		ID:        id,
		Phase:     PhaseLobby,
		HostID:    hostID,
		Players:   players,
		Revision:  1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
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
	if len(g.Players) >= Players {
		return fmt.Errorf("%w: at most %d players", ErrGameFull, Players)
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
	g.Players = append(g.Players, Player{ID: id, Name: n})
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

// Leave removes a player from the lobby. If the host leaves, the other player becomes host.
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

// EndGame closes the lobby for everyone (host only), so it no longer waits to be continued; a
// sole player cannot Leave, so this is how they get rid of it.
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase (not in the lobby). Emits GameEnded.
func (g *Game) EndGame(by PlayerID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: only a lobby can be ended", ErrWrongPhase)
	}
	g.Phase = PhaseEnded
	g.emit(GameEnded{ByPlayerID: by})
	g.commit()
	return nil
}
