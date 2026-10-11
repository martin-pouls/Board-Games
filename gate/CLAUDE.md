# Gate

Shared-password access gate for the board-game apps (root [CLAUDE.md](../CLAUDE.md) holds the common rules). No accounts: several passwords, each with a label (e.g. `family`, `friends`). A login cookie lasts about a year. Go only, no frontend, no database; the login page is `backend/internal/gate/login.html` (embedded, inline CSS); its password field is a text field masked with `-webkit-text-security`, so browsers don't offer to save the shared password.

Caddy asks the gate before serving a protected app (`forward_auth gate:8080 { uri /verify ... }`) and proxies `/gate/*` to it with the prefix stripped (`deploy/gate.caddy`).

## Routes (as the gate sees them)

- `GET /verify`: valid cookie 200. Otherwise page loads (GET/HEAD per `X-Forwarded-Method`, `Sec-Fetch-Mode: navigate` or `Accept: text/html`) get 302 to `<base>/login?next=<X-Forwarded-Uri>`; everything else (fetch, WebSocket, POST) gets 401 JSON with code `login_required` (the game frontends reload the page on it; games use `unauthorized` for bad player tokens).
- `GET /login`: the page, or 303 to `next` if already logged in. `POST /login`: checks the password against all configured ones in constant time, sets the cookie, 303 to `next`. Rate limited (10 attempts, then 1 per 30 s per client, `httpkit.Limiter`); 429/401 re-render the page.
- `GET /` 302 to the login, `GET /healthz`, everything else 404.
- `next` must be a same-origin path (`/...`, not `//`, `/\`, no control chars or backslash), else `GATE_DEFAULT_NEXT`.

## Config

Env: `ADDR` (default `:8083`; the image sets `:8080`), `GATE_SECRET_FILE` (`data/secret`), `GATE_PASSWORDS_FILE` (`data/passwords`), `GATE_BASE_PATH` (`/gate`), `GATE_DEFAULT_NEXT` (`/board-games/`).

- Secret file: hex, at least 32 bytes decoded (`openssl rand -hex 32`).
- Passwords file: one `label=password` per line; blank lines and `#` comments ignored. Label matches `^[a-z0-9][a-z0-9-]{0,31}$`, unique; the password is everything after the first `=`, trimmed, non-empty. Errors name the line, never the password.
- On the server both files live in `/etc/board-games/gate` (mounted at `/secrets`, uid 65532, mode 600).

## Cookie

`bg_gate=<label>.<issuedUnixSeconds>.<sig>`; `sig` is base64url (no padding) HMAC-SHA256 over `<label>.<issued>` with key HMAC-SHA256(secret, `board-games-gate/v1\0label\0password`). Changing or removing one password invalidates only that label's cookies; changing the secret invalidates all. Valid for 365 days, up to 5 min clock skew. Attributes: `Path=/; HttpOnly; Secure; SameSite=Lax`.

## Run locally

```sh
cd gate/backend
mkdir data; openssl rand -hex 32 > data/secret; printf 'family=change-me\n' > data/passwords
go run ./cmd/server
```

The cookie is `Secure`, so test on http://localhost only (browsers accept Secure cookies there). Checks: `go vet ./... && go test ./...`.
