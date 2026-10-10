# The Game — online multiplayer

A web implementation of the cooperative card game **"The Game"** (Steffen Benndorf, NSV), played live by 1–5 players in the browser.

- **Phase 1:** the regular game, including the Pro variant. Implemented.
- **Phase 2:** **The Game Extreme**, which adds command cards. Implemented (plan and design decisions: [docs/extreme-plan.md](docs/extreme-plan.md)). The host enables it with the *Extreme* lobby toggle, which combines freely with Pro and Reduced hand size.

Keep every domain concept open for further card kinds (the mini-expansions Joker, Zehner, On Fire), the way Extreme slotted in (see *Extensibility* below).

**Rules:** [docs/rules.md](docs/rules.md) is the source of truth. It contains the English rules summary, the Extreme card→command mapping, and the interpretations we chose for ambiguities. Read it before touching game logic. Do not re-read the PDFs or card photos. If a rule question is not answered there, ask, then record the answer in that file.

## Tech stack

| Area | Choice | Notes |
|---|---|---|
| Backend language | Go 1.27 | stdlib-first, minimal dependencies |
| HTTP | `net/http` `ServeMux` | Go 1.22+ method/path patterns; no web framework |
| Realtime | WebSockets via `github.com/coder/websocket` | for live gameplay; add when implementing `adapters/ws` |
| Persistence | SQLite via `modernc.org/sqlite` | pure Go, **no CGO / no C toolchain needed**; embedded file DB, no separate DB server |
| Migrations | Embedded `*.sql` files + tiny runner | `backend/internal/adapters/sqlite/migrations/NNNN_name.sql`, applied in order at startup |
| Logging | `log/slog` | |
| Frontend | SvelteKit 3 + Svelte 5 (runes) + TypeScript | SPA via `adapter-static` (`ssr = false`, fallback `index.html`) |
| UI | Tailwind CSS 4 + **shadcn-svelte** (bits-ui, lucide icons) | style `vega`, base color `zinc` |
| Frontend tooling | Vite, Vitest, ESLint, Prettier, svelte-check | |

## Repository layout

```
backend/                      Go module github.com/martin-pouls/the-game/backend
  cmd/server/                 main: config from env, wiring, graceful shutdown
  internal/core/              the hexagon — no imports of adapters, net/http or database/sql
    domain/                   pure game model + rule engine (stdlib only)
    ports/                    interfaces: driving (use cases) + driven (repo, publisher)
    service/                  use-case implementations of driving ports
  internal/adapters/          everything that talks to the outside world
    apidto/                   JSON wire contract (DTOs, mappers, error codes) shared by httpapi + ws
    httpapi/                  REST handlers (+ serves built SPA when STATIC_DIR set)
    ws/                       WebSocket hub: live game events/commands
    sqlite/                   repository implementations + migrations
    memory/                   in-memory repository (tests)
    repotest/                 shared repository contract test suite
  internal/e2e/               end-to-end test: sqlite + service + REST + WS over httptest
frontend/                     SvelteKit SPA
  src/lib/api.ts              REST client
  src/lib/types.ts            TS mirror of the Go DTOs in apidto
  src/lib/game/               WS protocol + GameConnection (runes), formatting helpers, Extreme command metadata (commands.ts), fixtures
  src/lib/components/game/    game screen components (Lobby, Board, GameOver, ...)
  src/lib/components/ui/      shadcn-svelte components (generated — don't hand-edit heavily)
  src/routes/                 pages
docs/rules.md                 rules reference (authoritative)
docs/plan.md                  Phase 1 implementation plan / wire contract
docs/extreme-plan.md          Phase 2 (Extreme) plan / wire contract changes
```

## Architecture: hexagonal (ports and adapters)

- **Dependency direction points inward:** `adapters → core/service → core/ports → core/domain`. The domain imports nothing from this module.
- **All game rules live in `core/domain`.** Services orchestrate: load aggregate → call domain method → save → publish events. Adapters translate only and contain no rules.
- **The server is authoritative.** Clients send *intents* (e.g. "play card X on pile Y", "end turn"). The domain validates them and returns errors for illegal moves. Never trust client-side validation.
- **Per-player views:** never send other players' hands to a client. Build a projection per player: own hand, piles, draw-pile count, other players' hand *counts*, whose turn, active commands, etc.
- **Transport split:**
  - **REST** (`/api/...`) for lobby and request/response actions: create game, join, list, start, fetch current view.
  - **WebSocket** (`/ws/...`) for live gameplay: server pushes state/events to everyone in a game, and clients may send turn actions over it too.
  - Both call the same driving ports.
- **Persistence:** each game is stored as a JSON snapshot of the domain aggregate (`games.state`) with a `version` column for optimistic concurrency. The aggregate must stay JSON-serialisable through explicit DTOs in the sqlite adapter, not through struct tags on domain types. Live connections and lobby presence may be in memory; game state must survive a restart.
- Each `*_test.go` sits next to its code. The domain must have thorough table-driven unit tests: every rule and edge case in docs/rules.md.

## Extensibility for The Game Extreme

These are built in and Extreme relies on them; keep them when adding card kinds:

- **Variant / ruleset config on the game:** variant (`classic` | `extreme`), min cards per turn, hand-size table, Pro flag. Rules read from the config and never hard-code 2/6/7/8.
- **Cards are structs, not ints.** For example `Card{Value int; Command *Command}` (or a kind field), because Extreme cards carry commands and the mini-expansions (Joker, Zehner, On Fire) would add other card kinds.
- **The deck is built by a variant-specific factory.** Extreme swaps in the 28 command cards listed in docs/rules.md.
- **Turn state is explicit.** Track cards played this turn and the piles touched, because commands constrain the *current turn*: STOP ends it, 3! needs an exact count, Single pile, Skull must be covered.
- **Placement validation is a composable rule check.** Base ascending/descending/backwards-trick rules, plus modifiers derived from the active (visible) Infinity commands and the lightning commands played this turn.
- **Hooks at turn end:** end-of-turn checks (Skull covered, 3! satisfied) and a draw-count policy (normal refill vs. Draw only one).
- **Communication restriction is game state.** The UI must be able to disable chat or hints while a "No communication" card is visible.
- **Illegal moves are blocked, never punished.** Every action that would break a rule or command is rejected with a typed domain error, and state stays unchanged. The game is lost only when the active player is *stuck*: no legal card play and ending the turn is blocked. See *Enforcement* in docs/rules.md.
- **Stuck detection has no lookahead** (owner decision): the game is lost only when the active player has no legal card play *and* cannot end the turn. A player who can still play a card but can no longer reach the minimum keeps playing until nothing fits.

## Commands

The toolchain was installed via winget. A shell opened before the install may need its PATH refreshed. Go is at `C:\Program Files\Go\bin` and Node at `C:\Program Files\nodejs`.

Backend (run from `backend/`):
```sh
go run ./cmd/server              # API on :8080, DB at backend/data/the-game.db
go test ./...
go vet ./...
go build -o bin/server.exe ./cmd/server
```
Env vars: `ADDR` (default `:8080`), `DB_PATH` (default `data/the-game.db`), `STATIC_DIR` (optional, e.g. `../frontend/build` to serve the SPA from Go), `WS_ORIGIN_PATTERNS` (optional, comma-separated extra allowed WebSocket origin host patterns), `GAME_RETENTION` (default `168h`: games idle longer are deleted), `CLEANUP_INTERVAL` (default `1h`: how often the janitor runs).

Frontend (run from `frontend/`):
```sh
npm run dev        # Vite on :5173, proxies /api and /ws to BACKEND_URL (default http://localhost:8080)
npm run check      # svelte-check (types)
npm run lint       # prettier --check + eslint
npm run format
npm test           # vitest
npm run build      # static SPA into frontend/build
```

Production-like local run: `npm run build` in `frontend/`, then in `backend/` run `STATIC_DIR=../frontend/build go run ./cmd/server` and open http://localhost:8080.

## Deployment

Live at https://yoshi-und-kruemli.de/the-game/. The server is a Hetzner arm64 Ubuntu box (root via `~/.ssh/hetzner_ed25519`).

- **Shared Caddy proxy** in `/opt/caddy` ([deploy/caddy/](deploy/caddy/)).
  - It is the only container with public ports, and it handles HTTPS automatically.
  - Apps join the Docker network `web` and add a snippet to `/opt/caddy/conf/sites/`.
- **The Game** is one image ([Dockerfile](Dockerfile): SPA + Go server) run by [deploy/compose.yaml](deploy/compose.yaml).
  - The SQLite DB lives in the volume `the-game_data`.
  - [deploy/the-game.caddy](deploy/the-game.caddy) strips `/the-game`, so the Go server stays path-agnostic.
- **Base path:** the SPA is built with `BASE_PATH=/the-game` (`paths.base`). Routes use `resolve()`; API and WS URLs use `__BASE_PATH__`. Never hard-code `/api` or `/ws` without it.
- **Deploying:** run `deploy/deploy.sh` from Git Bash.
  - It ships the *committed* HEAD via `git archive` and builds on the server (native arm64).
  - `deploy/setup-server.sh` is the one-time, idempotent bootstrap: Docker, ufw, Caddy.
- **Git Bash gotcha:** it rewrites `/the-game` in env vars into a Windows path. For a local base-path build use `MSYS2_ENV_CONV_EXCL=BASE_PATH BASE_PATH=/the-game npm run build`.

## Conventions and gotchas

- **SvelteKit 3 uses Node subpath imports (`#lib`), not `$lib`.** Imports need an explicit file, e.g. `#lib/api.js` (resolves to `api.ts`) or `#lib/components/ui/card/index.js`.
- **Svelte 5 runes only** (`$state`, `$derived`, `$props`, `$effect`). No legacy `export let` or stores for new code unless there's a reason.
- **Adding shadcn components:** `npx shadcn-svelte@latest add <name> -y`.
  - If it fails with `File '$app/tsconfig' not found`, run `npx svelte-kit sync` first; an npm install can wipe the generated `node_modules/$app`.
  - The generated components live in `src/lib/components/ui/` and are lint-exempt from `svelte/no-navigation-without-resolve`.
- **Go:** keep dependencies minimal and prefer the stdlib. Wrap errors with `%w`. Pass `context.Context` as the first param on anything doing I/O. Domain errors are typed sentinel errors so adapters can map them to HTTP / WS error codes.
- **API JSON uses camelCase.** Keep the TS types in `frontend/src/lib` in sync with the Go DTOs.
- Run `go vet ./... && go test ./...` and `npm run check && npm run lint` before declaring work done.
