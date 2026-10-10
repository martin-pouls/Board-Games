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

	"github.com/martin-pouls/regicide/backend/internal/core/domain"
	"github.com/martin-pouls/regicide/backend/internal/core/ports"
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
	locks *keyedMutex
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
	return &Service{repo: repo, pub: pub, cfg: cfg, locks: newKeyedMutex()}
}

// mutate runs fn on a freshly loaded record under the per-game lock, saves it with
// OCC retry and publishes the resulting events while still holding the lock.
func (s *Service) mutate(ctx context.Context, id domain.GameID, fn func(*ports.GameRecord) error) (*ports.GameRecord, error) {
	s.locks.lock(id)
	defer s.locks.unlock(id)
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

// act runs a domain mutator for player and returns the actor's view.
func (s *Service) act(ctx context.Context, id domain.GameID, player domain.PlayerID, fn func(*domain.Game) error) (domain.PlayerView, error) {
	rec, err := s.mutate(ctx, normalizeID(id), func(rec *ports.GameRecord) error { return fn(rec.Game) })
	if err != nil {
		return domain.PlayerView{}, err
	}
	return rec.Game.ViewFor(player)
}

// CreateGame implements ports.GameService.
func (s *Service) CreateGame(ctx context.Context, hostName string) (ports.Session, error) {
	hostID := s.cfg.NewPlayerID()
	token := s.cfg.NewToken()
	for range maxIDAttempts {
		g, err := domain.NewGame(s.cfg.NewGameID(), hostID, hostName)
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
	if _, _, ok := rec.Game.PlayerByID(found); !ok {
		return "", ports.ErrUnauthorized // the player left the game
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

// LeaveGame implements ports.GameService.
func (s *Service) LeaveGame(ctx context.Context, id domain.GameID, player domain.PlayerID) error {
	_, err := s.mutate(ctx, normalizeID(id), func(rec *ports.GameRecord) error { return rec.Game.Leave(player) })
	return err
}

// StartGame implements ports.GameService.
func (s *Service) StartGame(ctx context.Context, id domain.GameID, player, first domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.Start(player, first, s.cfg.NewRNG()) })
}

// Play implements ports.GameService.
func (s *Service) Play(ctx context.Context, id domain.GameID, player domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.Play(player, cards, s.cfg.NewRNG()) })
}

// Yield implements ports.GameService.
func (s *Service) Yield(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.Yield(player) })
}

// Discard implements ports.GameService.
func (s *Service) Discard(ctx context.Context, id domain.GameID, player domain.PlayerID, cards []domain.Card) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.Discard(player, cards) })
}

// ChooseNext implements ports.GameService.
func (s *Service) ChooseNext(ctx context.Context, id domain.GameID, player, next domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.ChooseNext(player, next) })
}

// UseSoloJester implements ports.GameService.
func (s *Service) UseSoloJester(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.UseSoloJester(player) })
}

// SetJesterSignal implements ports.GameService.
func (s *Service) SetJesterSignal(ctx context.Context, id domain.GameID, player domain.PlayerID, sig domain.JesterSignal) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.SetJesterSignal(player, sig) })
}

// GiveUp implements ports.GameService.
func (s *Service) GiveUp(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.GiveUp(player) })
}

// EndGame implements ports.GameService.
func (s *Service) EndGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	return s.act(ctx, id, player, func(g *domain.Game) error { return g.EndGame(player) })
}

// Rematch implements ports.GameService. It is idempotent: when the old game already
// has a NextGameID, the actor's view of the old game is returned unchanged.
func (s *Service) Rematch(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error) {
	id = normalizeID(id)
	s.locks.lock(id)
	defer s.locks.unlock(id)
	for range maxOCCAttempts {
		rec, err := s.repo.Load(ctx, id)
		if err != nil {
			return domain.PlayerView{}, err
		}
		if _, _, ok := rec.Game.PlayerByID(player); ok && rec.Game.NextGameID != "" {
			return rec.Game.ViewFor(player)
		}
		if err := s.createRematch(ctx, rec, player); err != nil {
			return domain.PlayerView{}, err
		}
		events := rec.Game.PullEvents()
		err = s.repo.Update(ctx, rec)
		if errors.Is(err, ports.ErrVersionConflict) {
			continue
		}
		if err != nil {
			return domain.PlayerView{}, fmt.Errorf("save game %s: %w", id, err)
		}
		s.pub.Publish(ctx, ports.GameUpdate{Game: rec.Game, Events: events})
		return rec.Game.ViewFor(player)
	}
	return domain.PlayerView{}, ports.ErrVersionConflict
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
