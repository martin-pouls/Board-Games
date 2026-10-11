# Board Games — monorepo

Web implementations of cooperative board and card games, played live in the browser, plus a start page that links to them. Live at https://yoshi-und-kruemli.de/board-games/ (also https://martin-pouls.de/board-games/, same site). The start page and the games are behind a shared password (the gate).

| App | Directory | URL path | What |
|---|---|---|---|
| Start page | [start-page/](start-page/) | `/board-games/` | Static page linking to every game, plus imprint and privacy policy. Frontend only. |
| The Game | [the-game/](the-game/) | `/the-game/` | "The Game" (NSV), 1–5 players, incl. Pro and Extreme. See [the-game/CLAUDE.md](the-game/CLAUDE.md). |
| Regicide | [regicide/](regicide/) | `/regicide/` | Regicide (Badgers From Mars), 1–4 players, incl. solo. See [regicide/CLAUDE.md](regicide/CLAUDE.md). |
| Fox Duet | [fox-duet/](fox-duet/) | `/fox-duet/` | "The Fox in the Forest: Duet" (Foxtrot Games), 2 players, 3 difficulty levels. See [fox-duet/CLAUDE.md](fox-duet/CLAUDE.md). |
| The Crew | [the-crew/](the-crew/) | `/the-crew/` | "The Crew: The Quest for Planet Nine" (Kosmos), 2–5 players (2 with JARVIS), 50 missions. See [the-crew/CLAUDE.md](the-crew/CLAUDE.md). |
| Gate | [gate/](gate/) | `/gate/` | Shared-password login in front of the start page and the games (Caddy `forward_auth`). Go server with a prerendered login page. See [gate/CLAUDE.md](gate/CLAUDE.md). |

This file holds the rules common to all apps. Each game's own CLAUDE.md adds its rules, domain model and look. **Read the game's CLAUDE.md before working on it.**

## Repository layout

```
CLAUDE.md, .mcp.json, .gitignore, .gitattributes    shared by all apps
.github/                      CD: workflows/deploy.yml, scripts/plan-deploy.sh, dependabot.yml;
                              workflows/cleanup-images.yml (weekly: old images on ghcr.io)
deploy/                       server-wide: shared Caddy proxy, server bootstrap, deploy scripts
  README.md                   CD setup, security model, rollback
  caddy/                      compose file + Caddyfile for the shared proxy (/opt/caddy on the server);
                              conf/common.caddy is the site itself (gate, snippets), shared with local/
  local/, local.sh            local prod-like stack in WSL2 + Docker (http://localhost:8000)
  server/board-games-deploy    runs on the server: the CI deploy key's forced command
  setup-server.sh             idempotent server bootstrap (Docker, ufw, Caddy, deploy user)
  deploy.sh                   fallback: builds and deploys from your PC without GitHub
shared/                       plumbing every game uses (see shared/README.md)
  backend/                    Go module github.com/martin-pouls/shared/backend
  frontend/                   Svelte/TS source, imported as @board-games/shared
start-page/                   static start page (SvelteKit, prerendered, no JS)
  frontend/  Dockerfile  Dockerfile.dockerignore  deploy/
the-game/                     one app per game
  backend/  frontend/  docs/rules.md  Dockerfile  Dockerfile.dockerignore  deploy/  CLAUDE.md
regicide/
  backend/  frontend/  docs/rules.md  Dockerfile  Dockerfile.dockerignore  deploy/  CLAUDE.md
fox-duet/
  backend/  frontend/  docs/rules.md  Dockerfile  Dockerfile.dockerignore  deploy/  CLAUDE.md
the-crew/
  backend/  frontend/  docs/rules.md  Dockerfile  Dockerfile.dockerignore  deploy/  CLAUDE.md
gate/                         password gate: Go server + login page (SvelteKit, prerendered, no JS)
  backend/  frontend/  Dockerfile  Dockerfile.dockerignore  deploy/  CLAUDE.md
```

**Apps are independent, except for `shared/`.** Each game has its own Go module, its own npm project (own `package.json`, lockfile, `node_modules`) and its own image, database and deploy. No app imports from another app. The lobby, session, WebSocket, persistence and ID plumbing that every game needs lives in [shared/](shared/README.md):

- **Backend:** each game's `go.mod` has `replace github.com/martin-pouls/shared/backend => ../../shared/backend`.
- **Frontend:** the shared code is compiled as source by each app via the alias `@board-games/shared` (Vite `resolve.alias` + tsconfig `paths`).
- **What may go there:** only what is the same in every game. Game rules, domain types and game-specific DTOs stay in the game. A change in `shared/` needs the checks in `shared/` and in every app that uses it.

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

## Tech stack (all games)

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

The start page uses the same SvelteKit/Tailwind/shadcn-svelte/tooling stack. It is fully prerendered with `csr = false`, so it ships HTML and CSS only, plus the small inline theme script in `src/app.html` (shadcn components render as static markup there; bits-ui interactivity is not available).

**One frame for every page:** the games, the start page and the gate's login page all render `PageFrame` from `shared/frontend` (games through `AppFrame`): the same top bar (`TopBar`: optional *All games* link, light/dark toggle top right), a `max-w-3xl` column (`wide` gives `max-w-6xl` for future games that need the room) and the imprint/privacy footer. Change the frame there, not per app.

**Light and dark mode:** every app has both, with the toggle in the top bar. The games use mode-watcher; the start page and the gate's login page ship no JS except an inline theme script in their `app.html` that handles clicks on `[data-theme-toggle]`. All use the same `localStorage` key (`mode-watcher-mode`), so on the shared production origin the choice carries over between apps.

## Architecture: hexagonal (ports and adapters)

- **Dependency direction points inward:** `adapters → core/service → core/ports → core/domain`. The domain imports nothing from its module.
- **Shared code:** adapters and `cmd/server` build on `shared/backend` (wire, httpkit, wskit, sqlitedb, serverkit). `core/service` may use only its `idgen` and `lockmap`. `core/domain` and `core/ports` stay stdlib-only.
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
go run ./cmd/server              # API on the game's port (table below), DB at data/<game>.db
go test ./...
go vet ./...
go build -o bin/server.exe ./cmd/server
```
Env vars: `ADDR` (default: the game's backend port below; the images set `:8080`), `DB_PATH` (default `data/<game>.db`), `STATIC_DIR` (optional, e.g. `../frontend/build` to serve the SPA from Go), `WS_ORIGIN_PATTERNS` (optional, comma-separated extra allowed WebSocket origin host patterns), `GAME_RETENTION` (default `168h`: games idle longer are deleted), `CLEANUP_INTERVAL` (default `1h`: how often the janitor runs).

**Local ports** are fixed, so all apps run side by side with plain `go run` / `npm run dev`:

| App | Frontend (Vite, `strictPort`) | Backend |
|---|---|---|
| The Game | `:5173` | `:8080` |
| Regicide | `:5174` | `:8081` |
| Fox Duet | `:5176` | `:8082` |
| The Crew | `:5177` | `:8084` |
| Start page | `:5175` | — |
| Gate | `:5178` (login page only) | `:8083` |

In dev the start page links each tile to the game's frontend port on the same hostname (`devPort` in `start-page/frontend/src/lib/games.ts`); production builds link to `/<game>/`. The other way round, each game's home screen has an *All games* link (`startPageHref` in `shared/frontend/src/startPage.ts`: port 5175 in dev, `/board-games/` in production). Changing a port means changing the app's `vite.config.ts`, its backend default or `BACKEND_URL` default, `games.ts` and `shared/frontend/src/startPage.ts` together.

Frontend (run from `<app>/frontend/`):
```sh
npm run dev        # Vite on the app's port; game frontends proxy /api and /ws to BACKEND_URL (default: the game's backend port)
npm run check      # svelte-check (types)
npm run lint       # prettier --check + eslint
npm run format
npm test           # vitest
npm run build      # static build into frontend/build
```

Production-like local run of a game: `npm run build` in `<game>/frontend/`, then in `<game>/backend/` run `STATIC_DIR=../frontend/build go run ./cmd/server` and open the backend port (e.g. http://localhost:8080).

## Deployment

Target: a Hetzner arm64 Ubuntu box serving https://yoshi-und-kruemli.de (root via `~/.ssh/hetzner_ed25519`).

- **One container per app**, all on the external Docker network `web`. CI builds the images on native arm64 runners and pushes them as `ghcr.io/martin-pouls/board-games/<app>:<commit sha>`:
  - `the-game`, `regicide`, `fox-duet` and `the-crew`: one image each (SPA + Go server), SQLite in the volumes `the-game_data`, `regicide_data`, `fox-duet_data` and `the-crew_data`.
  - `start-page`: the prerendered page served by `caddy file-server`.
  - `gate`: the password check, with its secret and passwords in `/etc/board-games/gate/` (bind mount, written by `setup-server.sh`).
- **Shared Caddy proxy** in `/opt/caddy` ([deploy/caddy/](deploy/caddy/)). It is the only container with public ports and handles HTTPS. One site block serves both `yoshi-und-kruemli.de` and `martin-pouls.de` (the `www.` names redirect). `common.caddy` (shared with the local stack) redirects `/` to `/board-games/` and imports every `sites/*.caddy`.
- **Password gate:** an app is protected when its snippet has `import gate` inside its `handle_path` (the `(gate)` snippet in `deploy/caddy/conf/common.caddy`, `forward_auth` to the gate). The start page keeps `/impressum`, `/datenschutz` and their assets public. Paths that don't import it (the site root, later a personal page) stay public. Without the login cookie, page loads redirect to `/gate/login` and API/WS requests get `401` with code `login_required`; the shared frontend then reloads the page (`shared/frontend/src/login.ts`). Passwords: [deploy/README.md](deploy/README.md#password-gate).
- **Each app ships a snippet** `<app>/deploy/<app>.caddy` that maps its path prefix to its container with `handle_path`. The prefix is stripped, so the servers stay path-agnostic.
- **Base path:** every frontend is built with `BASE_PATH=/<path>` (`paths.base`), set as a build arg in `<app>/deploy/compose.yaml`. Routes use `resolve()`; game API and WS URLs use `__BASE_PATH__`. Never hard-code `/api` or `/ws` without it. Links between apps are plain absolute paths with `rel="external"`.
- **Deploying:** push to `main`. [.github/workflows/deploy.yml](.github/workflows/deploy.yml) tests, builds and deploys the apps that changed since the last successful deploy. Manual runs (*Run workflow*) take `all` or app names. Setup, security model and rollback: [deploy/README.md](deploy/README.md).
  - `<app>/deploy/compose.yaml` is the build definition (context, `BASE_PATH`, image name with `${TAG}`). CI builds it with `docker buildx bake`, and the server runs it with `TAG=<sha>` in `/opt/board-games/<app>/deploy/.env`.
  - Every app uses `shared/` (the start page and the gate for the page frame), so they build with the repo root as context (`context: ../..`, `dockerfile: <app>/Dockerfile`, plus `<app>/Dockerfile.dockerignore`). A change in `shared/` deploys every app whose Dockerfile has `COPY shared/...`.
  - CI deploys as user `deploy`, whose key may only run [deploy/server/board-games-deploy](deploy/server/board-games-deploy) (forced command). `deploy` is root-equivalent (docker group), so that script takes only a commit SHA and app names from CI and reads `<app>/deploy/` from its own clone of `main`, refusing commits that are not on `main`. Keep it that way: never let it use files or config sent by the caller.
  - `deploy/setup-server.sh` is the idempotent bootstrap: Docker, ufw, Caddy, the deploy user and script. Re-run it after changing `deploy/caddy/` or `deploy/server/`; the repo owns the Caddyfile.
  - `deploy/local.sh [app...|down]` runs the same images and Caddy config locally in WSL2 + Docker at http://localhost:8000 (gate password `local`). Setup: [deploy/README.md](deploy/README.md#local-stack).
  - `deploy/deploy.sh [app...]` is the fallback from Git Bash: it ships the *committed* subtree and builds on the server as root (tag `local-<rev>`).
- **Adding an app:** a top-level directory with `Dockerfile`, `deploy/compose.yaml` (unique `name:` and `container_name:`, network `web`) and `deploy/<dir-name>.caddy` (with `import gate` for a game). CI and `deploy.sh` pick it up automatically. Add a tile in `start-page/frontend/src/lib/games.ts`. For a game, copy Regicide's `Dockerfile`, `Dockerfile.dockerignore` and `compose.yaml` (root context, `shared/`) and its shared wiring (`go.mod` replace; `vite.config.ts`, `tsconfig.json`, `layout.css`, see [shared/README.md](shared/README.md)).
- **Git Bash gotcha:** it rewrites `/regicide`-like values in env vars into Windows paths. For a local base-path build use `MSYS2_ENV_CONV_EXCL=BASE_PATH BASE_PATH=/regicide npm run build`.

## Conventions and gotchas

- **SvelteKit 3 uses Node subpath imports (`#lib`), not `$lib`.** Imports need an explicit file, e.g. `#lib/api.js` (resolves to `api.ts`) or `#lib/components/ui/card/index.js`.
- **Vite dev: deps are pre-scanned and pages warmed up** (`optimizeDeps.entries`, `server.warmup` in each game's `vite.config.ts`). Without that, Vite discovers dependencies on first navigation after a restart, re-bundles and reloads the page mid-click, which makes navigation feel slow. Keep it when adding routes outside `src/routes`.
- **Never run a second Vite instance of an app with `--force`** while its dev server runs: they share `node_modules/.vite` and the running server starts answering 500 until it restarts.
- **Svelte 5 runes only** (`$state`, `$derived`, `$props`, `$effect`). No legacy `export let` or stores for new code. A TS module that uses runes must be named `*.svelte.ts`.
- **Secure-context APIs** (`crypto.randomUUID`, `navigator.clipboard`, `navigator.share`) are missing when phones test over `http://192.168.x.x`: always feature-detect and fall back.
- The player token travels in the WS query string: don't log full URLs at the reverse proxy.
- **Abuse limits** (`shared/backend/httpkit/ratelimit.go`, used by every game's router) cap game creation, 404 game lookups (code guessing) and WebSocket connections per client IP, answering 429 `rate_limited`. The client IP is the last `X-Forwarded-For` entry (set by Caddy). Loopback clients are exempt, so local dev is never limited. New routes under `/api/games/` or `/ws/` are covered automatically.
- **Privacy policy:** [start-page/frontend/src/routes/datenschutz/+page.svelte](start-page/frontend/src/routes/datenschutz/+page.svelte) (German, linked from the footer of every app) describes exactly what data the apps process. Update it whenever that changes: new stored fields, browser storage, logging of IPs, cookies, or any third-party service or external resource.
- **Imprint:** [start-page/frontend/src/routes/impressum/+page.svelte](start-page/frontend/src/routes/impressum/+page.svelte) (German, linked from the footer of every app next to the privacy policy). The operator's name, address and e-mail for both pages live in `start-page/frontend/src/lib/site.ts`.
- **Adding shadcn components** (from `<game>/frontend/`): `npx shadcn-svelte@latest add <name> -y`.
  - If it fails with `File '$app/tsconfig' not found`, run `npx svelte-kit sync` first; an npm install can wipe the generated `node_modules/$app`.
  - The generated components live in `src/lib/components/ui/` and are lint-exempt from `svelte/no-navigation-without-resolve`.
- **bits-ui `AlertDialog.Action` does not close its dialog.** Set the bound `open` flag to `false` in its `onclick`, or the overlay stays up whenever the component stays mounted.
- **Go:** keep dependencies minimal and prefer the stdlib. Wrap errors with `%w`. Pass `context.Context` as the first param on anything doing I/O. Domain errors are typed sentinel errors so adapters can map them to HTTP / WS error codes.
- **API JSON uses camelCase.** Keep the TS types in `frontend/src/lib/types.ts` in sync with the Go DTOs.
- Files that run on the Linux server (`*.sh`, `Dockerfile`, `**/deploy/**`) keep LF line endings (`.gitattributes`).
- **Mobile-first:** everything must work on a phone in portrait.
- Before declaring work done, run in each app you touched: `go vet ./... && go test ./...` (backend) and `npm run check && npm run lint && npm test` (frontend).
