package apidto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

// Error codes on the wire (keep in sync with ApiErrorCode in frontend/src/lib/types.ts).
const (
	CodeBadRequest        = "bad_request"
	CodeUnauthorized      = "unauthorized"
	CodeNotHost           = "not_host"
	CodeGameNotFound      = "game_not_found"
	CodeNotFound          = "not_found"
	CodeWrongPhase        = "wrong_phase"
	CodeGameFull          = "game_full"
	CodeNameTaken         = "name_taken"
	CodeNotYourTurn       = "not_your_turn"
	CodeConflict          = "conflict"
	CodeInvalidName       = "invalid_name"
	CodeInvalidOptions    = "invalid_options"
	CodeUnknownPlayer     = "unknown_player"
	CodeCardNotInHand     = "card_not_in_hand"
	CodeInvalidPile       = "invalid_pile"
	CodeIllegalPlacement  = "illegal_placement"
	CodeMinimumNotReached = "minimum_not_reached"
	CodeRematchExists     = "rematch_exists"
	CodeInternal          = "internal"
)

// ErrBadRequest marks malformed input (bad JSON, missing field, unknown WS type).
// Adapters wrap it with detail via BadRequest.
var ErrBadRequest = errors.New("bad request")

// BadRequest returns an error wrapping ErrBadRequest with a human-readable reason.
func BadRequest(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, fmt.Sprintf(format, args...))
}

// InternalMessage is the message sent for unmapped errors (log the real error instead).
const InternalMessage = "internal error"

var errorTable = []struct {
	err    error
	code   string
	status int
}{
	{ErrBadRequest, CodeBadRequest, http.StatusBadRequest},
	{ports.ErrUnauthorized, CodeUnauthorized, http.StatusUnauthorized},
	{domain.ErrNotHost, CodeNotHost, http.StatusForbidden},
	{ports.ErrGameNotFound, CodeGameNotFound, http.StatusNotFound},
	{domain.ErrWrongPhase, CodeWrongPhase, http.StatusConflict},
	{domain.ErrGameFull, CodeGameFull, http.StatusConflict},
	{domain.ErrRematchExists, CodeRematchExists, http.StatusConflict},
	{domain.ErrNameTaken, CodeNameTaken, http.StatusConflict},
	{domain.ErrNotYourTurn, CodeNotYourTurn, http.StatusConflict},
	{ports.ErrVersionConflict, CodeConflict, http.StatusConflict},
	{domain.ErrInvalidName, CodeInvalidName, http.StatusUnprocessableEntity},
	{domain.ErrInvalidOptions, CodeInvalidOptions, http.StatusUnprocessableEntity},
	{domain.ErrUnknownPlayer, CodeUnknownPlayer, http.StatusUnprocessableEntity},
	{domain.ErrCardNotInHand, CodeCardNotInHand, http.StatusUnprocessableEntity},
	{domain.ErrInvalidPile, CodeInvalidPile, http.StatusUnprocessableEntity},
	{domain.ErrIllegalPlacement, CodeIllegalPlacement, http.StatusUnprocessableEntity},
	{domain.ErrMinimumNotReached, CodeMinimumNotReached, http.StatusUnprocessableEntity},
}

// ErrorFor maps an error to an HTTP status and wire error body. It is used by
// both REST and WS. Wrapped errors are matched with errors.Is / errors.As.
// Malformed JSON (*json.SyntaxError, *json.UnmarshalTypeError) and oversized
// bodies (*http.MaxBytesError) map to bad_request. Anything unknown maps to
// 500 "internal" with a generic message; the caller should log the real error.
func ErrorFor(err error) (int, ErrorBody) {
	if err == nil {
		return http.StatusInternalServerError, ErrorBody{Code: CodeInternal, Message: InternalMessage}
	}
	for _, e := range errorTable {
		if errors.Is(err, e.err) {
			return e.status, ErrorBody{Code: e.code, Message: err.Error()}
		}
	}
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		maxErr    *http.MaxBytesError
	)
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.As(err, &maxErr) {
		return http.StatusBadRequest, ErrorBody{Code: CodeBadRequest, Message: err.Error()}
	}
	return http.StatusInternalServerError, ErrorBody{Code: CodeInternal, Message: InternalMessage}
}
