# shared/

Plumbing every game shares. It knows nothing about any game: no rules, no domain types, no game-specific wire types. Game rules stay in each game's `core/domain`, which never imports anything from here.

## backend (`github.com/martin-pouls/shared/backend`)

A Go module of its own (overview in [backend/doc.go](backend/doc.go)):

- `wire`: error codes and message envelopes shared by REST and WS.
- `httpkit`: JSON, error mapping, the router skeleton, SPA handler, abuse limits, `Authenticate` (bearer token → player).
- `wskit`: the WebSocket handler and hub; `PublishViews` sends each connected player their own view and kicks players who left.
- `sqlitedb`: connection, migration runner, snapshot store and the generic `Repository` (a game supplies a `Codec` and its port errors).
- `idgen` (codes, IDs, tokens, RNG; `...As` variants return the game's own ID types), `lockmap` (per-game locks), `serverkit` (startup, shutdown, janitor).

A game uses it with a local replace in its `go.mod`:

```
require github.com/martin-pouls/shared/backend v0.0.0
replace github.com/martin-pouls/shared/backend => ../../shared/backend
```

Adapters and `cmd/server` may use any package. `core/service` may use `idgen` and `lockmap` (IDs, RNG, per-game locks). `core/domain` and `core/ports` use nothing from here.

## frontend (`@board-games/shared`)

Svelte/TS source compiled by each app as part of its own build, not a built package:

- **Plumbing:** wire and protocol helpers, `GameSocket` (WS with reconnect), the lobby REST client (`api.ts`), the session store, `prefs.ts` (one remembered per-device preference in localStorage, e.g. the host's last lobby options), `login.ts` (reloads the page when the gate answers 401 `login_required`).
- **Page frame:** `PageFrame` (`TopBar` with optional All games link and theme toggle, centred column with a `wide` option, imprint and privacy footer). `AppFrame` wires it up for a game; on the game screen `topBar={false}` hides the bar and `HeaderThemeToggle` goes into the game's own header row. The start page and the gate use `PageFrame` directly.
- **Lobby pages:** `HomePage` (host, join by code, continue and history; a `hero` snippet for the game's title card) and `JoinPage` (invite link). They take the game's REST calls, session store, error texts and navigation as props. `recent.ts` holds the remembered games and their status labels (`lobby`, `playing`, `won`, `lost`, `ended`, plus labels for a game's own phases).
- **Lobby and table:** `LobbyActions` (close / leave / end game with confirmations), `HeaderButton` (the icon and colour conventions are in its header comment), `ShareInvite`, `ConnectionBanner`, `cardMotion` (headless card flights, presentation only; a leaving card must take its key from its own list item, not from a recomputed value, or it fades instead of flying), new-card marks (`trackNewCards`, `NewCardMark` in the app's `--new-card` colour).
- **Dev fixtures:** `GameSocket.showFixture(view)` (a connection that shows a fixed view and never connects) and `FixtureIndex` (the list on `/dev/fixtures`). See "Dev fixtures" in the root CLAUDE.md.
- Copies of the shadcn components it needs (`src/ui/`).

An app wires it up in four places (see `fox-duet/frontend`):

- `vite.config.ts`: `resolve.alias` `{ [sharedAlias]: sharedSrc }`, `resolve.dedupe: sharedDedupe`, `server.fs.allow: [sharedRoot]`, and a `server.warmup` entry for `../../shared/frontend/src/**/*.{svelte,ts}`, all from [frontend/vite.ts](frontend/vite.ts).
- `tsconfig.json`: `"paths": { "@board-games/shared/*": ["../../shared/frontend/src/*"] }`.
- `src/routes/layout.css`: `@source '../../../../shared/frontend/src';` so Tailwind sees its classes.
- `package.json`: every `peerDependency` of `frontend/package.json` as a dependency. The shared code's imports resolve from the app's `node_modules`; `resolve.dedupe` keeps one copy of svelte, bits-ui, mode-watcher etc.

Imports use the alias with an explicit file: `@board-games/shared/AppFrame.svelte`, `@board-games/shared/connection.svelte.js`.

`frontend/` has its own `node_modules` only for its own checks (`npm run check`, `lint`, `test`); app builds and Docker images don't need it.

## Rules

- **Only plumbing that is the same in every game.** If a game needs something different, give the shared code a parameter or keep it in the game. Don't add game names or switches on them here.
- **A change here changes every game that uses it.** Run the checks in `shared/backend`, `shared/frontend` and in every app that uses it. CI deploys all those apps when `shared/` changes (`.github/scripts/plan-deploy.sh`).
- Apps that use it build their image with the repo root as context; see [deploy/README.md](../deploy/README.md#build-context).
