package httpapi

import (
	"net/http"

	"github.com/martin-pouls/shared/backend/httpkit"

	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

// authenticate resolves the bearer token to a player of game id.
// A missing token is ports.ErrUnauthorized; otherwise the service decides.
func (a *gamesAPI) authenticate(r *http.Request, id domain.GameID) (domain.PlayerID, error) {
	tok := httpkit.BearerToken(r)
	if tok == "" {
		return "", ports.ErrUnauthorized
	}
	return a.svc.Authenticate(r.Context(), id, tok)
}
