package httpkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/martin-pouls/shared/backend/wire"
)

func TestBearerToken(t *testing.T) {
	for _, tt := range []struct{ header, want string }{
		{"Bearer abc", "abc"},
		{"bearer  abc ", "abc"},
		{"BEARER abc", "abc"},
		{"Basic abc", ""},
		{"Bearer ", ""},
		{"", ""},
	} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.header != "" {
			r.Header.Set("Authorization", tt.header)
		}
		if got := BearerToken(r); got != tt.want {
			t.Errorf("%q: got %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestDecodeBody(t *testing.T) {
	decode := func(body string) error {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		var v struct{ Name string }
		return DecodeBody(httptest.NewRecorder(), r, &v)
	}
	if err := decode(`{"name":"Ann"}`); err != nil {
		t.Fatal(err)
	}
	if err := decode(""); !errors.Is(err, ErrEmptyBody) || !errors.Is(err, ErrBadRequest) {
		t.Fatalf("empty: %v", err)
	}
	if err := decode(`{"name":`); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("malformed: %v", err)
	}
	if err := decode(`{"name":"` + strings.Repeat("x", MaxBodyBytes) + `"}`); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("too big: %v", err)
	}
}

func TestErrorMapper(t *testing.T) {
	errGame := errors.New("illegal")
	m := append(ErrorMapper{{errGame, "illegal_play", http.StatusUnprocessableEntity}}, CommonErrors...)
	for _, tt := range []struct {
		err    error
		status int
		code   string
	}{
		{fmt.Errorf("wrapped: %w", errGame), 422, "illegal_play"},
		{BadRequest("missing %s", "x"), 400, wire.CodeBadRequest},
		{ErrRateLimited, 429, wire.CodeRateLimited},
		{&json.SyntaxError{}, 400, wire.CodeBadRequest},
		{errors.New("boom"), 500, wire.CodeInternal},
		{nil, 500, wire.CodeInternal},
	} {
		status, body := m.Map(tt.err)
		if status != tt.status || body.Code != tt.code {
			t.Errorf("%v: got %d %s, want %d %s", tt.err, status, body.Code, tt.status, tt.code)
		}
		if status == 500 && body.Message != wire.InternalMessage {
			t.Errorf("%v: internal message leaked: %q", tt.err, body.Message)
		}
	}
}

func TestRouterHealthCatchAllAndSPA(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("js"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := NewRouter(RouterDeps{StaticDir: dir})
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}
	if w := get("/api/health"); w.Code != 200 || !strings.Contains(w.Body.String(), `"ok"`) {
		t.Fatalf("health: %d %s", w.Code, w.Body)
	}
	if w := get("/api/nope"); w.Code != 404 || !strings.Contains(w.Body.String(), wire.CodeNotFound) {
		t.Fatalf("api catch-all: %d %s", w.Code, w.Body)
	}
	if w := get("/app.js"); w.Code != 200 || w.Body.String() != "js" {
		t.Fatalf("static file: %d %s", w.Code, w.Body)
	}
	if w := get("/g/ABCDE"); w.Code != 200 || !strings.Contains(w.Body.String(), "spa") {
		t.Fatalf("client route: %d %s", w.Code, w.Body)
	}
}
