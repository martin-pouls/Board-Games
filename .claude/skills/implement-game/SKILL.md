---
name: implement-game
description: Implement a new game (or an edition or variant of an existing one) end to end from reference material: rules, plan, mockup, scaffold, domain, adapters, UI, theme, audits. Orchestrates subagents; the user decides at four checkpoints (rules, plan + mockup, real UI, play + ship).
argument-hint: <game name or IDEAS.md entry> [rulebook / image paths or links]
disable-model-invocation: true
---

# Implement $ARGUMENTS

You are the **orchestrator**. Agents do the reading-heavy and writing-heavy work; you decide, brief, verify, commit and talk to the user. Your context should hold the rules summary, the plan, agent reports and check results, not code. Between checkpoints, work on without asking.

**Run this on Opus.** You make the calls agents can't: review gates, code vs test reconciliation, whether a question is answerable from rules.md. If the session isn't on Opus, ask the user to switch (`/model`) before step 1.

## Agents

| Agent | Model | Does |
|---|---|---|
| `rules-writer` | opus | reference material → `docs/rules.md` + open questions |
| `game-planner` | opus | rules.md → implementation plan with work packages (WPs) |
| `game-implementer` | sonnet; pass `model: "opus"` where the plan says opus | one WP |
| `rules-test-writer` | sonnet | black-box rule tests and data expectation tables, without seeing the code |
| `rules-auditor` | opus | domain vs rules.md, view leaks |
| `visual-reviewer` | sonnet | fixture screenshots vs design-guidelines |

**Models:** an agent file's `model` is its default; the Agent tool's `model` parameter overrides it per call. The plan's WP table sets the model per package (the planner decides at plan time; domain design, step machine, hidden information and the table layout are opus). Follow it. Raise a WP to opus before spawning when it turns out to touch the step machine, the views or objective/rule semantics; never lower a WP the plan put on opus.

**Briefs** (at most ~20 lines): the WP id and plan path (don't paste the plan), the rules.md §, the files to read and mirror, the files it may touch, the acceptance command, "report as your agent file says". Follow-ups to the same agent (fix this failure, fold in these answers) go through `SendMessage`, which keeps its context; spawn fresh only for new work, or when the agent's context is heavy with images (mockup and screenshot agents). Run independent WPs in parallel; give parallel writers `isolation: "worktree"`.

**Worktrees** hold only committed files, so:
- Commit everything the WP needs before spawning (rules.md, plan, mockup, the WPs it depends on).
- No `node_modules`: a WP that touches a frontend says in its brief "run `npm ci` in `<app>/frontend` (and `shared/frontend` if you touch it) first". Without it the format hook silently skips Prettier and the checks fail.
- No gitignored files: pass `<game-docs>/reference/` and screenshot directories as absolute paths into the main checkout.
- The agent doesn't commit. After review, commit inside the worktree (`git -C <worktree> commit`, the WP's message), merge its branch into `game/<dir>`, then `git worktree remove <worktree>` and delete the branch.

**Token discipline:** don't read code the agents wrote except at review gates (then the specific diff hunks); use `git diff --stat`, targeted `grep` and the check summary. Don't relay long reports to the user.

## Handover documents and resuming

The run must survive `/clear`, compaction and a new session. Everything it needs lives in committed files, never only in your context:

- **`<game-docs>/rules.md`**: the rules and every App decision, incl. those answered mid-build. Permanent: the source of truth for the game.
- **`<game-docs>/implementation-plan.md`**: design and WPs. Its §12 holds the run state: the status per WP (with commit), and **Open items**: unanswered questions, deferred findings, decisions taken during the build that belong nowhere else yet. After the run it stays as the design record, headed "Implemented; current state: `<dir>/CLAUDE.md`".
- **`<dir>/CLAUDE.md`**: written in step 4; the lasting summary for future work on the game.

Update §12 and commit after every WP, and before every checkpoint, so the user can `/clear` at any checkpoint (suggest it there when the context is large). **Resuming:** `/implement-game <game>` finds the plan (or only rules.md), reads it, §12 and `git log main..game/<dir>`, and continues at the first open step. Agents get these files by path, so a brief never depends on what you remember.

- Work on branch `game/<dir>` (create it from `main`). One commit per WP: `<Game>: WP<n> <title>`, with the attribution trailer. Never push without the user's yes at checkpoint D.
- After each WP, one line to the user: what's done, test count if it changed, what's next. On reaching a checkpoint, also send a `PushNotification` if that tool is available: the user is likely away.

## 0. Setup

1. Find the game: its `IDEAS.md` entry if any; the material from the arguments. Ask **once, now**, for anything missing that the run will need: rulebook (PDF or link), photos of the box and a few card faces (for the theme), the card data the rulebook lacks. Material goes to `<game-docs>/reference/`, gitignored: the repo is public, never commit publisher material. `<game-docs>` is `<dir>/docs/` for a new app, `<base>/docs/<edition>/` for an edition, `<base>/docs/` for a toggle (its rules become sections of the base rules.md, its plan `<variant>-plan.md`).
2. **Placement**, recommended now, decided at checkpoint A:
   - **Ruleset toggle in an existing app** (The Game's Extreme): an expansion or variant of the same game that players switch on per session; mostly the same components. Lobby option, ruleset config, rules follow the components.
   - **Edition** (Deep Sea): a standalone product with its own name, box and look whose rules engine is mostly (≳ ⅔) an existing app's. Same codebase, `EDITION` build, own URL, tile, theme and DB; `<dir>/` holds only `deploy/` and `CLAUDE.md`; rules and plan in `<base>/docs/<edition>/`.
   - **New app** (default): everything else. Plumbing from Fox Duet via `scaffold.sh`, the domain modelled after the closest game.
   - Between edition and toggle: edition if it deserves its own tile and look, toggle if players would switch between both in one sitting.
3. Pick the closest existing game (domain patterns) and, for a new app, the next free ports (root `CLAUDE.md` table).

## 1. Intake

Spawn `rules-writer` with the material, target path, closest game and tentative placement.

**Checkpoint A (rules)** (AskUserQuestion, several rounds if needed, recommendation first):
- placement, with the one-paragraph reason;
- v1 scope (player counts, variants, solo; what waits);
- every open rule question;
- title, URL path, licensing note.

Send the answers to the rules-writer (`SendMessage`) to turn into App decisions and drop the open-questions section. Commit rules.md.

## 2. Plan and mockup

In parallel:
- `game-planner` with rules.md, placement, decisions, closest game, plan path (`<game-docs>/implementation-plan.md`).
- `game-implementer` (`model: "opus"`) for the **look and the mockup**, no app code: `game-theme` steps 1–2 from the reference photos, where the preview of step 2 is a static `<game-docs>/mockup.html` (the agent doesn't publish it; you do at checkpoint B) (one self-contained file, sample data, design-guidelines applied). It shows the lobby, the game screen in its key states (your turn with a decision, waiting for someone else, the most crowded state at the largest player count), game over and the start-page tile, at 375×812 plus the game screen at 1280×800, in light and dark. Colours as CSS custom properties on `:root`, named like the future `layout.css` tokens, redefined for dark under `@media (prefers-color-scheme: dark)` (guarded by `:root:not([data-theme="light"])`) and under `:root[data-theme="dark"]`; an explicit `body` background; a 2–4-word `<title>`; no scripts from outside the file; fonts only from Google Fonts (mockup only, the app uses Fontsource); works at phone width without horizontal scroll. That keeps it publishable as an artifact. The agent serves it with `node .claude/skills/implement-game/serve-mockup.mjs <game-docs>` (background; Playwright blocks `file:` URLs), screenshots every screen × mode into `.playwright-mcp/implement-<dir>/mockup/`, and reports the theme plan with the file list. A toggle mocks only the screens it changes, in the base game's theme.

Read the plan yourself. Check: every rules.md § lands in a WP; the hidden information is listed per field; WPs are small, have models, files and acceptance; gates are marked; the UI track depends on the plan's wire contract, not on the adapters. Fix it via `SendMessage` to the planner.

**Checkpoint B (plan and mockup):** the plan in ≤ 15 lines (placement, scope, domain shape, WPs with models, gates, risks), the theme plan, and the mockup screenshots (Read them). Also publish `mockup.html` as a private artifact (Artifact tool, `icon: "cards"`) and give the link, so the user can open it on a real phone; locally it's at `http://localhost:5190/mockup.html` while the server runs. Commit the mockup as a baseline, then send each round of feedback (chat or artifact comments) to a **fresh** `game-implementer` (`model: "opus"`), not the mockup agent via `SendMessage`. That agent's transcript carries the reference photos and dozens of screenshots, so resuming it costs hundreds of thousands of tokens per step. The brief carries the feedback, the mockup path (its theme-plan section explains the earlier choices), the photos' path, and "shoot only the changed screens into `mockup<n>/`, view only what you need to check". Republish from the same path after each revision (same URL) until the user approves; then stop the server. Commit the plan and the mockup: they brief the theme and table UI WPs.

## 3. Build

Per WP, in the plan's order, parallel where it allows:
1. Brief and spawn. The scaffold WP runs `bash .claude/skills/implement-game/scaffold.sh <dir> "<Title>" <vite-port> <backend-port>` first; the agent then strips the copied game to a lobby-only one. The integration WP (in parallel) works through the `add-app` checklist and the plan's §8, with a plain placeholder tile; the theme WP styles the tile (`game-theme` step 4). The theme WP skips `game-theme`'s screenshot loop (step 5): the look was approved on the mockup and is reviewed on the real UI in step 4; its acceptance is the checks plus one lobby fixture shot at 375×812 in light and dark.
2. On the report: run `bash .claude/skills/check/check.sh <app>` (10-minute timeout; for an edition or toggle also the base app). Read only the summary, and the failure tail if any.
3. Failure: `SendMessage` the output to the same agent once. Still failing: rerun the WP with `model: "opus"`, or fix a small thing yourself. Then ask the user.
4. Rule questions from a report: answer from rules.md and send the answer back. Not answerable: collect them and keep building the WPs that don't depend on them; ask (AskUserQuestion) only when the critical path is blocked. Every answer becomes an App decision in rules.md.
5. Black-box test failures: decide per case with rules.md as the authority whether code or test is wrong, and send the fix to the agent that owns it.
6. **Review gates** (as marked in the plan): read the relevant diff hunks yourself, or run `rules-auditor` (domain) early. Always check by hand: the view → DTO mapping for leaks, any change to `deploy/`, `.github/` or `shared/`.
7. Commit and update §12.

Edition or toggle: the base game's existing tests stay unchanged and green (`git diff main --stat -- '<base>/**/*_test.go'` shows only additions), and its existing fixtures look the same.

## 4. UI checkpoint

As soon as the table UI WP renders every fixture (the backend track keeps going in parallel):
1. `visual-reviewer` (sonnet by default; start the app with the `run` skill; screenshots to `.playwright-mcp/implement-<dir>/ui/`). This is the one full visual review of the run. Fix blocking and gameplay-clarity findings via `game-implementer`, which reshoots the affected screens itself and reports the file paths.
2. **Checkpoint C (real UI):** the key fixtures at 375×812 in light and dark plus the game screen at 1280×800 in light, next to the approved mockup screens, the remaining consistency and theme findings with a recommendation each, and the fixtures URL (`/dev/fixtures`) for the user to click through. Implement the user's picks, reshoot, show again until approved. Record UI decisions that deviate from the mockup in the plan's §7.

The UI is now real but nothing is final-reviewed yet, so changes here are still cheap.

## 5. Verify

1. `rules-auditor` on the game. Wrong / Leak → fix WP. Untested → `rules-test-writer`. Undocumented or Ambiguous → an App decision you can justify from rules.md, else a question for checkpoint D.
2. `/code-review` on the branch (`code-review` skill, target the branch against `main`). Fix the confirmed findings.
3. No second visual review. If fixes since checkpoint C changed UI files, the agent that made them reshoots only the affected fixtures (375×812, light and dark); show those at checkpoint D.
4. `bash .claude/skills/check/check.sh --base main`.
5. Integration complete? Start-page tile and dev link, `run.sh` table, root `CLAUDE.md` tables, the game's `CLAUDE.md` (modelled on `fox-duet/CLAUDE.md`: intents and per-player view, domain modelling, look and feel), `IDEAS.md` (mark implemented), a `TODO.md` section, the gate's port comment.

## 6. Checkpoint D (play and ship)

Show: what was built, test counts, audit and review outcomes, the open items and deferred findings, and how to play locally (`bash .claude/skills/run/run.sh <app>`, the URL; two browser profiles for two players). Ask the open questions. The user plays a round.

Then, on an explicit yes only: delete `<game-docs>/mockup.html` (the fixtures replace it), merge into `main` and push (that deploys). Stop the apps you started and delete `.playwright-mcp/implement-<dir>/`.

**Retro:** if the run kept tripping over the same thing (a brief that always lacked something, an agent that overstepped, a missing script), fix this skill or the agent file and tell the user what you changed.
