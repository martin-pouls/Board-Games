// Package service implements the driving ports (use cases). It orchestrates
// domain operations, loads/saves aggregates via the repository port and
// publishes events via the publisher port. It holds no rules of its own.
package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/martin-pouls/shared/backend/idgen"
	"github.com/martin-pouls/shared/backend/lockmap"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
	"github.com/martin-pouls/the-crew/backend/internal/core/ports"
)

const (
	maxIDAttempts  = 5
	maxOCCAttempts = 3
)

// Config customises the service; nil fields get production defaults.
type Config struct {
	Edition     domain.Edition // the edition this deployment serves; default Planet Nine
	NewGameID   func() domain.GameID
	NewPlayerID func() domain.PlayerID
	NewToken    func() string
	NewRNG      func() domain.RNG
	Logger      *slog.Logger
}

// Service implements ports.GameService.
type Service struct {
	repo  ports.GameRepository
	pub   ports.EventPublisher
	cfg   Config
	locks *lockmap.Map[domain.GameID]
}

var _ ports.GameService = (*Service)(nil)

// New builds a Service.
func New(repo ports.GameRepository, pub ports.EventPublisher, cfg Config) *Service {
	if cfg.Edition == "" {
		cfg.Edition = domain.EditionPlanetNine
	}
	if cfg.NewGameID == nil {
		cfg.NewGameID = idgen.GameCodeAs[domain.GameID]
	}
	if cfg.NewPlayerID == nil {
		cfg.NewPlayerID = idgen.PlayerIDAs[domain.PlayerID]
	}
	if cfg.NewToken == nil {
		cfg.NewToken = idgen.Token
	}
	if cfg.NewRNG == nil {
		cfg.NewRNG = func() domain.RNG { return idgen.NewRNG() }
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	if pub == nil {
		pub = ports.NopPublisher{}
	}
	return &Service{repo: repo, pub: pub, cfg: cfg, locks: lockmap.New[domain.GameID]()}
}

// load fetches a game; a game of another edition is reported as not found, as if it didn't exist.
func (s *Service) load(ctx context.Context, id domain.GameID) (*ports.GameRecord, error) {
	rec, err := s.repo.Load(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Game.Edition != s.cfg.Edition {
		return nil, ports.ErrGameNotFound
	}
	return rec, nil
}

// mutate runs fn on a freshly loaded record under the per-game lock, saves it with
// OCC retry and publishes the resulting events while still holding the lock.
func (s *Service) mutate(ctx context.Context, id domain.GameID, fn func(*ports.GameRecord) error) (*ports.GameRecord, error) {
	s.locks.Lock(id)
	defer s.locks.Unlock(id)
	for range maxOCCAttempts {
		rec, err := s.load(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := fn(rec); err != nil {
			return nil, err
		}
		events := rec.Game.PullEvents()
		err = s.repo.Update(ctx, rec)
		if errors.Is(err, ports.ErrVersionConflict) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("save game %s: %w", id, err)
		}
		s.pub.Publish(ctx, ports.GameUpdate{Game: rec.Game, Events: events})
		return rec, nil
	}
	return nil, ports.ErrVersionConflict
}

// act runs a domain mutator on the game under the per-game lock and publishes the result.
func (s *Service) act(ctx context.Context, id domain.GameID, fn func(*domain.Game) error) error {
	_, err := s.mutate(ctx, idgen.NormalizeCodeAs(id), func(rec *ports.GameRecord) error { return fn(rec.Game) })
	return err
}

// CreateGame implements ports.GameService.
func (s *Service) CreateGame(ctx context.Context, hostName string, opts domain.Options) (ports.Session, error) {
	hostID := s.cfg.NewPlayerID()
	token := s.cfg.NewToken()
	for range maxIDAttempts {
		g, err := domain.NewGame(s.cfg.NewGameID(), s.cfg.Edition, hostID, hostName, opts)
		if err != nil {
			return ports.Session{}, err
		}
		rec := &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{hostID: idgen.HashToken(token)}}
		err = s.repo.Create(ctx, rec)
		if errors.Is(err, ports.ErrDuplicateGameID) {
			continue
		}
		if err != nil {
			return ports.Session{}, fmt.Errorf("create game: %w", err)
		}
		return newSession(g, hostID, token)
	}
	return ports.Session{}, fmt.Errorf("create game: no free game id after %d attempts: %w", maxIDAttempts, ports.ErrDuplicateGameID)
}

func newSession(g *domain.Game, player domain.PlayerID, token string) (ports.Session, error) {
	v, err := g.ViewFor(player)
	if err != nil {
		return ports.Session{}, err
	}
	return ports.Session{GameID: g.ID, PlayerID: player, Token: token, View: v}, nil
}

// JoinGame implements ports.GameService.
func (s *Service) JoinGame(ctx context.Context, id domain.GameID, name string) (ports.Session, error) {
	pid := s.cfg.NewPlayerID()
	token := s.cfg.NewToken()
	rec, err := s.mutate(ctx, idgen.NormalizeCodeAs(id), func(rec *ports.GameRecord) error {
		if err := rec.Game.AddPlayer(pid, name); err != nil {
			return err
		}
		rec.TokenHashes[pid] = idgen.HashToken(token)
		return nil
	})
	if err != nil {
		return ports.Session{}, err
	}
	return newSession(rec.Game, pid, token)
}

// Authenticate implements ports.GameService.
func (s *Service) Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error) {
	rec, err := s.load(ctx, idgen.NormalizeCodeAs(id))
	if err != nil {
		return "", err
	}
	got := []byte(idgen.HashToken(token))
	var found domain.PlayerID
	for pid, h := range rec.TokenHashes {
		if subtle.ConstantTimeCompare(got, []byte(h)) == 1 {
			found = pid
		}
	}
	if found == "" {
		return "", ports.ErrUnauthorized
	}
	if _, _, ok := rec.Game.PlayerByID(found); !ok {
		return "", ports.ErrUnauthorized // the player left the game
	}
	return found, nil
}

// Summary implements ports.GameService.
func (s *Service) Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error) {
	rec, err := s.load(ctx, idgen.NormalizeCodeAs(id))
	if err != nil {
		return domain.GameSummary{}, err
	}
	return rec.Game.Summary(), nil
}

// View implements ports.GameService.
func (s *Service) View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	rec, err := s.load(ctx, idgen.NormalizeCodeAs(id))
	if err != nil {
		return domain.PlayerView{}, err
	}
	return rec.Game.ViewFor(player)
}

// LeaveGame implements ports.GameService.
func (s *Service) LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	_, err := s.mutate(ctx, idgen.NormalizeCodeAs(id), func(rec *ports.GameRecord) error { return rec.Game.Leave(player) })
	return err
}

// SetOptions implements ports.GameService.
func (s *Service) SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SetOptions(player, opts) })
}

// SelectMission implements ports.GameService.
func (s *Service) SelectMission(ctx context.Context, id domain.GameID, player domain.PlayerID, mission int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SelectMission(player, mission) })
}

// StartMission implements ports.GameService.
func (s *Service) StartMission(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.StartMission(player, s.cfg.NewRNG()) })
}

// AdjustTokens implements ports.GameService.
func (s *Service) AdjustTokens(ctx context.Context, id domain.GameID, player domain.PlayerID, from, to int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.AdjustTokens(player, from, to) })
}

// SkipTokens implements ports.GameService.
func (s *Service) SkipTokens(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SkipTokens(player) })
}

// AnswerPoll implements ports.GameService.
func (s *Service) AnswerPoll(ctx context.Context, id domain.GameID, player domain.PlayerID, yes bool) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.AnswerPoll(player, yes) })
}

// PickSeats implements ports.GameService.
func (s *Service) PickSeats(ctx context.Context, id domain.GameID, player domain.PlayerID, seats []int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.PickSeats(player, seats) })
}

// PickTask implements ports.GameService.
func (s *Service) PickTask(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.PickTask(player, taskID) })
}

// SendDistress implements ports.GameService.
func (s *Service) SendDistress(ctx context.Context, id domain.GameID, player domain.PlayerID, dir domain.Direction) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SendDistress(player, dir) })
}

// SkipDistress implements ports.GameService.
func (s *Service) SkipDistress(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SkipDistress(player) })
}

// ChoosePassCard implements ports.GameService.
func (s *Service) ChoosePassCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.ChoosePassCard(player, card) })
}

// Handover implements ports.GameService.
func (s *Service) Handover(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID, toSeat int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.Handover(player, taskID, toSeat) })
}

// Communicate implements ports.GameService.
func (s *Service) Communicate(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card, pos domain.Position) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.Communicate(player, card, pos) })
}

// PlayCard implements ports.GameService.
func (s *Service) PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card domain.Card) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.PlayCard(player, card, s.cfg.NewRNG()) })
}

// GiveUp implements ports.GameService.
func (s *Service) GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.GiveUp(player) })
}

// PassTask implements ports.GameService.
func (s *Service) PassTask(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.PassTask(player) })
}

// ReplaceTask implements ports.GameService.
func (s *Service) ReplaceTask(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.ReplaceTask(player, taskID, s.cfg.NewRNG()) })
}

// Predict implements ports.GameService.
func (s *Service) Predict(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID, tricks int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.Predict(player, taskID, tricks) })
}

// AssignTask implements ports.GameService.
func (s *Service) AssignTask(ctx context.Context, id domain.GameID, player domain.PlayerID, taskID, toSeat int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.AssignTask(player, taskID, toSeat) })
}

// NextAttempt implements ports.GameService.
func (s *Service) NextAttempt(ctx context.Context, id domain.GameID, player domain.PlayerID, choice domain.NextAttemptChoice) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.NextAttempt(player, choice, s.cfg.NewRNG()) })
}

// NextMission implements ports.GameService.
func (s *Service) NextMission(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.NextMission(player, s.cfg.NewRNG()) })
}

// BackToLobby implements ports.GameService.
func (s *Service) BackToLobby(ctx context.Context, id domain.GameID, player domain.PlayerID, uncounted bool) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.BackToLobby(player, uncounted) })
}

// EndGame implements ports.GameService.
func (s *Service) EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.EndGame(player) })
}

// PurgeInactive deletes games not updated within maxAge and returns how many were removed.
// It is a maintenance entry point, not part of ports.GameService.
func (s *Service) PurgeInactive(ctx context.Context, maxAge time.Duration) (int, error) {
	n, err := s.repo.DeleteInactive(ctx, time.Now().Add(-maxAge))
	if err != nil {
		return 0, fmt.Errorf("purge inactive games: %w", err)
	}
	if n > 0 {
		s.cfg.Logger.Info("purged inactive games", "count", n, "maxAge", maxAge)
	}
	return n, nil
}
