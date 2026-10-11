# Access layer: handover

Status: planned, not started. Goal: keep bots and strangers out of every app with one shared password. There are no accounts, and a login lasts long enough that a browser rarely asks again.

## Behaviour

- **Without a valid cookie:** opening any page (start page or game, including invite links) redirects to `/auth/login?next=<original path>`. After the right password, the browser lands on `next`.
- **With a valid cookie:** no redirect, no visible change.
- **API and WebSocket requests** without a cookie get `401`, not a redirect. The frontend reloads on `401`, and the reload hits the login redirect.
- **Public without login:** `/auth/*`, the Impressum, the Datenschutz page, `favicon` and `robots.txt`. The login page links to both legal pages and is `noindex`.

## Design

A small Go app `gate/` in the monorepo with the usual app shape (`Dockerfile`, `deploy/compose.yaml`, `deploy/gate.caddy`). The shared Caddyfile puts `forward_auth gate:8080 { uri /verify }` in front of everything except the public paths, so no game changes for the check itself.

- **`GET /verify`:** `200` with a valid cookie. Without one: `302` to the login for page loads (`Sec-Fetch-Mode: navigate` or `Accept: text/html`), `401` otherwise. The original path comes from `X-Forwarded-Uri`.
- **`GET`/`POST /auth/login`:** the password form; on success it sets the cookie and redirects to `next`.
- **Cookie:** stateless HMAC token (issue time + signature), `HttpOnly; Secure; SameSite=Lax; Path=/`, max age about 1 year (Chrome caps at 400 days). The key is derived from a server secret plus the password, so changing the password logs everyone out. No database.
- **Open-redirect guard:** `next` must start with `/` and not with `//` or `/\`; otherwise it falls back to `/board-games/`.
- **Brute force:** per-IP rate limit on `POST /auth/login`, reusing `shared/backend/httpkit` (client IP = last `X-Forwarded-For` entry).
- **Secrets:** `/opt/board-games/gate/secrets.env` (`GATE_PASSWORD`, `GATE_SECRET`), created once on the server and loaded via compose `env_file`. Never in the repo or CI.

## Rejected alternatives

| Option | Why not |
|---|---|
| Caddy `basic_auth` | Browser popup, bad on phones, asks again after every browser restart. |
| Cloudflare Access | DNS moves to Cloudflare, all traffic via a US company (privacy policy), per-e-mail allow list. |
| Authelia / oauth2-proxy / tinyauth | Built around user accounts; more config than one password needs. |
| Check inside each app | The start page is static; the check would be duplicated in every game. |

## Local prod-like setup

Today nothing runs Caddy locally, so redirects are untested before deploy.

- **Windows 11 Home → WSL2 + Ubuntu** (`wsl --install` as admin, reboot), then the server's packages: `docker.io`, `docker-compose-v2`, `docker-buildx`. Docker Desktop also works.
- **Caddyfile with env placeholders** (`{$SITE_ADDRESS}`), so the real file serves `http://localhost:8000` locally. Browsers accept `Secure` cookies on localhost.
- **`deploy/local` script:** creates the `web` network, starts Caddy with all `*/deploy/*.caddy` snippets, builds and starts every app's `compose.yaml`, with a local `secrets.env` for the gate.

## Phases

1. **Local stack:** WSL, Docker, the script, the current site running locally (also checks the `/card-games/` → `/board-games/` redirect).
2. **Gate app:** unit tests for cookie sign/verify/expiry, the `next` guard, redirect vs. `401`, the rate limit.
3. **Wiring:** Caddyfile `forward_auth` and public paths. Shared frontend: reload on `401` from REST and WS.
4. **Server:** secrets file in `setup-server.sh` (create if missing), CI smoke test sends a valid cookie (otherwise it only sees the login redirect).
5. **Docs:** privacy policy (it says "keine Cookies"; the login cookie is strictly necessary under § 25 Abs. 2 Nr. 2 TDDDG, so no consent banner), CLAUDE.md, deploy/README.md.

## Open decisions

- Cookie lifetime of about 1 year, with a password change logging everyone out: OK?
- Login page in English (like the games) or German?
- Later, optionally: invite links that carry a login key, so invited friends skip the password.
