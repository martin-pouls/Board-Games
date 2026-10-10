# Card Games — monorepo

Web implementations of cooperative card games, played live in the browser, plus a start page that links to them. Live at https://yoshi-und-kruemli.de/card-games/.

| App | Directory | URL path | What |
|---|---|---|---|
| Start page | [start-page/](start-page/) | `/card-games/` | Static page linking to every game. Frontend only. |
| The Game | [the-game/](the-game/) | `/the-game/` | "The Game" (NSV), 1–5 players, incl. Pro and Extreme. See [the-game/CLAUDE.md](the-game/CLAUDE.md). |
| Regicide | [regicide/](regicide/) | `/regicide/` | Regicide (Badgers From Mars), 1–4 players, incl. solo. See [regicide/CLAUDE.md](regicide/CLAUDE.md). |

This file holds the rules common to all apps. Each game's own CLAUDE.md adds its rules, domain model and look. **Read the game's CLAUDE.md before working on it.**

## Repository layout

```
CLAUDE.md, .mcp.json, .gitignore, .gitattributes    shared by all apps
deploy/                       server-wide: shared Caddy proxy, server bootstrap, deploy script
  caddy/                      compose file + Caddyfile for the shared proxy (/opt/caddy on the server)
  setup-server.sh             idempotent server bootstrap (Docker, ufw, Caddy)
  deploy.sh                   deploys one, several or all apps
start-page/                   static start page (SvelteKit, prerendered, no JS)
  frontend/  Dockerfile  deploy/
the-game/                     one self-contained app per game
  backend/  frontend/  docs/rules.md  Dockerfile  deploy/  CLAUDE.md
regicide/
  backend/  frontend/  docs/rules.md  Dockerfile  deploy/  CLAUDE.md
```

**Apps are independent.** Each game has its own Go module, its own npm project (own `package.json`, lockfile, `node_modules`) and its own image, database and deploy. Nothing imports across app directories. Regicide's lobby, session, WebSocket and persistence plumbing was ported from The Game, so the two are deliberately parallel. When in doubt about a convention in one game, look at how the other does it, and keep them aligned when changing shared plumbing.

Every game app has the same shape:

```
<game>/
  backend/                    Go module (cmd/server + internal/...)
    cmd/server/               main: config from env, wiring, graceful shutdown, janitor
    internal/core/            the hexagon — no imports of adapters, net/http or database/sql
      domain/                 pure game model + rule engine (stdlib only)
      ports/                  interfaces: driving (use cases) + driven (repo, publisher)
      service/                use-case implementations of driving ports
    internal/adapters/        everything that talks to the outside world
      apidto/                 JSON wire contract (DTOs, mappers, error codes) shared by httpapi + ws
      httpapi/                REST handlers (+ serves built SPA when STATIC_DIR set)
      ws/                     WebSocket hub: live game events/commands
      sqlite/                 repository implementations + migrations
      memory/                 in-memory repository (tests)
      repotest/               shared repository contract test suite
    internal/e2e/             end-to-end test: sqlite + service + REST + WS over httptest
  frontend/                   SvelteKit SPA
    src/lib/api.ts            REST client
    src/lib/types.ts          TS mirror of the Go DTOs in apidto
    src/lib/game/             WS protocol + GameConnection (runes), formatting helpers, fixtures
    src/lib/components/game/  game screen components
    src/lib/components/ui/    shadcn-svelte components (generated — don't hand-edit heavily)
    src/routes/               pages
  docs/rules.md               rules reference (authoritative), including app decisions
  Dockerfile                  one image: built SPA + Go server
  deploy/compose.yaml         the app's container on the server
  deploy/<game>.caddy         the app's Caddy snippet (path prefix → container)
```

The wire contract is the code: `apidto` (Go) and `frontend/src/lib/types.ts` (TS mirror).

## Tech stack (both games)

| Area | Choice | Notes |
|---|---|---|
| Backend language | Go 1.27 | stdlib-first, minimal dependencies |
| HTTP | `net/http` `ServeMux` | Go 1.22+ method/path patterns; no web framework |
| Realtime | WebSockets via `github.com/coder/websocket` | live gameplay |
| Persistence | SQLite via `modernc.org/sqlite` | pure Go, **no CGO**; embedded file DB, no separate DB server |
| Migrations | Embedded `*.sql` files + tiny runner | `backend/internal/adapters/sqlite/migrations/NNNN_name.sql`, applied in order at startup |
| Logging | `log/slog` | |
| Frontend | SvelteKit 3 + Svelte 5 (runes) + TypeScript | SPA via `adapter-static` (`ssr = false`, fallback `index.html`) |
| UI | Tailwind CSS 4 + **shadcn-svelte** (bits-ui, lucide icons) | style `vega`; each game has its own theme |
| Frontend tooling | Vite, Vitest, ESLint, Prettier, svelte-check | |

The start page uses the same SvelteKit/Tailwind/tooling stack without shadcn. It is fully prerendered with `csr = false`, so it ships HTML and CSS only.

## Architecture: hexagonal (ports and adapters)

- **Dependency direction points inward:** `adapters → core/service → core/ports → core/domain`. The domain imports nothing from its module.
- **All game rules live in `core/domain`.** Services orchestrate: load aggregate → call domain method → save → publish events. Adapters translate only and contain no rules.
- **The server is authoritative.** Clients send *intents*. The domain validates them and returns typed errors for illegal moves. Never trust client-side validation; the client may pre-validate only for UX.
- **Per-player views:** never send other players' hands (or hidden deck order) to a client. Build a projection per player.
- **Transport split:**
  - **REST** (`/api/...`) for lobby and request/response actions: create game, join, start, fetch current view.
  - **WebSocket** (`/ws/...`) for live gameplay: server pushes per-player state/events to everyone in a game, and clients send turn actions over it.
  - Both call the same driving ports.
- **Persistence:** each game is stored as a JSON snapshot of the domain aggregate (`games.state`) with a `version` column for optimistic concurrency. Serialise through explicit DTOs in the sqlite adapter, not through struct tags on domain types. Live connections and presence may be in memory; game state must survive a restart.
- **Domain methods validate fully before mutating:** on error the state stays unchanged (tests assert this).
- **Illegal moves are rejected, never punished.** A game is lost only by its rules (or by give-up).
- Each `*_test.go` sits next to its code. The domain must have thorough table-driven unit tests: every rule, app decision and edge case in the game's docs/rules.md.
- **Rules:** each game's `docs/rules.md` is the source of truth. Read it before touching game logic. Do not re-read the rule PDFs. If a rule question is not answered there, ask, then record the answer in that file.

## Commands

The toolchain was installed via winget. Go is at `C:\Program Files\Go\bin` and Node at `C:\Program Files\nodejs`; a shell opened before the install may need its PATH refreshed.

Backend (run from `<game>/backend/`):
```sh
go run ./cmd/server              # API on :8080, DB at data/<game>.db
go test ./...
go vet ./...
go build -o bin/server.exe ./cmd/server
```
Env vars: `ADDR` (default `:8080`), `DB_PATH` (default `data/<game>.db`), `STATIC_DIR` (optional, e.g. `../frontend/build` to serve the SPA from Go), `WS_ORIGIN_PATTERNS` (optional, comma-separated extra allowed WebSocket origin host patterns), `GAME_RETENTION` (default `168h`: games idle longer are deleted), `CLEANUP_INTERVAL` (default `1h`: how often the janitor runs).

Both backends default to `:8080` and both frontends to Vite's `:5173`. To run both games at once, start one with e.g. `ADDR=:8081` and its frontend with `BACKEND_URL=http://localhost:8081 npm run dev -- --port 5174`.

Frontend (run from `<app>/frontend/`):
```sh
npm run dev        # Vite on :5173; game frontends proxy /api and /ws to BACKEND_URL (default http://localhost:8080)
npm run check      # svelte-check (types)
npm run lint       # prettier --check + eslint
npm run format
npm test           # vitest
npm run build      # static build into frontend/build
```

Production-like local run of a game: `npm run build` in `<game>/frontend/`, then in `<game>/backend/` run `STATIC_DIR=../frontend/build go run ./cmd/server` and open http://localhost:8080.

## Deployment

Target: a Hetzner arm64 Ubuntu box serving https://yoshi-und-kruemli.de (root via `~/.ssh/hetzner_ed25519`).

- **One container per app**, all on the external Docker network `web`, built natively on the server (arm64):
  - `the-game` and `regicide`: one image each (SPA + Go server), SQLite in the volumes `the-game_data` and `regicide_data`.
  - `start-page`: the prerendered page served by `caddy file-server`.
- **Shared Caddy proxy** in `/opt/caddy` ([deploy/caddy/](deploy/caddy/)). It is the only container with public ports and handles HTTPS. The Caddyfile redirects `/` to `/card-games/` and imports every `sites/*.caddy`.
- **Each app ships a snippet** `<app>/deploy/<app>.caddy` that maps its path prefix to its container with `handle_path`. The prefix is stripped, so the servers stay path-agnostic.
- **Base path:** every frontend is built with `BASE_PATH=/<path>` (`paths.base`), set as a build arg in `<app>/deploy/compose.yaml`. Routes use `resolve()`; game API and WS URLs use `__BASE_PATH__`. Never hard-code `/api` or `/ws` without it. Links between apps are plain absolute paths with `rel="external"`.
- **Deploying:** from Git Bash at the repo root.
  - `deploy/deploy.sh` deploys all apps; `deploy/deploy.sh regicide` (or several names) deploys only those.
  - It ships each app's *committed* subtree (`git archive HEAD:<app>`) to `/opt/card-games/<app>/src`, runs `docker compose up -d --build` there, installs the snippet and reloads Caddy once. Uncommitted changes are not deployed.
  - `deploy/setup-server.sh` is the idempotent bootstrap: Docker, ufw, Caddy. Re-run it after changing `deploy/caddy/`; the repo owns the Caddyfile.
- **Adding an app:** a top-level directory with `Dockerfile`, `deploy/compose.yaml` (unique `name:` and `container_name:`, network `web`) and `deploy/<dir-name>.caddy`. `deploy.sh` picks it up automatically. Add a tile in `start-page/frontend/src/lib/games.ts`.
- **Git Bash gotcha:** it rewrites `/regicide`-like values in env vars into Windows paths. For a local base-path build use `MSYS2_ENV_CONV_EXCL=BASE_PATH BASE_PATH=/regicide npm run build`.

## Conventions and gotchas

- **SvelteKit 3 uses Node subpath imports (`#lib`), not `$lib`.** Imports need an explicit file, e.g. `#lib/api.js` (resolves to `api.ts`) or `#lib/components/ui/card/index.js`.
- **Svelte 5 runes only** (`$state`, `$derived`, `$props`, `$effect`). No legacy `export let` or stores for new code. A TS module that uses runes must be named `*.svelte.ts`.
- **Secure-context APIs** (`crypto.randomUUID`, `navigator.clipboard`, `navigator.share`) are missing when phones test over `http://192.168.x.x`: always feature-detect and fall back.
- The player token travels in the WS query string: don't log full URLs at the reverse proxy.
- **Adding shadcn components** (from `<game>/frontend/`): `npx shadcn-svelte@latest add <name> -y`.
  - If it fails with `File '$app/tsconfig' not found`, run `npx svelte-kit sync` first; an npm install can wipe the generated `node_modules/$app`.
  - The generated components live in `src/lib/components/ui/` and are lint-exempt from `svelte/no-navigation-without-resolve`.
- **bits-ui `AlertDialog.Action` does not close its dialog.** Set the bound `open` flag to `false` in its `onclick`, or the overlay stays up whenever the component stays mounted.
- **Go:** keep dependencies minimal and prefer the stdlib. Wrap errors with `%w`. Pass `context.Context` as the first param on anything doing I/O. Domain errors are typed sentinel errors so adapters can map them to HTTP / WS error codes.
- **API JSON uses camelCase.** Keep the TS types in `frontend/src/lib/types.ts` in sync with the Go DTOs.
- Files that run on the Linux server (`*.sh`, `Dockerfile`, `**/deploy/**`) keep LF line endings (`.gitattributes`).
- **Mobile-first:** everything must work on a phone in portrait.
- Before declaring work done, run in each app you touched: `go vet ./... && go test ./...` (backend) and `npm run check && npm run lint && npm test` (frontend).
