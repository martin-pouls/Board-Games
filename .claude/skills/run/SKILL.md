---
name: run
description: Start, stop or check one or more apps locally (Go backend + Vite dev server on their fixed ports) in the background, waiting until they answer. Use before looking at a game in the browser (Playwright, screenshots, the dev fixtures route), for e2e poking, or when asked to start/stop the dev servers.
---

# Running apps locally

```bash
bash .claude/skills/run/run.sh fox-duet            # start backend + frontend, wait until both answer
bash .claude/skills/run/run.sh the-crew deep-sea   # several; deep-sea = the-crew with EDITION=deepSea
bash .claude/skills/run/run.sh status              # which app ports answer
bash .claude/skills/run/run.sh stop fox-duet       # stop what listens on its ports (stop --all: every app)
```

- Starting returns once the servers answer (backend `GET /api/health`, frontend `/`), usually within seconds; give it a 5-minute timeout for a cold Vite start. It prints the frontend URL and the fixtures index (`/dev/fixtures`).
- A port that already answers is left alone: the user often has their own dev servers running. **Stop only apps you started**, and stop them when you're done.
- The backend is built to `$TEMP/board-games-run/<app>-server.exe` and uses its normal dev database (`data/<game>.db`). Logs: `$TEMP/board-games-run/<app>-{backend,frontend}.log`; a failed start prints their tail.
- Supported: the games and the start page. Not the gate (its backend needs secrets, see gate/CLAUDE.md).

## Looking at game states: dev fixtures

Don't play a game into a state through the browser. Every game frontend has a dev-only route that renders `src/lib/game/fixtures.ts` on the real game screen:

- `http://localhost:<port>/dev/fixtures` lists them; `/dev/fixtures/<name>` shows one, with the in-game layout. Deep Sea (`:5179`) lists the Deep Sea fixtures.
- The view comes from the fixture, not the backend; buttons fail with "disconnected". The Crew's lobby still loads the mission list from its backend, so start the backend too (the default).
- Playwright MCP can only write inside the repo: save screenshots under `.playwright-mcp/` (gitignored) and delete them afterwards.
- A state you need isn't there? Add a fixture (see "Dev fixtures" in the root CLAUDE.md).
