package service

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// loadGame returns the stored game.
func (f *fixture) loadGame(t *testing.T, id domain.GameID) *domain.Game {
	t.Helper()
	rec, err := f.repo.Load(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return rec.Game
}

// TestMissionFlow runs a 2-player mission 1 through the service: every intent loads the game,
// calls the domain, saves, publishes the events once and returns the actor's view.
func TestMissionFlow(t *testing.T) {
	ctx := context.Background()
	must := func(v domain.PlayerView, err error) domain.PlayerView {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	var rngCalls atomic.Int64
	f := newFixture(t, func(c *Config) {
		c.NewRNG = func() domain.RNG { rngCalls.Add(1); return noShuffle{} }
	})
	id, ss := f.lobby(t)
	host, guest := ss[0], ss[1]

	// Lobby: only the host selects and starts; a rejected intent is neither saved nor published.
	base := len(f.rec.all())
	if _, err := f.svc.SelectMission(ctx, id, guest.PlayerID, 4); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("guest selects: %v", err)
	}
	if _, err := f.svc.SelectMission(ctx, id, host.PlayerID, 51); !errors.Is(err, domain.ErrInvalidMission) {
		t.Fatalf("mission 51: %v", err)
	}
	if _, err := f.svc.StartMission(ctx, id, guest.PlayerID); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("guest starts: %v", err)
	}
	if got := len(f.rec.all()); got != base {
		t.Fatalf("rejected intents published %d updates", got-base)
	}
	if v := must(f.svc.SelectMission(ctx, id, host.PlayerID, 4)); v.Mission != 4 {
		t.Fatalf("mission = %d", v.Mission)
	}
	v := must(f.svc.SelectMission(ctx, id, host.PlayerID, 1))
	if v.Mission != 1 {
		t.Fatalf("select failed: %+v", v)
	}
	pubs := f.rec.all()[base:]
	if len(pubs) != 2 || !reflect.DeepEqual(eventTypes(pubs[0].Events), []string{"MissionSelected"}) {
		t.Fatalf("publishes after select: %+v", pubs)
	}

	// Start: the deal uses the injected RNG, the view is the starter's.
	calls0 := rngCalls.Load()
	v = must(f.svc.StartMission(ctx, id, host.PlayerID))
	if v.Phase != domain.PhasePlaying || v.Attempt == nil || v.Attempt.Seat != 0 || rngCalls.Load() != calls0+1 {
		t.Fatalf("after start: %+v (rng calls %d)", v.Attempt, rngCalls.Load())
	}
	last := f.rec.all()[len(f.rec.all())-1]
	if got := eventTypes(last.Events); got[0] != "AttemptStarted" {
		t.Fatalf("events = %v", got)
	}
	g := f.loadGame(t, id)
	cmd := g.Attempt.Commander
	cmdID := g.Players[cmd].ID
	otherID := g.Players[1-cmd].ID

	// Steps the attempt is not in are rejected with the domain's error.
	for name, err := range map[string]error{
		"AdjustTokens": second(f.svc.AdjustTokens(ctx, id, cmdID, 0, 1)),
		"SkipTokens":   second(f.svc.SkipTokens(ctx, id, cmdID)),
		"AnswerPoll":   second(f.svc.AnswerPoll(ctx, id, otherID, true)),
		"PickSeats":    second(f.svc.PickSeats(ctx, id, cmdID, []int{1})),
	} {
		if !errors.Is(err, domain.ErrWrongStep) {
			t.Errorf("%s in the draft: %v", name, err)
		}
	}
	if _, err := f.svc.PlayCard(ctx, id, cmdID, domain.Card{Suit: domain.SuitPink, Rank: 1}); !errors.Is(err, domain.ErrWrongStep) {
		t.Errorf("PlayCard in the draft: %v", err)
	}

	// Draft: the commander takes the only task.
	before := len(f.rec.all())
	if _, err := f.svc.PickTask(ctx, id, otherID, 0); !errors.Is(err, domain.ErrNotYourTurn) {
		t.Fatalf("pick out of turn: %v", err)
	}
	v = must(f.svc.PickTask(ctx, id, cmdID, 0))
	if v.Attempt.Step != domain.StepPlay || v.Attempt.Tasks[0].Owner != cmd {
		t.Fatalf("after the draft: %+v", v.Attempt)
	}
	if len(f.rec.all()) != before+1 {
		t.Fatal("one publish per accepted intent")
	}

	// Before the first trick: handover is not for this mission; the commander sends the signal.
	if _, err := f.svc.Handover(ctx, id, cmdID, 0, 1-cmd); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("handover: %v", err)
	}
	if _, err := f.svc.SendDistress(ctx, id, otherID, domain.DirLeft); !errors.Is(err, domain.ErrNotCommander) {
		t.Fatalf("distress by a crew member: %v", err)
	}
	if _, err := f.svc.SendDistress(ctx, id, cmdID, "up"); !errors.Is(err, domain.ErrInvalidDirection) {
		t.Fatalf("distress up: %v", err)
	}
	v = must(f.svc.SendDistress(ctx, id, cmdID, domain.DirLeft))
	if v.Attempt.Step != domain.StepPass || v.Attempt.Pass == nil || !v.Attempt.Distress {
		t.Fatalf("after the signal: %+v", v.Attempt)
	}
	for _, pid := range []domain.PlayerID{cmdID, otherID} {
		hand := must(f.svc.View(ctx, id, pid)).Attempt.Hand
		var pick domain.Card
		for _, c := range hand {
			if !c.IsTrump() {
				pick = c
				break
			}
		}
		if _, err := f.svc.ChoosePassCard(ctx, id, pid, domain.Card{Suit: domain.SuitTrump, Rank: 4}); err == nil {
			t.Fatal("a trump cannot be passed")
		}
		v = must(f.svc.ChoosePassCard(ctx, id, pid, pick))
	}
	if v.Attempt.Step != domain.StepPlay || v.Attempt.Exchange == nil {
		t.Fatalf("after the pass: %+v", v.Attempt)
	}
	if _, err := f.svc.SkipDistress(ctx, id, cmdID); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("second decision: %v", err)
	}

	// A trick: the commander leads; communicate before the next one.
	v = must(f.svc.View(ctx, id, cmdID))
	lead := v.Attempt.Playable[0]
	v = must(f.svc.PlayCard(ctx, id, cmdID, lead))
	if len(v.Attempt.Trick) != 1 || v.Attempt.Trick[0].Card != lead {
		t.Fatalf("trick = %+v", v.Attempt.Trick)
	}
	if _, err := f.svc.Communicate(ctx, id, otherID, domain.Card{Suit: domain.SuitPink, Rank: 1}, domain.PosHighest); !errors.Is(err, domain.ErrCannotCommunicate) {
		t.Fatalf("communicate during a trick: %v", err)
	}

	// Give up: any player; then the host deals the next attempt.
	v = must(f.svc.GiveUp(ctx, id, otherID))
	if v.Phase != domain.PhaseDebrief || v.Attempt.Result == nil || v.Attempt.Result.Reason != domain.ReasonGaveUp {
		t.Fatalf("after giving up: %+v", v)
	}
	if _, err := f.svc.NextMission(ctx, id, host.PlayerID); !errors.Is(err, domain.ErrInvalidResultChoice) {
		t.Fatalf("next mission after a loss: %v", err)
	}
	if _, err := f.svc.NextAttempt(ctx, id, guest.PlayerID); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("next attempt by a guest: %v", err)
	}
	calls := rngCalls.Load()
	v = must(f.svc.NextAttempt(ctx, id, host.PlayerID))
	if v.Phase != domain.PhasePlaying || v.Attempt.Number != 2 || rngCalls.Load() != calls+1 {
		t.Fatalf("next attempt: %+v", v.Attempt)
	}
	if e := v.Logbook[1]; e.Attempts != 1 || e.Completed {
		t.Fatalf("logbook = %+v", e)
	}
	must(f.svc.GiveUp(ctx, id, host.PlayerID))
	v = must(f.svc.BackToLobby(ctx, id, host.PlayerID))
	if v.Phase != domain.PhaseLobby || v.Attempt != nil || v.Mission != 1 {
		t.Fatalf("back in the lobby: %+v", v)
	}
	v = must(f.svc.EndGame(ctx, id, host.PlayerID))
	if v.Phase != domain.PhaseEnded {
		t.Fatalf("phase = %s", v.Phase)
	}
}
