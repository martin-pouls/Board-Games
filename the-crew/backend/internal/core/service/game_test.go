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
// calls the domain, saves and publishes the events once.
func TestMissionFlow(t *testing.T) {
	ctx := context.Background()
	var rngCalls atomic.Int64
	f := newFixture(t, func(c *Config) {
		c.NewRNG = func() domain.RNG { rngCalls.Add(1); return noShuffle{} }
	})
	id, ss := f.lobby(t)
	host, guest := ss[0], ss[1]

	// do runs a command that must succeed; view reads a player's view afterwards.
	do := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	view := func(pid domain.PlayerID) domain.PlayerView {
		t.Helper()
		v, err := f.svc.View(ctx, id, pid)
		do(err)
		return v
	}

	// Lobby options: host only, saved and published; this flow plays with the distress signal on.
	if err := f.svc.SetOptions(ctx, id, guest.PlayerID, domain.Options{Distress: true}); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("guest sets options: %v", err)
	}
	do(f.svc.SetOptions(ctx, id, host.PlayerID, domain.Options{Distress: true}))
	if v := view(host.PlayerID); !v.Options.Distress {
		t.Fatalf("options = %+v", v.Options)
	}
	if got := eventTypes(f.rec.all()[len(f.rec.all())-1].Events); !reflect.DeepEqual(got, []string{"OptionsChanged"}) {
		t.Fatalf("events = %v", got)
	}
	if !f.loadGame(t, id).Options.Distress {
		t.Fatal("the option was not saved")
	}

	// Lobby: only the host selects and starts; a rejected intent is neither saved nor published.
	base := len(f.rec.all())
	if err := f.svc.SelectMission(ctx, id, guest.PlayerID, 4); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("guest selects: %v", err)
	}
	if err := f.svc.SelectMission(ctx, id, host.PlayerID, 51); !errors.Is(err, domain.ErrInvalidMission) {
		t.Fatalf("mission 51: %v", err)
	}
	if err := f.svc.StartMission(ctx, id, guest.PlayerID); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("guest starts: %v", err)
	}
	if got := len(f.rec.all()); got != base {
		t.Fatalf("rejected intents published %d updates", got-base)
	}
	do(f.svc.SelectMission(ctx, id, host.PlayerID, 4))
	if v := view(host.PlayerID); v.Mission != 4 {
		t.Fatalf("mission = %d", v.Mission)
	}
	do(f.svc.SelectMission(ctx, id, host.PlayerID, 1))
	v := view(host.PlayerID)
	if v.Mission != 1 {
		t.Fatalf("select failed: %+v", v)
	}
	pubs := f.rec.all()[base:]
	if len(pubs) != 2 || !reflect.DeepEqual(eventTypes(pubs[0].Events), []string{"MissionSelected"}) {
		t.Fatalf("publishes after select: %+v", pubs)
	}

	// Start: the deal uses the injected RNG.
	calls0 := rngCalls.Load()
	do(f.svc.StartMission(ctx, id, host.PlayerID))
	v = view(host.PlayerID)
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
		"AdjustTokens": f.svc.AdjustTokens(ctx, id, cmdID, 0, 1),
		"SkipTokens":   f.svc.SkipTokens(ctx, id, cmdID),
		"AnswerPoll":   f.svc.AnswerPoll(ctx, id, otherID, true),
		"PickSeats":    f.svc.PickSeats(ctx, id, cmdID, []int{1}),
	} {
		if !errors.Is(err, domain.ErrWrongStep) {
			t.Errorf("%s in the draft: %v", name, err)
		}
	}
	if err := f.svc.PlayCard(ctx, id, cmdID, domain.Card{Suit: domain.SuitPink, Rank: 1}); !errors.Is(err, domain.ErrWrongStep) {
		t.Errorf("PlayCard in the draft: %v", err)
	}

	// Draft: the commander takes the only task.
	before := len(f.rec.all())
	if err := f.svc.PickTask(ctx, id, otherID, 0); !errors.Is(err, domain.ErrNotYourTurn) {
		t.Fatalf("pick out of turn: %v", err)
	}
	do(f.svc.PickTask(ctx, id, cmdID, 0))
	v = view(cmdID)
	if v.Attempt.Step != domain.StepPlay || v.Attempt.Tasks[0].Owner != cmd {
		t.Fatalf("after the draft: %+v", v.Attempt)
	}
	if len(f.rec.all()) != before+1 {
		t.Fatal("one publish per accepted intent")
	}

	// Before the first trick: handover is not for this mission; the commander sends the signal.
	if err := f.svc.Handover(ctx, id, cmdID, 0, 1-cmd); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("handover: %v", err)
	}
	if err := f.svc.SendDistress(ctx, id, otherID, domain.DirLeft); !errors.Is(err, domain.ErrNotCommander) {
		t.Fatalf("distress by a crew member: %v", err)
	}
	if err := f.svc.SendDistress(ctx, id, cmdID, "up"); !errors.Is(err, domain.ErrInvalidDirection) {
		t.Fatalf("distress up: %v", err)
	}
	do(f.svc.SendDistress(ctx, id, cmdID, domain.DirLeft))
	v = view(cmdID)
	if v.Attempt.Step != domain.StepPass || v.Attempt.Pass == nil || !v.Attempt.Distress {
		t.Fatalf("after the signal: %+v", v.Attempt)
	}
	for _, pid := range []domain.PlayerID{cmdID, otherID} {
		hand := view(pid).Attempt.Hand
		var pick domain.Card
		for _, c := range hand {
			if !c.IsTrump() {
				pick = c
				break
			}
		}
		if err := f.svc.ChoosePassCard(ctx, id, pid, domain.Card{Suit: domain.SuitTrump, Rank: 4}); err == nil {
			t.Fatal("a trump cannot be passed")
		}
		do(f.svc.ChoosePassCard(ctx, id, pid, pick))
		v = view(pid)
	}
	if v.Attempt.Step != domain.StepPlay || v.Attempt.Exchange == nil {
		t.Fatalf("after the pass: %+v", v.Attempt)
	}
	if err := f.svc.SkipDistress(ctx, id, cmdID); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("second decision: %v", err)
	}

	// A trick: the commander leads; communicate before the next one.
	v = view(cmdID)
	lead := v.Attempt.Playable[0]
	do(f.svc.PlayCard(ctx, id, cmdID, lead))
	v = view(cmdID)
	if len(v.Attempt.Trick) != 1 || v.Attempt.Trick[0].Card != lead {
		t.Fatalf("trick = %+v", v.Attempt.Trick)
	}
	if err := f.svc.Communicate(ctx, id, otherID, domain.Card{Suit: domain.SuitPink, Rank: 1}, domain.PosHighest); !errors.Is(err, domain.ErrCannotCommunicate) {
		t.Fatalf("communicate during a trick: %v", err)
	}

	// Give up: any player; then the host deals the next attempt.
	do(f.svc.GiveUp(ctx, id, otherID))
	v = view(otherID)
	if v.Phase != domain.PhaseDebrief || v.Attempt.Result == nil || v.Attempt.Result.Reason != domain.ReasonGaveUp {
		t.Fatalf("after giving up: %+v", v)
	}
	if err := f.svc.NextMission(ctx, id, host.PlayerID); !errors.Is(err, domain.ErrInvalidResultChoice) {
		t.Fatalf("next mission after a loss: %v", err)
	}
	if err := f.svc.NextAttempt(ctx, id, guest.PlayerID, domain.NextAttemptChoice{}); !errors.Is(err, domain.ErrNotHost) {
		t.Fatalf("next attempt by a guest: %v", err)
	}
	calls := rngCalls.Load()
	do(f.svc.NextAttempt(ctx, id, host.PlayerID, domain.NextAttemptChoice{}))
	v = view(host.PlayerID)
	if v.Phase != domain.PhasePlaying || v.Attempt.Number != 2 || rngCalls.Load() != calls+1 {
		t.Fatalf("next attempt: %+v", v.Attempt)
	}
	if e := v.Logbook[1]; e.Attempts != 1 || e.Completed {
		t.Fatalf("logbook = %+v", e)
	}
	do(f.svc.GiveUp(ctx, id, host.PlayerID))
	do(f.svc.BackToLobby(ctx, id, host.PlayerID, false))
	v = view(host.PlayerID)
	if v.Phase != domain.PhaseLobby || v.Attempt != nil || v.Mission != 1 {
		t.Fatalf("back in the lobby: %+v", v)
	}
	do(f.svc.EndGame(ctx, id, host.PlayerID))
	v = view(host.PlayerID)
	if v.Phase != domain.PhaseEnded {
		t.Fatalf("phase = %s", v.Phase)
	}
}

// TestDistressIsOffByDefault: without the lobby option the commander cannot send or skip the
// signal, and the view does not offer it.
func TestDistressIsOffByDefault(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, func(c *Config) { c.NewRNG = func() domain.RNG { return noShuffle{} } })
	id, ss := f.lobby(t)
	if err := f.svc.StartMission(ctx, id, ss[0].PlayerID); err != nil {
		t.Fatal(err)
	}
	g := f.loadGame(t, id)
	if g.Options.Distress {
		t.Fatal("the signal must be off by default")
	}
	// Mission 1's draft comes first; let the commander take the task so the play step is reached.
	cmdID := g.Players[g.Attempt.Commander].ID
	if err := f.svc.PickTask(ctx, id, cmdID, 0); err != nil {
		t.Fatal(err)
	}
	v, err := f.svc.View(ctx, id, cmdID)
	if err != nil {
		t.Fatal(err)
	}
	if v.Attempt.Actions.SendDistress {
		t.Fatal("the commander is offered the signal")
	}
	if err := f.svc.SendDistress(ctx, id, cmdID, domain.DirLeft); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("send: %v", err)
	}
	if err := f.svc.SkipDistress(ctx, id, cmdID); !errors.Is(err, domain.ErrNotAllowed) {
		t.Fatalf("skip: %v", err)
	}
	// Options cannot change any more once the mission runs.
	if err := f.svc.SetOptions(ctx, id, ss[0].PlayerID, domain.Options{Distress: true}); !errors.Is(err, domain.ErrWrongPhase) {
		t.Fatalf("set options while playing: %v", err)
	}
}
