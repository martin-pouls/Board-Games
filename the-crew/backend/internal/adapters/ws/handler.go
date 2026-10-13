package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/martin-pouls/shared/backend/wire"
	"github.com/martin-pouls/shared/backend/wskit"

	"github.com/martin-pouls/the-crew/backend/internal/adapters/apidto"
	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

// Options configures the WebSocket handler.
type Options = wskit.Options

// NewHandler returns the endpoint for GET /ws/games/{code}?token=<token>.
func NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler {
	return wskit.NewHandler(hub.Hub, &backend{svc: svc}, opts)
}

// backend plugs The Crew into the shared endpoint: authentication, the first
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
	case apidto.CmdSetOptions:
		if msg.Options == nil {
			return apidto.BadRequest("missing options")
		}
		_, err = b.svc.SetOptions(ctx, id, p, msg.Options.ToDomain())
	case apidto.CmdSelectMission:
		if msg.Mission == nil {
			return apidto.BadRequest("missing mission")
		}
		_, err = b.svc.SelectMission(ctx, id, p, *msg.Mission)
	case apidto.CmdStartMission:
		_, err = b.svc.StartMission(ctx, id, p)
	case apidto.CmdAdjustTokens:
		if msg.From == nil || msg.To == nil {
			return apidto.BadRequest("missing from or to")
		}
		_, err = b.svc.AdjustTokens(ctx, id, p, *msg.From, *msg.To)
	case apidto.CmdSkipTokens:
		_, err = b.svc.SkipTokens(ctx, id, p)
	case apidto.CmdAnswerPoll:
		if msg.Yes == nil {
			return apidto.BadRequest("missing yes")
		}
		_, err = b.svc.AnswerPoll(ctx, id, p, *msg.Yes)
	case apidto.CmdPick:
		if len(msg.Seats) == 0 {
			return apidto.BadRequest("missing seats")
		}
		_, err = b.svc.PickSeats(ctx, id, p, msg.Seats)
	case apidto.CmdPickTask:
		if msg.TaskID == nil {
			return apidto.BadRequest("missing taskId")
		}
		_, err = b.svc.PickTask(ctx, id, p, *msg.TaskID)
	case apidto.CmdSendDistress:
		if msg.Direction == "" {
			return apidto.BadRequest("missing direction")
		}
		_, err = b.svc.SendDistress(ctx, id, p, domain.Direction(msg.Direction))
	case apidto.CmdSkipDistress:
		_, err = b.svc.SkipDistress(ctx, id, p)
	case apidto.CmdChoosePassCard:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.ChoosePassCard(ctx, id, p, c)
		}
	case apidto.CmdHandover:
		if msg.TaskID == nil || msg.Seat == nil {
			return apidto.BadRequest("missing taskId or seat")
		}
		_, err = b.svc.Handover(ctx, id, p, *msg.TaskID, *msg.Seat)
	case apidto.CmdCommunicate:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.Communicate(ctx, id, p, c, domain.Position(msg.Position))
		}
	case apidto.CmdPlayCard:
		var c domain.Card
		if c, err = cardOf(msg); err == nil {
			_, err = b.svc.PlayCard(ctx, id, p, c)
		}
	case apidto.CmdGiveUp:
		_, err = b.svc.GiveUp(ctx, id, p)
	case apidto.CmdNextAttempt:
		_, err = b.svc.NextAttempt(ctx, id, p)
	case apidto.CmdNextMission:
		_, err = b.svc.NextMission(ctx, id, p)
	case apidto.CmdBackToLobby:
		_, err = b.svc.BackToLobby(ctx, id, p)
	case apidto.CmdEndGame:
		_, err = b.svc.EndGame(ctx, id, p)
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
