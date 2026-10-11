package gate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

const (
	// CookieName is the login cookie.
	CookieName = "bg_gate"
	// MaxAge is how long a login lasts.
	MaxAge    = 365 * 24 * time.Hour
	clockSkew = 5 * time.Minute
)

// Tokens signs and verifies login cookies.
type Tokens struct {
	secret    []byte
	passwords map[string]string
	now       func() time.Time
}

// NewTokens returns a signer for the configured passwords. now is the clock (nil: time.Now).
func NewTokens(secret []byte, passwords []Password, now func() time.Time) *Tokens {
	if now == nil {
		now = time.Now
	}
	m := make(map[string]string, len(passwords))
	for _, p := range passwords {
		m[p.Label] = p.Password
	}
	return &Tokens{secret: secret, passwords: m, now: now}
}

// key is bound to the label and its password, so changing a password invalidates only its cookies.
func (t *Tokens) key(label, password string) []byte {
	m := hmac.New(sha256.New, t.secret)
	m.Write([]byte("board-games-gate/v1\x00" + label + "\x00" + password))
	return m.Sum(nil)
}

func (t *Tokens) sign(label, password string, issued int64) string {
	payload := label + "." + strconv.FormatInt(issued, 10)
	m := hmac.New(sha256.New, t.key(label, password))
	m.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// Issue returns the cookie value for label, or false if the label is not configured.
func (t *Tokens) Issue(label string) (string, bool) {
	pw, ok := t.passwords[label]
	if !ok {
		return "", false
	}
	return t.sign(label, pw, t.now().Unix()), true
}

// Verify reports the label a cookie value was issued for, if it is valid.
func (t *Tokens) Verify(value string) (label string, ok bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return "", false
	}
	label = parts[0]
	pw, known := t.passwords[label]
	if !known {
		return "", false
	}
	issued, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", false
	}
	now := t.now()
	if issued > now.Add(clockSkew).Unix() || now.Sub(time.Unix(issued, 0)) >= MaxAge {
		return "", false
	}
	want := t.sign(label, pw, issued)
	if !hmac.Equal([]byte(want), []byte(value)) {
		return "", false
	}
	return label, true
}
