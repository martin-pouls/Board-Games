package domain

import (
	"reflect"
	"testing"
)

// Extreme card values used throughout the command tests (rules.md §2):
//
//	STOP 6 35 57 95 · Skull 17 41 61 87 · 3! 19 28 53 81
//	No communication 26 65 78 92 · No backwards trick 21 32 59 70
//	Single pile 9 47 68 84 · Draw only one 13 38 43 76
//
// Piles 0 and 1 are ascending (start 1), piles 2 and 3 descending (start 100).

func TestDeriveCommands(t *testing.T) {
	none := commandState{boundPile: -1}
	tests := []struct {
		name  string
		piles [][]int
		plays []Play
		want  commandState
	}{
		{"no tops", nil, nil, none},
		{"plain tops", [][]int{{10}, {20}, {90}, {80}}, nil, none},
		{"no backwards trick top", [][]int{nil, nil, {70}}, nil,
			commandState{active: []ActiveCommand{{CommandNoBackwardsTrick, 2}}, noBackwardsTrick: true, boundPile: -1}},
		{"single pile top: unbound before the first card", [][]int{nil, {9}}, nil,
			commandState{active: []ActiveCommand{{CommandSinglePile, 1}}, singlePile: true, boundPile: -1}},
		{"draw only one top", [][]int{{13}}, nil,
			commandState{active: []ActiveCommand{{CommandDrawOnlyOne, 0}}, drawOnlyOne: true, boundPile: -1}},
		{"no communication top", [][]int{nil, nil, nil, {92}}, nil,
			commandState{active: []ActiveCommand{{CommandNoCommunication, 3}}, noCommunication: true, boundPile: -1}},
		{"covered infinity cards mean nothing", [][]int{{21, 25}, {9, 10}, {76, 75}, {92, 91}}, nil, none},
		{"two different and two identical infinity tops, pile order", [][]int{{21}, {32}, {92}, {76}}, nil,
			commandState{
				active: []ActiveCommand{
					{CommandNoBackwardsTrick, 0}, {CommandNoBackwardsTrick, 1},
					{CommandNoCommunication, 2}, {CommandDrawOnlyOne, 3},
				},
				noBackwardsTrick: true, noCommunication: true, drawOnlyOne: true, boundPile: -1,
			}},
		{"lightning tops from an earlier turn mean nothing", [][]int{{41}, {95}, {81}, {87}}, nil, none},
		{"skull played this turn on top", [][]int{{30}, {41}}, []Play{play(30, 0), play(41, 1)},
			commandState{skullPiles: []int{1}, boundPile: -1}},
		{"skull played this turn, covered", [][]int{{41, 45}}, []Play{play(41, 0), play(45, 0)}, none},
		{"skull on skull: the new skull counts", [][]int{nil, {41, 61}}, []Play{play(41, 1), play(61, 1)},
			commandState{skullPiles: []int{1}, boundPile: -1}},
		{"earlier skull covered by a skull this turn", [][]int{nil, {41, 61}}, []Play{play(61, 1)},
			commandState{skullPiles: []int{1}, boundPile: -1}},
		{"two skulls this turn on different piles", [][]int{{17}, {30, 41}}, []Play{play(17, 0), play(41, 1)},
			commandState{skullPiles: []int{0, 1}, boundPile: -1}},
		{"skull from an earlier turn, another pile touched this turn", [][]int{{41}, {30}}, []Play{play(30, 1)}, none},
		{"STOP in plays", [][]int{{95}}, []Play{play(95, 0)}, commandState{stopPlayed: true, boundPile: -1}},
		{"STOP in plays, covered", [][]int{{95, 96}}, []Play{play(95, 0), play(96, 0)}, commandState{stopPlayed: true, boundPile: -1}},
		{"3! in plays", [][]int{{81}}, []Play{play(81, 0)}, commandState{exactCount: 3, boundPile: -1}},
		{"3! in plays, covered", [][]int{{81, 85}}, []Play{play(81, 0), play(85, 0)}, commandState{exactCount: 3, boundPile: -1}},
		{"single pile at start, first play on another pile binds it", [][]int{{30}, {9}}, []Play{play(30, 0)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 1}}, singlePile: true, boundPile: 0}},
		{"single pile at start, first play covers it: inactive (Q1)", [][]int{nil, {9, 10}}, []Play{play(10, 1)}, none},
		{"single pile played mid-turn binds to its own pile", [][]int{{30}, nil, {68}}, []Play{play(30, 0), play(68, 2)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 2}}, singlePile: true, boundPile: 2}},
		{"single pile played mid-turn, then covered: free (Q1)", [][]int{{30}, nil, {68, 60}}, []Play{play(30, 0), play(68, 2), play(60, 2)}, none},
		{"two single piles at start, one covered by the first card: first card's pile", [][]int{nil, {9, 10}, nil, {84}}, []Play{play(10, 1)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 3}}, singlePile: true, boundPile: 1}},
		{"single pile at start, another played on the bound pile", [][]int{{30, 47}, nil, nil, {84}}, []Play{play(30, 0), play(47, 0)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 0}, {CommandSinglePile, 3}}, singlePile: true, boundPile: 0}},
		{"single pile at start, another played and covered on the bound pile", [][]int{{30, 47, 50}, nil, nil, {84}}, []Play{play(30, 0), play(47, 0), play(50, 0)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 3}}, singlePile: true, boundPile: 0}},
		{"single pile played as the first card", [][]int{{47}, nil, nil, {84}}, []Play{play(47, 0)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 0}, {CommandSinglePile, 3}}, singlePile: true, boundPile: 0}},
		{"single pile played on a skull pile, skull covered", [][]int{{41, 47}}, []Play{play(41, 0), play(47, 0)},
			commandState{active: []ActiveCommand{{CommandSinglePile, 0}}, singlePile: true, boundPile: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := table{opts: extreme, hands: [][]int{{50}}, piles: tt.piles, plays: tt.plays}.build(t)
			if got := g.commands(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("commands()\n got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

// TestDeriveCommandsClassic: classic cards never derive any command state.
func TestDeriveCommandsClassic(t *testing.T) {
	g := table{hands: [][]int{{50}}, piles: [][]int{{41}, {9, 95}, {70}, {92}}, plays: []Play{play(95, 1)}}.build(t)
	if got := g.commands(); !reflect.DeepEqual(got, commandState{boundPile: -1}) {
		t.Fatalf("classic commands() = %+v", got)
	}
}

func TestPlayedThisTurn(t *testing.T) {
	plays := []Play{play(30, 0), play(41, 2)}
	for pile, want := range []bool{true, false, true, false} {
		if got := playedThisTurn(plays, pile); got != want {
			t.Errorf("playedThisTurn(pile %d) = %v, want %v", pile, got, want)
		}
	}
	if playedThisTurn(nil, 0) {
		t.Error("playedThisTurn(nil) = true")
	}
}

// TestWithPlayDoesNotMutate: the projection never changes the game it was taken from.
func TestWithPlayDoesNotMutate(t *testing.T) {
	g := table{opts: extreme, hands: [][]int{{30, 41, 95}, {50}}, draw: []int{60}, piles: [][]int{{20}}, plays: []Play{play(20, 0)}}.build(t)
	g.Turn.Plays = append(make([]Play, 0, 8), g.Turn.Plays...) // spare capacity must not be shared
	g.Piles[0].Cards = append(make([]Card, 0, 8), g.Piles[0].Cards...)
	before := g.Clone()
	after := g.withPlay(Placement{Card: xcard(41), PileIndex: 0, Pile: g.Piles[0], Kind: PlacementNormal})
	if !reflect.DeepEqual(g, before) {
		t.Fatalf("withPlay mutated the game:\n%+v\n%+v", before, g)
	}
	if got := values(after.Players[0].Hand); !reflect.DeepEqual(got, []int{30, 95}) {
		t.Fatalf("after hand %v", got)
	}
	if got := values(after.Piles[0].Cards); !reflect.DeepEqual(got, []int{20, 41}) {
		t.Fatalf("after pile %v", got)
	}
	if len(after.Turn.Plays) != 2 || after.Turn.Plays[1] != (Play{Card: xcard(41), Pile: 0}) || after.events != nil {
		t.Fatalf("after turn %+v", after.Turn)
	}
	// A second projection from the same game is independent of the first.
	other := g.withPlay(Placement{Card: xcard(30), PileIndex: 0, Pile: g.Piles[0], Kind: PlacementNormal})
	if after.Piles[0].Cards[1].Value != 41 || other.Piles[0].Cards[1].Value != 30 || after.Turn.Plays[1].Card.Value != 41 {
		t.Fatal("projections share their appended state")
	}
}
