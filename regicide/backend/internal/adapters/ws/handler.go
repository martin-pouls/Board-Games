package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/martin-pouls/shared/backend/wire"
	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/regicide/backend/internal/adapters/apidto"
	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
)

// Options configures the WebSocket handler.
type Options = wskit.Options

// NewHandler returns the endpoint for GET /ws/games/{code}?token=<token>.
func NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler {
	return wskit.NewHandler(hub.Hub, &backend{svc: svc}, opts)
}

// backend plugs Regicide into the shared endpoint: authentication, the first
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
		_, err = b.svc.StartGame(ctx, id, p, domain.PlayerID(msg.FirstPlayerID))
	case apidto.CmdPlay:
		var cards []domain.Card
		if cards, err = cardsOf(msg); err == nil {
			_, err = b.svc.Play(ctx, id, p, cards)
		}
	case apidto.CmdYield:
		_, err = b.svc.Yield(ctx, id, p)
	case apidto.CmdDiscard:
		var cards []domain.Card
		if cards, err = cardsOf(msg); err == nil {
			_, err = b.svc.Discard(ctx, id, p, cards)
		}
	case apidto.CmdChooseNext:
		if msg.PlayerID == "" {
			return apidto.BadRequest("missing playerId")
		}
		_, err = b.svc.ChooseNext(ctx, id, p, domain.PlayerID(msg.PlayerID))
	case apidto.CmdUseSoloJester:
		_, err = b.svc.UseSoloJester(ctx, id, p)
	case apidto.CmdSetSignal:
		if msg.Signal == nil {
			return apidto.BadRequest("missing signal")
		}
		_, err = b.svc.SetJesterSignal(ctx, id, p, apidto.SignalFromWire(*msg.Signal))
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

// cardsOf maps the cards of a play/discard message; a missing or null list is bad_request.
func cardsOf(msg apidto.ClientMessage) ([]domain.Card, error) {
	if msg.Cards == nil {
		return nil, apidto.BadRequest("cards are required")
	}
	return apidto.CardsToDomain(msg.Cards)
}
