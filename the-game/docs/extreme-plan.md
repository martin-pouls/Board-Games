# Phase 2 Implementation Plan: The Game Extreme (command cards)

> **Status:** implemented (waves A, B and C: domain, wire/transports, frontend). The manual 360 px visual check and the manual run against the server (wave C acceptance) are still open. Baseline: the working tree on `main`, i.e. Phase 1 plus the uncommitted host "end game early" change (`OutcomeAborted`, `EndGame`, `/end-game`, `endGame` WS command).
> **Audience:** the implementation agents for waves A, B and C (see §8). This document is the contract. If it conflicts with code written before this plan, this plan wins. If it conflicts with code written *to* this plan in an earlier wave, the code wins; report the mismatch in your handback.
> Read `CLAUDE.md`, `docs/rules.md` §2 and `docs/plan.md` (§0 ground rules, §1 domain, §6 wire) first.
> **Before wave A starts, the owner answers §3 (Open rule questions).** Implement the **recommended** answer for every question the owner has not answered, and list those in your handback.

---

## 0. Ground rules (in addition to plan.md §0)

- **The rules follow the cards, not the variant.** No rule code checks `Rules.Variant == VariantExtreme`. Classic decks have no commands, so every command rule is a no-op in classic. The Phase 1 test suite is the regression suite for classic and must pass **unchanged**, except the rows listed in wave A that assert "extreme is rejected".
- **All command state is derived, never stored.** It is computed from `Piles`, `Turn.Plays` and `Turn.StartHandSize`, which are already persisted. So there is no new persisted field, no DTO change and no migration (§4).
- **Illegal moves are blocked, never punished.** Every veto returns a typed sentinel and leaves the state unchanged: `expectErr` atomicity applies to every new error. A game is lost only through the existing paths: `checkStuck` (no legal move and the turn cannot end) and the "no cards left but the turn cannot end" branch in `PlayCard`.
- **No lookahead** (plan.md §11.1). The one exception is a single-card look at "the state right after this card" (`withPlay`, §2.6). STOP needs it because STOP ends the turn the moment it lands. It is *not* a search.
- Done means: `go vet ./... && go test ./...` in `backend/`; `npm run format && npm run check && npm run lint && npm test` in `frontend/`.

---

## 1. Summary of design decisions

| Topic | Decision |
|---|---|
| **Command constants** | Seven `Command` string constants in `card.go`. Their values are the camelCase wire strings: `stop`, `skull`, `exactlyThree`, `noCommunication`, `noBackwardsTrick`, `singlePile`, `drawOnlyOne`. `Command.Class()` returns `lightning` / `infinity` / `""`. Extreme cards keep `Kind = CardKindNumber`. |
| **Deck factory** | `NewDeck` gets a `VariantExtreme` branch. It builds the same 98 number cards and sets `Command` from an unexported table `extremeCommandValues` in `deck.go`, copied from the rules.md table (7 × 4 values). The table is deck data, not a ruleset parameter, so `Ruleset` and its DTO do not change. |
| **Options** | `Options.Validate` accepts `extreme`. **Pro and Reduced hand size combine with Extreme** exactly as with classic (Pro: min 3; Reduced: only with Pro). This is Q4, recommended "allowed". |
| **Derived command state** | New file `commands.go`, `func deriveCommands(piles []Pile, plays []Play) commandState`. **Infinity:** a command is active iff its card is the **top** card of a pile, in any turn. **Lightning:** a command counts iff its card is in `Turn.Plays` (this turn), whether or not it is covered. A lightning card on top from an earlier turn means nothing. |
| **Active Infinity commands** | `commandState.active []ActiveCommand{Command, Pile}`, one entry per pile whose top card is an Infinity card, in pile order. Duplicates (two `noBackwardsTrick` piles) are listed twice; the flags are idempotent. Exposed as `PlayerView.ActiveCommands` → JSON `activeCommands`. |
| **STOP** | **Placement rule:** a STOP may be played only if the turn could legally end right after it. The rule runs every end-turn check on `withPlay(p)`; `minimumCheck` passes there because STOP overrides the minimum. **Follow-up:** after a STOP, `PlayCard` auto-ends the turn (`TurnEnded{Auto:true}`), so "no card after STOP" holds by construction. `ErrTurnOver` (reserved in plan.md) is **not** needed and is not defined. |
| **3!** | `commandState.exactCount = 3` once a 3! is in `Turn.Plays`. **Placement:** a card after 3 plays, or a 3! that would itself be the 4th or later card, is blocked with `ErrExactCountExceeded`. A 3! when `Turn.StartHandSize < 3` is blocked with `ErrExactCountUnreachable`. **STOP before the 3rd card** fails the end-turn check on the after-state and is mapped to `ErrStopTooEarly`. **End of turn:** `exactCountCheck` requires `len(Plays) == 3` (`ErrExactCountRequired`). The turn does not auto-end at exactly 3 (app decision D3). |
| **Skull coverage** | `commandState.skullPiles` lists the piles whose top is a Skull played **this turn**. `skullCheck` (end of turn) fails with `ErrSkullUncovered` while the list is non-empty. A STOP that would leave a skull showing is blocked with the same error. A Skull that can never be covered because 3! allows no further card is blocked with `ErrSkullUncoverable` (Q5, recommended "block"). |
| **Single pile binding** | Active iff some pile's top card is `singlePile`. The bound pile is derived. (1) The **latest Single pile card played this turn that is still on top** binds to its own pile (owner decision). (2) Otherwise, every visible Single pile card has been visible since the turn started, so the **first card of the turn** picked the pile: `Plays[0].Pile`. (3) With no plays yet, nothing is bound and the next card picks. **Covering the Single pile card cancels the binding immediately** (the brief, plus "covering cancels its effect"). See Q1 for the consequence and the alternative. |
| **No backwards trick** | A constraint vetoes `PlacementBackwardsTrick` while any `noBackwardsTrick` card is visible. It is checked at placement time, so it automatically affects only cards played **after** the NBT card. Earlier trick plays this turn stay. Covering the NBT card lifts it at once. |
| **Draw only one** | `drawCount()` is evaluated in `endTurn`, i.e. at turn end with the final piles: `n = refill; if drawOnlyOne visible { n = min(n, 1) }; n = min(n, len(DrawPile))`. Once the card is covered, refill semantics restore the full hand size automatically. |
| **No communication** | `Game.CommunicationRestricted()` returns `Phase == playing && a noCommunication card is visible`. It already flows to `PlayerView.CommunicationRestricted` → JSON `communicationRestricted`. The UI shows a banner and a toast. There is no chat; a future chat must gate on this flag **server-side**. |
| **Mid-turn covering** | Nothing is cached. Constraints, end-turn checks, the draw policy and the view call `g.commands()` on the current state, so covering an Infinity card takes effect for the very next action. |
| **Errors** | **Separate sentinels, one per blocked situation**, each with its own wire code. None of them wraps `ErrIllegalPlacement` or `ErrMinimumNotReached`. Justification in §2.5. |
| **Per-card reason** | `HandCard.Blocked error` (domain) → `blocked?: ApiErrorCode` (wire). It is the first command veto for a hand card with **no** legal pile that the base rules alone would allow somewhere. The UI shows `errorText(blocked)` instead of "No legal pile". |
| **Turn constraints in the view** | `TurnView.Constraints` holds `exactCount`, `singlePile`, `boundPile`, `skullPiles`, `stopPlayed`. `PileView.CommandActive` says the top card's command is in effect: an Infinity top, or a lightning top played this turn. The client renders these facts and never re-derives rules. |
| **Events** | No new event types. `cardPlayed.card.command` and `turnEnded.auto` / `cardsDrawn` are enough for toasts: "STOP", "draw only one". |

---

## 2. Domain changes (`backend/internal/core/domain`)

### 2.1 `card.go`
```go
const (
    CommandStop             Command = "stop"
    CommandSkull            Command = "skull"
    CommandExactlyThree     Command = "exactlyThree"
    CommandNoCommunication  Command = "noCommunication"
    CommandNoBackwardsTrick Command = "noBackwardsTrick"
    CommandSinglePile       Command = "singlePile"
    CommandDrawOnlyOne      Command = "drawOnlyOne"
)
type CommandClass string
const (
    CommandClassNone      CommandClass = ""
    CommandClassLightning CommandClass = "lightning" // effect for the rest of the turn of the player who played it
    CommandClassInfinity  CommandClass = "infinity"  // effect for everyone while visible on top of a pile
)
func (c Command) Class() CommandClass // switch; unknown/none → CommandClassNone
func (c Command) IsLightning() bool
func (c Command) IsInfinity() bool
```
Update the `Command` doc comment, which currently says "Phase 1 only uses CommandNone".

### 2.2 `deck.go`
```go
// extremeCommandValues: rules.md §2 "Command → card values".
var extremeCommandValues = map[Command][4]int{
    CommandStop: {6, 35, 57, 95}, CommandSkull: {17, 41, 61, 87}, CommandExactlyThree: {19, 28, 53, 81},
    CommandNoCommunication: {26, 65, 78, 92}, CommandNoBackwardsTrick: {21, 32, 59, 70},
    CommandSinglePile: {9, 47, 68, 84}, CommandDrawOnlyOne: {13, 38, 43, 76},
}
```
- `NewDeck`: the `VariantExtreme` case builds the classic deck, then sets `deck[v-LowestCard].Command` for every table value. A value outside `[LowestCard, HighestCard]` returns `ErrInvalidOptions` (defensive).
- Update the doc comment.

### 2.3 `ruleset.go`
- `Validate`: `case VariantClassic, VariantExtreme:`. Keep the reduced-without-Pro check.
- Remove the "rejected for now" comment.
- `NewRuleset` is unchanged: same hand sizes and minimums (rules.md: "Hand sizes and minimums are the same").

### 2.4 `commands.go` (new)
```go
// Command parameters printed on the cards (not table config, so not in Ruleset).
const (
    exactlyThreeCount = 3 // 3!
    drawOnlyOneCount  = 1 // Draw only one
)

// ActiveCommand is a visible Infinity command and the pile it lies on (view type).
type ActiveCommand struct {
    Command Command
    Pile    int
}

// commandState is everything the Extreme commands imply right now.
// It is derived from Piles and Turn.Plays and never stored.
type commandState struct {
    active           []ActiveCommand // visible Infinity commands, pile order; nil if none
    noBackwardsTrick bool
    singlePile       bool
    boundPile        int  // -1 = not bound (no Single pile, or the next card picks)
    drawOnlyOne      bool
    noCommunication  bool
    exactCount       int   // exactlyThreeCount once a 3! was played this turn, else 0
    stopPlayed       bool
    skullPiles       []int // piles whose top card is a Skull played this turn; nil if none
}

func (g *Game) commands() commandState { return deriveCommands(g.Piles, g.Turn.Plays) }
func deriveCommands(piles []Pile, plays []Play) commandState
func singlePileBinding(plays []Play) int  // only called when singlePile is active
func playedThisTurn(plays []Play, pile int) bool // some play this turn went on pile, so its top is this turn's card
```
Derivation sketch:
```go
cs := commandState{boundPile: -1}
for i, p := range piles {
    top, ok := p.TopCard()
    if !ok { continue }
    if top.Command.IsInfinity() { cs.active = append(cs.active, ActiveCommand{top.Command, i}) }
    switch top.Command {
    case CommandNoBackwardsTrick: cs.noBackwardsTrick = true
    case CommandSinglePile:       cs.singlePile = true
    case CommandDrawOnlyOne:      cs.drawOnlyOne = true
    case CommandNoCommunication:  cs.noCommunication = true
    case CommandSkull:
        if playedThisTurn(plays, i) { cs.skullPiles = append(cs.skullPiles, i) }
    }
}
for _, p := range plays {
    switch p.Card.Command {
    case CommandStop:         cs.stopPlayed = true
    case CommandExactlyThree: cs.exactCount = exactlyThreeCount
    }
}
if cs.singlePile { cs.boundPile = singlePileBinding(plays) }
```
The `singlePileBinding` sketch shows the decided rule:
```go
// 1. The latest Single pile card played this turn that is still on top binds to its own pile.
for i := len(plays) - 1; i >= 0; i-- {
    p := plays[i]
    stillTop := !slices.ContainsFunc(plays[i+1:], func(q Play) bool { return q.Pile == p.Pile })
    if p.Card.Command == CommandSinglePile && stillTop { return p.Pile }
}
// 2. Otherwise the visible Single pile card(s) were visible since the turn started:
//    the first card picked the pile, and every play since went there.
if len(plays) > 0 { return plays[0].Pile }
return -1 // 3. Active but unbound: the next card picks.
```
`playedThisTurn` uses pile membership, not card values. This keeps it correct for future duplicate-value kinds (Joker).

### 2.5 `errors.go`: new sentinels (all `errors.New`, wrapped with detail via `%w`)

| Sentinel | Raised by | Situation |
|---|---|---|
| `ErrBackwardsTrickBlocked` | PlayCard | backwards trick while a No backwards trick card is visible |
| `ErrSinglePileViolation` | PlayCard | card not on the bound pile |
| `ErrExactCountExceeded` | PlayCard | a 4th card after 3!, or a 3! that would be the 4th or later card |
| `ErrExactCountUnreachable` | PlayCard | 3! when the turn started with fewer than 3 cards |
| `ErrStopTooEarly` | PlayCard | STOP after 3! as the 1st or 2nd card of the turn |
| `ErrSkullUncovered` | EndTurn; PlayCard (STOP) | turn would end with a skull from this turn on top |
| `ErrExactCountRequired` | EndTurn | turn would end with fewer than 3 cards after 3! |
| `ErrSkullUncoverable` | PlayCard | Q5: any 3rd card after 3! (Skull, 3! or plain) that leaves a skull from this turn uncovered |

**Why separate sentinels rather than one `ErrCommandViolation` with reasons:**
1. They fit the existing `apidto.errorTable` (one sentinel ↔ one code ↔ one status), so REST and WS need no new mechanism.
2. The TS `ApiErrorCode` union plus `Record<ApiErrorCode, string>` in `errors.ts` makes a missing message a compile error.
3. `HandCard.blocked` can carry a plain code.
4. A single sentinel would need a reason sub-field on `ErrorBody`, a typed error struct and `errors.As` in `ErrorFor`. That is more surface for the same information.

None of them wraps `ErrIllegalPlacement` or `ErrMinimumNotReached`, so `errorTable` order stays irrelevant. Code that needs "any veto" already checks `err != nil` (`LegalMoves`). Update the plan.md reserved-names line: `ErrTurnOver` is not needed, and `ErrExactCountRequired` / `ErrSkullUncovered` are used as reserved.

### 2.6 `placement.go`
```go
type noBackwardsTrickRule struct{ cs commandState }
type singlePileRule       struct{ cs commandState }
type exactCountRule       struct{ cs commandState }
type stopRule             struct{}
type skullConflictRule    struct{} // Q5

func (g *Game) placementConstraints() []PlacementConstraint {
    cs := g.commands()
    return []PlacementConstraint{
        exactCountRule{cs},       // count first: pile-independent
        singlePileRule{cs},
        noBackwardsTrickRule{cs},
        stopRule{},               // needs the after-state
        skullConflictRule{},      // needs the after-state
    }
}

// withPlay returns a read-only projection of g after placing p (card out of the active hand,
// onto the pile, appended to Turn.Plays). It shares unchanged slices and maps with g:
// never mutate it, emit on it or commit it.
func (g *Game) withPlay(p Placement) *Game {
    after := *g
    after.events = nil
    after.Piles = slices.Clone(g.Piles)
    after.Piles[p.PileIndex].Cards = append(slices.Clip(g.Piles[p.PileIndex].Cards), p.Card)
    after.Turn.Plays = append(slices.Clip(g.Turn.Plays), Play{Card: p.Card, Pile: p.PileIndex})
    seat := g.Turn.ActiveSeat
    after.Players = slices.Clone(g.Players)
    after.Players[seat].Hand = slices.DeleteFunc(slices.Clone(g.Players[seat].Hand),
        func(c Card) bool { return c.Value == p.Card.Value })
    return &after
}
```
Rule bodies:
- **noBackwardsTrickRule:** `cs.noBackwardsTrick && p.Kind == PlacementBackwardsTrick` → `ErrBackwardsTrickBlocked`. This includes a trick that would cover the NBT card itself (Q2).
- **singlePileRule:** `cs.singlePile && cs.boundPile >= 0 && p.PileIndex != cs.boundPile` → `ErrSinglePileViolation` ("all cards this turn go on pile %d").
- **exactCountRule:** with `played := len(g.Turn.Plays)`:
  - `cs.exactCount > 0 && played >= cs.exactCount` → `ErrExactCountExceeded`.
  - If the card is a 3!: `played+1 > exactlyThreeCount` → `ErrExactCountExceeded`; else `g.Turn.StartHandSize < exactlyThreeCount` → `ErrExactCountUnreachable`.
- **stopRule:** if the card is a STOP, `err := g.withPlay(p).runEndTurnChecks()`. Then:
  - `nil` → allowed.
  - `errors.Is(err, ErrExactCountRequired)` → `fmt.Errorf("%w: STOP must be the third card after 3!", ErrStopTooEarly)`.
  - Anything else (`ErrSkullUncovered`) → `fmt.Errorf("STOP would end the turn: %w", err)`.
  - A STOP on the skull's own pile covers it, so it is allowed.
- **skullConflictRule** (every card; skipped when it cannot be the 3rd card): `after := g.withPlay(p); cs := after.commands()`. If `cs.exactCount > 0 && len(after.Turn.Plays) >= cs.exactCount && len(cs.skullPiles) > 0` → `ErrSkullUncoverable`.

`checkPlacement` is unchanged: base rule first, so base-illegal moves still return `ErrIllegalPlacement`, then the constraints in order.

### 2.7 `play.go`
- **End-turn checks:**
  ```go
  func (g *Game) endTurnChecks() []EndTurnCheck {
      return []EndTurnCheck{skullCheck{}, exactCountCheck{}, minimumCheck{}} // most specific message first
  }
  // runEndTurnChecks runs every check without the phase check (used by CanEndTurn and on withPlay states).
  func (g *Game) runEndTurnChecks() error
  func (g *Game) CanEndTurn() error // phase check, then runEndTurnChecks()
  ```
  - `skullCheck`: `len(cs.skullPiles) > 0` → `ErrSkullUncovered` ("skull on pile %d").
  - `exactCountCheck`: `cs.exactCount > 0 && len(Plays) != cs.exactCount` → `ErrExactCountRequired` ("played %d, need exactly %d").
  - `minimumCheck`: add `if g.commands().stopPlayed { return nil }` (STOP overrides the minimum). Otherwise unchanged.
- **Draw policy:**
  ```go
  func (g *Game) drawCount() int {
      hand := len(g.Players[g.Turn.ActiveSeat].Hand)
      n := max(0, g.Rules.HandSize(len(g.Players))-hand) // refill
      if g.commands().drawOnlyOne {                       // evaluated at turn end (rules.md)
          n = min(n, drawOnlyOneCount)
      }
      return min(n, len(g.DrawPile))
  }
  ```
- **PlayCard follow-ups.** Only the second case changes:
  ```go
  case g.noCardsLeft():                       // unchanged: won if CanEndTurn()==nil, else lost (skull last card)
  case (card.Command == CommandStop || len(g.Players[seat].Hand) == 0) && g.CanEndTurn() == nil:
      g.endTurn(true)                         // STOP always lands here: stopRule pre-validated the after-state
  default:
      g.checkStuck()
  ```
- **Legality per card** (used by `LegalMoves` and `ViewFor`):
  ```go
  // cardLegality checks c against every pile: the legal piles, and, when there are none,
  // the first command veto (in pile order) of a placement the base rules allowed (nil otherwise).
  func (g *Game) cardLegality(c Card) (piles []int, blocked error)
  ```
  `LegalMoves` loops the hand through `cardLegality`. Its output stays identical.
- Update doc comments. `PlayCard` errors: + `ErrBackwardsTrickBlocked`, `ErrSinglePileViolation`, `ErrExactCountExceeded`, `ErrExactCountUnreachable`, `ErrStopTooEarly`, `ErrSkullUncovered`, `ErrSkullUncoverable`. `EndTurn` errors: + `ErrSkullUncovered`, `ErrExactCountRequired`.

### 2.8 `stuck.go`
No logic change. `ActivePlayerStuck` already composes `LegalMoves` (now command-aware) and `CanEndTurn` (now with skull and 3! checks). Update the doc comment to say commands are included.

### 2.9 `game.go`
- `CommunicationRestricted()`: `return g.Phase == PhasePlaying && g.commands().noCommunication`.
- Fix the `OutcomeLost` comment, which currently says "score > 0". It becomes: "active player stuck; score ≥ 0 (0 only when the last card was an uncovered Skull)".

### 2.10 `view.go`
```go
type PlayerView struct { /* ... */ ActiveCommands []ActiveCommand } // nil if none
type HandCard   struct { Card Card; LegalPiles []int; Blocked error } // Blocked only for the active viewer
type PileView   struct { /* ... */ CommandActive bool }  // top card's command is in effect: Infinity top, or lightning top played this turn
type TurnView   struct { /* ... */ Constraints TurnConstraints }
type TurnConstraints struct {
    ExactCount int   // 3 once a 3! was played this turn; 0 = none
    SinglePile bool  // a visible Single pile card restricts this turn
    BoundPile  *int  // pile all further cards must go on; nil = unbound (the next card picks) or no Single pile
    SkullPiles []int // piles showing a Skull played this turn; nil = none
    StopPlayed bool  // only ever true in a finished game (STOP auto-ends the turn)
}
```
- The zero values mean "no command", so classic `reflect.DeepEqual` expectations in `view_test.go` keep passing unchanged.
- `ViewFor` computes `cs := g.commands()` once.
- `ViewFor` copies slices so the view never aliases game state. Existing test pattern: `slices.Clone(cs.skullPiles)`, `slices.Clone(cs.active)`, and a fresh `*int` for `BoundPile`.

### 2.11 `events.go`
No change.

### 2.12 Edge cases (decided by rules.md, the brief, or §3 recommendations)

| Case | Behaviour |
|---|---|
| STOP is the last card overall | `stopRule` passed, so `CanEndTurn()==nil` and the `noCardsLeft` branch gives **won**, score 0. |
| Skull is the last card overall | `noCardsLeft` and `CanEndTurn` fails with skull, so **lost** with score 0 and `StuckPlayerID = actor` (rules.md). UI: §6.7. |
| Skull is the last **hand** card but others still hold cards | No auto-end (`CanEndTurn` fails). `checkStuck`: no hand, cannot end, so **lost**. |
| 3!, then STOP as the 3rd card | Allowed. Auto-ends with exactly 3 plays. |
| 3!, then STOP as the 1st or 2nd card | `ErrStopTooEarly`. |
| 3! with `StartHandSize` 1 or 2 | `ErrExactCountUnreachable`. If 3! is the player's only card at the start of a turn: no legal move and cannot end, so **lost** at turn start (consistent with rules.md's "game is lost" for this case). |
| Hand empties with constraints outstanding | Skull: stuck, so lost. 3!: cannot happen (3! needs `StartHandSize ≥ 3`, so the hand cannot empty before 3 plays). Minimum: stuck, so lost (existing; reachable through Draw only one, see Q7). Single pile: no end-of-turn obligation. |
| Draw pile empty (min 1) + 3! | 3! still requires exactly 3. Min 1 is irrelevant. STOP is still allowed only as the 3rd card. |
| Single pile + backwards trick | The trick is allowed on the bound pile. If a NBT card is also visible, the trick is blocked there too. |
| Several Infinity cards at once | All apply independently. Duplicates are idempotent. There is one binding. `activeCommands` lists every one. |
| Covering an Infinity card mid-turn | The effect ends for the next action: NBT covered means the trick is legal again; Single pile covered means free; Draw only one covered means refill at the end; No communication covered means the flag is false in the same view. |
| Covering a lightning card mid-turn | No effect: 3! and STOP count from `Plays`. A skull counts only while on top (by definition). |
| STOP with a skull from this turn showing on another pile | `ErrSkullUncovered`. A STOP placed on the skull pile is allowed. |
| Skull covered by another Skull | The new skull must be covered. |
| Draw only one played by the active player this turn, visible at turn end | They draw 1 (evaluated at turn end, rules.md). |
| STOP in solo | Auto-end, draw, and the same player starts the next turn. |
| 3 cards reached after 3! with cards left | The 4th card is blocked. `LegalMoves` is empty but `CanEndTurn` is nil, so the player is **not stuck**. No auto-end (D3). |

---

## 3. Open rule questions (owner to confirm; implement the recommended answer by default)

> **Owner answers (2026-10-10):** all questions are decided as follows. Q1: **covering ends the binding at once** (the recommended answer). Q2: no. Q3: yes. Q4: allowed. Q5: **block**; later extended (2026-10-10): a **plain** 3rd card that leaves a skull from this turn uncovered after 3! is blocked too, so `skullConflictRule` runs for every card. Q6: the flag is false outside playing. Q7: **literal rule** (the game is lost). Wave A records these in rules.md §2 *Interpretations (decided)*.

1. **Does covering a Single pile card end its binding?** rules.md says both "covering the card cancels its effect" and "all *remaining* cards of that turn must go on the pile where the Single pile card was played". The brief says covering removes the effect immediately.
   - **Recommended: yes, covering ends it.** Consequences: a Single pile card played mid-turn only forces the **next** card onto its pile, because that card covers it. A Single pile card visible at turn start is neutralised if the first card goes on top of it.
   - *Alternative:* the binding lasts the whole turn. It is still derivable without persistence: bind to the latest Single pile card played this turn (covered or not), else to `Plays[0].Pile` if any pile's top-before-this-turn (`Cards[len-1-playsOn(i)]`) was a Single pile card.
2. **May a backwards trick cover the No backwards trick card itself?** **Recommended: no.** The card is active while visible, and the placement happens while it is visible.
3. **May an Infinity card itself be placed by a move its own effect forbids?** For example, the NBT card placed via the backwards trick. **Recommended: yes.** It takes effect only once it is on the pile ("from the moment it is played"). Another visible NBT card would still block it.
4. **Extreme + Pro (and + Reduced hand size)?** **Recommended: allowed**, as independent toggles.
5. **A Skull as the 3rd card after 3!**, or a 3! as the 3rd card while a skull from this turn is uncovered. No further card is allowed, so the skull can never be covered.
   - **Recommended: block** with `skull_uncoverable`. It is a conflict between two commands, decidable without lookahead, and analogous to the decided "STOP after 3! only as the 3rd card".
   - *Alternative:* allow it. The player is then immediately stuck and the game is lost, consistent with the Skull-as-last-card ruling.
6. **No communication outside playing.** **Recommended:** the flag is `false` in lobby, choosing and finished. Talking after the game is fine.
7. **Draw only one can shrink a hand below the minimum while the draw pile still has cards.** For example, a 1-card hand when the minimum is 2. **Recommended: literal base rule.** The player plays what they can and is then stuck, so the game is lost. The engine already does this (`TestStuckWhenHandEmptiesBelowMinimum`).
   - *Alternative:* cap the minimum at `StartHandSize`.

**App (UX) decisions, not rule questions:**
- **D1** No new event types.
- **D2** A "lost with score 0" headline reads "So close!…".
- **D3** No auto-end when exactly 3 is reached.
- **D4** Lightning tops from earlier turns are rendered muted.

After the owner answers, wave A records Q1–Q7 under rules.md §2 *Interpretations (decided)*. It also records the derived items as interpretations: "3! as the 4th+ card is blocked" and "STOP that would leave a skull uncovered is blocked".

---

## 4. Persistence (`adapters/sqlite`, `adapters/memory`)

- **No production change, no migration, `schemaVersion` stays 1.**
  - `cardDTO` already has `Command string json:"command"`, and `cardToDTO` / `cardFromDTO` round-trip it. Hands, draw pile, piles and `Turn.Plays` all use `cardDTO`.
  - `optionsDTO.Variant` and `rulesDTO.Variant` are strings, and the `games.variant` column is `TEXT NOT NULL` with no CHECK constraint.
  - All Extreme turn state is derived from `Piles`, `Turn.Plays` and `Turn.StartHandSize`, which are already persisted. No new `Ruleset` fields are added, because the deck table and the 3/1 constants stay in the domain.
- **memory repo:** `Clone` deep-copies `Card` values. No change.
- **New tests only (wave B):**
  - `sqlite/games_test.go` `TestRoundTripExtreme`: a playing game with `Options{Variant: extreme, Pro: true}`. Command cards in hands, the draw pile, on piles (Infinity tops) and in `Turn.Plays` (a 3! and an uncovered skull). The test asserts JSON equality of `toDTO` before and after, `variant` column `= 'extreme'`, and that `g.ViewFor` gives the same `ActiveCommands` / `Constraints` after load.
  - The e2e restart check in §7.3.

---

## 5. Wire contract changes (Go `adapters/apidto`, TS `frontend/src/lib/types.ts`)

### 5.1 Go DTOs (`dto.go`)
```go
type Card struct {
    Value   int    `json:"value"`
    Kind    string `json:"kind"`              // "number"
    Command string `json:"command,omitempty"` // stop|skull|exactlyThree|noCommunication|noBackwardsTrick|singlePile|drawOnlyOne
}
type HandCard struct {
    Card
    LegalPiles []int  `json:"legalPiles"`
    Blocked    string `json:"blocked,omitempty"` // error code of the command that blocks this card (only when legalPiles is empty)
}
type GameOptions struct {
    Variant string `json:"variant"` // "classic" | "extreme"; empty on input means classic
    // Pro, ReducedHandSize unchanged
}
type Pile struct {
    // existing fields ...
    CommandActive bool `json:"commandActive"` // topCard.command is in effect now
}
type ActiveCommand struct {
    Command string `json:"command"`
    Pile    int    `json:"pile"`
}
type TurnConstraints struct {
    ExactCount *int  `json:"exactCount"` // 3 after a 3!, else null
    SinglePile bool  `json:"singlePile"`
    BoundPile  *int  `json:"boundPile"`  // null = not bound
    SkullPiles []int `json:"skullPiles"` // never null
    StopPlayed bool  `json:"stopPlayed"`
}
type Turn struct {
    // existing fields ...
    Constraints TurnConstraints `json:"constraints"`
}
type PlayerView struct {
    // existing fields ...
    ActiveCommands []ActiveCommand `json:"activeCommands"` // never null
}
```
- `map.go`:
  - `FromView` fills the new fields. `ExactCount` 0 → `nil`. `SkullPiles` and `ActiveCommands` are never nil (`make(…, 0)`). Copy `BoundPile` into a fresh pointer.
  - New helper `blockedCode(err error) string`: `""` if nil, else `ErrorFor(err)` code, and `""` if that code is `internal`.
  - New helper `FromActiveCommand`.
- `ToDomain` already passes `"extreme"` through.

### 5.2 Error codes (`errors.go`), all HTTP **422**

| Domain sentinel | `Code*` const | code |
|---|---|---|
| `ErrBackwardsTrickBlocked` | `CodeBackwardsTrickBlocked` | `backwards_trick_blocked` |
| `ErrSinglePileViolation` | `CodeSinglePileViolation` | `single_pile_violation` |
| `ErrExactCountExceeded` | `CodeExactCountExceeded` | `exact_count_exceeded` |
| `ErrExactCountUnreachable` | `CodeExactCountUnreachable` | `exact_count_unreachable` |
| `ErrExactCountRequired` | `CodeExactCountRequired` | `exact_count_required` |
| `ErrStopTooEarly` | `CodeStopTooEarly` | `stop_too_early` |
| `ErrSkullUncovered` | `CodeSkullUncovered` | `skull_uncovered` |
| `ErrSkullUncoverable` | `CodeSkullUncoverable` | `skull_uncoverable` (drop if Q5 = allow) |

### 5.3 Events
Unchanged. `cardPlayed.card` now carries `command` for command cards.

### 5.4 Example JSON (extreme, mid-turn)
```jsonc
"piles": [{ "index": 2, "direction": "down", "startValue": 100, "topValue": 59,
            "topCard": { "value": 59, "kind": "number", "command": "noBackwardsTrick" },
            "cardCount": 4, "commandActive": true }, ...],
"activeCommands": [{ "command": "noBackwardsTrick", "pile": 2 }],
"turn": { ..., "constraints": { "exactCount": 3, "singlePile": false, "boundPile": null, "skullPiles": [0], "stopPlayed": false } },
"hand": [{ "value": 81, "kind": "number", "command": "exactlyThree", "legalPiles": [], "blocked": "exact_count_exceeded" }]
```

### 5.5 TypeScript (`types.ts`)
```ts
export type Variant = 'classic' | 'extreme';
export type CommandName =
	| 'stop' | 'skull' | 'exactlyThree'
	| 'noCommunication' | 'noBackwardsTrick' | 'singlePile' | 'drawOnlyOne';
export interface Card { value: number; kind: 'number'; command?: CommandName }
export interface HandCard extends Card { legalPiles: number[]; blocked?: ApiErrorCode }
export interface GameOptions { variant: Variant; pro: boolean; reducedHandSize: boolean }
export interface Pile { /* existing */ commandActive: boolean }
export interface ActiveCommand { command: CommandName; pile: number }
export interface TurnConstraints {
	exactCount: number | null; singlePile: boolean; boundPile: number | null;
	skullPiles: number[]; stopPlayed: boolean;
}
export interface Turn { /* existing */ constraints: TurnConstraints }
export interface PlayerView { /* existing */ activeCommands: ActiveCommand[] }
export type ApiErrorCode = /* existing */
	| 'backwards_trick_blocked' | 'single_pile_violation' | 'exact_count_exceeded'
	| 'exact_count_unreachable' | 'exact_count_required' | 'stop_too_early'
	| 'skull_uncovered' | 'skull_uncoverable';
```

---

## 6. Frontend changes (mobile-first, 360 px target, matching existing components)

### 6.1 `src/lib/game/commands.ts` (new, pure TS; no Svelte imports, so it is node-testable)
```ts
export type CommandClass = 'lightning' | 'infinity';
export interface CommandInfo { kind: CommandClass; label: string; short: string; description: string }
export const COMMANDS: Record<CommandName, CommandInfo>;
export function commandInfo(c: CommandName): CommandInfo;
/** Target dots for TurnBar: exactCount ?? requiredMinimum. */
export function targetCount(turn: Turn): number;
/** Hints for TurnBar, from server facts only (constraints + activeCommands). */
export function turnHints(view: PlayerView, isMyTurn: boolean): { command: CommandName; text: string }[];
```
Command metadata:

| command | kind | label | short | description |
|---|---|---|---|---|
| stop | lightning | STOP | STOP | Ends the turn immediately. |
| skull | lightning | Skull | Skull | Must be covered in the same turn. |
| exactlyThree | lightning | 3! | 3! | Exactly 3 cards this turn. |
| noCommunication | infinity | No communication | Silence | No talking about the game while it is visible. |
| noBackwardsTrick | infinity | No backwards trick | No trick | No backwards trick while it is visible. |
| singlePile | infinity | Single pile | One pile | All cards of a turn on one pile while it is visible. |
| drawOnlyOne | infinity | Draw only one | Draw 1 | Draw only 1 card at the end of the turn while it is visible. |

`turnHints` rules, in this order:
- `skullPiles.length` → "Cover the skull before ending your turn" (others' turn: "{name} must cover the skull").
- `exactCount` → "3!: exactly 3 cards this turn".
- `singlePile`: if `boundPile === null` → "Single pile: the first card picks the pile"; else "Single pile: only the marked pile".
- `noBackwardsTrick` in `activeCommands` → "No backwards trick".
- `drawOnlyOne` → "Draw only 1 at turn end".

### 6.2 New components (`src/lib/components/game/`)
- **`CommandIcon.svelte`** `{ command: CommandName; class?: string }`, using lucide icons:

  | command | icon |
  |---|---|
  | stop | `OctagonX` |
  | skull | `Skull` |
  | exactlyThree | text "3!" (`font-black`), since lucide has no glyph for it |
  | noCommunication | `MessageCircleOff` |
  | noBackwardsTrick | `Undo2` with an absolutely positioned `Slash` overlay |
  | singlePile | `Target` |
  | drawOnlyOne | `ArrowDownToLine` |

  Kind marks: `Zap` (lightning), `Infinity` (infinity).
- **`CommandChip.svelte`** `{ command; muted?: boolean; size?: 'sm'|'xs' }`: a pill with the kind mark, the icon and the short label.
  - Lightning uses red tones (`bg-red-500/15 text-red-800 dark:text-red-200`). Infinity uses violet (`bg-violet-500/15 text-violet-800 dark:text-violet-200`). These avoid clashing with the sky (up) and amber (down) pile colours.
  - `muted` gives `opacity-50`, with the title "No effect any more".
- **`ActiveCommands.svelte`** `{ commands: ActiveCommand[]; piles: Pile[] }`: a wrapping row of `CommandChip`s with a tiny ↑/↓ for the pile direction. Tapping a chip shows `toast(info.description, { id })`. It renders nothing when the list is empty.
- **`NoCommunicationBanner.svelte`**: a full-width violet bar with `MessageCircleOff` and the text "No communication: don't talk about the game until the card is covered." `role="status"`.

### 6.3 Updated components
- **`Lobby.svelte`** (Options card): a new first row, a Switch `opt-extreme`, "Extreme (command cards)", calling `setOptions({ variant: on ? 'extreme' : 'classic' })`.
  - Pro and Reduced stay independent toggles (Q4).
  - Under it, when on: "28 cards carry commands: STOP, Skull, 3!, No communication, No backwards trick, Single pile, Draw only one."
  - The summary line appends "· Extreme" when on.
- **`Hand.svelte`**: a command card gets a kind-tinted border (`border-red-400` / `border-violet-400`) and a `CommandIcon` (`size-3.5`) under the value. The value stays `text-lg`, which fits the ~40 px cell at 8 cards.
  - `aria-label`: `Card 95, STOP` (+ ", no legal pile").
- **`PileCard.svelte`**: new props `bound?: boolean`, `skull?: boolean`.
  - When `pile.topCard?.command` is set, the bottom line shows a `CommandChip` (`muted={!pile.commandActive}`) at the left and "+N this turn" at the right.
  - `skull`: red ring plus a "Cover!" chip, pulsing unless `prefersReducedMotion`.
  - `bound`: a `Target` icon in the header and a dashed emerald outline.
  - The aria label adds the command and its state.
- **`TurnBar.svelte`**:
  - Dots use `targetCount(turn)`. The label reads "{played}/{target} played", plus " · exactly" when `exactCount` is set.
  - Below the status line, a wrapping row of `text-xs` hint chips from `turnHints` (icon and text), shown on both your turn and others' turns.
  - The End turn button logic is unchanged (`canEndTurn`); the hints explain why it is disabled.
- **`Board.svelte`**:
  - Header: an `EXTREME` badge (`bg-violet-600 text-white`) next to `PRO`.
  - `NoCommunicationBanner` under the header when `view.communicationRestricted`.
  - `ActiveCommands` between `PlayerStrip` and the pile grid.
  - Pass `bound={turn?.constraints.boundPile === pile.index}` and `skull={turn?.constraints.skullPiles.includes(pile.index)}`.
  - `onselect`: if a card has no legal pile → `toast(card.blocked ? errorText(card.blocked) : 'No legal pile for this card', { id: 'no-legal-pile' })`.
  - Event toasts (existing `$effect` on `client.lastEvents`):
    - `cardPlayed` STOP: "STOP: your turn is over." or "{name} played STOP."
    - `cardPlayed` noCommunication: "No communication until the card is covered." (id `no-comm`)
- **`GameOver.svelte`**: revealed-hand chips show a tiny `CommandIcon` for command cards.
- **`ChooseFirstPlayer`** renders the read-only Board, so it gets the icons for free.

### 6.4 `errors.ts` messages
| code | message |
|---|---|
| `backwards_trick_blocked` | "No backwards trick while that card is visible." |
| `single_pile_violation` | "Single pile: all your cards this turn go on the same pile." |
| `exact_count_exceeded` | "3!: exactly 3 cards this turn, no more." |
| `exact_count_unreachable` | "You can't play 3!: you started this turn with fewer than 3 cards." |
| `exact_count_required` | "3!: play exactly 3 cards before ending your turn." |
| `stop_too_early` | "After 3!, STOP can only be your third card." |
| `skull_uncovered` | "Cover the skull first: it can't stay on top when your turn ends." |
| `skull_uncoverable` | "After 3! no 4th card is allowed, so this skull could never be covered." |

### 6.5 `format.ts`
- `resultHeadline`: check `outcome` before `score`. `lost && score === 0` → "So close! All cards played, but a skull stayed uncovered."
- `stuckLine`: if `view.turn?.constraints.skullPiles.length` → "{name} couldn't cover the skull."; else if `exactCount !== null && cardsPlayed < exactCount` → "{name} couldn't complete 3!."; else the existing "had no playable card". Use "You…" for self.

### 6.6 `fixtures.ts`
- `card(value, command?)` helper.
- Every `Pile` gets `commandActive`, every `Turn` gets `constraints` (an empty constant `noConstraints`), and every view gets `activeCommands: []`.
- New fixtures, all added to `fixtures`:
  - `lobbyHostExtreme` (variant extreme + pro).
  - `extremeMyTurn`: `activeCommands` noBackwardsTrick on pile 2 and singlePile on pile 1, `boundPile` 0, a STOP and a 3! in hand, one card with `blocked: 'single_pile_violation'`.
  - `extremeSkull`: `exactCount` 3, `skullPiles` [3], `canEndTurn` false.
  - `extremeOthersNoComm`: `communicationRestricted` true, others' turn.
  - `finishedLostSkull`: lost, score 0, `turn.constraints.skullPiles` [1].

### 6.7 Layout (360 px)
```
| K7MQ2 (o) PRO EXTREME   Deck 64  End |
| [No communication: don't talk ...]   |  only while restricted
| [*Anna 6] [Ben 6] [You 6] [Cleo 6]   |
| [∞ No trick ↓] [∞ One pile ↑]        |  active commands (hidden if none)
| +-----------+ +-----------+          |
| | ^ 1 (o)   | | ^ 1       |          |  (o) = bound pile marker
| |    34     | |    12     |          |
| | [∞ Draw 1]| |  +1       |          |  command chip on the top card
| +-----------+ +-----------+          |
| Your turn ●●○ 2/3 · exactly [End]    |
| [3! exactly 3] [Cover the skull]     |  hint chips
| [7][15][23ø][41][56][62⚡][80][93]    |  hand with command icons
```

---

## 7. Test plan

### 7.1 Domain (wave A), table-driven, every error case through `expectErr` (atomicity)

**Builder change (`helpers_test.go`):**
- `table.build` maps every value (hands, draw, piles, `plays[i].Card`) through a lookup built from `NewDeck(rules)`, falling back to `NumberCard(v)`. With `opts: extreme` the real command values apply automatically (95 = STOP, 41 = Skull, ...). Classic is unchanged.
- Add `extreme := Options{Variant: VariantExtreme}`, `extremePro`, and `play(v, pile int) Play`.

| Area | Test (file) | Cases |
|---|---|---|
| Deck | `TestNewDeckExtreme` (deck_test) | 98 cards, 2–99 once each, ascending, `Kind` number; exactly 28 with a command, 4 per command, values equal the rules.md table; classic deck has 0 commands |
| Options | `TestOptionsValidate` (ruleset_test) | extreme, extreme+pro, extreme+pro+reduced valid; extreme+reduced without Pro invalid. Flip the "extreme rejected" rows in `ruleset_test.go`, `lifecycle_test.go:43` and `:137` to success cases |
| Start | `TestStartExtremeDealsCommands` (lifecycle_test) | noShuffle solo hand = 92..99 with 92 `noCommunication`, 95 `stop` |
| Derivation | `TestDeriveCommands` (commands_test, new) | no tops gives zero state; each Infinity top sets its flag and an `active` entry; covered Infinity gives nothing; 2 different plus 2 identical Infinity tops (pile order); lightning top from an earlier turn has no effect; skull this turn on top → `skullPiles`; skull covered → none; skull on skull → new pile; STOP/3! in plays (even covered) → `stopPlayed`/`exactCount`; Single pile binding: at start, no plays → unbound; first play on q → q; first play covers S → inactive; S played mid-turn → own pile; S played then covered → free; two S at start, one covered by first card → first card's pile; S at start, then S' played on the bound pile → bound |
| No backwards trick | `TestNoBackwardsTrick` (extreme_test, new) | trick blocked on every pile (`ErrBackwardsTrickBlocked`), normal plays allowed; trick onto the NBT card itself blocked (Q2); NBT card placed via trick allowed (Q3), then the next trick blocked; NBT played mid-turn: earlier trick play stays, later trick blocked; NBT covered mid-turn → trick legal again |
| Single pile | `TestSinglePile` | at start: first card free (LegalMoves on all piles), second card elsewhere → `ErrSinglePileViolation`, same pile OK; first card covers S → free (Q1); S mid-turn: earlier cards stay, next card only on S's pile; after that card covers S → free (Q1); Single pile + trick on bound pile OK; + NBT → trick blocked on bound pile |
| Several Infinity | `TestSeveralInfinityActive` | NBT + SP + DOO + NC visible together: all effects, `ActiveCommands` lists 4, draw 1, `CommunicationRestricted` true |
| 3! placement | `TestExactlyThreePlacement` | 4th card after 3! (3! as 1st/2nd/3rd) → `ErrExactCountExceeded`; 3! as 4th card → `ErrExactCountExceeded`; `StartHandSize` 1 and 2 → `ErrExactCountUnreachable`; 3 → OK; draw pile empty (min 1) still exactly 3; Pro (min 3) consistent; two 3! in one turn OK |
| 3! end turn | `TestExactlyThreeEndTurn` | after 3!, EndTurn with 1/2 plays → `ErrExactCountRequired`, 3 → OK; hand empties at exactly 3 → auto-end; 3 reached with cards left: LegalMoves empty, `CanEndTurn` nil, not stuck, no auto-end (D3); stranded 3! (only card, start hand 1) → lost at turn start |
| STOP | `TestStop` | STOP as 1st card: auto-end, events `[CardPlayed, TurnEnded{Auto, CardsDrawn:1}, TurnStarted]`, Revision +1, minimum overridden; STOP as 2nd card with min 3 (Pro); STOP in solo → same player, new turn; STOP as last card overall → won, score 0; STOP as last hand card (draw empty) → exactly one TurnEnded; after STOP the former player's PlayCard → `ErrNotYourTurn`; STOP after 3! as 1st/2nd → `ErrStopTooEarly`; 3! then STOP as 3rd → OK; STOP with this turn's skull on another pile → `ErrSkullUncovered`; STOP on the skull pile → OK |
| Skull | `TestSkull` | after a skull, EndTurn → `ErrSkullUncovered` (atomic); cover → can end; skull on skull → still blocked; no covering card → stuck → lost (`StuckPlayerID`, events `CardPlayed`, `GameFinished`); skull as last card overall → lost, score 0; skull as last hand card with others holding cards → no auto-end, lost; skull covered but minimum unmet → `ErrMinimumNotReached`; Q5: skull as 3rd after 3! → `ErrSkullUncoverable`, 3! as 3rd with a skull uncovered → `ErrSkullUncoverable`, 3! as 3rd **on** the skull pile → OK |
| Draw only one | `TestDrawOnlyOne` (+ rows in `TestDrawCount`) | visible at end → draw 1; played by the active player this turn → 1; visible at start but covered → refill; earlier shrunk hand, DOO gone → refill to hand size; draw pile empty → 0; one card left → 1; Q7: hand of 1, draw pile non-empty, min 2 → plays it, then lost |
| No communication | `TestCommunicationRestricted` | true while visible in playing; false after covered (same command); false in lobby/choosing/finished |
| Blocked reason | `TestHandCardBlocked` (view_test) | 3! with start hand 2 → `Blocked` is `ErrExactCountUnreachable`; base-illegal card → nil; card legal somewhere → nil; non-active viewer → nil |
| View | `TestViewForExtreme` (view_test) | `ActiveCommands`, `PileView.CommandActive` (Infinity top true; lightning top this turn true, from an earlier turn false), `Turn.Constraints` (each field), no aliasing of `SkullPiles`/`BoundPile`; classic views unchanged (existing tests untouched) |
| Random | `TestRandomGames` (play_test) | add `extreme` and `extremePro` to `optsList` (e.g. 250 seeds). Existing invariants, plus: before every `EndTurn` `CanEndTurn()==nil`; after any STOP play the turn advanced or the game finished; a `TurnEnded` while DOO was visible has `CardsDrawn ≤ 1`; `lost` with score 0 only if a skull is on top; no panic. Relax the `lost ⇒ score != 0` assertion accordingly |

Coverage goal: `go test ./internal/core/domain/ -cover` ≥ 90 %.

### 7.2 Adapters and service (wave B)
- **`apidto_test.go`:**
  - `TestFromViewMapsEveryField`: add `ActiveCommands`, `CommandActive`, `Constraints` (`exactCount` 3, `boundPile` 1, `skullPiles` [0], `stopPlayed`), `Blocked` → code. Expected `want` gets `ActiveCommands: []ActiveCommand{…}`.
  - `TestFromViewEmptySlicesMarshalAsArrays`: `activeCommands: []`, `constraints.skullPiles: []`, `exactCount: null`, `boundPile: null`.
  - Command card JSON `{"value":95,"kind":"number","command":"stop"}`; a plain card has no `command` key; a hand card without `blocked` has no key.
  - `TestErrorFor`: 8 new rows, also wrapped (`fmt.Errorf("%w: …")`).
- **`httpapi/games_test.go`:** `TestErrorTable` + 8 rows (422). `TestCreateGame` with `{"options":{"variant":"extreme","pro":true}}` passes `domain.Options{Variant: extreme, Pro: true}` to the fake.
- **`ws/ws_test.go`:** `TestCommandErrors` + one row (`skull_uncovered` on `endTurn`) proving new codes travel over WS.
- **`sqlite/games_test.go`:** `TestRoundTripExtreme` (§4).
- **`service/service_test.go`:**
  - `TestExtremeGame`: CreateGame with extreme options; noShuffle solo start → hand has 92 `noCommunication` and 95 `stop`.
  - `PlayCard(95, 0)` → one publish with `[CardPlayed, TurnEnded, TurnStarted]`, view turn 2.
  - `PlayCard(92, 2)` → view `CommunicationRestricted` true.
  - `SetOptions` to extreme+pro accepted.

### 7.3 E2E (wave B, `internal/e2e/e2e_test.go`)
- Add `startAppRNG(t, dbPath, newRNG func() domain.RNG)`; `startApp` delegates to it. Add `type identityRNG struct{}` (no shuffle).
- **`TestExtremeSoloDeterministic`** (identity RNG, deck ascending, solo hand = 92..99, draw top = 91):
  1. Create with `{variant:"extreme"}` and start. `hand` has `command` on 92 and 95 only. `activeCommands == []`. `constraints == {exactCount:null, singlePile:false, boundPile:null, skullPiles:[], stopPlayed:false}`.
  2. Over WS (dial first), REST `end-turn` → 422 `minimum_not_reached`.
  3. `plays {card:95, pile:0}` → 200. Turn 2, `cardsPlayed` 0, hand `[91..94, 96..99]`. `piles[0].topCard.command == "stop"`, `commandActive == false`. WS `view` events: `cardPlayed` (card.command `stop`), `turnEnded` (auto, cardsDrawn 1), `turnStarted`.
  4. `plays {card:92, pile:2}` → `communicationRestricted == true`, `activeCommands == [{command:"noCommunication", pile:2}]`.
  5. Restart the app on the same DB and re-fetch the view with the same token: options.variant `extreme`, `activeCommands` / `communicationRestricted` / `topCard.command` are unchanged.
- **`TestExtremeGamesToCompletion`:** seeds {1,2,3} × {solo, 3 players} with extreme (+ one Pro). `playToEnd` works unchanged (it uses `legalPiles` / `canEndTurn`). `checkFinished` is relaxed: `won ⇒ score 0`, and `lost` with score 0 is allowed only when some pile top has `command == "skull"`.

### 7.4 Frontend (wave C, vitest, node env)
- `commands.test.ts`: `COMMANDS` covers every `CommandName` with non-empty texts; `targetCount`; `turnHints` for every fixture (empty for classic; skull / 3! / single pile unbound and bound / NBT / DOO; others'-turn wording).
- `errors.test.ts`: add the 8 codes.
- `format.test.ts`: lost with score 0 headline; aborted and won unchanged; `stuckLine` skull and 3! variants and fallback.
- Manual: render each extreme fixture at 360×740 and on desktop. 8 command cards must fit without horizontal scroll, and the chips must not overflow the piles.

---

## 8. Work breakdown (sequential waves; each leaves the repo green)

Shell setup on this machine (Git Bash): `export PATH="/c/Program Files/Go/bin:/c/Program Files/nodejs:$PATH"`.

### Wave 0: owner (no agent)
Answer §3 Q1–Q7. Whatever is unanswered defaults to the recommendation.

### Wave A: domain rules and domain tests (strong model, high effort)
- **Goal:** §2 complete, §7.1 complete.
- **Files owned:**
  - `backend/internal/core/domain/{card.go, deck.go, ruleset.go, commands.go (new), errors.go, placement.go, play.go, stuck.go (comment), game.go, view.go}`
  - Tests: `{helpers_test.go, deck_test.go, ruleset_test.go, lifecycle_test.go, commands_test.go (new), extreme_test.go (new), placement_test.go, play_test.go, view_test.go}`
  - `docs/rules.md` §2 *Interpretations*: only the owner-confirmed answers; otherwise list them in the handback.
- **Depends on:** wave 0 (soft).
- **Must not touch:** adapters, service, frontend. Domain signatures used by adapters stay as they are; the view types are only extended.
- **Acceptance:**
  - Every §7.1 row exists and passes.
  - The Phase 1 tests pass unchanged, except the three "extreme rejected" rows.
  - Domain coverage ≥ 90 %. The whole backend still builds and its tests pass: apidto does not map the new fields yet, which is fine.
- **Verify** (from `backend/`):
  ```sh
  go vet ./... && go test ./... && go test ./internal/core/domain/ -count=1 -cover
  ```

### Wave B: wire, persistence checks, transports, service, e2e (Sonnet/strong, medium)
- **Goal:** §4, §5.1–§5.4, §7.2, §7.3.
- **Files owned:**
  - `backend/internal/adapters/apidto/{dto.go, map.go, errors.go, apidto_test.go}`
  - `backend/internal/adapters/sqlite/games_test.go`
  - `backend/internal/adapters/httpapi/games_test.go`
  - `backend/internal/adapters/ws/ws_test.go`
  - `backend/internal/core/service/service_test.go`
  - `backend/internal/e2e/e2e_test.go`
  - No production changes are expected in `sqlite`, `memory`, `httpapi`, `ws` or `service`. If you need one, explain it in the handback.
- **Depends on:** A.
- **Acceptance:**
  - JSON matches §5 exactly (camelCase, never-null slices, `null` for `exactCount` / `boundPile`).
  - All 8 codes map to 422 on REST and travel over WS.
  - The extreme game round-trips through sqlite, and the e2e restart keeps commands.
  - No migration was added.
- **Verify** (from `backend/`):
  ```sh
  go vet ./... && go test ./... && go test ./internal/e2e/ -count=1 -v -run Extreme
  ```

### Wave C: frontend (strong model, medium; the UX matters)
- **Goal:** §5.5 and §6. Also update `CLAUDE.md` (Phase 2 implemented) and this plan's status line.
- **Files owned:**
  - `frontend/src/lib/{types.ts, errors.ts, errors.test.ts}`
  - `frontend/src/lib/game/{commands.ts (new), commands.test.ts (new), format.ts, format.test.ts, fixtures.ts}`
  - `frontend/src/lib/components/game/{CommandIcon.svelte, CommandChip.svelte, ActiveCommands.svelte, NoCommunicationBanner.svelte (all new), Hand.svelte, PileCard.svelte, TurnBar.svelte, Board.svelte, Lobby.svelte, GameOver.svelte}`
  - `CLAUDE.md`
  - `docs/extreme-plan.md` (status only)
- **Depends on:** B (wire contract live; TS mirrors §5.5 verbatim).
- **Acceptance:**
  - Type-check, lint and tests pass.
  - Every fixture renders at 360×740 and on desktop: 8 command cards fit, piles stay prominent, End turn is thumb-reachable, hints and chips wrap without overflow.
  - Manual run against `go run ./cmd/server`: an Extreme+Pro lobby toggle round-trips; in a solo extreme game, STOP auto-ends with a toast; a No communication card shows the banner; a blocked card shows its specific toast.
- **Verify** (from `frontend/`):
  ```sh
  npm run format && npm run check && npm run lint && npm test
  ```
  Optionally `npm run build` and a manual run with `STATIC_DIR=../frontend/build go run ./cmd/server`.

---

### Critical Files for Implementation
- C:\Users\marti\github\The-Game\backend\internal\core\domain\commands.go (new; derivation of all command state) together with C:\Users\marti\github\The-Game\backend\internal\core\domain\placement.go (constraint composition, `withPlay`)
- C:\Users\marti\github\The-Game\backend\internal\core\domain\play.go (end-turn checks, draw policy, STOP auto-end, `cardLegality`)
- C:\Users\marti\github\The-Game\backend\internal\core\domain\view.go and C:\Users\marti\github\The-Game\backend\internal\core\domain\helpers_test.go (deck-aware `table{}` builder)
- C:\Users\marti\github\The-Game\backend\internal\adapters\apidto\dto.go / map.go / errors.go (wire contract and error codes)
- C:\Users\marti\github\The-Game\frontend\src\lib\types.ts plus C:\Users\marti\github\The-Game\frontend\src\lib\components\game\Board.svelte (and PileCard/Hand/TurnBar/Lobby)

**Things I verified in the code that drive the plan:**
- C:\Users\marti\github\The-Game\backend\internal\adapters\sqlite\dto.go already persists `cardDTO.Command`, so no migration and no DTO change are needed.
- `apidto.errorTable` is first-match, which is why the new sentinels do not wrap `ErrIllegalPlacement`.
- These tests currently assert that Extreme is rejected and must flip in wave A:
  - `lifecycle_test.go` lines 43 and 137
  - `ruleset_test.go` row "extreme rejected"
  - `deck_test.go` `TestNewDeckExtremeRejected`
- `TestRandomGames` and the e2e `checkFinished` both assume "lost implies score > 0". A Skull as the last card breaks that assumption, so both must be relaxed. `resultHeadline` in C:\Users\marti\github\The-Game\frontend\src\lib\game\format.ts would currently say "You won!" for a lost game with score 0.
