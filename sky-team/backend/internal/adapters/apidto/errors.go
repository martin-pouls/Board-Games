package apidto

import (
	"net/http"

	"github.com/martin-pouls/shared/backend/httpkit"
	"github.com/martin-pouls/shared/backend/wire"

	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
	"github.com/martin-pouls/sky-team/backend/internal/core/ports"
)

// Error codes on the wire (keep in sync with ApiErrorCode in frontend/src/lib/types.ts).
// The codes every game shares come from the shared wire package.
const (
	CodeBadRequest    = wire.CodeBadRequest
	CodeUnauthorized  = wire.CodeUnauthorized
	CodeNotHost       = "not_host"
	CodeGameNotFound  = wire.CodeGameNotFound
	CodeNotFound      = wire.CodeNotFound
	CodeWrongPhase    = "wrong_phase"
	CodeGameFull      = "game_full"
	CodeNameTaken     = "name_taken"
	CodeCannotLeave   = "cannot_leave"
	CodeConflict      = wire.CodeConflict
	CodeInvalidName   = "invalid_name"
	CodeUnknownPlayer = "unknown_player"
	CodeRateLimited   = wire.CodeRateLimited
	CodeInternal      = wire.CodeInternal
)

// ErrBadRequest marks malformed input (bad JSON, missing field, unknown WS type).
// Adapters wrap it with detail via BadRequest.
var ErrBadRequest = httpkit.ErrBadRequest

// BadRequest returns an error wrapping ErrBadRequest with a human-readable reason.
func BadRequest(format string, args ...any) error { return httpkit.BadRequest(format, args...) }

// ErrRateLimited marks a request refused by the per-client abuse limits (429).
var ErrRateLimited = httpkit.ErrRateLimited

// InternalMessage is the message sent for unmapped errors (log the real error instead).
const InternalMessage = wire.InternalMessage

// Errors maps every error the adapters can see to its wire code and status.
var Errors = append(httpkit.ErrorMapper{
	{Err: ports.ErrUnauthorized, Code: CodeUnauthorized, Status: http.StatusUnauthorized},
	{Err: domain.ErrNotHost, Code: CodeNotHost, Status: http.StatusForbidden},
	{Err: ports.ErrGameNotFound, Code: CodeGameNotFound, Status: http.StatusNotFound},
	{Err: domain.ErrWrongPhase, Code: CodeWrongPhase, Status: http.StatusConflict},
	{Err: domain.ErrGameFull, Code: CodeGameFull, Status: http.StatusConflict},
	{Err: domain.ErrNameTaken, Code: CodeNameTaken, Status: http.StatusConflict},
	{Err: domain.ErrCannotLeave, Code: CodeCannotLeave, Status: http.StatusConflict},
	{Err: ports.ErrVersionConflict, Code: CodeConflict, Status: http.StatusConflict},
	{Err: domain.ErrInvalidName, Code: CodeInvalidName, Status: http.StatusUnprocessableEntity},
	{Err: domain.ErrUnknownPlayer, Code: CodeUnknownPlayer, Status: http.StatusUnprocessableEntity},
}, httpkit.CommonErrors...)

// ErrorFor maps an error to an HTTP status and wire error body. It is used by
// both REST and WS. Wrapped errors are matched with errors.Is / errors.As.
// Malformed JSON and oversized bodies map to bad_request. Anything unknown maps to
// 500 "internal" with a generic message; the caller should log the real error.
func ErrorFor(err error) (int, ErrorBody) { return Errors.Map(err) }
