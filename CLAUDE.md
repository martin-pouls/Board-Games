# Board Games — monorepo

Web implementations of cooperative board and card games, played live in the browser, plus a start page that links to them. Live at https://yoshi-und-kruemli.de/board-games/ (also https://martin-pouls.de/board-games/, same site). The start page and the games are behind a shared password (the gate).

| App | Directory | URL path | What |
|---|---|---|---|
| Start page | [start-page/](start-page/) | `/board-games/` | Static page linking to every game, plus imprint and privacy policy. Frontend only. |
| The Game | [the-game/](the-game/) | `/the-game/` | "The Game" (NSV), 1–5 players, incl. Pro and Extreme. See [the-game/CLAUDE.md](the-game/CLAUDE.md). |
| Regicide | [regicide/](regicide/) | `/regicide/` | Regicide (Badgers From Mars), 1–4 players, incl. solo. See [regicide/CLAUDE.md](regicide/CLAUDE.md). |
| Fox Duet | [fox-duet/](fox-duet/) | `/fox-duet/` | "The Fox in the Forest: Duet" (Foxtrot Games), 2 players, 3 difficulty levels. See [fox-duet/CLAUDE.md](fox-duet/CLAUDE.md). |
| The Crew | [the-crew/](the-crew/) | `/the-crew/` | "The Crew: The Quest for Planet Nine" (Kosmos), 2–5 players (2 with JARVIS), 50 missions. See [the-crew/CLAUDE.md](the-crew/CLAUDE.md). |
| Gate | [gate/](gate/) | `/gate/` | Shared-password login in front of the start page and the games (Caddy `forward_auth`). See [gate/CLAUDE.md](gate/CLAUDE.md). |

This file holds the rules common to all apps. Each game's CLAUDE.md adds its domain model, per-player view and look. **Read it before working on that game.**

**Reference game: [fox-duet/](fox-duet/).** It has the newest pattern without special cases. When a convention isn't written down, look at how Fox Duet does it, and copy it when starting a new game.

## Repository layout

Top level: `.github/` (CD), `deploy/` (shared Caddy proxy, server scripts, local prod-like stack), `shared/`, `start-page/`, `gate/`, and one directory per game. Every game has the same shape: `backend/` (Go, hexagonal, see Architecture), `frontend/` (SvelteKit SPA; `src/lib/components/ui/` is generated shadcn-svelte, don't hand-edit heavily), `docs/rules.md`, `Dockerfile`, `deploy/`.

The wire contract is the code: `apidto` (Go) and `frontend/src/lib/types.ts` (TS mirror). API JSON uses camelCase.

## Shared code

**Apps are independent, except for [shared/](shared/README.md).** Each game has its own Go module, npm project, image, database and deploy. No app imports from another app. Backend: `replace github.com/martin-pouls/shared/backend => ../../shared/backend` in `go.mod`. Frontend: compiled as source via the alias `@board-games/shared`.

- **What goes there:** code that is identical in every game, changes only for plumbing reasons and has no game-specific behaviour. Today: the lobby REST/WS/session/persistence plumbing (incl. WS view publishing, bearer auth, typed IDs, the SQLite snapshot repository), the page frame, the home and join pages, the remembered-games list, the lobby's leave/end actions, connection banner, card motion and new-card marks.
- **What stays in the game:** rules, domain and DTOs, the service, WS command dispatch, and the Lobby/Table/GameOver screens: everything that defines how a game plays or feels. Differences are props, snippets or parameters, never switches on game names.
- **A change in `shared/` changes every game:** run the checks in `shared/` and in every app, and CI redeploys them all.

## Tech stack (all games)

Stdlib-first Go (`net/http` `ServeMux`, no web framework, minimal dependencies); WebSocket messages compressed from 512 bytes, no delta updates; pure-Go SQLite, **no CGO**, embedded `migrations/NNNN_name.sql` applied at startup. Frontend: SvelteKit SPA via `adapter-static` (`ssr = false`, fallback `index.html`); shadcn-svelte style `vega`, and each game has its own theme (tokens in `src/routes/layout.css`, `:root` light, `.dark` dark).

The start page and the gate's login page use the same stack, fully prerendered with `csr = false` (HTML and CSS, plus an inline theme script in `src/app.html`).

**One frame, light and dark everywhere:** every page renders `PageFrame` from `shared/frontend` (details in [shared/README.md](shared/README.md)). Every app has light and dark mode with the same `localStorage` key (`mode-watcher-mode`), so the choice carries over between apps.

## Architecture: hexagonal (ports and adapters)

- **Dependency direction points inward:** `adapters → core/service → core/ports → core/domain`. `core/domain` and `core/ports` are stdlib-only; `core/service` may use only `idgen` and `lockmap` from `shared/backend`.
- **All game rules live in `core/domain`.** Services orchestrate: load aggregate → call domain method → save → publish events. Adapters translate only.
- **The server is authoritative.** Clients send *intents*; the domain validates them and returns typed sentinel errors for illegal moves, which adapters map to error codes. The client may pre-validate only for UX.
- **Domain methods validate fully before mutating:** on error the state stays unchanged (tests assert this). Illegal moves are rejected, never punished.
- **Per-player views:** never send other players' hands (or hidden deck order) to a client. Build a projection per player.
- **Transport split:** REST (`/api/...`) only for the lobby: health, create, summary, join, view, leave (plus static data such as The Crew's `GET /api/missions`). Everything from *start* to *rematch* goes over the WebSocket (`/ws/...`): clients send actions, the server pushes each player their view and the events. No gameplay routes over REST. Gameplay service methods return only `error`; the actor's new view arrives with the broadcast.
- **Persistence:** each game is a JSON snapshot of the domain aggregate (`games.state`, explicit DTOs in the sqlite adapter, no struct tags on domain types) with a `version` column for optimistic concurrency. Game state survives a restart; connections and presence are in memory. Old snapshots must keep loading (map renamed values in the DTO).
- **Tests:** each `*_test.go` sits next to its code. The domain has thorough table-driven tests for every rule, app decision and edge case in the game's docs/rules.md.
- **Rules:** each game's `docs/rules.md` is the source of truth. Read it before touching game logic; don't re-read the rule PDFs. If it doesn't answer a rule question, ask, then record the answer there (marked **App decision**).

## Game lifecycle (every game)

- **Phases:** `lobby → playing → won | lost`, plus `lobby → ended` when the host closes the lobby. A game may add its own in-between phases (The Game's `choosingFirstPlayer`, The Crew's `debrief`); `won`, `lost` and `ended` are terminal (history on the home screen). Every action checks the phase and the actor first.
- **Lobby:** alone, the host can *Close lobby* (no confirmation). Otherwise anyone can *Leave* (seat freed; if the host leaves, the next seat becomes host) and the host can *End game* for everyone (→ `ended`), both with confirmation. Use the shared `LobbyActions`.
- **Running game:** any player can *Give up* at any time, with confirmation → `lost`. A game is otherwise lost only by its rules. After `won`/`lost` anyone can start a rematch (The Crew: next attempt / next mission / back to lobby).
- **Header buttons** (`HeaderButton`): X = close lobby, DoorOpen = leave, Flag = give up or end. Text labels in the lobby header; icon only with a tooltip in the in-game header (no room on phones). Red only on the confirm button inside the dialog, never in a header. On the game screen (`/g/[code]`) the layout passes `topBar={false}` and the headers render `HeaderThemeToggle` at the right end.
- **In-game help:** every game has a `RulesHelp` dialog in the table header.

## Commands

The toolchain was installed via winget: Go at `C:\Program Files\Go\bin`, Node at `C:\Program Files\nodejs`; a shell opened before the install may need its PATH refreshed.

Backend scripts run from `<game>/backend/`, frontend scripts from `<app>/frontend/` (game frontends proxy `/api` and `/ws` to `BACKEND_URL`, default the game's backend port).
Backend env vars: `ADDR` (default: the game's backend port below; the images set `:8080`), `DB_PATH` (default `data/<game>.db`), `STATIC_DIR` (optional, e.g. `../frontend/build` to serve the SPA from Go), `WS_ORIGIN_PATTERNS` (optional extra allowed WebSocket origin host patterns), `GAME_RETENTION` (default `168h`: idle games are deleted), `CLEANUP_INTERVAL` (default `1h`).

**Local ports** are fixed, so all apps run side by side:

| App | Frontend (Vite, `strictPort`) | Backend |
|---|---|---|
| The Game | `:5173` | `:8080` |
| Regicide | `:5174` | `:8081` |
| Fox Duet | `:5176` | `:8082` |
| The Crew | `:5177` | `:8084` |
| Start page | `:5175` | — |
| Gate | `:5178` (login page only) | `:8083` |

In dev the start page links to each game's frontend port (`devPort` in `start-page/frontend/src/lib/games.ts`) and the games link back to port 5175 (`shared/frontend/src/startPage.ts`). Changing a port means changing the app's `vite.config.ts`, its backend default or `BACKEND_URL` default, `games.ts` and `startPage.ts` together.

**Before declaring work done,** run in each app you touched: `go vet ./... && go test ./...` (backend) and `npm run check && npm run lint && npm test` (frontend).

## Deployment

Hetzner arm64 Ubuntu box (root via `~/.ssh/hetzner_ed25519`), one container per app behind a shared Caddy proxy. **Push to `main` deploys** what changed (CI tests, builds arm64 images to ghcr.io, deploys). Setup, security model, gate passwords, rollback and the local stack: [deploy/README.md](deploy/README.md). What matters when changing code:

- **Base path:** every frontend is built with `BASE_PATH=/<path>` (build arg in `<app>/deploy/compose.yaml`). Routes use `resolve()`; game API and WS URLs use `__BASE_PATH__`, never a bare `/api` or `/ws`. Links between apps are absolute paths with `rel="external"`.
- **Each app ships a Caddy snippet** `<app>/deploy/<app>.caddy` (`handle_path`, prefix stripped, so servers stay path-agnostic). A game's snippet has `import gate`, which puts it behind the password gate.
- **The deploy script trusts nothing from the caller** ([deploy/server/board-games-deploy](deploy/server/board-games-deploy)): only a commit SHA on `main` and app names; config comes from its own clone of `main`. Keep it that way. After changing `deploy/server/` or `deploy/setup-server.sh`, re-run `deploy/setup-server.sh`.
- **Adding an app:** use the `add-app` skill (`.claude/skills/add-app/`). CI picks up any top-level directory with a Dockerfile and `deploy/`.
- **Git Bash gotcha:** it rewrites `/regicide`-like env values into Windows paths. For a local base-path build use `MSYS2_ENV_CONV_EXCL=BASE_PATH BASE_PATH=/regicide npm run build`.

## Conventions and gotchas

- **SvelteKit 3 uses Node subpath imports (`#lib`), not `$lib`.** Imports need an explicit file, e.g. `#lib/api.js` (resolves to `api.ts`) or `#lib/components/ui/card/index.js`. Shared code is imported as `@board-games/shared/<file>.js` and never imports `$app/*` or `#lib/*`.
- **Svelte 5 runes only** (`$state`, `$derived`, `$props`, `$effect`). A TS module that uses runes must be named `*.svelte.ts`.
- **Vite dev: deps are pre-scanned and pages warmed up** (`optimizeDeps.entries`, `server.warmup` in each `vite.config.ts`), so navigation after a restart doesn't re-bundle and reload mid-click. Keep it when adding routes outside `src/routes`. Never run a second Vite instance of an app with `--force` while its dev server runs (shared `node_modules/.vite`).
- **Secure-context APIs** (`crypto.randomUUID`, `navigator.clipboard`, `navigator.share`) are missing when phones test over `http://192.168.x.x`: feature-detect and fall back.
- **Tokens:** the player token travels in the WS query string: never log full URLs (the apps log paths only, Caddy writes no access log).
- **Abuse limits** (`shared/backend/httpkit/ratelimit.go`) cap game creation, 404 lookups and WS connections per client IP (last `X-Forwarded-For` entry), answering 429 `rate_limited`. Loopback is exempt. New routes under `/api/games/` or `/ws/` are covered automatically.
- **Privacy policy and imprint** ([datenschutz](start-page/frontend/src/routes/datenschutz/+page.svelte), [impressum](start-page/frontend/src/routes/impressum/+page.svelte), German; operator data in `start-page/frontend/src/lib/site.ts`). Update the privacy policy whenever the processed data changes: stored fields, browser storage, IP logging, cookies, third-party resources.
- **Adding shadcn components** (from `<app>/frontend/`): `npx shadcn-svelte@latest add <name> -y`. If it fails with `File '$app/tsconfig' not found`, run `npx svelte-kit sync` first.
- **bits-ui `AlertDialog.Action` does not close its dialog.** Set the bound `open` flag to `false` in its `onclick`.
- **Go:** wrap errors with `%w`; `context.Context` first on anything doing I/O; domain errors are typed sentinels.
- Files that run on the Linux server (`*.sh`, `Dockerfile`, `**/deploy/**`) keep LF line endings (`.gitattributes`).
- **Mobile-first:** everything must work on a phone in portrait.
