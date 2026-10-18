# The Crew: Mission Deep Sea — implementation plan

Status (2026-10-11): **ready to implement**, all decisions in §6. Sources gathered, architecture chosen (§2), rules draft with the mission
table and all 96 task cards written ([rules.md](rules.md)). Next step: WP1, with WP2 and WP3 in
parallel (§5).

## 0. Sources (outside the repo, not committed)

`C:\Users\marti\Documents\the-crew-deep-sea-sources\`:

- `deep-sea-rules-en.pdf`: the English rulebook (https://boardgame.bg/the%20crew%20mission%20deep%20sea%20rules.pdf).
  Everything we need from it is in [rules.md](rules.md); don't re-read it.
- `mission-sheet-v1d.pdf` / `.png`: the fan-made mission sheet by David Fox (v1d, July 2024). It covers
  missions 1–32, the side quests Q1/Q2 and the epilogue 33–41, all transcribed in rules.md §8.
- `task-cards-screenshots/scribd-task-cards-1..6.png`: a community list of the 96 task cards with
  difficulty triples (3/4/5 players), from https://de.scribd.com/document/594466041. Screenshots 4–6
  repeat 1–3. Transcribed in rules.md §9 (96 unique cards, the count matches the rulebook).

## 1. What Deep Sea is (vs. Planet Nine)

The same: the 40-card deck (rockets → submarines), trick rules, communication with highest/only/lowest,
the distress signal, the 2-player dummy (JARVIS → Tonoja, the same setup), the commander (→ captain:
holds trump 4, picks the first task, leads the first trick), the attempt/logbook session.

New or different (details in rules.md):

- **96 task cards with difficulty per player count** (3/4/5; 2 players use the 3 column). A mission
  sets a total difficulty; cards are drawn, skipping overshooting ones, until the total is hit exactly.
- **About 20 new objective kinds** (trick-count comparisons, trick value sums, odd/even tricks,
  predictions, "don't open with", colour comparisons, …).
- **No task tokens, no five-player handover.**
- **Selection:** captain first, clockwise; passing allowed only while there are fewer tasks than
  members; everything must be assigned after one round.
- **Re-deal rules** for some submarine tasks (alignments that make them impossible, not an attempt).
- **After a failed attempt:** same tasks again or draw new ones.
- **Communication modes:** normal, currents (= Planet Nine's dead zone), rapture of the deep (n−2
  shared sonar tokens, first come, repeat use allowed), unfamiliar terrain (random colour card picks
  the mode per attempt), no communication.
- **Mission modes:** volunteer missions (14, 15, 16, 26), "all tasks to one member" (6), captain
  takes all or passes to a volunteer (10, 13), captain takes the hardest task first (19), captain gets
  no task (25), free selection (17, 28–31, epilogue), fixed tasks (32), real-time missions (14, 15,
  16, 26, each with an untimed alternative).

## 2. Architecture decision

**One codebase, two apps.** Deep Sea is built from `the-crew/` (same Go module, same SvelteKit
project) and deployed as its own app: own URL, own container, own database, own start-page tile, own
theme. The edition is fixed per deployment by the env var `EDITION` (`planetNine` | `deepSea`), at
build time for the frontend and at runtime for the server.

| | Planet Nine | Deep Sea |
|---|---|---|
| URL | `/the-crew/` | `/deep-sea/` |
| Container / volume | `the-crew` / `the-crew_data` | `deep-sea` / `deep-sea_data` |
| Image | `board-games/the-crew` | `board-games/deep-sea` (same Dockerfile, `EDITION=deepSea`) |
| Local ports (Vite / Go) | 5177 / 8084 | 5178 / 8085 |
| DB | `data/the-crew.db` | `data/deep-sea.db` |

Why:

- **The rules engine is ~80% shared** (tricks, dummy, communication, distress, polls, attempts,
  logbook, views, sim test). A copy would mean every fix twice in ~10k lines of domain + frontend.
  The Crew already has the seams for this (`Game.Edition`, mission table per edition, open
  `ObjectiveKind` set, edition-neutral names in `trick.go`/`objective.go`/`seat.go`).
- **A separate app on the outside** gives what the user wants: its own tile, URL, theme and home
  screen, and its own lobby codes, retention and logbook. A lobby-level edition switch in one app
  (the earlier idea in [../implementation-plan.md](../implementation-plan.md) §11) can't give Deep Sea its own
  home screen and look.
- Rejected: forking into `deep-sea/` with its own code (duplication); moving the crew engine to
  `shared/` (shared is for what *every* game uses); one container serving both paths (Caddy would
  have to tag requests, and both editions would share one DB and code namespace).

Consequences:

- The repo-root rule "apps are independent" gets one documented exception: `deep-sea/` holds
  only `deploy/` (compose + Caddy snippet) and a `CLAUDE.md` pointer; its image builds
  `the-crew/Dockerfile` with `EDITION=deepSea` and `BASE_PATH=/deep-sea`.
- **CI:** `.github/scripts/plan-deploy.sh` must redeploy `deep-sea` when `the-crew/` (or `shared/`)
  changes. Generalise its rule: an app depends on every top-level directory its Dockerfile `COPY`s
  from, where the Dockerfile is the one named in its `deploy/compose.yaml` (`dockerfile:`). Check
  `deploy/deploy.sh` (fallback) and `deploy/local.sh` for the same assumption.
- A change in the-crew deploys both apps: intended (shared engine).

### 2.1 Edition plumbing

- **Server:** `cmd/server` reads `EDITION` (default `planetNine`); `DB_PATH` default by edition.
  `NewGame` takes the edition from config (today hard-coded `EditionPlanetNine` in `lifecycle.go:26`
  and `httpapi/missions.go:15`). A game of the other edition is not found (as if it didn't exist).
- **Domain:** `Edition` selects the mission table, the task source (Planet Nine: 36 card tasks with
  tokens; Deep Sea: the 96-card deck with difficulties), naming-free rule differences (selection
  passing, no tokens, no handover) and the comm modes. Keep edition switches in a few functions
  (`MissionFor`, task drawing, `nextItem`), not spread across files.
- **Frontend:** Vite `define: { __EDITION__ }` from env `EDITION` (default `planetNine`);
  `src/lib/edition.ts` holds the per-edition terms (rocket/submarine, JARVIS/Tonoja,
  commander/captain, radio/sonar, title, favicon, start-page href), and the theme. Theme: two token
  sets in `layout.css`, selected by a `data-edition` attribute on `<html>` set in `app.html` at build
  time (or a Vite alias `#theme` → `theme-planet-nine.css` / `theme-deep-sea.css`). Components keep
  using tokens. `npm run dev:deep-sea` = `EDITION=deepSea vite --port 5178` with `BACKEND_URL` :8085.
- **Docker:** `ARG EDITION=planetNine` in the Dockerfile → passed to `npm run build` and set as
  `ENV EDITION`, `DB_PATH=/data/<edition db>`.

### 2.2 Look: deep-sea theme

Distinct from the space theme. Light: pale sea-foam / sand "logbook paper", deep teal-navy ink, coral
accent. Dark: abyssal blue-black gradient fading downward, sparse bioluminescent dots (instead of the
starfield), cyan/turquoise accent. Submarine trump cards (dark, submarine glyph), sonar-ping glyph
for communication, a buoy for the distress signal. Same layout as Planet Nine (mobile-first).

## 3. Domain changes (the-crew/backend/internal/core/domain)

- **`missions_deepsea.go`:** the mission table (rules.md §8) as `MissionDef`s. Extend `MissionDef`:
  `Difficulty int`, `Comm` gains `Mode` (normal | currents | rapture | unfamiliar | none) and
  `OnlyBeforeFirstTrick` (10/13 when passed on), `Selection` (open | oneMemberChosen (6) |
  volunteers N (14–16: 1, 26: 2) | captainOrVolunteer (10, 13) | free | captainSkipped (25) |
  captainHardestFirst (19)), `FixedTasks` (32), `Objectives` (8, 12, 21, 23, 27).
- **`tasks_deepsea.go`:** the 96 task cards (rules.md §9): `TaskCard{ID, Difficulty [3]int,
  Objective, Redeal RedealRule}`. A test compares it with an independently written table (as for the
  Planet Nine missions).
- **Task drawing:** shuffle the deck, take cards in order, skip a card that overshoots, until the sum
  is exact (2 players: the 3-player column). App decision: keep a per-game "used" set so new tasks
  come first; reshuffle when exhausted (rulebook suggestion). The RNG stays injected.
- **`Task`:** `Token` zero and `Card()` only valid for Planet Nine; add `CardID` (Deep Sea card
  number) and `Prediction *int` (+ `PredictionOpen`). Views must hide a hidden prediction from others.
- **Re-deal check** right after drawing (and after the distress pass? no: the pass cannot move
  submarines): if any displayed task's `RedealRule` matches a hand, re-deal the playing cards
  automatically; the attempt count is not raised; an event says so.
- **Selection with passing:** draft turns from the captain clockwise; a member may pass only if at
  the start of the round there were fewer tasks than members, and only as long as the remaining
  members can still take the remaining tasks (all assigned after one round). Planet Nine keeps
  no-passing.
- **New objective kinds** (rules.md §9, column *Kind*): each a case in `Objective.Status`, pure
  over the history plus an evaluation context (`players`, `captain seat`, `prediction`). History
  already has leads (`Plays[0]`) and winners. Fail early, fulfil once it can no longer fail.
- **Communication:** rapture pool (n−2 shared tokens, a member may use several), unfamiliar terrain
  (draw a random non-submarine card from the shuffled deck before the deal, show it, mix it back),
  `none`, `OnlyBeforeFirstTrick`. Currents reuse dead zone.
- **Prediction step:** after the selection, owners of prediction tasks submit a number (0..hand
  size) before the first trick; communication and play wait for it.
- **After a loss:** the host chooses *same tasks* or *new tasks* for the next attempt.
- **Epilogue (33–41):** attempts are not counted, the distress signal is.

## 4. Adapters, service, frontend

- `apidto`: task cards go over the wire as `{id, difficulty, kind, params}`; the client renders the
  text (one formatter per kind, tested against all 96 cards). New intents on WS: `passTask`,
  `predict`, `chooseVolunteer…` (polls exist), `captainTakeAll|captainPass` (10/13),
  `assignTask` (free selection: captain assigns), `nextAttempt{newTasks bool}`.
- Frontend: `TaskCard.svelte` gets a Deep Sea variant (text + difficulty pip, no token), selection
  panel with *Pass*, prediction input, rapture pool badge, unfamiliar-terrain reveal, mission
  briefing per mission (rules.md §8 texts), the theme and terms.
- Start page: a new tile *The Crew: Mission Deep Sea* (`games.ts`, `devPort` 5178); privacy policy:
  one more app with the same 30-day retention.

## 5. Work packages and agents

Same working mode as Planet Nine ([../implementation-plan.md](../implementation-plan.md) §10): the
implementing agent (main session, Opus) is the **orchestrator and owns the hard parts**. It hands
clearly scoped, well-specified work to **Sonnet** subagents (`Agent` tool with `model: "sonnet"`) to
save cost. Use Haiku only for read-only chores (run all checks and summarise failures). Every subagent
brief contains:

- the work package text below and the rows of rules.md it implements;
- the files to read (rules.md, this plan, and the existing Planet Nine files it extends or mirrors);
- the exact files it may touch;
- the acceptance commands (`go vet ./... && go test ./...`, `npm run check && npm run lint && npm test`).

The orchestrator **reviews every diff** before building on it. Parallel subagents that touch
different directories run with `isolation: "worktree"`; otherwise one after another. Rule of thumb:
data transcription, mechanical plumbing, tests written from a spec, and leaf UI components go to
Sonnet; anything that changes the step machine, the objective semantics or the hidden-information
rules stays with Opus.

| WP | Work | Model | Depends on | Acceptance |
|---|---|---|---|---|
| **1** | **Edition plumbing (backend).** `EDITION` env in `cmd/server` (default `planetNine`), `DB_PATH` default per edition, `NewGame` and the missions endpoint take the edition from config, `EditionDeepSea` constant, empty Deep Sea mission table. | Sonnet | – | Planet Nine unchanged, all checks green; `EDITION=deepSea go run ./cmd/server` boots. |
| **2** | **Deployment + integration.** Dockerfile `ARG EDITION`; `deep-sea/deploy/compose.yaml` (builds `the-crew/Dockerfile`, `BASE_PATH=/deep-sea`, `EDITION=deepSea`, own volume, 720h retention) and `deep-sea/deploy/deep-sea.caddy` (with `import gate`); `plan-deploy.sh` dependency rule (§2); check `deploy.sh` / `local.sh`; start-page tile (`devPort` 5178); `startPage.ts` and the ports table; root CLAUDE.md (apps table, ports, the one-codebase exception); privacy policy (one more app, same data, 30 days); `deep-sea/CLAUDE.md` pointer. | Sonnet (worktree, parallel to 1) | – | start-page checks green; `plan-deploy.sh` with a BASE before a the-crew change lists both apps. The orchestrator reviews the CI/deploy-script change closely (security model in deploy/README.md). |
| **3** | **Frontend edition skeleton.** `__EDITION__` define, `src/lib/edition.ts` (terms, title, favicon), `npm run dev:deep-sea` (5178 → :8085), the Deep Sea theme token set (§2.2) selected per edition, submarine/sonar/buoy glyphs. Every UI string naming rocket/JARVIS/commander/radio goes through `edition.ts`. | Sonnet | 1 | Both editions render the lobby in both themes; Planet Nine looks unchanged. |
| **4** | **Task-card and mission data.** `tasks_deepsea.go` from rules.md §9 (96 cards: number, difficulty triple, objective template, re-deal rule) and `missions_deepsea.go` from §8. **Separately**, expectation-table tests written from rules.md without reading the implementation. | Sonnet (two agents: data / tests) | 5 (types) | Both green; the orchestrator compares 10 random rows by hand. |
| **5** | **Domain core (Deep Sea).** Extend `MissionDef` (difficulty, comm mode, selection mode, fixed tasks, crew objectives) and `Task` (card id, face-down, prediction); evaluation context; draw-to-difficulty with the used-card pile; re-deal check; selection with passing and the captain-comparison exclusion; replace task; captain skipped (25); hardest first (19); volunteers / captain-or-volunteer (10, 13, 14–16, 26) on the existing poll; free selection (captain assigns); same/new tasks and "don't count this attempt" after a loss; epilogue logbook; Q1 (two stages, captain = submarine 1); Q2 (face-down tasks). Fix the public signatures early so WP4 and the WP6 tests can start. | **Opus** | 1 | Rules tests green; Planet Nine tests untouched and green. |
| **6** | **New objective kinds.** All bold kinds of rules.md §9 plus the mission-level kinds (8, 12, 21, 23, 27) and the prediction step. Opus writes the evaluation; the per-card test table (`rules_deepsea_tasks_test.go`, one block per card 1–96 with fulfilled / failed / still-open cases) goes to Sonnet, written from rules.md only. | **Opus** (evaluation) + Sonnet (tests, worktree) | 5 | All 96 cards covered; failures reconciled by the orchestrator (rules.md is the authority). |
| **7** | **Communication modes.** Rapture pool (n−2 shared tokens, repeat use), unfamiliar-terrain draw per attempt (shown to all), no communication, only-before-first-trick (10/13), from trick 2 (23). | **Opus** | 5 | Rules tests per mode. |
| **8** | **Simulation test** extended to Deep Sea: random legal actions through every Deep Sea mission (incl. Q1, Q2, epilogue) × 2–5 players × ≥20 seeds; card conservation, actors never empty, hidden predictions and face-down tasks never in other players' views. | Sonnet | 5, 6, 7 | Green. |
| **9** | **Adapters, service, e2e.** apidto (task cards as `{id, difficulty, kind, params}`, comm mode, pool, prediction visibility), WS intents (`passTask`, `replaceTask`, `predict`, `assignTask`, `nextAttempt{newTasks, uncounted}`), sqlite DTO fields, e2e with leak checks. Follow the existing adapters file by file. | Sonnet | 5–7 | `go vet` / `go test` green; the orchestrator reviews view → DTO for leaks. |
| **10** | **Frontend plumbing.** `types.ts` mirror, client commands, the task-text formatter per kind with a test over all 96 cards (expected strings from rules.md §9), fixtures for every new step × 2/3/5 players, mission briefings (rules.md §8). | Sonnet | 9 | check/lint/test green. |
| **11** | **Game table UI.** Deep Sea `TaskCard` (text + difficulty, face-down variant for Q2), selection panel with Pass / Replace, prediction input, rapture pool badge, unfamiliar-terrain reveal, debrief options (same/new tasks, don't count). Opus does the decision-panel changes; leaf components go to Sonnet once the pattern stands. | **Opus** (+ Sonnet leaves) | 10 | Fixtures render at 360–430 px in both themes; a 2-player game with Tonoja plays end to end. |
| **12** | **QA and review.** Playwright session on both editions, `/code-review`, fixes, all checks in the-crew, start-page and shared (if touched). | **Opus** | 11 | Green; the user plays a session locally. |

**Critical path:** 1 → 5 → 6/7 → 9 → 10 → 11 → 12. **In parallel:** 2 and 3 alongside 1/5; 4 once the
types of 5 stand; 8 alongside 9.

**Review gates:** after 2 (deploy/CI security), after 4 (data vs rules.md), after 6 (objective
semantics), after 9 (no leaks: hidden predictions, face-down tasks, hands), after 11 (a real
playthrough).

**Commit cadence:** one commit per work package on a feature branch; push to `main` only after WP12,
with the user's go-ahead.

## 6. Decisions (user, 2026-10-11)

All questions are answered; the rules-level ones are in rules.md as **App decision**.

1. Path `/deep-sea/`, title "The Crew: Mission Deep Sea".
2. **No timers:** real-time missions use their untimed variant.
3. **Impossible task combinations:** no automatic check (a hand-aware check is infeasible and would
   leak information); the captain may **replace** displayed tasks before the first pick.
4. **"Decide together" and free selection:** the captain decides / assigns in the app, as in Planet
   Nine; the crew coordinates outside the game. Volunteer polls work like Planet Nine's polls.
5. **Side quests Q1/Q2:** first included, later removed (they did not play well; rules.md §8).
6. **Unavoidable failures:** after a loss the host may choose "don't count this attempt".
7. All card interpretations confirmed (rules.md §9).
8. Tonoja works exactly like JARVIS (3-player difficulty column).

Local ports changed since the plan: the gate took 5178, so Deep Sea uses
Vite 5179 (backend 8085).
