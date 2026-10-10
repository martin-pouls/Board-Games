package domain

import (
	"reflect"
	"slices"
	"testing"
)

func TestViewLobby(t *testing.T) {
	g := lobbyWith(t, 2)
	v, err := g.ViewFor(pid(1))
	mustOK(t, err)
	want := PlayerView{
		GameID:   "LOBBY",
		Revision: 2,
		Phase:    PhaseLobby,
		You:      pid(1),
		HostID:   pid(0),
		Rules:    RulesView{MinPlayers: 1, MaxPlayers: 4, MaxHandSize: 7},
		Players: []PlayerSummary{
			{ID: pid(0), Name: "Player 0", Seat: 0, IsHost: true},
			{ID: pid(1), Name: "Player 1", Seat: 1, IsYou: true},
		},
		Hand: []Card{},
	}
	if !reflect.DeepEqual(v, want) {
		t.Fatalf("view =\n%+v\nwant\n%+v", v, want)
	}
	solo, err := lobbyWith(t, 1).ViewFor(pid(0))
	mustOK(t, err)
	if solo.Rules != (RulesView{MinPlayers: 1, MaxPlayers: 4, MaxHandSize: 8, SoloJesters: 2}) || solo.SoloJestersLeft != 2 {
		t.Fatalf("solo rules %+v", solo.Rules)
	}
}

func TestViewPlaying(t *testing.T) {
	g := table{
		hands:   []string{"5C 7D", "KH 2S 3S", "X"},
		enemy:   "QS",
		castle:  "KH KD",
		tavern:  "2H 3H 4H",
		discard: "9C",
		played:  []played{{1, "4S"}, {2, "X"}, {0, "3C 3H"}},
		seat:    1,
		step:    StepDiscard,
		yielded: []int{2},
	}.build(t)
	g.Players[0].Signal = SignalGoodPlay
	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	if !slices.Equal(v.Hand, cs("5C 7D")) || v.Phase != PhasePlaying || v.CastleCount != 2 || v.TavernCount != 3 ||
		v.DiscardCount != 1 || v.SoloJestersLeft != 0 || v.Result != nil || v.Rules.MaxHandSize != 6 || v.Rules.TavernJesters != 1 {
		t.Fatalf("view %+v", v)
	}
	for i, p := range v.Players {
		if p.RevealedHand != nil || p.HandCount != len(g.Players[i].Hand) || p.Seat != i || p.IsActive != (i == 1) ||
			p.IsYou != (i == 0) || p.IsHost != (i == 0) || p.YieldedLastTurn != (i == 2) {
			t.Fatalf("player %d: %+v", i, p)
		}
	}
	if v.Players[0].Signal != SignalGoodPlay {
		t.Fatal("signal")
	}
	wantEnemy := &EnemyView{
		Card:              c("QS"),
		Attack:            15,
		Health:            30,
		Damage:            4 + 12,
		Shield:            4,
		EffectiveAttack:   11,
		Immunity:          SuitSpades,
		ImmunityCancelled: true,
		Plays: []PlayedSetView{
			{By: pid(1), Cards: cs("4S"), Value: 4, Damage: 4},
			{By: pid(2), Cards: cs("X"), Jester: true},
			{By: pid(0), Cards: sorted("3C 3H"), Value: 6, Damage: 12, Doubled: true},
		},
	}
	if !reflect.DeepEqual(v.Enemy, wantEnemy) {
		t.Fatalf("enemy =\n%+v\nwant\n%+v", v.Enemy, wantEnemy)
	}
	if *v.Turn != (TurnView{Number: 1, ActivePlayerID: pid(1), Step: StepDiscard, DamageToCover: 11}) {
		t.Fatalf("turn %+v", v.Turn)
	}
	// The view never aliases the game.
	v.Hand[0] = c("KS")
	v.Enemy.Plays[0].Cards[0] = c("KS")
	if g.Players[0].Hand[0] != c("5C") || g.Played[0].Cards[0] != c("4S") {
		t.Fatal("view aliases the game")
	}
}

func TestViewTurnCanYield(t *testing.T) {
	g := table{hands: []string{"5C", "6C"}, enemy: "JH", yielded: []int{1}}.build(t)
	v, _ := g.ViewFor(pid(1))
	if v.Turn.CanYield || v.Turn.DamageToCover != 0 {
		t.Fatalf("turn %+v", v.Turn)
	}
	g = table{hands: []string{"5C", "6C"}, enemy: "JH"}.build(t)
	v, _ = g.ViewFor(pid(1))
	if !v.Turn.CanYield {
		t.Fatalf("turn %+v", v.Turn)
	}
	g.Step = StepChooseNext
	v, _ = g.ViewFor(pid(1))
	if v.Turn.CanYield {
		t.Fatal("CanYield outside Step play")
	}
}

func TestViewEnemyBlockedSuits(t *testing.T) {
	g := table{hands: []string{"5C", "6C"}, enemy: "JH", played: []played{{0, "5H 5C"}}}.build(t)
	v, _ := g.ViewFor(pid(0))
	if p := v.Enemy.Plays[0]; !slices.Equal(p.BlockedSuits, []Suit{SuitHearts}) || !p.Doubled || p.Damage != 20 ||
		v.Enemy.Immunity != SuitHearts || v.Enemy.ImmunityCancelled {
		t.Fatalf("play %+v", p)
	}
}

func TestViewLost(t *testing.T) {
	g := table{hands: []string{"5C 7D", "KH"}, enemy: "QS", castle: "KH", seat: 1}.build(t)
	mustOK(t, g.GiveUp(pid(1)))
	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	if !slices.Equal(v.Players[1].RevealedHand, cs("KH")) || !slices.Equal(v.Players[0].RevealedHand, cs("5C 7D")) {
		t.Fatalf("revealed %+v", v.Players)
	}
	want := ResultView{Outcome: PhaseLost, Reason: LossGaveUp, PlayerID: pid(1), EnemiesDefeated: 10}
	if v.Result == nil || *v.Result != want || v.Enemy == nil || v.Turn.Step != StepNone || v.Players[1].IsActive {
		t.Fatalf("result %+v enemy %v turn %+v", v.Result, v.Enemy, v.Turn)
	}
}

func TestViewWonSolo(t *testing.T) {
	g := table{hands: []string{"2H"}, enemy: "KS", played: []played{{0, "10H"}, {0, "10D"}, {0, "9C"}}, soloUsed: 2}.build(t)
	mustOK(t, g.Play(pid(0), cs("2H"), noShuffle{}))
	v, err := g.ViewFor(pid(0))
	mustOK(t, err)
	want := ResultView{Outcome: PhaseWon, Grade: GradeBronze, EnemiesDefeated: 12, SoloJestersUsed: 2}
	if v.Result == nil || *v.Result != want || v.Enemy != nil || v.Players[0].RevealedHand == nil || v.SoloJestersLeft != 0 {
		t.Fatalf("view %+v result %+v", v, v.Result)
	}
}

func TestViewUnknownPlayer(t *testing.T) {
	if _, err := started(t, 2).ViewFor("zz"); err == nil {
		t.Fatal("no error")
	} else {
		wantErr(t, err, ErrUnknownPlayer)
	}
}
