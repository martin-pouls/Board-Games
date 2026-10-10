package domain

// GameID identifies a game. It is also the join code: 5 chars without 0/O/1/I, generated and
// normalised (upper-cased, trimmed) by the service.
type GameID string

// PlayerID identifies a player within a game.
type PlayerID string
