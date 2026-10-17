// Package gametest drives a domain.Game through the setup and the tricks with simple legal
// choices. It is a test helper for the adapters' tests: it builds mid-attempt and debrief games
// without a service, a repository or a socket.
package gametest

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/martin-pouls/the-crew/backend/internal/core/domain"
)

// NewRNG returns the deterministic RNG for seed.
func NewRNG(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)) }

// PlayerID is the ID NewGame gives the player in seat: p1, p2, ...
func PlayerID(seat int) domain.PlayerID { return domain.PlayerID(fmt.Sprintf("p%d", seat+1)) }

// NewGame creates a lobby of n players (Ann, Ben, Cleo, Dan, Eve; p1 hosts), selects mission and
// starts it with rng. The options are the defaults (distress signal off). The returned game has no
// pending events.
func NewGame(t testing.TB, n, mission int, rng domain.RNG) *domain.Game {
	t.Helper()
	return NewGameWith(t, n, mission, domain.Options{}, rng)
}

// NewGameWith is NewGame with the lobby options opts set by the host.
func NewGameWith(t testing.TB, n, mission int, opts domain.Options, rng domain.RNG) *domain.Game {
	t.Helper()
	names := []string{"Ann", "Ben", "Cleo", "Dan", "Eve"}
	g, err := domain.NewGame("GAME1", domain.EditionPlanetNine, PlayerID(0), names[0], domain.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < n; i++ {
		if err := g.AddPlayer(PlayerID(i), names[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetOptions(PlayerID(0), opts); err != nil {
		t.Fatal(err)
	}
	if err := g.SelectMission(PlayerID(0), mission); err != nil {
		t.Fatal(err)
	}
	if err := g.StartMission(PlayerID(0), rng); err != nil {
		t.Fatal(err)
	}
	g.PullEvents()
	return g
}

// Step makes one legal move for the first actor of the attempt: skips the distress signal and the
// token swap, answers polls with yes, picks the first seat that is accepted, takes the first free
// task, passes the first non-trump card and plays the first legal card. It fails the test when
// the game is not playing or the move is rejected.
func Step(t testing.TB, g *domain.Game, rng domain.RNG) {
	t.Helper()
	if g.Phase != domain.PhasePlaying {
		t.Fatalf("step in phase %s", g.Phase)
	}
	a := g.Attempt
	seat := g.Actors()[0]
	by := g.Players[seat].ID
	var err error
	switch a.Step {
	case domain.StepAdjustTokens:
		err = g.SkipTokens(by)
	case domain.StepPoll:
		err = g.AnswerPoll(by, true)
	case domain.StepPick:
		err = pickSeats(g, by)
	case domain.StepDraft:
		for _, task := range a.Tasks {
			if task.Owner() == domain.OwnerUnassigned {
				err = g.PickTask(by, task.ID)
				break
			}
		}
	case domain.StepPass:
		for _, c := range a.Members[seat].Hand {
			if !c.IsTrump() {
				err = g.ChoosePassCard(by, c)
				break
			}
		}
	case domain.StepPlay:
		if g.CanSendDistress(seat) == nil {
			err = g.SkipDistress(by)
			break
		}
		turn := (a.Trick.Leader + len(a.Trick.Plays)) % len(a.Members)
		err = g.PlayCard(by, g.LegalCards(turn)[0], rng)
	default:
		t.Fatalf("unknown step %q", a.Step)
	}
	if err != nil {
		t.Fatalf("step %s (seat %d): %v", a.Step, seat, err)
	}
}

// pickSeats tries the seats in order until the domain accepts one (a rejected pick changes
// nothing), and for the roles of mission 50 the first accepted pair.
func pickSeats(g *domain.Game, by domain.PlayerID) error {
	n := len(g.Attempt.Members)
	var last error
	for s := range n {
		if last = g.PickSeats(by, []int{s}); last == nil {
			return nil
		}
	}
	for s := range n {
		for u := range n {
			if last = g.PickSeats(by, []int{s, u}); last == nil {
				return nil
			}
		}
	}
	return last
}

// Until steps until done reports true. It fails the test when the attempt ends first or after
// 5000 steps.
func Until(t testing.TB, g *domain.Game, rng domain.RNG, done func(*domain.Game) bool) {
	t.Helper()
	if !TryUntil(t, g, rng, done) {
		t.Fatal("condition not reached before the attempt ended")
	}
}

// TryUntil steps until done reports true and returns true. It returns false when the attempt
// ends first.
func TryUntil(t testing.TB, g *domain.Game, rng domain.RNG, done func(*domain.Game) bool) bool {
	t.Helper()
	for range 5000 {
		if done(g) {
			g.PullEvents()
			return true
		}
		if g.Phase != domain.PhasePlaying {
			return false
		}
		Step(t, g, rng)
	}
	t.Fatal("condition never reached")
	return false
}

// ToEnd plays the attempt to the debrief.
func ToEnd(t testing.TB, g *domain.Game, rng domain.RNG) {
	t.Helper()
	for range 5000 {
		if g.Phase == domain.PhaseDebrief {
			g.PullEvents()
			return
		}
		Step(t, g, rng)
	}
	t.Fatal("attempt never ended")
}
