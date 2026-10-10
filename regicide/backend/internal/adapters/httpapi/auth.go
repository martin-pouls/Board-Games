package httpapi

import (
	"net/http"
	"strings"

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

// bearerToken extracts the token from "Authorization: Bearer <token>".
// The scheme is matched case-insensitively. It returns "" when absent or malformed.
func bearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

// authenticate resolves the bearer token to a player of game id.
// A missing token is ports.ErrUnauthorized; otherwise the service decides.
func (a *gamesAPI) authenticate(r *http.Request, id domain.GameID) (domain.PlayerID, error) {
	tok := bearerToken(r)
	if tok == "" {
		return "", ports.ErrUnauthorized
	}
	return a.svc.Authenticate(r.Context(), id, tok)
}
