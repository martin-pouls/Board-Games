package domain

import (
	"fmt"
	"slices"
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

// newLobby builds a fresh lobby game with the given seated players: seat 0 is the Pilot, the
// default scenario and no abilities (§4.1).
func newLobby(id GameID, hostID PlayerID, players []Player) *Game {
	return &Game{
		ID:        id,
		Phase:     PhaseLobby,
		HostID:    hostID,
		Players:   players,
		Scenario:  DefaultScenario,
		Abilities: []AbilityID{},
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

// Leave removes a player from the lobby. If the host leaves, the other player becomes host. The
// remaining player keeps their role; the free seat takes the other one.
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
	if seat == 0 {
		g.PilotSeat = 1 - g.PilotSeat // the remaining player moves from seat 1 to seat 0
	}
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

// Configure sets the scenario and the ability cards (host only, lobby, §4.1, §13). The selection
// may be partial; Start requires the scenario's full number.
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrUnknownScenario, ErrInvalidAbilities
// (unknown, duplicate or more than the scenario allows). Emits FlightConfigured.
func (g *Game) Configure(by PlayerID, scenario ScenarioID, abilities []AbilityID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: the flight can only be configured in the lobby", ErrWrongPhase)
	}
	sc, ok := ScenarioFor(scenario)
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownScenario, scenario)
	}
	if err := validateAbilities(abilities, sc.Abilities); err != nil {
		return err
	}

	// All validation done; mutate.
	g.Scenario = scenario
	g.Abilities = append([]AbilityID{}, abilities...)
	g.emit(FlightConfigured{Scenario: scenario, Abilities: slices.Clone(g.Abilities)})
	g.commit()
	return nil
}

// validateAbilities checks a card selection: known cards, no duplicates, at most n.
func validateAbilities(abilities []AbilityID, n int) error {
	if len(abilities) > n {
		return fmt.Errorf("%w: %d cards, the scenario allows %d", ErrInvalidAbilities, len(abilities), n)
	}
	for i, a := range abilities {
		if !slices.Contains(Abilities(), a) {
			return fmt.Errorf("%w: unknown card %q", ErrInvalidAbilities, a)
		}
		if slices.Contains(abilities[:i], a) {
			return fmt.Errorf("%w: %q twice", ErrInvalidAbilities, a)
		}
	}
	return nil
}

// SwapRoles swaps Pilot and Co-Pilot between the seats (host only, lobby, §4.1).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase. Emits RolesSwapped.
func (g *Game) SwapRoles(by PlayerID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: roles can only be swapped in the lobby", ErrWrongPhase)
	}
	g.PilotSeat = 1 - g.PilotSeat
	g.emit(RolesSwapped{PilotSeat: g.PilotSeat})
	g.commit()
	return nil
}

// Start sets up the flight (§4.1, §12) and starts round 1 (host only, both seats taken, the
// scenario's full number of abilities): axis 0, all switches red, the approach track with its
// plane tokens, kerosene at StartFuel, the wind token at the top, the intern row shuffled, then
// the round start (reroll pickup, traffic rolls) and the discussion.
// RNG use order: the intern row (Shuffle(6), intern module only), then the traffic rolls of
// space 1 (one IntN(6) each).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrNotEnoughPlayers, ErrUnknownScenario,
// ErrInvalidAbilities, ErrAbilitiesIncomplete. Emits GameStarted, RoundStarted.
func (g *Game) Start(by PlayerID, rng RNG) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: game already started", ErrWrongPhase)
	}
	if len(g.Players) != Players {
		return fmt.Errorf("%w: the game needs %d players", ErrNotEnoughPlayers, Players)
	}
	sc, ok := ScenarioFor(g.Scenario)
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownScenario, g.Scenario)
	}
	if err := validateAbilities(g.Abilities, sc.Abilities); err != nil {
		return err
	}
	if len(g.Abilities) < sc.Abilities {
		return fmt.Errorf("%w: %d of %d cards chosen", ErrAbilitiesIncomplete, len(g.Abilities), sc.Abilities)
	}
	t, ok := TrackFor(sc.Track)
	if !ok {
		return fmt.Errorf("%w: %q has no track %q", ErrUnknownScenario, sc.ID, sc.Track)
	}

	// All validation done; mutate.
	f := &Flight{Round: 1, Position: 1, Planes: make([]int, t.Len())}
	for i, s := range t.Spaces {
		f.Planes[i] = s.Planes
	}
	if sc.hasKerosene() {
		f.Fuel = StartFuel
	}
	if sc.Has(ModuleIntern) {
		f.Intern = newInternRow(rng)
		f.InternRight = len(f.Intern) - 1
	}
	g.Flight = f
	g.Result = nil
	g.Phase = PhasePlaying
	g.emit(GameStarted{Scenario: g.Scenario})
	g.startRound(rng)
	g.commit()
	return nil
}

// GiveUp ends the game as a loss for both (any player, any step, §7.2).
// Errors: ErrUnknownPlayer, ErrWrongPhase. Emits GameLost.
func (g *Game) GiveUp(by PlayerID) error {
	if _, err := g.requireSeated(by); err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: only a running game can be given up", ErrWrongPhase)
	}
	g.lose(LossGaveUp, nil, by)
	g.commit()
	return nil
}

// Rematch creates the lobby for a new game with the same players (same IDs, names, seats and
// roles), the same scenario and abilities (§7.2); by becomes host. Any player may call it once
// the game is over.
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
	next := newLobby(newID, by, slices.Clone(g.Players))
	next.PilotSeat = g.PilotSeat
	next.Scenario = g.Scenario
	next.Abilities = append([]AbilityID{}, g.Abilities...)

	g.NextGameID = newID
	g.emit(RematchCreated{GameID: newID})
	g.commit()
	return next, nil
}
