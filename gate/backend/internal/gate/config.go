// Package gate is the shared-password access gate of the board-game apps: it checks a login
// cookie for Caddy's forward_auth and serves the login page.
package gate

import (
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const minSecretBytes = 32

var labelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

// Password is one accepted password and the label it logs in as.
type Password struct {
	Label    string
	Password string
}

// Config is the gate's configuration.
type Config struct {
	Secret      []byte
	Passwords   []Password
	BasePath    string // public path prefix of the gate, e.g. /gate
	DefaultNext string // where to go after login without a (valid) next
}

// ParseSecret decodes a hex secret of at least 32 bytes.
func ParseSecret(s string) ([]byte, error) {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("secret is not valid hex: %w", err)
	}
	if len(b) < minSecretBytes {
		return nil, fmt.Errorf("secret is too short: %d bytes, need at least %d", len(b), minSecretBytes)
	}
	return b, nil
}

// ParsePasswords parses lines of "label=password". Errors name the line, never the password.
func ParsePasswords(s string) ([]Password, error) {
	var out []Password
	seen := map[string]int{}
	for i, line := range strings.Split(s, "\n") {
		n := i + 1
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		label, pw, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected label=password", n)
		}
		label = strings.TrimSpace(label)
		pw = strings.TrimSpace(pw)
		if !labelPattern.MatchString(label) {
			return nil, fmt.Errorf("line %d: invalid label (want %s)", n, labelPattern)
		}
		if pw == "" {
			return nil, fmt.Errorf("line %d: empty password", n)
		}
		if first, dup := seen[label]; dup {
			return nil, fmt.Errorf("line %d: duplicate label %q (first on line %d)", n, label, first)
		}
		seen[label] = n
		out = append(out, Password{Label: label, Password: pw})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no passwords configured")
	}
	return out, nil
}

// LoadConfig reads the secret and password files and the remaining settings.
func LoadConfig(secretFile, passwordsFile, basePath, defaultNext string) (Config, error) {
	raw, err := os.ReadFile(secretFile)
	if err != nil {
		return Config{}, fmt.Errorf("read secret file: %w", err)
	}
	secret, err := ParseSecret(string(raw))
	if err != nil {
		return Config{}, fmt.Errorf("secret file %s: %w", secretFile, err)
	}
	raw, err = os.ReadFile(passwordsFile)
	if err != nil {
		return Config{}, fmt.Errorf("read passwords file: %w", err)
	}
	pws, err := ParsePasswords(string(raw))
	if err != nil {
		return Config{}, fmt.Errorf("passwords file %s: %w", passwordsFile, err)
	}
	return Config{
		Secret:      secret,
		Passwords:   pws,
		BasePath:    strings.TrimRight(basePath, "/"),
		DefaultNext: defaultNext,
	}, nil
}
