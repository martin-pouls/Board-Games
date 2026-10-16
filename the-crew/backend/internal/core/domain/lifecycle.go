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

// NewGame creates a game in the lobby with the host in seat 0, mission 1 selected and opts.
// Errors: ErrInvalidName. Revision = 1; no events.
func NewGame(id GameID, hostID PlayerID, hostName string, opts Options) (*Game, error) {
	name, err := normalizeName(hostName)
	if err != nil {
		return nil, err
	}
	return &Game{
		ID:        id,
		Phase:     PhaseLobby,
		HostID:    hostID,
		Players:   []Player{{ID: hostID, Name: name}},
		Edition:   EditionPlanetNine,
		Mission:   1,
		Options:   opts,
		Logbook:   map[int]LogEntry{},
		Revision:  1,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}, nil
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

// requireHostIn checks that by is the host and the game is in phase p.
func (g *Game) requireHostIn(by PlayerID, p Phase) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != p {
		return fmt.Errorf("%w: needs phase %s", ErrWrongPhase, p)
	}
	return nil
}

// Leave removes a player from the lobby. If the host leaves, the new seat 0 becomes host.
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

// SelectMission picks the mission to play next (host, lobby).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidMission. Emits MissionSelected.
func (g *Game) SelectMission(by PlayerID, n int) error {
	if err := g.requireHostIn(by, PhaseLobby); err != nil {
		return err
	}
	if _, ok := MissionFor(g.Edition, n); !ok {
		return fmt.Errorf("%w: %d", ErrInvalidMission, n)
	}
	g.Mission = n
	g.emit(MissionSelected{Mission: n})
	g.commit()
	return nil
}

// SetOptions changes the lobby options (host only, lobby only).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase. Emits OptionsChanged.
func (g *Game) SetOptions(by PlayerID, opts Options) error {
	if err := g.requireHostIn(by, PhaseLobby); err != nil {
		return err
	}
	g.Options = opts
	g.emit(OptionsChanged{Options: opts})
	g.commit()
	return nil
}

// StartMission deals the first attempt of the selected mission (host, lobby, 2–5 players).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrNotEnoughPlayers.
// Emits AttemptStarted (and whatever the setup emits).
func (g *Game) StartMission(by PlayerID, rng RNG) error {
	if err := g.requireHostIn(by, PhaseLobby); err != nil {
		return err
	}
	if len(g.Players) < MinPlayers {
		return fmt.Errorf("%w: a mission needs %d to %d players", ErrNotEnoughPlayers, MinPlayers, MaxPlayers)
	}
	g.startAttempt(g.firstAttemptNumber(), rng)
	g.commit()
	return nil
}

// firstAttemptNumber is the number of the next attempt of the selected mission when it is started
// from the lobby or as the next mission: the logbook continues an unfinished mission, a replay of
// a completed one counts from 1.
func (g *Game) firstAttemptNumber() int {
	e := g.Logbook[g.Mission]
	if e.Completed {
		return 1
	}
	return e.Attempts + 1
}

// NextAttempt deals a new attempt of the same mission after a lost attempt (host, debrief).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidResultChoice (the attempt was won).
func (g *Game) NextAttempt(by PlayerID, rng RNG) error {
	if err := g.requireHostIn(by, PhaseDebrief); err != nil {
		return err
	}
	if g.Attempt.Result.Success {
		return fmt.Errorf("%w: the mission is completed", ErrInvalidResultChoice)
	}
	g.startAttempt(g.Attempt.Number+1, rng)
	g.commit()
	return nil
}

// NextMission starts the next mission after a won attempt (host, debrief).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidResultChoice (the attempt was
// lost), ErrInvalidMission (this was the last mission).
func (g *Game) NextMission(by PlayerID, rng RNG) error {
	if err := g.requireHostIn(by, PhaseDebrief); err != nil {
		return err
	}
	if !g.Attempt.Result.Success {
		return fmt.Errorf("%w: the mission is not completed", ErrInvalidResultChoice)
	}
	if _, ok := MissionFor(g.Edition, g.Mission+1); !ok {
		return fmt.Errorf("%w: mission %d is the last one", ErrInvalidMission, g.Mission)
	}
	g.Mission++
	g.startAttempt(g.firstAttemptNumber(), rng)
	g.commit()
	return nil
}

// BackToLobby returns the crew to the lobby (host, debrief). After a won attempt, the next
// unfinished mission is selected; after a loss the mission stays selected (rules §16). Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase.
// Emits BackInLobby.
func (g *Game) BackToLobby(by PlayerID) error {
	if err := g.requireHostIn(by, PhaseDebrief); err != nil {
		return err
	}
	if r := g.Attempt.Result; r != nil && r.Success {
		g.Mission = g.nextUnfinished()
	}
	g.Phase = PhaseLobby
	g.Attempt = nil
	g.emit(BackInLobby{})
	g.commit()
	return nil
}

// GiveUp loses the current attempt (any player, playing).
// Errors: ErrUnknownPlayer, ErrWrongPhase. Emits AttemptEnded.
func (g *Game) GiveUp(by PlayerID) error {
	seat, err := g.requireSeated(by)
	if err != nil {
		return err
	}
	if g.Phase != PhasePlaying {
		return fmt.Errorf("%w: only a running attempt can be given up", ErrWrongPhase)
	}
	g.endAttempt(&AttemptResult{Reason: ReasonGaveUp, Task: -1, Objective: -1, Trick: len(g.Attempt.Tricks), Seat: seat})
	g.commit()
	return nil
}

// EndGame closes the session for everyone (host, lobby or debrief); a running attempt must be
// given up first. Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase. Emits GameEnded.
func (g *Game) EndGame(by PlayerID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby && g.Phase != PhaseDebrief {
		return fmt.Errorf("%w: only a lobby or a debrief can be ended", ErrWrongPhase)
	}
	g.Phase = PhaseEnded
	g.emit(GameEnded{ByPlayerID: by})
	g.commit()
	return nil
}
