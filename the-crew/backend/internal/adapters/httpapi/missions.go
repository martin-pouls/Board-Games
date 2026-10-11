package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// newMissionsHandler serves GET /api/missions: the static catalogue of the 50 missions for the
// lobby's mission picker. It is the same for everybody, so it is encoded once. Like the whole
// app it sits behind the gate; it needs no game token.
func newMissionsHandler() http.Handler {
	body, err := json.Marshal(apidto.Catalogue(domain.EditionPlanetNine))
	if err != nil {
		panic("httpapi: encode mission catalogue: " + err.Error())
	}
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(append(body[:len(body):len(body)], '\n'))
	})
}
