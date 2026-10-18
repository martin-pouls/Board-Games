# <Game>: implementation plan

<!-- Template for the game-planner agent. Keep the numbered sections; drop a subsection that has
nothing to say. Aim for 200–350 lines: decisions and seams, not code. Worked examples:
the-crew/docs/implementation-plan.md (new app), the-crew/docs/deep-sea/implementation-plan.md (edition). -->

Read before working on a package: root `CLAUDE.md`, `<rules.md path>` (§ numbers below refer to it),
`shared/README.md`, and the files each package names. Don't re-read the rulebook. A rule question
rules.md doesn't answer goes back to the orchestrator, not into the code.

## 1. Scope

- **Version 1:** player counts, variants, modes, solo.
- **Later, not in v1:** with the seam that keeps it possible (§5.8).

## 2. Placement and identity

- **Placement:** new app | edition of `<base>` | ruleset toggle in `<base>`, and why (one paragraph).
- Table: directory, URL path, Go module, DB file, Vite / backend ports, container / volume, start-page
  tile (yes/no), game retention, UI language (English).
- **Template sources:** plumbing from Fox Duet; the domain modelled after `<closest game>` (which parts).
  For an edition or toggle: the existing files that change and how existing games and snapshots stay
  unaffected.

## 3. Decisions

User decisions with date, then the App decisions taken in rules.md (§ refs), one line each.

## 4. Shared code

What moves to `shared/` (usually nothing) and why. Plumbing that the scaffold copied unchanged a
third time is a candidate: name it, don't move it in v1.

## 5. Domain design (`backend/internal/core/domain`)

### 5.1 Files
One line per file: what it holds.
### 5.2 Aggregate and phases
The Game struct (fields, what is derived instead of stored), phases and in-between phases.
### 5.3 Steps and actors
The step machine inside `playing`: each step, who acts, what ends it. Automatic resolutions.
### 5.4 Data tables
Cards, enemies, missions, boards as Go tables transcribed from rules.md, with stable IDs.
### 5.5 Views and hidden information
Per viewer: what is visible, what only as a count, what never. Private events. The game-over view.
### 5.6 Intents
Domain method → service method → WS command, one row each, with its sentinel errors.
### 5.7 Randomness and persistence
Injected RNG; snapshot DTO notes; what must keep loading.
### 5.8 Seams for later
What the design leaves open (§1 "later") without building it.

## 6. Service, ports, adapters

Deviations from Fox Duet only. New WS commands, REST additions (static data only).

**Wire contract:** the view and event DTOs and the WS commands, field by field (camelCase). The
frontend track writes `types.ts` and the fixtures from this section, in parallel with the adapters,
so the UI can reach the UI checkpoint early. A later change here updates both sides.

## 7. Frontend

Screens and components (Lobby options, Table regions top to bottom, decision panel per step,
GameOver, RulesHelp) as approved in `mockup.html`, the `types.ts` mirror, format helpers, fixtures
(one per phase, step, pending decision and game-over variant, at the relevant player counts).
UI decisions taken at the UI checkpoint that deviate from the mockup are recorded here.

## 8. Integration checklist

Every place outside the game's directory: start page (`games.ts`, tile in `+page.svelte`, tile tokens
in `layout.css`), `shared/frontend/src/startPage.ts`, privacy policy, `.claude/skills/run/run.sh`
ports table, root `CLAUDE.md` (apps table, ports table), gate port comment, `IDEAS.md`, `TODO.md`
section, the game's own `CLAUDE.md`.

## 9. Testing

Domain table tests per rules.md §; validate-before-mutate assertions; black-box rule tests
(rules-test-writer); data tables against an independent expectation table; simulation test
(random legal actions × player counts × seeds: no panic, conservation of cards, every game ends,
actors never empty, views hide what §5.5 says); adapters; e2e incl. a leak check; frontend helpers.

## 10. Work packages

| WP | Work | Agent (model) | Depends on | May touch | Acceptance |
|---|---|---|---|---|---|

Rules for the table:
- One package = one agent run: ≲ 15 files or ≲ 1500 changed lines. Split bigger ones.
- `game-implementer` (sonnet): scaffold stripping, integration, data transcription, adapters,
  service, e2e, frontend plumbing, fixtures, leaf UI. `game-implementer` (**opus**): domain
  foundation and flows, views and hidden information, the table layout and decision panel, the theme.
  `rules-test-writer` (sonnet): black-box tests (package `domain_test`, helpers via `export_test.go`), and expectation tables for transcribed data.
- Fix the public domain signatures and `export_test.go` early (stubs returning `ErrNotImplemented` are fine), so tests,
  adapters and frontend plumbing can start in parallel.
- Parallel packages touch disjoint directories and run with `isolation: "worktree"`.
- Acceptance is a command (`bash .claude/skills/check/check.sh <app>`), plus what to verify by hand.

**Critical path** and **parallel tracks**, one line each.

**Review gates** (the orchestrator stops and checks): domain semantics after the foundation, data
tables against rules.md, hidden information after the views, no leaks on the wire after the
adapters, deploy/CI changes, the table UI on fixtures.

## 11. Risks

| Risk | Handling |
|---|---|

## 12. Status

The orchestrator keeps this section current and commits it after every WP; it is the state a resumed
run (after `/clear` or in a new session) starts from.

| WP | Status | Commit | Notes |
|---|---|---|---|

**Open items:** unanswered rule questions, deferred review findings, build-time decisions not yet in
rules.md or the game's CLAUDE.md. Each moves to its permanent home (rules.md App decision,
CLAUDE.md, TODO.md) before the run ends.
