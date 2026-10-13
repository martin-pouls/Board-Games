# shared/

Plumbing every game shares. It knows nothing about any game: no rules, no domain types, no game-specific wire types. Game rules stay in each game's `core/domain`, which never imports anything from here.

## backend (`github.com/martin-pouls/shared/backend`)

A Go module of its own (overview in [backend/doc.go](backend/doc.go)): `wire`, `httpkit` (incl. rate limits), `wskit`, `sqlitedb`, `idgen`, `lockmap`, `serverkit`.

A game uses it with a local replace in its `go.mod`:

```
require github.com/martin-pouls/shared/backend v0.0.0
replace github.com/martin-pouls/shared/backend => ../../shared/backend
```

Adapters and `cmd/server` may use any package. `core/service` may use `idgen` and `lockmap` (IDs, RNG, per-game locks). `core/domain` and `core/ports` use nothing from here.

## frontend (`@board-games/shared`)

Svelte/TS source compiled by each app as part of its own build, not a built package: wire and protocol helpers, `GameSocket` (WS connection with reconnect), the lobby REST client, `login.ts` (reloads the page when the access gate answers 401 `login_required`, so the browser hits the login redirect), the session store, the page frame (`PageFrame`: `TopBar` with optional All games link and the theme toggle, a centred column with a `wide` option, the imprint and privacy footer; `AppFrame` wires it up for a game, and its `topBar={false}` hides the top bar on the game screen, where `HeaderThemeToggle` puts the toggle into the game's own header row to save a row on phones; the start page and the gate's login page use `PageFrame` directly and handle the toggle in their no-JS theme script), `ShareInvite`, `ConnectionBanner`, `cardMotion` (headless card flights between hand, table and off-screen, presentation only), new-card marks (`trackNewCards` in `newCards.svelte.ts` remembers which cards just came into a hand for a few seconds, `NewCardMark` draws the dot and glow in the app's `--new-card` colour), and copies of the shadcn `button`/`input` it needs.

An app wires it up in four places (see `regicide/frontend`):

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
