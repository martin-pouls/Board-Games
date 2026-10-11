package gate

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

// The login page, built from gate/frontend into ui/build (npm run build there; the Dockerfile does
// it before go build). ui/README.md keeps the directory in a fresh checkout.
//
//go:embed all:ui
var uiFiles embed.FS

// fallbackPage stands in for the login page when the frontend is not built (go test and go run in
// a fresh checkout): the same form, unstyled. The image always has the real page.
const fallbackPage = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="robots" content="noindex"><title>Board Games</title></head>
<body><p>The login page is not built: run npm run build in gate/frontend.</p>
<form method="post" action="{{.Action}}">
<input name="password" type="text" autocomplete="off" required>
<input type="hidden" name="next" value="{{.Next}}">
<button type="submit">Enter</button>
{{if .Message}}<p role="alert">{{.Message}}</p>{{end}}
</form></body></html>`

// loadUI returns the login page as a template (fields Action, Next, Message) and the files it
// links to (CSS, fonts, favicon), or the fallback page and nil without a build.
func loadUI(files fs.FS) (*template.Template, fs.FS, error) {
	build, err := fs.Sub(files, "ui/build")
	if err != nil {
		return nil, nil, err
	}
	page, err := fs.ReadFile(build, "index.html")
	if errors.Is(err, fs.ErrNotExist) {
		t, err := template.New("login").Parse(fallbackPage)
		return t, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	t, err := template.New("login").Parse(string(page))
	return t, build, err
}

// serveAssets serves the page's files: /_app/... (hashed, cached for good) and /favicon.svg. No
// directory listings; everything else is 404.
func serveAssets(assets fs.FS) http.Handler {
	if assets == nil {
		return http.NotFoundHandler()
	}
	files := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/_app/immutable/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}
