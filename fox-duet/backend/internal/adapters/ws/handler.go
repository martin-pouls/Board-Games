package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/martin-pouls/shared/backend/wire"
	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/fox-duet/backend/internal/adapters/apidto"
	"github.com/martin-pouls/fox-duet/backend/internal/core/domain"
	"github.com/martin-pouls/fox-duet/backend/internal/core/ports"
)

// Options configures the WebSocket handler.
type Options = wskit.Options

// NewHandler returns the endpoint for GET /ws/games/{code}?token=<token>.
func NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler {
	return wskit.NewHandler(hub.Hub, &backend{svc: svc}, opts)
}

// backend plugs Fox Duet into the shared endpoint: authentication, the first
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
	case apidto.CmdStartGame:
		_, err = b.svc.StartGame(ctx, id, p, apidto.LevelOrDefault(msg.Level))
	case apidto.CmdPlayCard:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.PlayCard(ctx, id, p, c)
		}
	case apidto.CmdChooseFoxesPlayer:
		if msg.PlayerID == "" {
			return apidto.BadRequest("missing playerId")
		}
		_, err = b.svc.ChooseFoxesPlayer(ctx, id, p, domain.PlayerID(msg.PlayerID))
	case apidto.CmdSwapDecree:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.SwapDecree(ctx, id, p, c)
		}
	case apidto.CmdKeepDecree:
		_, err = b.svc.KeepDecree(ctx, id, p)
	case apidto.CmdGiveGift:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.GiveGift(ctx, id, p, c)
		}
	case apidto.CmdChooseMove:
		if msg.Delta == nil {
			return apidto.BadRequest("missing delta")
		}
		_, err = b.svc.ChooseMove(ctx, id, p, *msg.Delta)
	case apidto.CmdChooseForestEnd:
		if msg.End == "" {
			return apidto.BadRequest("missing end")
		}
		_, err = b.svc.ChooseForestEnd(ctx, id, p, apidto.EndFromWire(msg.End))
	case apidto.CmdGiveUp:
		_, err = b.svc.GiveUp(ctx, id, p)
	case apidto.CmdEndGame:
		_, err = b.svc.EndGame(ctx, id, p)
	case apidto.CmdRematch:
		_, err = b.svc.Rematch(ctx, id, p)
	default:
		return apidto.BadRequest("unknown message type %q", msg.Type)
	}
	return err
}

// cardOf maps the card of a message; a missing card is bad_request.
func cardOf(msg apidto.ClientMessage) (domain.Card, error) {
	if msg.Card == nil {
		return domain.Card{}, apidto.BadRequest("card is required")
	}
	return msg.Card.ToDomain()
}
