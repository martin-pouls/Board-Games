package domain

import (
	"reflect"
	"testing"
)

func TestPlayEffectAt(t *testing.T) {
	tests := []struct {
		name   string
		enemy  string
		played []played
		i      int
		want   PlayEffect
	}{
		{"single", "JS", []played{{0, "7H"}}, 0,
			PlayEffect{Value: 7, Suits: []Suit{SuitHearts}, Damage: 7}},
		{"clubs double", "JS", []played{{0, "7C"}}, 0,
			PlayEffect{Value: 7, Suits: []Suit{SuitClubs}, Doubled: true, Damage: 14}},
		{"same suit once", "JS", []played{{0, "AC 8C"}}, 0,
			PlayEffect{Value: 9, Suits: []Suit{SuitClubs}, Doubled: true, Damage: 18}},
		{"rules example 1", "JH", []played{{0, "3D 3S 3C"}}, 0,
			PlayEffect{Value: 9, Suits: []Suit{SuitDiamonds, SuitClubs, SuitSpades}, Doubled: true, Damage: 18}},
		{"clubs vs clubs", "JC", []played{{0, "7C"}}, 0,
			PlayEffect{Value: 7, Suits: []Suit{SuitClubs}, BlockedSuits: []Suit{SuitClubs}, Damage: 7}},
		{"clubs vs clubs after jester", "JC", []played{{1, "X"}, {0, "7C"}}, 1,
			PlayEffect{Value: 7, Suits: []Suit{SuitClubs}, Doubled: true, Damage: 14}},
		{"clubs vs clubs before jester stays single", "JC", []played{{0, "7C"}, {1, "X"}}, 0,
			PlayEffect{Value: 7, Suits: []Suit{SuitClubs}, BlockedSuits: []Suit{SuitClubs}, Damage: 7}},
		{"hearts vs hearts before jester stays blocked", "JH", []played{{0, "5H 5D"}, {1, "X"}}, 0,
			PlayEffect{Value: 10, Suits: []Suit{SuitHearts, SuitDiamonds}, BlockedSuits: []Suit{SuitHearts}, Damage: 10}},
		{"spades vs spades blocked", "QS", []played{{0, "4S 4H"}}, 0,
			PlayEffect{Value: 8, Suits: []Suit{SuitHearts, SuitSpades}, BlockedSuits: []Suit{SuitSpades}, Damage: 8}},
		{"spades vs spades retroactive after jester", "QS", []played{{0, "4S 4H"}, {1, "X"}}, 0,
			PlayEffect{Value: 8, Suits: []Suit{SuitHearts, SuitSpades}, Damage: 8}},
		{"jester", "QS", []played{{1, "X"}}, 0, PlayEffect{Jester: true}},
		{"companion with king", "KD", []played{{0, "AD KC"}}, 0,
			PlayEffect{Value: 21, Suits: []Suit{SuitDiamonds, SuitClubs}, BlockedSuits: []Suit{SuitDiamonds}, Doubled: true, Damage: 42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{hands: []string{"", ""}, enemy: tt.enemy, played: tt.played}.build(t)
			if got := g.PlayEffectAt(tt.i); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("PlayEffectAt = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDerivedTotals(t *testing.T) {
	tests := []struct {
		name                   string
		enemy                  string
		played                 []played
		damage, shield, attack int
		immune                 Suit
		jester                 bool
	}{
		{"fresh", "JH", nil, 0, 0, 10, SuitHearts, false},
		{"shield cumulative across players", "JH", []played{{0, "3S"}, {1, "4S 4D"}}, 11, 11, 0, SuitHearts, false},
		{"shield capped at 0 attack", "JD", []played{{0, "10S"}, {1, "5S"}}, 15, 15, 0, SuitDiamonds, false},
		{"partial shield", "KC", []played{{0, "7S"}}, 7, 7, 13, SuitClubs, false},
		{"spades immune", "QS", []played{{0, "7S"}, {1, "5C"}}, 17, 0, 15, SuitSpades, false},
		{"spades retroactive after jester", "QS", []played{{0, "7S"}, {1, "X"}, {0, "2S 2H"}}, 11, 11, 4, SuitNone, true},
		{"clubs doubling", "JS", []played{{0, "5C"}, {1, "AC 4D"}}, 20, 0, 10, SuitSpades, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{hands: []string{"", ""}, enemy: tt.enemy, played: tt.played}.build(t)
			if g.Damage() != tt.damage || g.Shield() != tt.shield || g.EffectiveAttack() != tt.attack ||
				g.ImmuneSuit() != tt.immune || g.JesterPlayed() != tt.jester {
				t.Fatalf("damage %d shield %d attack %d immune %q jester %v", g.Damage(), g.Shield(),
					g.EffectiveAttack(), g.ImmuneSuit(), g.JesterPlayed())
			}
		})
	}
	g := lobbyWith(t, 2)
	if g.Damage() != 0 || g.Shield() != 0 || g.EffectiveAttack() != 0 || g.ImmuneSuit() != SuitNone {
		t.Fatal("derived values without an enemy")
	}
}

func TestCanYield(t *testing.T) {
	tests := []struct {
		name    string
		players int
		yielded []int
		seat    int
		want    bool
	}{
		{"solo always", 1, []int{0}, 0, true},
		{"2p fresh", 2, nil, 0, true},
		{"2p other yielded", 2, []int{1}, 0, false},
		{"2p self yielded only", 2, []int{0}, 0, true},
		{"3p two others yielded", 3, []int{1, 2}, 0, false},
		{"3p one other yielded", 3, []int{1}, 0, true},
		{"4p all yielded", 4, []int{0, 1, 2, 3}, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{hands: make([]string, tt.players), enemy: "JH", yielded: tt.yielded}.build(t)
			if got := g.CanYield(tt.seat); got != tt.want {
				t.Fatalf("CanYield = %v", got)
			}
		})
	}
}

func TestGrade(t *testing.T) {
	for used, want := range []Grade{GradeGold, GradeSilver, GradeBronze} {
		g := table{hands: []string{""}, soloUsed: used}.build(t)
		g.Phase = PhaseWon
		if got := g.Grade(); got != want {
			t.Errorf("used %d: %q", used, got)
		}
		g.Phase = PhaseLost
		if g.Grade() != GradeNone {
			t.Errorf("lost solo has a grade")
		}
	}
	g := table{hands: []string{"", ""}}.build(t)
	g.Phase = PhaseWon
	if g.Grade() != GradeNone {
		t.Fatal("multiplayer win has a grade")
	}
}

func TestEnemiesDefeated(t *testing.T) {
	if lobbyWith(t, 2).EnemiesDefeated() != 0 {
		t.Fatal("lobby")
	}
	if n := started(t, 2).EnemiesDefeated(); n != 0 {
		t.Fatalf("fresh game %d", n)
	}
	g := table{hands: []string{""}, enemy: "QH", castle: "KH KD KC KS QD QS"}.build(t)
	if n := g.EnemiesDefeated(); n != 5 {
		t.Fatalf("mid game %d", n)
	}
	g = table{hands: []string{""}}.build(t)
	g.Phase = PhaseWon
	if n := g.EnemiesDefeated(); n != 12 {
		t.Fatalf("won %d", n)
	}
}
