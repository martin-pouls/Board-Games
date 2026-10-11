package gate

import (
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadUIFallback(t *testing.T) {
	page, assets, err := loadUI(fstest.MapFS{"ui/README.md": {Data: []byte("x")}})
	if err != nil {
		t.Fatal(err)
	}
	if assets != nil {
		t.Error("assets without a build")
	}
	var b strings.Builder
	if err := page.Execute(&b, loginView{Action: "/gate/login", Next: "/a", Message: "Wrong password."}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`action="/gate/login"`, `value="/a"`, `role="alert">Wrong password.`, "npm run build"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("fallback lacks %q", want)
		}
	}
}

func TestLoadUIBuild(t *testing.T) {
	files := fstest.MapFS{
		"ui/build/index.html":               {Data: []byte(`<form action="{{.Action}}"><input value="{{.Next}}">{{if .Message}}<p>{{.Message}}</p>{{end}}</form>`)},
		"ui/build/favicon.svg":              {Data: []byte(`<svg></svg>`)},
		"ui/build/_app/immutable/a.css":     {Data: []byte(`body{}`)},
		"ui/build/_app/version.json":        {Data: []byte(`{}`)},
		"ui/build/_app/immutable/x/b.woff2": {Data: []byte(`font`)},
	}
	page, assets, err := loadUI(files)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	if err := page.Execute(&b, loginView{Action: "/gate/login", Next: `/a"><script>`}); err != nil {
		t.Fatal(err)
	}
	if got := b.String(); got != `<form action="/gate/login"><input value="/a&#34;&gt;&lt;script&gt;"></form>` {
		t.Errorf("page = %s", got)
	}

	h := serveAssets(assets)
	tests := []struct {
		path      string
		wantCode  int
		immutable bool
	}{
		{"/_app/immutable/a.css", 200, true},
		{"/_app/immutable/x/b.woff2", 200, true},
		{"/_app/version.json", 200, false},
		{"/favicon.svg", 200, false},
		{"/_app/", 404, false},
		{"/_app/immutable/", 404, false},
		{"/_app/immutable/missing.css", 404, false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := do(h, "GET", tt.path, nil, "")
			if w.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantCode)
			}
			if got := w.Header().Get("Cache-Control") != ""; got != tt.immutable {
				t.Errorf("Cache-Control = %q", w.Header().Get("Cache-Control"))
			}
		})
	}
}

// TestBuiltLoginPage checks the page built from gate/frontend; skipped without a build. The image
// build runs the tests after the frontend build, so it always runs there.
func TestBuiltLoginPage(t *testing.T) {
	if uiAssets == nil {
		t.Skip("login page not built (npm run build in gate/frontend)")
	}
	h, _, _ := newTestHandler(t)
	body := do(h, "GET", "/login", nil, "").Body.String()
	for _, want := range []string{
		`/board-games/impressum/`,
		`/board-games/datenschutz/`,
		`mode-watcher-mode`,
		`data-theme-toggle`,
		`[-webkit-text-security:disc]`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	// The stylesheet the page links to is served (the base path is Caddy's, which strips it).
	css := regexp.MustCompile(`href="(?:/gate)?(/_app/immutable/[^"]+\.css)"`).FindStringSubmatch(body)
	if css == nil {
		t.Fatal("no stylesheet link")
	}
	if w := do(h, "GET", css[1], nil, ""); w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/css") {
		t.Errorf("%s: %d %q", css[1], w.Code, w.Header().Get("Content-Type"))
	}
	if w := do(h, "GET", "/favicon.svg", nil, ""); w.Code != 200 {
		t.Errorf("favicon: %d", w.Code)
	}
}
