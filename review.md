# Review: architecture, security, performance, coherence

As of 2026-10-15 (`d59af13`). I sampled the code rather than reading every file. I read the shared plumbing (`shared/backend`, `shared/frontend`) in full, plus the gate, Caddy, compose, CI, one service and one WebSocket handler per pattern, and each game's lobby, table and game route. I did not review the domain logic and rules.

## Verdict

**There are no large problems.** The architecture holds up: the hexagonal split is real, the shared plumbing is careful, and the security edge (gate, tokens, logging) is thought through. What remains is copy-paste between the games and UI differences that grew because each game copied the one before it. The quick fixes are done (see [Done](#done) at the bottom); this file keeps what is still open.

## 1. Code reuse: trade-offs

What is copied today (the diffs are only names and imports):

| Copied code | Where | Differences |
|---|---|---|
| `ws/hub.go` | all 4 | identical |
| `httpapi/auth.go`, `service/ids.go` | all 4 | identical |
| `httpapi/router.go` | all 4 | The Crew adds `/api/missions` |
| `sqlite/games.go` | 3 identical, The Game ~28 lines | small |
| `service.go` lifecycle (create, join, leave, view, authenticate, `act`, rematch) | all 4 | ~190 identical lines per game |
| `routes/join/[code]/+page.svelte` | all 4 | 2 lines (title) |
| `routes/+page.svelte` (home) | all 4 | hero and options; The Game more |
| `lib/game/recent.ts` | all 4 | near-identical |

**For sharing:** a bug fix lands once instead of four times. The planned fifth game (Deep Sea, possibly more) starts thinner. Differences that nobody intended stop growing.

**Against sharing:**

- **Coupling.** Every change in `shared/` redeploys and has to be re-checked in all apps.
- **Generic Go is harder to read.** A `Service[G]` needs an interface the domain type satisfies (`PlayerByID`, `ViewFor`, `PullEvents`, `Leave`, `Clone`, `Rematch`, …). That's fine structurally, since the domain doesn't import it, but tracing an action then means jumping between modules. It also breaks today's rule that `core/service` uses only `idgen`/`lockmap`.
- **Intentional divergence becomes a parameter or switch.** The Crew's rematch is "next mission / back to lobby", and The Game's end-game semantics differ (§2). Each such difference either becomes a parameter or the game drops out of the shared code again.
- **Shared tests then need a fake game.**

**Recommendation:** share code that is identical, changes only for plumbing reasons, and has no visible game-specific behaviour. Keep everything that defines how a game plays or feels in the game.

| Do it | Maybe | Don't |
|---|---|---|
| WS hub publisher (`wskit.Publisher` taking a view builder func) | Service lifecycle skeleton: only if a generic version stays readable; otherwise accept the copy | Lobby / Table / GameOver components |
| `auth.go` (`httpkit` helper taking an authenticate func) | Home page: a shared shell with snippets for hero and options | WS command dispatch, `apidto`, `client.ts` |
| `ids.go` | | Domain, rules, game DTOs |
| SQLite games repository (generic over a snapshot codec) | | |
| Join page (props: title, tagline) | | |
| `recent.ts` | | |

The "Do it" column is low-risk, because those parts are already identical, and it removes most of the copy-paste. Do it before Deep Sea starts copying again. Decide on the service skeleton afterwards: if the "Do it" column already feels like enough, stop there.

## 2. Coherence between the 4 games

The pattern: each game copied the previous one (The Game → Regicide → Fox Duet → The Crew), and later improvements didn't flow back. The Game in particular is a generation behind.

### Ending, leaving, giving up: the rules differ

| | The Game | Regicide, Fox Duet, The Crew |
|---|---|---|
| Lobby, non-host player | **no way to leave** (only Home, which keeps the seat) | *Leave* (removes the player; if the host leaves, the next seat becomes host) |
| Lobby, alone | *End game* + confirmation | *Close lobby*, no confirmation |
| Lobby, host with others | *End game* (game finishes as `aborted`, result screen, rematch possible) | *End game* (phase `ended`, "Lobby closed" screen) |
| Running game | host only: *End game* at any time; *Give up* only for the active player and only while **stuck** | any player: *Give up* at any time → lost |

**Your question about The Game:** right, the *host* can end the game from the lobby. What's missing is **Leave** for everyone else. A player who joined by mistake or wants to leave can only press Home, which keeps them seated. When the host starts, that player is dealt in, and the game blocks on their turn until they come back or the host ends it. In the other three games they can leave and their seat frees up. The domain has no `Leave`, and there's no `/leave` route or button.

**Proposal:** adopt the newer model in The Game: `Leave` in the domain, the route and the lobby buttons as in Regicide/Fox Duet/The Crew. Whether *Give up* in The Game stays "only when stuck" is a rules decision. It fits the "no lookahead" decision in its CLAUDE.md, but "anyone can give up anytime" is more consistent. Decide that, then record it in `the-game/docs/rules.md`.

### The leave/give-up buttons look different everywhere

| Where | The Game | Regicide | Fox Duet | The Crew |
|---|---|---|---|---|
| Lobby | Flag "End game" with text (host) | X "Close lobby" / DoorOpen "Leave" / Flag "End game", all with text | same as Regicide | same as Regicide |
| Table header | Flag **icon only**, means *End game* (host) | Flag icon only, *Give up* | Flag **with text** "Give up" (text hidden below 24rem) | Flag icon only, tooltip "Give up this attempt" |
| Elsewhere | *Give up* as a **solid red** text button in the turn bar (when stuck) | — | — | — |

What's inconsistent:

- **The Flag icon means two things:** *End game* in The Game's table, *Give up* elsewhere.
- **Text vs. icon-only** varies at the same spot.
- **Red is hit and miss.** `HeaderButton destructive` turns red only on hover, which phones don't have, so on phones it is always grey, while The Game's turn-bar button is permanently red.

**Proposal:**

- Icons: X = close lobby, DoorOpen = leave, Flag = give up or end a running game.
- Labels: text in the lobby (there's room); icon only plus tooltip in the table header (no room on phones).
- Colour: always red only inside the confirmation dialog, never in the header.
- One shared `LobbyActions` component (alone / leave / end game + the two dialogs) would enforce this. The lobby versions in the three newer games are already nearly identical.

### Other differences

| | The Game | Regicide | Fox Duet | The Crew | Proposal |
|---|---|---|---|---|---|
| Rules help in game (`RulesHelp`) | ❌ | ✅ | ✅ | ✅ | Add to The Game (the Extreme commands especially need it) |
| Card motion (`cardMotion`) | ❌ | ❌ | ✅ | ✅ | Nice to have in The Game and Regicide; not urgent |
| New-card mark | ✅ | ✅ | ✅ | ✅ | — |
| Hand sort menu | not needed (numbers) | ✅ | no menu | no menu | Fine as it is: game-specific |
| Rematch | ✅ | ✅ | ✅ | next mission / back to lobby | Fine: a Crew session is several missions |
| Phase names | `lobby`, `choosingFirstPlayer`, `playing`, `finished` + `outcome` | `lobby`, `playing`, `won`/`lost`, `ended` | like Regicide | `lobby`, `playing`, `debrief`, `ended` | Optional: settle on `lobby → playing → won/lost`, `ended` for a closed lobby. Only worth it together with the Leave/End-game rework in The Game |

## 3. CLAUDE.md files

They should reflect the decisions above, and they can be much shorter.

- **One reference game.** Today the chain is Regicide → "look at the-game", Fox Duet → "follow regicide", The Crew → "follow fox-duet", while the root says "copy Regicide's Dockerfile". Name **one** reference game in the root (The Crew or Fox Duet, the newest pattern) and drop the chain.
- **Write down the conventions that drifted:**
  - The lobby/table action model and button conventions (§2).
  - What goes into `shared/` (the "Do it" column of §1).
  - Phase naming, if you align it.
- **Make them concise:**
  - Root CLAUDE.md (202 lines): the deployment section largely repeats [deploy/README.md](deploy/README.md), and the page-frame description repeats [shared/README.md](shared/README.md). Keep one line plus a link.
  - Game files: drop what each one repeats or what the code shows anyway. Examples: the rules paragraph (move it to the root once), the "Light/dark toggle via mode-watcher; theme tokens live in …" sentence, the "Deployment: container …, volume …, Caddy snippet …" line, and local ports already in the root table.
  - Keep in each game file only: its rules pointer, the domain modelling decisions, and the per-player view (the security-relevant part) and the look and feel.

## Suggested order

1. The Game: Leave in the lobby, rules help, end/give-up semantics (decide, then implement).
2. Align the header buttons (shared `LobbyActions`).
3. Move the "Do it" column of §1 into `shared/`, before Deep Sea.
4. Update and shorten the CLAUDE.md files (after 1–3, so they describe the new state).
5. Optional: card motion in The Game/Regicide, phase names.

## Done

Done on 2026-10-15, after the review. Relevant for §3: the CLAUDE.md files should describe this state.

- **WebSocket compression.** [wskit/handler.go](shared/backend/wskit/handler.go) accepts with `CompressionNoContextTakeover` (messages from 512 bytes, so acks and pings stay uncompressed). No delta updates: compression gets most of that saving.
- **Security headers** for every response, in the `(site)` snippet of [common.caddy](deploy/caddy/conf/common.caddy): `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: same-origin`, `X-Frame-Options: DENY` (deferred, so they replace the gate's own instead of doubling them). A strict CSP is still open, but not urgent without `{@html}`. Goes live with the next run of `deploy/setup-server.sh`.
- **CI builds in parallel with the tests.** `build` needs only `plan`; `deploy` waits for all tests and builds, so nothing untested is deployed.
- **CI checks `shared/` once.** A `test-shared` job checks `shared/backend` and `shared/frontend` only when they changed (`plan-deploy.sh` reports which), instead of in every app's job. The app jobs only install `shared/frontend` for their type check.
- **No gameplay REST.** Every game's REST is now only health, create, summary, join, view and leave (The Game has no leave yet; The Crew adds `GET /api/missions`). Everything else goes over WebSocket. The handlers, their tests, the TS functions and `giveUp`/`rematch` in the shared `lobbyApi` are gone; the e2e tests drive gameplay over WebSocket, and each httpapi test asserts the old routes answer 404/405. The gameplay service methods return only `error`. The transport split in the root CLAUDE.md says so.
- **Dropped:** compose hardening, a per-socket message rate limit and separate Docker networks. Behind the gate, among friends, not worth it.
