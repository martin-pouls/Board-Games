package idgen

import (
	"strings"
	"testing"
)

func TestGameCode(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		c := GameCode()
		if len(c) != CodeLength {
			t.Fatalf("code %q: length %d, want %d", c, len(c), CodeLength)
		}
		for _, r := range c {
			if !strings.ContainsRune(CodeAlphabet, r) {
				t.Fatalf("code %q: %q is not in the alphabet", c, r)
			}
		}
		seen[c] = true
	}
	if len(seen) < 190 {
		t.Fatalf("only %d distinct codes out of 200", len(seen))
	}
}

func TestPlayerIDAndToken(t *testing.T) {
	id := PlayerID()
	if !strings.HasPrefix(id, "p_") || len(id) != 12 || strings.ToLower(id) != id {
		t.Fatalf("player id %q", id)
	}
	a, b := Token(), Token()
	if len(a) != 43 || a == b {
		t.Fatalf("tokens %q, %q", a, b)
	}
}

func TestHashToken(t *testing.T) {
	// sha256("abc")
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := HashToken("abc"); got != want {
		t.Fatalf("HashToken = %s, want %s", got, want)
	}
}

func TestNormalizeCode(t *testing.T) {
	if got := NormalizeCode("  abc2d \n"); got != "ABC2D" {
		t.Fatalf("NormalizeCode = %q", got)
	}
}

func TestNewRNGIsSeeded(t *testing.T) {
	a, b := NewRNG(), NewRNG()
	if a.Uint64() == b.Uint64() && a.Uint64() == b.Uint64() {
		t.Fatal("two generators produced the same sequence")
	}
}

type testID string

func TestTypedHelpers(t *testing.T) {
	if c := GameCodeAs[testID](); len(c) != CodeLength || strings.Trim(string(c), CodeAlphabet) != "" {
		t.Errorf("GameCodeAs = %q", c)
	}
	if p := PlayerIDAs[testID](); !strings.HasPrefix(string(p), "p_") || len(p) != 12 {
		t.Errorf("PlayerIDAs = %q", p)
	}
	if got := NormalizeCodeAs(testID("  ab2cd ")); got != "AB2CD" {
		t.Errorf("NormalizeCodeAs = %q, want AB2CD", got)
	}
}
