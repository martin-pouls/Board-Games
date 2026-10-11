package gate

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParsePasswords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Password
	}{
		{"single", "family=secret1", []Password{{"family", "secret1"}}},
		{"two with trailing newline", "family=a\nfriends=b\n", []Password{{"family", "a"}, {"friends", "b"}}},
		{"comments and blanks", "# hi\n\n  # indented\nfamily=a\n\n", []Password{{"family", "a"}}},
		{"crlf", "family=a\r\nfriends=b\r\n", []Password{{"family", "a"}, {"friends", "b"}}},
		{"equals in password", "family=a=b==c", []Password{{"family", "a=b==c"}}},
		{"whitespace trimmed", "  family  =  pass word  ", []Password{{"family", "pass word"}}},
		{"hash inside password", "family=a#b", []Password{{"family", "a#b"}}},
		{"label with dash and digits", "kids-2=x", []Password{{"kids-2", "x"}}},
		{"32 char label", strings.Repeat("a", 32) + "=x", []Password{{strings.Repeat("a", 32), "x"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePasswords(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePasswordsErrors(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantMsg string
	}{
		{"empty file", "", "no passwords"},
		{"only comments", "# a\n\n", "no passwords"},
		{"no equals", "family", "line 1"},
		{"uppercase label", "\nFamily=hunter2", "line 2"},
		{"leading dash", "-a=hunter2", "line 1"},
		{"space in label", "a b=hunter2", "line 1"},
		{"empty label", "=hunter2", "line 1"},
		{"too long label", strings.Repeat("a", 33) + "=hunter2", "line 1"},
		{"empty password", "family=   ", "line 1"},
		{"duplicate", "family=hunter2\n\nfamily=other", "line 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePasswords(tt.in)
			if err == nil {
				t.Fatal("want error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error %q lacks %q", err, tt.wantMsg)
			}
			for _, secret := range []string{"hunter2", "other"} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("error %q leaks the password", err)
				}
			}
		})
	}
}

func TestParseSecret(t *testing.T) {
	good := strings.Repeat("ab", 32)
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"32 bytes", good, true},
		{"trailing newline", good + "\n", true},
		{"crlf", good + "\r\n", true},
		{"longer", good + "cd", true},
		{"31 bytes", strings.Repeat("ab", 31), false},
		{"empty", "", false},
		{"not hex", strings.Repeat("zz", 32), false},
		{"odd length", good + "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSecret(tt.in)
			if (err == nil) != tt.ok {
				t.Errorf("err = %v, want ok=%v", err, tt.ok)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret")
	pws := filepath.Join(dir, "passwords")
	write := func(path, s string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := LoadConfig(secret, pws, "/gate", "/x/"); err == nil {
		t.Error("missing secret file: want error")
	}
	write(secret, strings.Repeat("ab", 32)+"\n")
	if _, err := LoadConfig(secret, pws, "/gate", "/x/"); err == nil {
		t.Error("missing passwords file: want error")
	}
	write(pws, "family=a\n")
	cfg, err := LoadConfig(secret, pws, "/gate/", "/x/")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BasePath != "/gate" || cfg.DefaultNext != "/x/" || len(cfg.Passwords) != 1 || len(cfg.Secret) != 32 {
		t.Errorf("unexpected config %+v", cfg)
	}
	write(secret, "abcd")
	if _, err := LoadConfig(secret, pws, "/gate", "/x/"); err == nil {
		t.Error("short secret: want error")
	}
	write(secret, strings.Repeat("ab", 32))
	write(pws, "family=topsecret\nfamily=again\n")
	_, err = LoadConfig(secret, pws, "/gate", "/x/")
	if err == nil || strings.Contains(err.Error(), "topsecret") || strings.Contains(err.Error(), "again") {
		t.Errorf("bad passwords file: err = %v", err)
	}
}
