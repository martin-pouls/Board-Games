package domain

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxNameRunes is the maximum length of a player name in runes.
const maxNameRunes = 20

// NewGame creates a game in the lobby with the host in seat 0.
// Errors: ErrInvalidName, ErrInvalidOptions. Piles = NewPiles; Revision = 1; no events.
func NewGame(id GameID, hostID PlayerID, hostName string, opts Options) (*Game, error) {
	name, err := normalizeName(hostName)
	if err != nil {
		return nil, err
	}
	rules, err := NewRuleset(opts)
	if err != nil {
		return nil, err
	}
	return newLobby(id, hostID, opts, rules, []Player{{ID: hostID, Name: name, Hand: []Card{}}}), nil
}

// newLobby builds a fresh lobby game with the given seated players.
func newLobby(id GameID, hostID PlayerID, opts Options, rules Ruleset, players []Player) *Game {
	return &Game{
		ID:       id,
		Options:  opts,
		Rules:    rules,
		Phase:    PhaseLobby,
		HostID:   hostID,
		Players:  players,
		DrawPile: []Card{},
		Piles:    NewPiles(rules),
		Revision: 1,
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
// Errors: ErrWrongPhase, ErrGameFull, ErrInvalidName, ErrNameTaken, ErrDuplicatePlayer. Emits PlayerJoined.
func (g *Game) AddPlayer(id PlayerID, name string) error {
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: players can only join in the lobby", ErrWrongPhase)
	}
	if _, _, ok := g.PlayerByID(id); ok {
		return fmt.Errorf("%w: %s", ErrDuplicatePlayer, id)
	}
	if len(g.Players) >= g.Rules.MaxPlayers {
		return fmt.Errorf("%w: at most %d players", ErrGameFull, g.Rules.MaxPlayers)
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

// requireHost checks that by is seated and is the host.
func (g *Game) requireHost(by PlayerID) error {
	if _, _, ok := g.PlayerByID(by); !ok {
		return fmt.Errorf("%w: %s", ErrUnknownPlayer, by)
	}
	if !g.IsHost(by) {
		return fmt.Errorf("%w: only the host may do this", ErrNotHost)
	}
	return nil
}

// SetOptions changes the options in the lobby (host only).
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidOptions. Emits OptionsChanged.
func (g *Game) SetOptions(by PlayerID, opts Options) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: options can only change in the lobby", ErrWrongPhase)
	}
	rules, err := NewRuleset(opts)
	if err != nil {
		return err
	}
	g.Options = opts
	g.Rules = rules
	g.emit(OptionsChanged{Options: opts})
	g.commit()
	return nil
}

// Start deals the cards (host only). Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase.
// Emits GameStarted (+ solo: FirstPlayerChosen, TurnStarted).
func (g *Game) Start(by PlayerID, rng RNG) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseLobby {
		return fmt.Errorf("%w: game already started", ErrWrongPhase)
	}
	rules, err := NewRuleset(g.Options)
	if err != nil {
		return err
	}
	n := len(g.Players) // 1..MaxPlayers: the host is seated and AddPlayer enforces the maximum
	deck, err := NewDeck(rules)
	if err != nil {
		return err
	}
	handSize := rules.HandSize(n)
	if handSize*n > len(deck) {
		return fmt.Errorf("%w: deck too small to deal", ErrInvalidOptions)
	}
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	// All validation done; mutate.
	g.Rules = rules
	hands := make([][]Card, n)
	for i := range hands {
		hands[i] = make([]Card, 0, handSize)
	}
	for range handSize {
		for seat := range n {
			top := len(deck) - 1
			hands[seat] = append(hands[seat], deck[top])
			deck = deck[:top]
		}
	}
	for seat := range g.Players {
		sortCards(hands[seat])
		g.Players[seat].Hand = hands[seat]
	}
	g.DrawPile = deck
	g.Piles = NewPiles(rules)
	g.Turn = Turn{}
	g.Result = nil
	g.Phase = PhaseChoosingFirstPlayer
	g.emit(GameStarted{})
	if n == 1 {
		// Solo: the only possible choice, not a random default.
		g.chooseFirst(0)
	}
	g.commit()
	return nil
}

// ChooseFirstPlayer picks the starting player (host only).
// Errors: ErrUnknownPlayer (by or first), ErrNotHost, ErrWrongPhase. Emits FirstPlayerChosen, TurnStarted.
func (g *Game) ChooseFirstPlayer(by, first PlayerID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseChoosingFirstPlayer {
		return fmt.Errorf("%w: not choosing the first player", ErrWrongPhase)
	}
	_, seat, ok := g.PlayerByID(first)
	if !ok {
		return fmt.Errorf("%w: first player %s", ErrUnknownPlayer, first)
	}
	g.chooseFirst(seat)
	g.commit()
	return nil
}

// chooseFirst moves to playing and begins the first turn at seat.
func (g *Game) chooseFirst(seat int) {
	g.Phase = PhasePlaying
	g.emit(FirstPlayerChosen{PlayerID: g.Players[seat].ID})
	g.beginTurn(seat)
}

// EndGame ends a running game early (host only), while choosing the first player or playing.
// The game finishes with OutcomeAborted and the cards left at this moment as the score, so the
// result screen and Rematch work as after a regular end.
// Errors: ErrUnknownPlayer, ErrNotHost, ErrWrongPhase. Emits GameFinished.
func (g *Game) EndGame(by PlayerID) error {
	if err := g.requireHost(by); err != nil {
		return err
	}
	if g.Phase != PhaseChoosingFirstPlayer && g.Phase != PhasePlaying {
		return fmt.Errorf("%w: only a running game can be ended", ErrWrongPhase)
	}
	g.finish(OutcomeAborted, "")
	g.commit()
	return nil
}

// Rematch creates the lobby for a new game with the same players (same IDs, names, seat order),
// Options and an empty table; by becomes host. Any player may call it once the game is finished.
// Errors: ErrUnknownPlayer, ErrWrongPhase (not finished), ErrRematchExists (NextGameID already set).
// On success sets g.NextGameID = newID, emits RematchCreated, commits and returns the new game.
func (g *Game) Rematch(by PlayerID, newID GameID) (*Game, error) {
	if _, _, ok := g.PlayerByID(by); !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPlayer, by)
	}
	if g.Phase != PhaseFinished {
		return nil, fmt.Errorf("%w: rematch needs a finished game", ErrWrongPhase)
	}
	if g.NextGameID != "" {
		return nil, fmt.Errorf("%w: %s", ErrRematchExists, g.NextGameID)
	}
	rules, err := NewRuleset(g.Options)
	if err != nil {
		return nil, err
	}
	players := make([]Player, len(g.Players))
	for i, p := range g.Players {
		players[i] = Player{ID: p.ID, Name: p.Name, Hand: []Card{}}
	}
	next := newLobby(newID, by, g.Options, rules, players)

	g.NextGameID = newID
	g.emit(RematchCreated{GameID: newID})
	g.commit()
	return next, nil
}

// sortCards sorts cards ascending by value in place.
func sortCards(cs []Card) {
	slices.SortFunc(cs, func(a, b Card) int { return a.Value - b.Value })
}
