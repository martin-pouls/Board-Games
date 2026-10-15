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
	"maps"
	"time"

	"github.com/martin-pouls/shared/backend/lockmap"

	"github.com/martin-pouls/the-game/backend/internal/core/domain"
	"github.com/martin-pouls/the-game/backend/internal/core/ports"
)

const (
	maxIDAttempts  = 5
	maxOCCAttempts = 3
)

// Config customises the service; nil fields get production defaults.
type Config struct {
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
	if cfg.NewGameID == nil {
		cfg.NewGameID = defaultGameID
	}
	if cfg.NewPlayerID == nil {
		cfg.NewPlayerID = defaultPlayerID
	}
	if cfg.NewToken == nil {
		cfg.NewToken = defaultToken
	}
	if cfg.NewRNG == nil {
		cfg.NewRNG = defaultRNG
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	if pub == nil {
		pub = ports.NopPublisher{}
	}
	return &Service{repo: repo, pub: pub, cfg: cfg, locks: lockmap.New[domain.GameID]()}
}

// mutate runs fn on a freshly loaded record under the per-game lock, saves it with
// OCC retry and publishes the resulting events while still holding the lock.
func (s *Service) mutate(ctx context.Context, id domain.GameID, fn func(*ports.GameRecord) error) (*ports.GameRecord, error) {
	s.locks.Lock(id)
	defer s.locks.Unlock(id)
	for range maxOCCAttempts {
		rec, err := s.repo.Load(ctx, id)
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

// act runs a domain mutator for the game, saves it and publishes the result.
func (s *Service) act(ctx context.Context, id domain.GameID, fn func(*domain.Game) error) error {
	_, err := s.mutate(ctx, normalizeID(id), func(rec *ports.GameRecord) error { return fn(rec.Game) })
	return err
}

// CreateGame implements ports.GameService.
func (s *Service) CreateGame(ctx context.Context, hostName string, opts domain.Options) (ports.Session, error) {
	hostID := s.cfg.NewPlayerID()
	token := s.cfg.NewToken()
	for range maxIDAttempts {
		g, err := domain.NewGame(s.cfg.NewGameID(), hostID, hostName, opts)
		if err != nil {
			return ports.Session{}, err
		}
		rec := &ports.GameRecord{Game: g, TokenHashes: map[domain.PlayerID]string{hostID: hashToken(token)}}
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
	rec, err := s.mutate(ctx, normalizeID(id), func(rec *ports.GameRecord) error {
		if err := rec.Game.AddPlayer(pid, name); err != nil {
			return err
		}
		rec.TokenHashes[pid] = hashToken(token)
		return nil
	})
	if err != nil {
		return ports.Session{}, err
	}
	return newSession(rec.Game, pid, token)
}

// Authenticate implements ports.GameService.
func (s *Service) Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error) {
	rec, err := s.repo.Load(ctx, normalizeID(id))
	if err != nil {
		return "", err
	}
	got := []byte(hashToken(token))
	var found domain.PlayerID
	for pid, h := range rec.TokenHashes {
		if subtle.ConstantTimeCompare(got, []byte(h)) == 1 {
			found = pid
		}
	}
	if found == "" {
		return "", ports.ErrUnauthorized
	}
	return found, nil
}

// Summary implements ports.GameService.
func (s *Service) Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error) {
	rec, err := s.repo.Load(ctx, normalizeID(id))
	if err != nil {
		return domain.GameSummary{}, err
	}
	return rec.Game.Summary(), nil
}

// View implements ports.GameService.
func (s *Service) View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	rec, err := s.repo.Load(ctx, normalizeID(id))
	if err != nil {
		return domain.PlayerView{}, err
	}
	return rec.Game.ViewFor(player)
}

// SetOptions implements ports.GameService.
func (s *Service) SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.SetOptions(player, opts) })
}

// StartGame implements ports.GameService.
func (s *Service) StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.Start(player, s.cfg.NewRNG()) })
}

// ChooseFirstPlayer implements ports.GameService.
func (s *Service) ChooseFirstPlayer(ctx context.Context, id domain.GameID, player, first domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.ChooseFirstPlayer(player, first) })
}

// PlayCard implements ports.GameService.
func (s *Service) PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card, pile int) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.PlayCard(player, card, pile) })
}

// UndoPlay implements ports.GameService.
func (s *Service) UndoPlay(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.UndoPlay(player) })
}

// GiveUp implements ports.GameService.
func (s *Service) GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.GiveUp(player) })
}

// EndTurn implements ports.GameService.
func (s *Service) EndTurn(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.EndTurn(player) })
}

// EndGame implements ports.GameService.
func (s *Service) EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	return s.act(ctx, id, func(g *domain.Game) error { return g.EndGame(player) })
}

// Rematch implements ports.GameService. It is idempotent: when the old game already
// has a NextGameID, nothing changes and nothing is published.
func (s *Service) Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	id = normalizeID(id)
	s.locks.Lock(id)
	defer s.locks.Unlock(id)
	for range maxOCCAttempts {
		rec, err := s.repo.Load(ctx, id)
		if err != nil {
			return err
		}
		if _, _, ok := rec.Game.PlayerByID(player); ok && rec.Game.NextGameID != "" {
			return nil
		}
		if err := s.createRematch(ctx, rec, player); err != nil {
			return err
		}
		events := rec.Game.PullEvents()
		err = s.repo.Update(ctx, rec)
		if errors.Is(err, ports.ErrVersionConflict) {
			continue
		}
		if err != nil {
			return fmt.Errorf("save game %s: %w", id, err)
		}
		s.pub.Publish(ctx, ports.GameUpdate{Game: rec.Game, Events: events})
		return nil
	}
	return ports.ErrVersionConflict
}

// createRematch creates the new lobby game and, on success, replaces rec.Game with the
// updated old game (carrying NextGameID).
func (s *Service) createRematch(ctx context.Context, rec *ports.GameRecord, player domain.PlayerID) error {
	for range maxIDAttempts {
		old := rec.Game.Clone() // fresh clone per attempt: Rematch mutates its receiver
		next, err := old.Rematch(player, s.cfg.NewGameID())
		if err != nil {
			return err
		}
		err = s.repo.Create(ctx, &ports.GameRecord{Game: next, TokenHashes: maps.Clone(rec.TokenHashes)})
		if errors.Is(err, ports.ErrDuplicateGameID) {
			continue
		}
		if err != nil {
			return fmt.Errorf("create rematch: %w", err)
		}
		rec.Game = old
		return nil
	}
	return fmt.Errorf("create rematch: no free game id after %d attempts: %w", maxIDAttempts, ports.ErrDuplicateGameID)
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
