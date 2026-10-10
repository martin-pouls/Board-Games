package domain

import (
	"errors"
	"reflect"
	"testing"
)

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{"default", DefaultOptions(), false},
		{"pro", Options{Variant: VariantClassic, Pro: true}, false},
		{"pro reduced", Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}, false},
		{"reduced without pro", Options{Variant: VariantClassic, ReducedHandSize: true}, true},
		{"extreme", Options{Variant: VariantExtreme}, false},
		{"extreme pro", Options{Variant: VariantExtreme, Pro: true}, false},
		{"extreme pro reduced", Options{Variant: VariantExtreme, Pro: true, ReducedHandSize: true}, false},
		{"extreme reduced without pro", Options{Variant: VariantExtreme, ReducedHandSize: true}, true},
		{"empty variant", Options{}, true},
		{"unknown variant", Options{Variant: "bogus"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidOptions) {
					t.Fatalf("Validate() = %v, want ErrInvalidOptions", err)
				}
				if _, rerr := NewRuleset(tt.opts); !errors.Is(rerr, ErrInvalidOptions) {
					t.Fatalf("NewRuleset() = %v, want ErrInvalidOptions", rerr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestDefaultOptionsIsExtreme(t *testing.T) {
	if got := DefaultOptions(); got != (Options{Variant: VariantExtreme}) {
		t.Fatalf("DefaultOptions() = %+v", got)
	}
}

func TestRulesetHandSizes(t *testing.T) {
	classic := classic
	pro := Options{Variant: VariantClassic, Pro: true}
	proReduced := Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}
	tests := []struct {
		opts    Options
		players int
		want    int
	}{
		{classic, 1, 8}, {classic, 2, 7}, {classic, 3, 6}, {classic, 4, 6}, {classic, 5, 6},
		{pro, 1, 8}, {pro, 2, 7}, {pro, 3, 6}, {pro, 4, 6}, {pro, 5, 6},
		{proReduced, 1, 7}, {proReduced, 2, 6}, {proReduced, 3, 5}, {proReduced, 4, 5}, {proReduced, 5, 5},
		{classic, 0, 0}, {classic, 6, 0},
	}
	for _, tt := range tests {
		r, err := NewRuleset(tt.opts)
		if err != nil {
			t.Fatal(err)
		}
		if got := r.HandSize(tt.players); got != tt.want {
			t.Errorf("%+v HandSize(%d) = %d, want %d", tt.opts, tt.players, got, tt.want)
		}
	}
}

func TestRulesetMinimums(t *testing.T) {
	tests := []struct {
		name          string
		opts          Options
		drawPileEmpty bool
		want          int
	}{
		{"classic", classic, false, 2},
		{"classic empty draw pile", classic, true, 1},
		{"pro", Options{Variant: VariantClassic, Pro: true}, false, 3},
		{"pro empty draw pile", Options{Variant: VariantClassic, Pro: true}, true, 1},
		{"pro reduced", Options{Variant: VariantClassic, Pro: true, ReducedHandSize: true}, false, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRuleset(tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := r.MinCardsToPlay(tt.drawPileEmpty); got != tt.want {
				t.Fatalf("MinCardsToPlay(%v) = %d, want %d", tt.drawPileEmpty, got, tt.want)
			}
		})
	}
}

// TestRulesetExtremeSameAsClassic: hand sizes and minimums are the same (rules.md §2).
func TestRulesetExtremeSameAsClassic(t *testing.T) {
	pairs := [][2]Options{
		{classic, extreme},
		{{Variant: VariantClassic, Pro: true}, extremePro},
		{{Variant: VariantClassic, Pro: true, ReducedHandSize: true}, extremeProReduced},
	}
	for _, p := range pairs {
		c, x := mustRules(t, p[0]), mustRules(t, p[1])
		if x.Variant != VariantExtreme {
			t.Fatalf("variant %q", x.Variant)
		}
		x.Variant = VariantClassic
		if !reflect.DeepEqual(c, x) {
			t.Fatalf("extreme ruleset %+v differs from classic %+v", x, c)
		}
	}
}

func TestRulesetConstants(t *testing.T) {
	r, err := NewRuleset(classic)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variant != VariantClassic || r.MinPlayers != 1 || r.MaxPlayers != 5 ||
		r.BackwardsTrickStep != 10 || r.AscendingStart != 1 || r.DescendingStart != 100 ||
		r.LowestCard != 2 || r.HighestCard != 99 || r.AscendingPiles != 2 || r.DescendingPiles != 2 ||
		r.HandSizeReduction != 0 || r.MinCardsDrawPileEmpty != 1 {
		t.Fatalf("unexpected ruleset %+v", r)
	}
}
