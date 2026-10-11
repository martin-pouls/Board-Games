package ports

import "errors"

// Port-level sentinel errors (not game rules).
var (
	ErrGameNotFound    = errors.New("game not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrVersionConflict = errors.New("version conflict")
	ErrDuplicateGameID = errors.New("duplicate game id")
)
