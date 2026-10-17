---
name: add-app
description: Checklist for adding a new app (game, page or service) to the board-games monorepo: files, deploy wiring, start-page tile. Use when asked to add or scaffold a new game or app.
---

# Adding an app

- A top-level directory with `Dockerfile`, `Dockerfile.dockerignore`, `deploy/compose.yaml` (unique `name:` and `container_name:`, network `web`, repo root as build context) and `deploy/<dir-name>.caddy`. CI picks it up.
- Add a tile in `start-page/frontend/src/lib/games.ts`.
- For a game, copy Fox Duet's files and shared wiring (`go.mod` replace; `vite.config.ts`, `tsconfig.json`, `layout.css`, see `shared/README.md`).
- Pick fixed local ports and update `vite.config.ts`, the backend default or `BACKEND_URL` default, `games.ts` and `shared/frontend/src/startPage.ts` together.
- A game's Caddy snippet starts with `import gate`.
- Add the game's ports to the table in `.claude/skills/run/run.sh`.
- A game gets Fox Duet's dev fixtures wiring: `GameScreen.svelte`, `src/routes/dev/fixtures/`, the layout's `topBar` condition, `fixtures.ts` with one fixture per UI state and `fixtures.test.ts` (root CLAUDE.md, "Dev fixtures").
- Done when `bash .claude/skills/check/check.sh <dir> start-page` passes. The check skill finds the new app on its own (any top-level `backend/go.mod` or `frontend/package.json`).
