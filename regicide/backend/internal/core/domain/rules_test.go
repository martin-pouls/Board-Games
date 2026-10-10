package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestClassifyPlay(t *testing.T) {
	tests := []struct {
		cards  string
		want   PlayKind
		reason string // "" = legal
	}{
		{"", "", "empty"},
		// singles
		{"7H", PlaySingle, ""},
		{"AS", PlaySingle, ""},
		{"JC", PlaySingle, ""},
		{"QD", PlaySingle, ""},
		{"KH", PlaySingle, ""},
		{"X", PlayJester, ""},
		// combos
		{"2H 2S", PlayCombo, ""},
		{"3H 3S", PlayCombo, ""},
		{"4H 4S", PlayCombo, ""},
		{"5H 5S", PlayCombo, ""},
		{"6H 6S", "", "comboTooHigh"},
		{"10H 10S", "", "comboTooHigh"},
		{"2H 2S 2D", PlayCombo, ""},
		{"3H 3S 3D", PlayCombo, ""},
		{"4H 4S 4D", "", "comboTooHigh"},
		{"2H 2S 2D 2C", PlayCombo, ""},
		{"3H 3S 3D 3C", "", "comboTooHigh"},
		{"2H 3S", "", "mixedRanks"},
		{"JH JS", "", "royalCombo"},
		{"2H KS", "", "royalCombo"},
		// animal companions
		{"AH 9S", PlayCompanion, ""},
		{"AH AS", PlayCompanion, ""},
		{"AH KS", PlayCompanion, ""},
		{"AD JD", PlayCompanion, ""},
		{"AH X", "", "jesterNotAlone"},
		{"AH 2S 2D", "", "companionPair"},
		{"AH AS AD", "", "companionPair"},
		// jester
		{"X 5H", "", "jesterNotAlone"},
		{"X X", "", "jesterNotAlone"},
	}
	for _, tt := range tests {
		t.Run(tt.cards, func(t *testing.T) {
			got, err := ClassifyPlay(cs(tt.cards))
			if tt.reason == "" {
				if err != nil || got != tt.want {
					t.Fatalf("ClassifyPlay = %q, %v; want %q", got, err, tt.want)
				}
				return
			}
			if !errors.Is(err, ErrIllegalPlay) || got != "" || !strings.HasSuffix(err.Error(), ": "+tt.reason) {
				t.Fatalf("ClassifyPlay = %q, %v; want reason %q", got, err, tt.reason)
			}
		})
	}
	// More than four of a rank cannot occur with one deck; the check still exists (order: after mixedRanks).
	if _, err := ClassifyPlay([]Card{c("2H"), c("2H"), c("2H"), c("2H"), c("2H")}); err == nil || !strings.HasSuffix(err.Error(), "tooManyCards") {
		t.Fatalf("five 2s: %v", err)
	}
}

func TestValidateDiscard(t *testing.T) {
	tests := []struct {
		name   string
		cards  string
		attack int
		reason string
	}{
		{"exact single", "10H", 10, ""},
		{"single over", "KH", 10, ""},
		{"two needed", "5H 7C", 10, ""},
		{"two needed with ace", "AH 9C", 10, ""},
		{"superfluous small card", "AH 10C", 10, "superfluous"},
		{"superfluous big set", "5H 5C 5D", 10, "superfluous"},
		{"jester in set", "X 10C", 10, "superfluous"},
		{"royal values", "JH QD", 25, ""},
		{"king alone covers 20", "KS", 20, ""},
		{"queen short", "QS", 20, "notEnough"},
		{"not enough", "5H 4C", 10, "notEnough"},
		{"empty", "", 10, "empty"},
		{"three needed", "3H 3C 4D", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiscard(cs(tt.cards), tt.attack)
			if tt.reason == "" {
				mustOK(t, err)
				return
			}
			if !errors.Is(err, ErrIllegalDiscard) || !strings.Contains(err.Error(), tt.reason) {
				t.Fatalf("err = %v, want %s", err, tt.reason)
			}
		})
	}
}

func TestRemoveCards(t *testing.T) {
	hand := cs("X X 5H 5C KS")
	rest, err := removeCards(hand, cs("X 5C"))
	mustOK(t, err)
	if !slices.Equal(rest, cs("X 5H KS")) || !slices.Equal(hand, cs("X X 5H 5C KS")) {
		t.Fatalf("rest %v hand %v", rest, hand)
	}
	rest[0] = c("2H")
	if hand[0] != Jester() {
		t.Fatal("removeCards aliases the hand")
	}
	if rest, err := removeCards(hand, cs("X X")); err != nil || !slices.Equal(rest, cs("5H 5C KS")) {
		t.Fatalf("both jesters: %v %v", rest, err)
	}
	if _, err := removeCards(hand, cs("X X X")); !errors.Is(err, ErrCardNotInHand) {
		t.Fatalf("three jesters: %v", err)
	}
	if _, err := removeCards(hand, cs("5H 5H")); !errors.Is(err, ErrCardNotInHand) {
		t.Fatalf("duplicate: %v", err)
	}
	if _, err := removeCards(hand, cs("2D")); !errors.Is(err, ErrCardNotInHand) {
		t.Fatalf("missing: %v", err)
	}
	if rest, err := removeCards(nil, nil); err != nil || rest == nil || len(rest) != 0 {
		t.Fatalf("empty: %v %v", rest, err)
	}
}
