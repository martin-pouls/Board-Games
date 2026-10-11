# The Crew: implementation plan

Starting point for the agent that builds **The Crew: The Quest for Planet Nine** as the fourth game of this monorepo. Read in this order before writing code:

1. [CLAUDE.md](../../CLAUDE.md) at the repo root: stack, hexagonal architecture, conventions, deployment.
2. [rules.md](rules.md): the source of truth for every rule and app decision. Section numbers (§) below refer to it.
3. [fox-duet/CLAUDE.md](../../fox-duet/CLAUDE.md) and the Fox Duet code. It is the closest existing game (trick-taking, newest code), and **the template for everything that is not game logic**. Where this plan is silent on a convention, do what Fox Duet does.
4. [shared/README.md](../../shared/README.md).

Do **not** re-read the rulebook PDFs. If a rule question is not answered in rules.md, ask the user, then record the answer there.

---

## 1. Scope

**Version 1 (this plan):**
- Planet Nine, all **50 missions** from rules.md §14.
- **2–5 players.** 2 players with JARVIS is the most important case: the user mainly plays as a pair.
- Tasks, all task tokens, the special objectives, open distribution, the commander's decision and distribution, polls, communication (incl. dead zone, disruption, silenced member), distress signal, five-player handover, mission 12's card draw.
- Session model with a mission picker and a logbook (§16).
- Light/dark theme, mobile-first, deployed at `/the-crew/` behind the gate, with a start page tile.

**Later, not in v1:**
- *Mission Deep Sea* (§11 of this plan). The design must leave room for it.
- The *Challenge for three* variant.

## 2. Decisions

Settled with the user (2026-10-11):

| Topic | Decision |
|---|---|
| Campaign | One game = one sitting. Mission picker in the lobby, logbook kept in the game, retention stays at 7 days. |
| Player counts | 2–5 in v1, including JARVIS. 2 players is the priority. |
| "Decide together" calls | The **commander** decides in the app (distress signal, handover, token swap/move, mission 50 roles). The players talk outside the app. |
| Deep Sea | Wanted later. Same app or not is open, but the overlap is large, so plan for the same app (§11). |

App decisions taken while writing rules.md. They are recorded there; **show this list to the user once and adjust rules.md if they object**:

1. Same-trick rule for tokens, generalised (§7).
2. A singleton must be communicated as "only" (§8).
3. A mission with a "can still fail until the end" objective is played to the last trick, e.g. mission 17 (§13).
4. Mission 34: the commander owns "first & last trick" (§14).
5. JARVIS takes no part in the distress pass or mission 12's draw. A communicated card cannot be drawn (§9, §16).
6. JARVIS must follow suit among its face-up cards only (§12).
7. Mission 11: the commander may silence themselves. Mission 46: the owner is fixed at the deal (§13, §14).
8. Polls are answered by every human except the commander, in any order, and are visible live (§10).
9. All hands are revealed after an attempt (§16).
10. The 3-player extra card goes to a random seat. JARVIS sits in seat 3 (§3, §4).

## 3. App identity

| Item | Value |
|---|---|
| Directory | `the-crew/` |
| URL path | `/the-crew/` |
| Go module | `github.com/martin-pouls/the-crew/backend` |
| Default DB | `data/the-crew.db` |
| Local ports | Vite **`:5177`**, backend **`:8084`** (`:8083` is the gate) |
| Container / volume | `the-crew` / `the-crew_data` |
| Image | `ghcr.io/martin-pouls/board-games/the-crew:<sha>` |
| Caddy snippet | `the-crew/deploy/the-crew.caddy`, with `import gate` |
| UI language | English, like the other games |

**Look:** a space theme, distinct from the others.
- **Light:** pale "mission paper" with deep indigo ink and a coral/orange accent.
- **Dark:** midnight navy with a subtle starfield and a cyan accent.

Cards are drawn by us. The four colours pink, blue, green and yellow each get a distinct glyph (colour-blind support). Rockets are dark cards with a rocket glyph. The theme tokens live in `src/routes/layout.css` (`:root` light, `.dark` dark), and components use the tokens, not raw colours.

## 4. Shared code: what moves, what stays

The user asked to consider `shared/` without forcing it. Verdict: **v1 moves nothing new into `shared/`.** Reasons:

- **Trick engine (follow suit, winner with trump):** this is the obvious candidate, since Fox Duet, The Crew, Deep Sea, Sail and LOTR all have one. But the root rule says `core/domain` stays stdlib-only and never imports `shared/`. The engines also differ in substance:
  - Fox Duet: 2 seats, decree trump, the Royal Heir exception.
  - The Crew: N seats, fixed rocket trump, and a dummy that plays only from its face-up cards.
  - LOTR: cards that cannot win or cannot be led.

  The common core is about 40 lines. Sharing it would need an architecture exception for little gain.
  **Instead:** write `domain/trick.go` so that it depends on nothing Crew-specific: a pure function over `[]Card`, the lead and a trump predicate. Then a later extraction (when a third trick-taker arrives, and only if the rule is relaxed then) is a file copy. Deep Sea reuses it directly because it lives in the same app.
- **Objective engine** (tasks, tokens, special objectives): Crew-specific. Deep Sea reuses it inside the same app. Nothing else needs it.
- **Lobby lifecycle in the domain** (NewGame, AddPlayer, Leave, names, host): already copied in every game. It is domain code, so it stays per game. Copy Fox Duet's.
- **Frontend `recent.ts`:** nearly identical in all games, but the phase names and labels differ per game. The Crew's phases (`lobby`, `playing`, `debrief`, `ended`) differ again. Copy and adapt. Sharing it would need two callbacks and saves about 60 lines.
- **Frontend `Hand.svelte`** (overlapping fan, dimmed illegal cards, flip and card motion): The Crew's hand is nearly Fox Duet's (single select, illegal cards dimmed). Regicide's differs (multi-select). **Copy it in v1.** Once The Crew's hand has settled, an optional follow-up (WP12) can extract a generic `CardHand.svelte` (generic over the card type, with a card snippet and a key function) into `shared/frontend`, used by Fox Duet and The Crew. Two concrete users then define the API, and future trick-takers inherit it.

Everything that *is* already shared (wire, httpkit, wskit, sqlitedb, idgen, lockmap, serverkit, AppFrame, GameSocket, session, ShareInvite, ConnectionBanner, cardMotion, login) is used as Fox Duet uses it.

## 5. Domain design (`backend/internal/core/domain`)

Rules for the domain from the root CLAUDE.md still apply:
- stdlib only;
- validate fully before mutating;
- typed sentinel errors;
- an injected RNG;
- derived values are computed, not stored;
- the snapshot is minimal.

The sketches below fix structure and names, not every field. Adjust details while implementing, but keep the shape.

### 5.1 Files

| File | Content |
|---|---|
| `card.go` | `Suit` (`pink`, `blue`, `green`, `yellow`, `trump`), `Card{Suit, Rank}`, `NewDeck()` (40), `SortCards`. The trump suit is called `trump` in the domain; the UI names it "rocket" (Deep Sea: "submarine"). |
| `trick.go` | Edition-agnostic trick rules: `legalCards(hand, lead *Card)`, `trickWinner(cards []Card) int` (index into the played cards). |
| `seat.go` | `Member` (a seat in an attempt: a human or the dummy), the dummy's 7 stacks, `DummyPlayable`. |
| `mission.go` | `MissionDef` and its enums (distribution, communication, setup actions, objective specs), the `missions` table lookup. |
| `missions_planet9.go` | The 50 `MissionDef` entries, transcribed from rules.md §14. **Data only.** |
| `objective.go` | `Objective`, `ObjectiveKind`, `Status`, and evaluation over the trick history (§5.4). |
| `tokens.go` | Task tokens and the order check (rules.md §7). |
| `attempt.go` | `Attempt`: dealing, commander, the step machine, trick play, end detection. |
| `setup.go` | Token swap/move, polls, commander picks, open draft, commander's decision and distribution. |
| `comm.go` | Communication, distress signal and pass, five-player handover, mission 12 draw. |
| `game.go` | The aggregate `Game`: phase, players, host, selected mission, logbook, the current `Attempt`, `Actors()`. |
| `lifecycle.go` | Lobby: NewGame, AddPlayer, Leave, SelectMission, StartMission, NextAttempt, NextMission, BackToLobby, EndGame, GiveUp. Copy Fox Duet's name and lobby code. |
| `view.go` | `PlayerView` projection (§5.7). |
| `events.go`, `errors.go`, `ids.go`, `rng.go` | As in Fox Duet. |

### 5.2 Aggregate and phases

```
Phase:  lobby ──start──▶ playing ──attempt ends──▶ debrief ──nextAttempt──▶ playing
          ▲                                          │ nextMission ─────────▶ playing
          └──────────────── backToLobby ─────────────┘
        lobby/debrief ──endGame──▶ ended
```

```go
type Game struct {
    ID         GameID
    Phase      Phase            // lobby | playing | debrief | ended
    HostID     PlayerID
    Players    []Player         // humans, join order = seat order
    Edition    Edition          // "planetNine" (v1). Deep Sea later adds "deepSea".
    Mission    int              // selected mission number (lobby) / the one being played
    Logbook    map[int]LogEntry // per mission: Attempts, Distress, Completed
    Attempt    *Attempt         // nil in the lobby; kept in debrief for the reveal
    Revision   int64
    CreatedAt  time.Time
    events     []Event
}
```

`Attempt` holds everything that one deal needs, so starting a new attempt means replacing it:

```go
type Attempt struct {
    Number      int        // attempt count for this mission, 1-based
    Members     []Member   // seat order; 2 players: [p0, p1, dummy]
    Commander   int        // seat holding the trump 4
    Step        Step       // see 5.3
    Tasks       []Task     // in layout order (left to right)
    Objectives  []Objective // special objectives from the mission (owner may be set during setup)
    Poll        *Poll      // the open poll, if any
    Draft       *Draft     // open distribution: whose turn
    Reveal      int        // commander's distribution: index of the task being decided
    Comms       []*Communication // per seat
    Silenced    int        // mission 11: seat, or -1
    Pass        *Pass      // distress pass in progress: direction, chosen cards per seat
    Exchange    *Exchange  // last pass/draw, for each player's own view
    HandoverUsed, TokensAdjusted, DistressPassedThisAttempt bool
    Trick       Trick      // in progress: Leader, Plays []Play{Seat, Card}
    Tricks      []TrickRecord // completed: Leader, Plays, Winner (seat), WinningCard
    Leftover    []Card     // set at the end: unplayed cards
    Result      *AttemptResult // Success, Reason, the failed task/objective ID, trick number
}
```

**Task = objective + token + owner.** A Planet Nine task is an objective of kind `winCards` with one card. Deep Sea tasks are other kinds. So:

```go
type Task struct { ID int; Objective Objective; Token Token; Owner int /* seat, -1 = undistributed */; FaceDown bool }
```

### 5.3 Steps (inside `playing`)

Each step has a clear actor set, derived by `Actors()`; it is never stored. The commander acts **for the dummy**.

| Step | Who acts | Leaves to | Used by |
|---|---|---|---|
| `adjustTokens` | commander (swap/move, or skip) | next setup step | 23, 40 |
| `poll` | every human except the commander, any order | `pick` | 5, 20, 27, 33, 37, 41, and each task in 24, 32, 36, 43 |
| `pick` | commander | next setup step or the next revealed task | after each poll; also 11 (silenced) and 50 (roles), which need no poll |
| `draft` | the seat whose turn it is (dummy: the commander) | `play` when all tasks are taken | open distribution |
| `pass` | every human who has not chosen a card yet | `play` | distress signal |
| `play` | the seat whose turn it is (dummy: the commander) | `play`, or `debrief` (phase) when the attempt ends | tricks |

The **setup sequence** after the deal is derived from the `MissionDef`, in this order:
1. Fix the mission 46 owner (no step).
2. `adjustTokens` (23, 40).
3. Mission-specific picks: `pick` for mission 11's silenced member and mission 50's roles; `poll` → `pick` for 5, 33 and 41.
4. Task distribution:
   - open: `draft`;
   - commander's decision: `poll` → `pick` (all tasks to one member), then reveal;
   - commander's distribution: per task, reveal → `poll` → `pick`.
5. `play`.

**Actions in the pre-play window** (while `Step == play`, trick 1 and no card played) are not steps. They are allowed actions:
- `sendDistress(direction)` / `skipDistress`. Commander only, before any communication, once per attempt. Moves to `pass`.
- `handover(taskID, toSeat)`. Commander only, 5 players, golden missions, once.
- `communicate(card, position)`. Any eligible human. Also allowed before later tricks (rules.md §8).

**Mission 12's draw** runs automatically inside the mutator that completes trick 1, with the RNG.

### 5.4 Objective evaluation

Evaluation is a pure function of the completed tricks:

```go
type Status int // open | fulfilled | failed
func (o Objective) Status(h History) Status
type History struct {
    Tricks    []TrickRecord // completed so far
    Total     int           // tricks in this attempt (13/13/10/8)
    Seats     int
    Leftover  []Card        // known only when the attempt is over; nil before
}
```

- Kinds needed for Planet Nine (rules.md §13):
  - `winCards` (a task),
  - `winNoTricks`,
  - `winExactlyTricks{N, NotWithTrump}`,
  - `winOnlyTricks{Set, NotWithTrump}` (41, 50),
  - `winTricks{Set}` (34),
  - `winAllOfSuit{Suit}` (46),
  - `crewWinWithAny{Cards, Count}` (9, 26),
  - `crewWinWithEach{Cards, Ordered}` (13, 44),
  - `crewNeverWinWith{Rank, ColourOnly}` (16, 17),
  - `crewBalance{MaxLead: 1}` (29, 34),
  - the task modifier `InLastTrick` (48).
- Trick sets reference "last" symbolically. It resolves against `Total`.
- **Fail early, fulfil only when it can no longer fail** (rules.md §13). Each kind documents both conditions in a doc comment that points to §13. At the end of the attempt, every open objective counts as failed.
- `tokens.go` checks the order conditions over the trick index at which each task was fulfilled (rules.md §7, the *b*/*s* formula). It runs after each trick, together with the objective statuses.
- After each trick:
  - any task or objective failed, or an order condition broken → the attempt is lost, with the reason;
  - all fulfilled → the mission is completed;
  - otherwise the next trick is led by the winner.
- Keep the evaluation stateless: statuses are derived from `Tricks`, never stored. The view and the end check both call it.

### 5.5 Mission data

```go
type MissionDef struct {
    Number       int
    Tasks        int          // task cards to draw (0 = none)
    Tokens       []Token      // placement order
    Distribution Distribution // open | commanderDecision | commanderDistribution
    Comm         CommRule     // {DeadZone bool; FromTrick int}
    Golden       bool         // five-player handover allowed
    Setup        []SetupAction// swapTokens | moveToken | silenceMember | assignRoles50 | drawAfterFirstTrick
    Poll         PollKind     // none | wellIll | yesNo (for chosen-member missions)
    Objectives   []ObjectiveSpec // kind + owner rule: crew | commander | chosenByCommander | roleA | roleB | leftOfHolder(pink 9)
}
```

The table in `missions_planet9.go` mirrors rules.md §14 row by row. A test asserts all 50 rows (count, tokens, flags) against an independently written expectation table (WP4).

### 5.6 JARVIS (the dummy)

- `Member{PlayerID: "", Dummy: true, Stacks [7]DummyStack{Down, Up *Card}}`. Humans have `Hand []Card`.
- **Playable for the dummy:** the face-up cards, filtered by follow suit among them (rules.md §12).
- **Flip:** at the end of a trick, the stack whose face-up card was played turns its hidden card up.
- **Acting:** `PlayCard(by, card)` acts for the seat whose turn it is. If that seat is the dummy, `by` must be the commander. The same rule applies to `PickTask`.
- `Dummy` must not be special-cased in the trick and objective code. Only `legal cards` and `who acts` know about it.
- **Naming:** the domain says `dummy`. The UI shows "JARVIS" (Deep Sea: "Tonoja").

### 5.7 Views and hidden information

`PlayerView` for a viewer (seat *v*) contains:
- the phase, step, mission number and `MissionDef` summary, the attempt number, the commander's seat, and the actors (`mustAct` per seat);
- the members: name, seat, dummy flag, hand count, tricks won, the communication (card + position, or the card only in a dead zone, or nothing), and whether they have communicated;
- the dummy's face-up cards per stack, and whether a hidden card lies under each (never the hidden card itself);
- the viewer's own hand, and the viewer's **playable** cards. For the commander on the dummy's turn, also the dummy's playable cards;
- the tasks: card, token, owner, status. Face-down tasks (commander's decision) show only a count until they are revealed;
- the objectives with their owner and status;
- the poll question and the answers so far;
- the pick options for the commander, the draft turn, and the remaining pass choosers (not their cards);
- the current trick, the last completed trick (only that one), and per-seat trick counts;
- the viewer's own side of the last pass or draw (`Exchange`: gave X, got Y);
- the distress state and the actions allowed now (`canCommunicate` with the eligible cards and positions, `canSendDistress`, `canHandover`);
- the logbook;
- in debrief: the result and **all hands revealed**, including the dummy's hidden cards and leftovers.

**Never** sent before debrief:
- other humans' hands;
- the dummy's hidden cards;
- the deck or task deck order;
- other players' pass choices;
- the undrawn tasks;
- any other player's side of an exchange.

Events are public and must not carry those cards. Fox Duet's events are not filtered per viewer, and the shared wire has no per-viewer events, so **private facts go into the per-player view, not into events**: e.g. `CardsPassed{Direction}` without cards. Test this in `view_test.go` and the e2e test (a second player's view and events never contain the first player's hand).

### 5.8 Intents (domain methods → service → WS commands)

| Domain method | WS command | Notes |
|---|---|---|
| `SelectMission(by, n)` | `selectMission` | host, lobby |
| `StartMission(by, rng)` | `startMission` | host, lobby, 2–5 players |
| `AdjustTokens(by, swap/move or skip)` | `adjustTokens` | commander |
| `AnswerPoll(by, yes bool)` | `answerPoll` | |
| `Pick(by, seats []int)` | `pick` | commander; 1 seat, or 2 for mission 50 |
| `PickTask(by, taskID)` | `pickTask` | draft; dummy's turn → commander |
| `SendDistress(by, dir)` / `SkipDistress(by)` | `sendDistress` / `skipDistress` | commander |
| `ChoosePassCard(by, card)` | `choosePassCard` | each human |
| `Handover(by, taskID, toSeat)` | `handover` | commander, 5 players, golden |
| `Communicate(by, card, pos)` | `communicate` | `pos` is ignored in a dead zone |
| `PlayCard(by, card)` | `playCard` | the seat to play; dummy → commander |
| `GiveUp(by)` | `giveUp` | anyone, playing |
| `NextAttempt(by)` / `NextMission(by)` / `BackToLobby(by)` | same names | host, debrief |
| `EndGame(by)` | `endGame` | host, lobby/debrief |

Errors are typed sentinels (`ErrWrongPhase`, `ErrWrongStep`, `ErrNotYourTurn`, `ErrNotCommander`, `ErrCardNotInHand`, `ErrMustFollowSuit`, `ErrCannotCommunicate`, `ErrAlreadyCommunicated`, `ErrUnevenDistribution`, `ErrCannotPassTrump`, …). Each maps to a code in `apidto/errors.go`. On the wire, members are referenced by **seat index** (the dummy has no PlayerID).

### 5.9 Deep Sea seams (keep them, don't build them)

- `Game.Edition`, with the mission table and the task source keyed by edition.
- Task setup through a small interface or switch:
  - Planet Nine: draw *n* colour-card tasks and place tokens;
  - Deep Sea: draw task cards until the difficulty sum is reached.
- `ObjectiveKind` is an open enum. New kinds are added as cases with their own tests.
- `CommRule` grows: `Shared` (rapture of the deep: a pool of n−2 tokens), `Random` (unfamiliar terrain).
- `Distribution` grows: `openWithPass` (Deep Sea allows passing when there are fewer tasks than members), `volunteers`, `free`.
- No Planet Nine names in `trick.go`, `objective.go` or `seat.go`: say trump, dummy, commander. The UI maps them per edition (rocket/submarine, JARVIS/Tonoja, commander/captain).

## 6. Service, ports, adapters

Mirror Fox Duet file by file; only the game-specific parts change.

- **ports:** the `GameService` methods from the intent table, plus Create/Join/Authenticate/Summary/View/Leave (copied). The repository and publisher are unchanged.
- **service:** `act(...)` wrappers as in Fox Duet. No `Rematch`: the session continues instead.
- **sqlite:** `0001_init.sql` copied. `dto.go` serialises `Game` and `Attempt` explicitly (no struct tags on domain types). Missions are not persisted; they are looked up by number.
- **apidto:**
  - DTOs and mappers for `PlayerView`, `Event` and errors;
  - `ClientMessage` gains `mission`, `seats`, `taskId`, `card`, `position`, `direction`, `yes`, `swap`;
  - `apidto_test.go` covers every DTO mapping.
- **ws:** command dispatch for every command in §5.8.
- **httpapi:** REST for create/join/summary/view (copied).
- **e2e:** 2 players (dummy) and 4 players over httptest. Play a scripted attempt of mission 1 to a win with a seeded RNG, and check that no view leaks another hand.
- **cmd/server:** `serverkit.LoadConfig(":8084", "data/the-crew.db")`.

## 7. Frontend (`frontend/`)

Copy Fox Duet's skeleton: routes, `api.ts`, `session.ts`, `game/client.ts`, `game/connection.svelte.ts`, `recent.ts`, `errors.ts`, the shared wiring in `vite.config.ts` / `tsconfig.json` / `layout.css` / `package.json`, and shadcn components as needed.

**Screens:**
- **Home** (`/`): create or join, plus the recent games list (copied).
- **Lobby:** the players (with "JARVIS joins as the third crew member" when there are 2), the **mission picker** (a 1–50 grid with each mission's icons, the logbook state per mission and the next one preselected), the selected mission's briefing (the §14 row in words), the logbook total, and Start / End for the host.
- **Game table** (mobile portrait, top to bottom):
  1. **Top bar:** mission number, attempt, a commander badge, communication rule badges (dead zone, "comms from trick N"), and a rules/help button.
  2. **Crew strip:** one compact chip per other member: name, commander/silenced badges, hand count, tricks won, mini task chips with status, and the communicated card with its token position (top/middle/bottom marker, or "?" in a dead zone). The **JARVIS chip** expands into its 7 stacks: the face-up card on each, with a marker when a hidden card lies under it.
  3. **Mission panel:** the task cards with token, owner and status (open, done, failed), and the special objectives as one-line chips with owner and status.
  4. **Trick area:** up to 5 played cards labelled by seat, the leader marked, and the winner highlighted after the trick. A "last trick" peek.
  5. **Decision panel:** depends on the step: draft (tap a task), poll (yes/no or well/ill), commander pick (choose seat(s)), token swap/move, distress (send left/right, skip), pass (choose a card), handover, or "waiting for X". When it is JARVIS's turn, the commander gets a prompt to tap a JARVIS card.
  6. **Own area:** my tasks and the **Communicate** button (enabled only when allowed). It opens a sheet: pick an eligible card, then the allowed positions.
  7. **Hand:** a fan (copied from Fox Duet, adapted) with illegal cards dimmed. 13–14 cards must fit at 360 px width.
- **Debrief:** the result with its reason ("Blue 7 was won by Anna, but it is Ben's task", "Token 2 was fulfilled third"), all hands revealed, the logbook line, and the host buttons Next attempt / Next mission / Lobby.
- **RulesHelp:** a short in-game summary of tricks, communication and the current mission's symbols.

**Cards:** our own SVG/CSS card face: the value large, the colour glyph and the colour. Rockets are dark with a rocket glyph. Task cards are a smaller version with the token badge. Card motion (deal, play, collect the trick, pass) uses `cardMotion` from shared, as in Fox Duet.

**Fixtures:** `game/fixtures.ts` has a `PlayerView` for every step and player count. The table can then be developed and screenshot-tested without a backend (Fox Duet does the same).

## 8. Integration checklist (deployment, start page, docs)

Every place another app is named (checked against how Fox Duet was added):

- [ ] `the-crew/Dockerfile`, `the-crew/Dockerfile.dockerignore`: copy Fox Duet's (root context, `COPY shared/...`) and rename.
- [ ] `the-crew/deploy/compose.yaml`: `name: the-crew`, `container_name: the-crew`, `BASE_PATH: /the-crew`, image `.../the-crew:${TAG:-latest}`, volume `data`, network `web`.
- [ ] `the-crew/deploy/the-crew.caddy`: `redir /the-crew /the-crew/`, `handle_path /the-crew/*` with `import gate` and `reverse_proxy the-crew:8080`.
- [ ] CI (`.github/scripts/plan-deploy.sh`), `deploy/deploy.sh`, `deploy/local.sh` and the server script pick up apps automatically (a top-level directory with `deploy/compose.yaml`). Nothing to edit there; verify with a manual workflow run if unsure.
- [ ] `start-page/frontend/src/lib/games.ts`: add `'the-crew'` to the `id` union, and an entry with `path: '/the-crew/'`, `devPort: 5177`, `players: '2–5 players'`, a tagline, and features (`['50 missions', '2 players with JARVIS']`).
- [ ] `start-page/frontend/src/routes/+page.svelte`: tile branches (background, ring, hover colours, header art: a small rocket/planet SVG).
- [ ] `start-page/frontend/src/routes/layout.css`: The Crew's tile colour tokens (light and dark), and the comment listing the themes.
- [ ] `start-page/frontend/vite.config.ts`: the port comment.
- [ ] `start-page/frontend/src/routes/datenschutz/+page.svelte`: add „The Crew" (/the-crew/) to the list of games. No new data categories: same lobby, session token and `localStorage` use as the other games. Re-check if that changes.
- [ ] `gate/backend/cmd/server/main.go`: the port comment (`:8084` The Crew).
- [ ] Root `CLAUDE.md`: the app table, the layout tree, the local ports table, the deployment bullet (images, volumes).
- [ ] `the-crew/CLAUDE.md`: new, modelled on `fox-duet/CLAUDE.md`. Content: app specifics (ports, intents, per-player view), domain modelling rules (from §5 of this plan), look and feel (§3, §7).
- [ ] `IDEAS.md`: mark The Crew as implemented, and point the Deep Sea entry at §11 of this plan.
- [ ] `deploy/README.md`: no per-app list. Check only.
- [ ] After merge, push to `main`. CI tests, builds and deploys `the-crew` and `start-page` (and `gate`, if its comment changed). The server needs no `setup-server.sh` re-run: no change in `deploy/caddy/` or `deploy/server/`.

## 9. Testing

- **Domain (thorough, table-driven):**
  - every section of rules.md: tricks and follow suit incl. trump and the dummy; dealing per player count; the commander; open draft; commander's decision and distribution incl. the evenness rule; polls; token order incl. the same-trick cases; every objective kind (fulfil, fail early, fail at the end); communication (eligibility, timing, dead zone, disruption, silenced); distress and pass (no trump, direction, 2 players); handover; mission 12's draw; give up; logbook counting incl. distress +1;
  - each mutator's errors leave the state unchanged (assert with `Clone` + `reflect.DeepEqual`, as Fox Duet does).
- **Mission table:** all 50 rows against an independent expectation table. For every mission with special rules, at least one scripted win and one scripted loss.
- **Simulation (`sim_test.go`):** for every mission × player count 2–5 × several seeds, play random legal actions until debrief. Invariants:
  - no panic;
  - the 40 cards are conserved (hands + dummy + tricks + leftover);
  - every view hides what §5.7 says;
  - every attempt ends;
  - the step's actors are never empty while playing.
- **Adapters:** repotest (copied), sqlite round trip of a mid-attempt game with the dummy, apidto mapping, ws dispatch, httpapi.
- **e2e:** §6.
- **Frontend:** vitest for format helpers (objective and reason texts, token labels), the recent-games logic and the communicate eligibility display. Fixtures render in `npm run check`.
- **Manual QA:** Playwright at 390×844 in both themes. Play 2 players (two browser contexts) through missions 1, 5, 12, 20, 24 and 50. Spot-check 3 and 5 players.
- Before declaring done, in `the-crew/backend`: `go vet ./... && go test ./...`; in `the-crew/frontend` and `start-page/frontend`: `npm run check && npm run lint && npm test`; plus `gate/backend` if touched.

## 10. Work breakdown and agents

The implementing agent (main session, Opus) is the **orchestrator and owner of the hard parts**. It delegates mechanical, well-specified work to **Sonnet** subagents (`Agent` tool with `model: "sonnet"`). Every subagent brief contains:
- the work package text below;
- the files to read (always rules.md, this plan, and the Fox Duet files it mirrors);
- the exact files it may touch;
- the acceptance commands.

The orchestrator **reviews every diff** before building on it. Parallel subagents that touch different directories run with `isolation: "worktree"`; otherwise run them one after another in the main tree. Haiku is not needed. If you want it, use it only for read-only chores such as "run all checks and summarise failures".

| WP | Work | Model | Depends on | Acceptance |
|---|---|---|---|---|
| **1** | **Scaffold.** Copy `fox-duet/` → `the-crew/`. Rename module, ports, DB, container, base path, titles. Strip the Fox Duet domain down to a compiling **lobby-only** skeleton (Game with lobby lifecycle, view, events), with matching service/adapters/frontend. Add the Dockerfile, dockerignore, compose and caddy files. | Sonnet | – | All backend and frontend checks pass. `go run` + `npm run dev` show create/join/lobby at :5177. |
| **2** | **Integration.** Everything in §8 except the-crew's own CLAUDE.md: start page tile, privacy page, root CLAUDE.md, gate comment, IDEAS.md. | Sonnet (worktree, parallel to 1) | – | start-page checks pass; the tile links to :5177 in dev. |
| **3** | **Domain foundation.** `card.go`, `trick.go`, `seat.go` (dummy), the `mission.go` types, `objective.go` + `tokens.go` with the full evaluation, the `Attempt` struct, dealing per player count, the commander, `play` with end detection. Tests for all of it. Fix the public method signatures for §5.8 (stubs that return `ErrNotImplemented` are fine) so WP4, WP6 and WP7 can start. | **Opus** | 1 | `go test ./internal/core/domain` green; the objective and token tables cover rules.md §7 and §13. |
| **4** | **Mission table.** Transcribe rules.md §14 into `missions_planet9.go`. Separately write the expectation table test from rules.md, **without** reading the implementation table. | Sonnet | 3 (types) | Test green; the orchestrator compares both against §14 row by row. |
| **5** | **Domain flows.** Setup sequence, `adjustTokens`, polls, picks, draft (incl. dummy), commander's decision and distribution with evenness, communication, distress and pass, handover, mission 12's draw, give up, debrief, logbook, the lobby mission flow, `PlayerView` with the hiding rules. | **Opus** | 3, 4 | Domain tests green; view tests prove the hiding rules. |
| **6** | **Black-box rule tests.** From rules.md §5–§13 only, write table-driven tests against the public domain API (signatures from WP3), without reading the flow implementation. Failures go to the orchestrator, who decides whether code or test is wrong (rules.md is the authority) and fixes the loser. | Sonnet (worktree) | 3 | Merged and green after reconciliation. |
| **7** | **Simulation test.** `sim_test.go` per §9. | Sonnet | 5 | Green for all missions × 2–5 players × ≥20 seeds. |
| **8** | **Adapters + service + e2e.** Ports, service, sqlite DTO, apidto (+ errors, ws commands), ws dispatch, httpapi, e2e. Follow Fox Duet file by file. | Sonnet | 5 | `go vet`/`go test` green. The orchestrator reviews the view → DTO mapping for leaks. |
| **9** | **Frontend plumbing.** The `types.ts` mirror of apidto, client commands, connection, format helpers (objective, reason, token texts) with tests, `fixtures.ts` for every step × 2/3/5 players, the lobby mission picker and logbook. | Sonnet | 8 (DTOs) | `npm run check/lint/test` green; the lobby works end to end. |
| **10** | **Game table UI.** Theme, card faces, Table layout, CrewStrip with the JARVIS stacks, MissionPanel, TrickArea, DecisionPanel (all steps), the Communicate sheet, Debrief, RulesHelp. Opus builds the Table, DecisionPanel and Communicate flow; the simpler leaf components (RulesHelp, Debrief list, theme tokens) may go to Sonnet once the patterns stand. | **Opus** (+ Sonnet for leaves) | 9 | Fixtures render at 360–430 px in both themes. Playable end to end with 2 players. |
| **11** | **QA and review.** Manual Playwright session per §9, `/code-review` on the branch, fixes, final checks in every touched app. Show the user the decisions list in §2 if not done yet. | **Opus** | 10 | All checks green; the user plays a 2-player session locally. |
| 12 | *(Optional, after v1)* Extract the hand fan into `shared/frontend/CardHand.svelte` and use it in Fox Duet and The Crew (§4). | Sonnet | 11 | Checks green in shared, fox-duet and the-crew. Fox Duet behaves the same. |

**Critical path:** 1 → 3 → 5 → 8 → 9 → 10 → 11. **In parallel:** 2 with 1; 4 and 6 with the end of 3 and with 5; 7 after 5, alongside 8.

**Review gates** (the orchestrator stops and checks):
- after WP3: are the objective and token semantics exactly §7/§13?
- after WP4: all 50 rows match §14;
- after WP5: the hidden-information rules;
- after WP8: no leaks on the wire;
- after WP10: a real 2-player playthrough.

**Commit cadence:** one commit per work package (or smaller) on a feature branch. Push to `main` only after WP11, with the user's go-ahead.

## 11. Deep Sea outlook

From the Deep Sea rulebook (read once for this plan; not needed again until Deep Sea is built):

**The same as Planet Nine:**
- the 40-card deck (rockets are called submarines);
- tricks;
- communication (sonar tokens);
- the distress signal;
- the 2-player dummy (called Tonoja; the same setup and rules);
- the commander (called captain), who holds trump 4 and starts the selection and the first trick.

**New:**
- **96 task cards**, each with a difficulty for 3, 4 and 5 players. A mission sets a total difficulty, and cards are drawn (skipping cards that overshoot) until the sum is reached exactly. Task cards are "I will win …" conditions: specific cards, exactly or at least N of a kind, none of a kind, the first or last tricks, trick counts and "in a row", more or fewer tricks than others, trick value sums, a prediction (secret or open), equal numbers of two colours, more of one colour in hand at the end, all of a colour or one of each colour, odd/even tricks, never opening with certain colours.
- **No task tokens.**
- **Distribution:** passing is allowed when there are fewer tasks than members. After a failed attempt, keep the same tasks or draw new ones. Impossible combinations get a redraw. Some trump alignments force a re-deal that does not count as an attempt.
- **Communication modes:** currents (= dead zone), rapture of the deep (n−2 shared tokens, first come, multiple uses allowed), unfamiliar terrain (a random colour card picks normal, currents or rapture).
- **Mission modes:** real-time missions (a timer, optional), volunteer missions (14, 15, 16, 26: yes/no until 1 or 2 volunteers), free task selection.

**Data we do not have yet:** the 96 task-card texts with their difficulty triples, and the 32-mission logbook. They are on the physical cards and in the separate logbook, not in the rulebook PDF. **Ask the user for a source** (photos or a community list) before planning Deep Sea in detail.

**Recommended home:** the same app, with an edition choice in the lobby. Deep Sea then needs:
- a second mission table;
- a task deck with difficulties;
- the new objective kinds (each a case in `objective.go`, all evaluable over the trick history plus, for some, the final hands, the leads and a prediction stored per task);
- the new communication and distribution modes;
- the edition naming in the UI.

The seams are listed in §5.9.

## 12. Risks and how to handle them

| Risk | Handling |
|---|---|
| Rule misreadings in objectives and tokens | Two independent test sets (WP3 and WP6), the rules.md formulas, and the sim test. |
| Hidden-information leaks (other hands, the dummy's hidden cards, pass and draw cards) | Private data lives only in per-viewer view fields, and events are card-free for private moves. View tests, the e2e leak check, and the WP8 review gate. |
| The step machine gets tangled (many setup variants) | The setup sequence is derived from `MissionDef` in one function. Each step has one actor rule. The sim test covers every mission. |
| Busy phone layout with 5 players + tasks + communications | Compact chips, task details on tap, fixtures for 5 players, and a manual check at 360 px. |
| Live races (a lead vs. a communication before a trick) | First come, first served on the server (one lock per game). The UI disables the action as soon as the view changes. Documented in rules.md §8. |
| Licensing | Private behind the gate, our own art and text (rules.md header). |
