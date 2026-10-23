package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/martin-pouls/shared/backend/wire"
	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/sky-team/backend/internal/adapters/apidto"
	"github.com/martin-pouls/sky-team/backend/internal/core/domain"
	"github.com/martin-pouls/sky-team/backend/internal/core/ports"
)

// Options configures the WebSocket handler.
type Options = wskit.Options

// NewHandler returns the endpoint for GET /ws/games/{code}?token=<token>.
func NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler {
	return wskit.NewHandler(hub.Hub, &backend{svc: svc}, opts)
}

// backend plugs Sky Team into the shared endpoint: authentication, the first
// view and the commands, each a call on the driving port.
type backend struct {
	svc ports.GameService
}

var _ wskit.Backend = (*backend)(nil)

func (b *backend) Authenticate(ctx context.Context, gameID, token string) (string, error) {
	p, err := b.svc.Authenticate(ctx, domain.GameID(gameID), token)
	return string(p), err
}

func (b *backend) Hello(ctx context.Context, gameID, playerID string) ([]byte, error) {
	v, err := b.svc.View(ctx, domain.GameID(gameID), domain.PlayerID(playerID))
	if err != nil {
		return nil, err
	}
	return json.Marshal(apidto.ViewMessage(apidto.FromView(v), nil))
}

func (b *backend) MapError(err error) (int, wire.ErrorBody) { return apidto.ErrorFor(err) }

// Handle decodes the full message and runs the matching service call. The
// service publishes before returning, so the broadcast view is queued before the ack.
func (b *backend) Handle(ctx context.Context, gameID, playerID string, _ wire.Command, data []byte) error {
	var msg apidto.ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return apidto.BadRequest("malformed message: %v", err)
	}
	id, p := domain.GameID(gameID), domain.PlayerID(playerID)
	var err error
	switch msg.Type {
	case apidto.CmdEndGame:
		err = b.svc.EndGame(ctx, id, p)
	default:
		return apidto.BadRequest("unknown message type %q", msg.Type)
	}
	return err
}
