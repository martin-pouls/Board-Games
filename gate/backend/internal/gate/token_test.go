package gate

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	testSecret = []byte(strings.Repeat("s", 32))
	testPws    = []Password{{"family", "pw-family"}, {"friends", "pw-friends"}}
	t0         = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newTestTokens(c *clock) *Tokens { return NewTokens(testSecret, testPws, c.now) }

func TestTokenRoundTrip(t *testing.T) {
	c := &clock{t0}
	tk := newTestTokens(c)
	for _, p := range testPws {
		v, ok := tk.Issue(p.Label)
		if !ok {
			t.Fatal("issue failed")
		}
		if label, ok := tk.Verify(v); !ok || label != p.Label {
			t.Errorf("Verify(%q) = %q, %v", v, label, ok)
		}
	}
	if _, ok := tk.Issue("nobody"); ok {
		t.Error("issued for unknown label")
	}
}

func TestTokenVerify(t *testing.T) {
	c := &clock{t0}
	tk := newTestTokens(c)
	valid, _ := tk.Issue("family")
	parts := strings.Split(valid, ".")
	issued := t0.Unix()
	sigFor := func(label string, pw string, at int64) string { // build a token as another gate would
		return NewTokens(testSecret, []Password{{label, pw}}, nil).sign(label, pw, at)
	}

	tests := []struct {
		name  string
		value string
		now   time.Time
		ok    bool
	}{
		{"fresh", valid, t0, true},
		{"just under a year", valid, t0.Add(MaxAge - time.Second), true},
		{"exactly a year", valid, t0.Add(MaxAge), false},
		{"after a year", valid, t0.Add(MaxAge + time.Hour), false},
		{"issued within skew", sigFor("family", "pw-family", issued+240), t0, true},
		{"issued beyond skew", sigFor("family", "pw-family", issued+301), t0, false},
		{"wrong label signature", parts[0] + "." + parts[1] + "." + strings.Split(sigFor("friends", "pw-friends", issued), ".")[2], t0, false},
		{"label swapped", "friends." + parts[1] + "." + parts[2], t0, false},
		{"unknown label", sigFor("ghost", "x", issued), t0, false},
		{"wrong password", sigFor("family", "other", issued), t0, false},
		{"tampered issued", parts[0] + "." + strconv.FormatInt(issued+1, 10) + "." + parts[2], t0, false},
		{"tampered signature", valid[:len(valid)-2] + "AA", t0, false},
		{"truncated signature", valid[:len(valid)-3], t0, false},
		{"empty signature", parts[0] + "." + parts[1] + ".", t0, false},
		{"empty", "", t0, false},
		{"two parts", parts[0] + "." + parts[1], t0, false},
		{"four parts", valid + ".x", t0, false},
		{"non-numeric issued", parts[0] + ".abc." + parts[2], t0, false},
		{"padded issued", parts[0] + ".0" + parts[1] + "." + parts[2], t0, false},
		{"garbage", "!!!", t0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.t = tt.now
			if _, ok := tk.Verify(tt.value); ok != tt.ok {
				t.Errorf("Verify = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestTokenInvalidation(t *testing.T) {
	c := &clock{t0}
	old := newTestTokens(c)
	family, _ := old.Issue("family")
	friends, _ := old.Issue("friends")

	tests := []struct {
		name                      string
		secret                    []byte
		pws                       []Password
		familyValid, friendsValid bool
	}{
		{"unchanged", testSecret, testPws, true, true},
		{"family password changed", testSecret, []Password{{"family", "new"}, {"friends", "pw-friends"}}, false, true},
		{"family removed", testSecret, []Password{{"friends", "pw-friends"}}, false, true},
		{"friends removed", testSecret, []Password{{"family", "pw-family"}}, true, false},
		{"new label added", testSecret, append([]Password{{"kids", "k"}}, testPws...), true, true},
		{"secret changed", []byte(strings.Repeat("t", 32)), testPws, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewTokens(tt.secret, tt.pws, c.now)
			if _, ok := tk.Verify(family); ok != tt.familyValid {
				t.Errorf("family valid = %v, want %v", ok, tt.familyValid)
			}
			if _, ok := tk.Verify(friends); ok != tt.friendsValid {
				t.Errorf("friends valid = %v, want %v", ok, tt.friendsValid)
			}
		})
	}
}

func TestTokenFormat(t *testing.T) {
	tk := newTestTokens(&clock{t0})
	v, _ := tk.Issue("family")
	parts := strings.Split(v, ".")
	if len(parts) != 3 || parts[0] != "family" || parts[1] != strconv.FormatInt(t0.Unix(), 10) {
		t.Fatalf("unexpected token %q", v)
	}
	if len(parts[2]) != 43 || strings.ContainsAny(parts[2], "=+/") {
		t.Errorf("signature %q is not unpadded base64url of 32 bytes", parts[2])
	}
}
