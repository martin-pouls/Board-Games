# Phase 1 Implementation Plan: The Game (base and Pro), online multiplayer

> **Status (2026-10-10): Phase 1 is implemented** (WP1–WP9, including the §11 amendments). Kept as the reference for the wire contract and design decisions.

> Audience: the implementation sub-agents. This document is the contract. If a contract here conflicts with what you find in the code, the code from an earlier wave wins only if it was written to this document. Otherwise follow this document and report the mismatch in your handback.
> Read `CLAUDE.md` and `docs/rules.md` first. Both are authoritative for conventions and rules.

---

## 0. Ground rules for every agent

- **Dependency direction:** `adapters → core/service → core/ports → core/domain`. Domain is stdlib only. All rules live in the domain. Adapters only translate.
- **Frozen contracts:** the exported Go types and signatures in §1, §2 and §4–§6, and the TS types in §6, are frozen after WP1/WP2. If you really must change one, change the Go DTO, the TS type and this document together, and say so in your handback.
- **Done means:** in `backend/`, `go vet ./... && go test ./...` pass. In `frontend/`, `npm run check && npm run lint && npm test` pass. Run `npm run format` before lint.
- **Frontend gotchas:**
  - Import with `#lib/...js` (not `$lib`). Runes only.
  - Any rune-using TS module must be named `*.svelte.ts`.
  - `href` and `goto` must go through `resolve()` from `$app/paths` (lint rule `svelte/no-navigation-without-resolve`), e.g. `goto(resolve('/g/[code]', { code }))`.
  - Use `page` from `$app/state` for params.
  - `crypto.randomUUID`, `navigator.clipboard` and `navigator.share` need a secure context. Phones testing over `http://192.168.x.x` do not have one, so always feature-detect and fall back.
- **Go gotchas:**
  - Wrap errors with `%w`. `ctx` is the first param on anything doing I/O.
  - Domain errors are sentinels, matched with `errors.Is`. Domain methods must **validate fully before mutating**: on error, state stays unchanged.
- **Game ID = join code.** One identifier everywhere (`games.id`, URLs, WS). It is 5 chars from the alphabet `23456789ABCDEFGHJKLMNPQRSTUVWXYZ` (32 symbols, no 0/O/1/I). The service normalises input with `strings.ToUpper(strings.TrimSpace(code))`.

---

## 1. Domain model (`backend/internal/core/domain`)

### 1.1 Files and ownership

| File | Content | Written by |
|---|---|---|
| `ids.go` | `GameID`, `PlayerID` | WP1 |
| `card.go` | `Card`, `CardKind`, `Command`, `NumberCard` | WP1 |
| `deck.go` | `NewDeck` (variant factory) | WP1 |
| `ruleset.go` | `Variant`, `Options`, `Ruleset`, `NewRuleset`, helpers | WP1 |
| `pile.go` | `Direction`, `Pile`, `NewPiles` | WP1 |
| `rng.go` | `RNG` interface | WP1 |
| `errors.go` | sentinel errors | WP1 |
| `events.go` | domain events | WP1 |
| `game.go` | `Game`, `Player`, `Phase`, `Turn`, `Play`, `Result`, `Outcome`, `Move`; `Clone`, `PullEvents`, `Score`, small accessors, `Summary` | WP1 |
| `view.go` | view types (WP1); `ViewFor` body (WP3) | WP1 + WP3 |
| `lifecycle.go` | `NewGame`, `AddPlayer`, `SetOptions`, `Start`, `ChooseFirstPlayer` | WP1 stubs, WP3 bodies |
| `play.go` | `PlayCard`, `EndTurn`, `LegalMoves`, `CanEndTurn`, `RequiredMinimum`, turn advance | WP1 stubs, WP3 bodies |
| `placement.go` | composable placement check | WP3 |
| `stuck.go` | stuck check: no legal move and cannot end turn (see §11.1) | WP3 |
| `*_test.go` | table-driven tests | WP3 (WP1 tests only for its helpers) |

WP1 stubs are `panic("not implemented: WP3")`.

### 1.2 Types (exact)

Fields are **exported** so the sqlite DTO mapper and tests can read and build state. By convention, code outside `domain` must **not mutate** a `Game` except through its methods (tests excepted). No struct tags.

```go
type GameID string
type PlayerID string

// ---- cards ----
type CardKind string
const CardKindNumber CardKind = "number" // Extreme/mini-expansions add kinds later (joker, tens, ...)

type Command string
const CommandNone Command = ""
// Phase 2 (DO NOT define now): stop, skull, exactlyThree, noCommunication, noBackwardsTrick, singlePile, drawOnlyOne.

type Card struct {
    Value   int      // 2..99; identity of a card in classic/extreme (values are unique)
    Kind    CardKind
    Command Command  // CommandNone in Phase 1
}
func NumberCard(v int) Card            // Card{Value: v, Kind: CardKindNumber}

// ---- variant / config ----
type Variant string
const (
    VariantClassic Variant = "classic"
    VariantExtreme Variant = "extreme" // NewRuleset/Validate reject it in Phase 1 (ErrInvalidOptions)
)

type Options struct {
    Variant         Variant
    Pro             bool // min 3 cards per turn
    ReducedHandSize bool // hand size -1; only allowed together with Pro (rules.md); else ErrInvalidOptions
}
func DefaultOptions() Options          // {VariantClassic, false, false}
func (o Options) Validate() error      // ErrInvalidOptions (wrapped with reason)

type Ruleset struct {
    Variant                Variant
    MinPlayers, MaxPlayers int          // 1, 5
    MinCardsPerTurn        int          // 2, or 3 if Pro
    MinCardsDrawPileEmpty  int          // 1 (also for Pro; rules.md interpretation)
    HandSizes              map[int]int  // base table: {1:8, 2:7, 3:6, 4:6, 5:6}
    HandSizeReduction      int          // 1 if Pro && ReducedHandSize, else 0
    BackwardsTrickStep     int          // 10
    AscendingStart         int          // 1
    DescendingStart        int          // 100
    LowestCard, HighestCard int         // 2, 99
    AscendingPiles, DescendingPiles int // 2, 2
}
func NewRuleset(o Options) (Ruleset, error)       // validates o
func (r Ruleset) HandSize(players int) int        // HandSizes[players] - HandSizeReduction
func (r Ruleset) MinCardsToPlay(drawPileEmpty bool) int
```
No rule code may hard-code 2/3/6/7/8/10/1/100/99. Read the values from `Ruleset`.

```go
// ---- deck ----
// NewDeck builds the unshuffled deck for the ruleset's variant (variant-specific factory).
// classic: NumberCard(LowestCard..HighestCard) ascending. extreme: ErrInvalidOptions (Phase 2).
func NewDeck(r Ruleset) ([]Card, error)

// ---- piles ----
type Direction string
const (
    DirectionUp   Direction = "up"   // ascending
    DirectionDown Direction = "down" // descending
)
type Pile struct {
    Direction  Direction
    StartValue int    // 1 or 100 (the start card; never in hand)
    Cards      []Card // played cards, last = top
}
func (p Pile) TopValue() int          // last card's value, or StartValue if empty
func (p Pile) TopCard() (Card, bool)  // false if only the start card is showing
func NewPiles(r Ruleset) []Pile       // index 0,1 = up (start 1); index 2,3 = down (start 100)

// ---- RNG (injected; *math/rand/v2.Rand satisfies it) ----
type RNG interface{ Shuffle(n int, swap func(i, j int)) }

// ---- game ----
type Phase string
const (
    PhaseLobby               Phase = "lobby"
    PhaseChoosingFirstPlayer Phase = "choosingFirstPlayer"
    PhasePlaying             Phase = "playing"
    PhaseFinished            Phase = "finished"
)
type Outcome string
const (
    OutcomeWon  Outcome = "won"  // score 0
    OutcomeLost Outcome = "lost" // active player stuck; score > 0
)

type Player struct {
    ID   PlayerID
    Name string // trimmed, 1..20 runes, no control chars, unique case-insensitively within the game
    Hand []Card // kept sorted ascending by Value
}
type Play struct {
    Card Card
    Pile int
}
type Turn struct {
    Number        int    // 1-based; 0 before the first turn
    ActiveSeat    int    // index into Game.Players
    Plays         []Play // cards played this turn, in order ("piles touched" = distinct Play.Pile)
    StartHandSize int    // hand size at turn start (Extreme 3! hook)
}
func (t Turn) PilesTouched() []int

type Result struct {
    Outcome             Outcome
    Score               int // CardsLeftInHands + CardsLeftInDrawPile
    CardsLeftInHands    int
    CardsLeftInDrawPile int
    StuckPlayerID       PlayerID // "" when won
}
type Move struct{ Card, Pile int } // card value, pile index

type Game struct {
    ID       GameID
    Options  Options
    Rules    Ruleset  // re-derived on SetOptions in lobby; frozen from Start on
    Phase    Phase
    HostID   PlayerID
    Players  []Player // seat order = join order = clockwise
    DrawPile []Card   // last element = top
    Piles    []Pile   // len 4, see NewPiles
    Turn     Turn     // meaningful in playing; kept as-is in finished (shows who was stuck)
    Result   *Result  // non-nil iff Phase == finished
    Revision int64    // starts at 1 in NewGame; +1 exactly once per successful mutating method
    events   []Event  // pending, not persisted
}
```

### 1.3 Methods (exact signatures)

```go
// lifecycle.go
func NewGame(id GameID, hostID PlayerID, hostName string, opts Options) (*Game, error) // ErrInvalidName, ErrInvalidOptions; host = seat 0; Phase lobby; Piles=NewPiles; Revision=1; no events
func (g *Game) AddPlayer(id PlayerID, name string) error   // ErrWrongPhase (not lobby), ErrGameFull, ErrInvalidName, ErrNameTaken, ErrDuplicatePlayer; emits PlayerJoined
func (g *Game) SetOptions(by PlayerID, opts Options) error // ErrUnknownPlayer, ErrNotHost, ErrWrongPhase, ErrInvalidOptions; emits OptionsChanged
func (g *Game) Start(by PlayerID, rng RNG) error           // ErrUnknownPlayer, ErrNotHost, ErrWrongPhase; see 1.5; emits GameStarted (+ solo: FirstPlayerChosen, TurnStarted)
func (g *Game) ChooseFirstPlayer(by, first PlayerID) error // ErrUnknownPlayer (by or first), ErrNotHost, ErrWrongPhase; emits FirstPlayerChosen, TurnStarted

// play.go
func (g *Game) PlayCard(by PlayerID, value int, pile int) error // ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrCardNotInHand, ErrInvalidPile, ErrIllegalPlacement
func (g *Game) EndTurn(by PlayerID) error                       // ErrUnknownPlayer, ErrWrongPhase, ErrNotYourTurn, ErrMinimumNotReached
func (g *Game) LegalMoves(by PlayerID) []Move                   // nil unless playing && by is active
func (g *Game) CanEndTurn() error                               // nil if the active player may end the turn now (runs all end-turn checks)
func (g *Game) RequiredMinimum() int                            // Rules.MinCardsToPlay(len(DrawPile)==0)
func (g *Game) ActivePlayerStuck() bool                         // §11.1: len(LegalMoves(active))==0 && CanEndTurn()!=nil

// game.go (WP1)
func (g *Game) Clone() *Game              // deep copy incl. pending events; preserves nil vs empty slices
func (g *Game) PullEvents() []Event       // returns and clears pending events
func (g *Game) Score() (hands, drawPile int)
func (g *Game) PlayerByID(id PlayerID) (Player, int, bool) // player, seat, ok
func (g *Game) IsHost(id PlayerID) bool
func (g *Game) CommunicationRestricted() bool // Phase 1: always false (Extreme hook: visible "No communication")
func (g *Game) Summary() GameSummary

// view.go
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) // ErrUnknownPlayer
```

### 1.4 Composable placement rule (`placement.go`)

```go
type PlacementKind int
const (
    PlacementNormal PlacementKind = iota + 1 // up: card > top; down: card < top
    PlacementBackwardsTrick                  // up: card == top-Step; down: card == top+Step
)
type Placement struct {
    Card      Card
    PileIndex int
    Pile      Pile // current state of the target pile
    Kind      PlacementKind
}
// PlacementConstraint can veto a placement the base rule allowed.
// Phase 2 adds constraints derived from visible Infinity commands (no backwards trick, single pile)
// and from lightning commands played this turn (3! count, STOP).
type PlacementConstraint interface {
    Check(g *Game, p Placement) error // nil = allowed; else wrap ErrIllegalPlacement (or a future specific error)
}
func basePlacement(r Ruleset, pile Pile, c Card) (PlacementKind, bool)
func (g *Game) placementConstraints() []PlacementConstraint // Phase 1: return nil
func (g *Game) checkPlacement(c Card, pileIdx int) (Placement, error) // base rule, then every constraint
```

End-of-turn hooks (in `play.go`):
```go
type EndTurnCheck interface{ Check(g *Game) error }
func (g *Game) endTurnChecks() []EndTurnCheck // Phase 1: []EndTurnCheck{minimumCheck{}}  (ErrMinimumNotReached)
func (g *Game) drawCount() int                // draw policy. Phase 1 "refill": max(0, HandSize(len(Players)) - len(hand)), capped at len(DrawPile)
```
Use **refill to hand size**, not "number of cards played". The two are equal in the base game, but Extreme's "Draw only one … once covered, refills to full hand size" needs refill semantics.

### 1.5 Behaviour (WP3 must implement exactly)

- **Start:**
  1. Freeze `Rules = NewRuleset(Options)`.
  2. Build the deck with `NewDeck` and shuffle it with `rng.Shuffle`. That shuffled deck becomes `DrawPile`.
  3. Deal `HandSize(n)` rounds. In each round, for each seat in order, pop the top card (last element) into that hand. Keep hands sorted.
  4. Set `Phase = choosingFirstPlayer`.
  5. **Solo (1 player):** skip the choice. Immediately do `ChooseFirstPlayer` semantics for the only player (the only possible choice, not a random default).
- **ChooseFirstPlayer:** `Phase = playing`, then `beginTurn(seat of first)`.
- **beginTurn(seat):** `Turn = {Number+1, seat, nil, len(hand)}`, emit `TurnStarted`, then `checkStuck()`.
- **PlayCard:**
  1. Validate: phase playing, `by` is active, card is in hand, pile index is in `[0,len(Piles))`, then `checkPlacement`.
  2. Apply: remove the card from the hand, append it to `Piles[i].Cards` and to `Turn.Plays`. Emit `CardPlayed`.
  3. Afterwards, in this order:
     1. If no cards are left anywhere (all hands and the draw pile are empty): if `CanEndTurn()==nil`, `finish(won)`; otherwise `finish(lost)`. The second branch is the Extreme "skull on last card" hook.
     2. Else, if the actor's hand is empty and `CanEndTurn()==nil`: **auto end turn** (same as `EndTurn`, event `TurnEnded{Auto:true}`).
     3. Else `checkStuck()`.
- **EndTurn:**
  1. Validate, then run `CanEndTurn()`.
  2. Draw `drawCount()` cards from the top into the hand (sorted). Emit `TurnEnded{CardsDrawn}`.
  3. `advanceTurn()`.
- **advanceTurn:** walk seats clockwise from `ActiveSeat+1`, wrapping and including the current seat last (for solo). Pick the first seat with a non-empty hand. Emit `PlayerSkipped` for each seat passed over because its hand is empty (this only happens once the draw pile is empty). If no seat has cards, `finish(won)`. This is defensive: the PlayCard step 3.1 normally catches it.
- **checkStuck:** if `ActivePlayerStuck()`, `finish(lost)` with `StuckPlayerID = active`. **Superseded by §11.1: there is no lookahead, so ignore the DFS bullet below.**
- ~~ActivePlayerCanFinishTurn (stuck search, brute force)~~, REMOVED by §11.1:
  1. Clone the game **once**.
  2. Run a DFS. If `CanEndTurn()==nil`, return true. Otherwise, for each hand card × pile where `checkPlacement` is OK: apply in place (remove from hand, push to pile, append play), recurse, then **undo**.
  3. Return false if nothing works.
  - Optional pruning: skip a pile whose direction and top equal an earlier pile's.
  - Phase 1 depth ≤ 3. It runs at every turn start and after every play. So the game ends **as soon as the minimum is provably unreachable**, including before any card of the turn is played. This gives the physical-game score.
- **finish(outcome):** `Phase = finished`, `Result` filled from `Score()`, emit `GameFinished`.
- **Required minimum** is `RequiredMinimum()`. The draw pile does not change mid-turn, so it is constant during a turn.
- **Atomicity:** every exported mutator validates first and mutates only on success. On success it calls `g.commit()`, which does `Revision++` exactly once, even when it chains auto-end, advance and finish.

### 1.6 Errors (`errors.go`), all `errors.New` sentinels, wrapped with detail via `fmt.Errorf("%w: ...")`

`ErrInvalidName`, `ErrNameTaken`, `ErrGameFull`, `ErrDuplicatePlayer`, `ErrInvalidOptions`, `ErrWrongPhase`, `ErrNotHost`, `ErrUnknownPlayer`, `ErrNotYourTurn`, `ErrCardNotInHand`, `ErrInvalidPile`, `ErrIllegalPlacement`, `ErrMinimumNotReached`.

Phase 2 (Extreme, see [extreme-plan.md](extreme-plan.md) §2.5) adds `ErrBackwardsTrickBlocked`, `ErrSinglePileViolation`, `ErrExactCountExceeded`, `ErrExactCountUnreachable`, `ErrStopTooEarly`, `ErrSkullUncovered`, `ErrExactCountRequired`, `ErrSkullUncoverable`. `ErrTurnOver` is not needed: STOP ends the turn automatically, so no card can follow it.

### 1.7 Events (`events.go`)

```go
type Event interface{ isEvent() }
type PlayerJoined      struct{ PlayerID PlayerID; Name string }
type OptionsChanged    struct{ Options Options }
type GameStarted       struct{}
type FirstPlayerChosen struct{ PlayerID PlayerID }
type TurnStarted       struct{ PlayerID PlayerID; Number int }
type CardPlayed        struct{ PlayerID PlayerID; Card Card; Pile int }
type TurnEnded         struct{ PlayerID PlayerID; CardsDrawn int; Auto bool }
type PlayerSkipped     struct{ PlayerID PlayerID }
type GameFinished      struct{ Outcome Outcome; Score int; StuckPlayerID PlayerID }
```
Events carry only public information. Never put hand contents in an event.

### 1.8 View types (`view.go`), the projection per player

```go
type PlayerView struct {
    GameID                  GameID
    Revision                int64
    Phase                   Phase
    You                     PlayerID
    HostID                  PlayerID
    Options                 Options
    Rules                   RulesView
    Players                 []PlayerSummary // seat order
    Hand                    []HandCard      // ONLY the viewer's hand, sorted ascending; empty in lobby
    Piles                   []PileView
    DrawPileCount           int
    Turn                    *TurnView       // nil in lobby / choosingFirstPlayer
    CommunicationRestricted bool
    Result                  *ResultView     // nil unless finished
}
type RulesView struct {
    MinCardsPerTurn, MinCardsDrawPileEmpty int
    HandSize                               int // for the current player count (preview in lobby)
    MinPlayers, MaxPlayers                 int
}
type PlayerSummary struct {
    ID        PlayerID
    Name      string
    Seat      int
    HandCount int
    IsHost, IsYou, IsActive bool
}
type HandCard struct {
    Card       Card
    LegalPiles []int // pile indices this card may legally go on now; empty unless viewer is active in playing
}
type PileView struct {
    Index      int
    Direction  Direction
    StartValue int
    TopValue   int
    TopCard    *Card // nil when only the start card shows
    CardCount  int   // played cards on the pile
}
type TurnView struct {
    Number          int
    ActivePlayerID  PlayerID
    CardsPlayed     int
    RequiredMinimum int
    CanEndTurn      bool // CanEndTurn()==nil (public info)
    Plays           []Play
}
type ResultView struct {
    Outcome             Outcome
    Score               int
    Excellent           bool // Score < 10
    CardsLeftInHands    int
    CardsLeftInDrawPile int
    StuckPlayerID       PlayerID
}
type GameSummary struct { // public, pre-join
    GameID      GameID
    Phase       Phase
    HostName    string
    PlayerCount int
    MaxPlayers  int
    Joinable    bool // Phase == lobby && PlayerCount < MaxPlayers
}
```
**Recommendation (decided):** the server sends `LegalPiles` per hand card, and the client never re-implements rules. This also keeps Phase 2 (command-dependent legality) free on the client.

### 1.9 Domain tests (WP3), table-driven, one test per rules.md item

| Area | Cases to cover |
|---|---|
| Placement | up: greater ok, equal/less rejected; down mirror; backwards trick exactly ±10 on both directions; ±9/±11 rejected; trick on the start card (1, 100) impossible |
| Rules | hand sizes for 1–5 players × {classic, pro, pro+reduced}; reduced without Pro is invalid; extreme variant rejected; minimums 2 / 3 (Pro) / 1 when the draw pile is empty (Pro too) |
| Start | deck has 98 unique cards 2–99; hands dealt correctly; draw-pile count = 98 − n·hand; deterministic with a fixed RNG; solo auto-starts in playing |
| Lobby | host-only checks; phase errors; full at 5; name validation and case-insensitive duplicates; join refused after start |
| Turn flow | not your turn; card not in hand; invalid pile; end turn below minimum; refill draws exactly to hand size; draw capped when the draw pile runs short; next seat clockwise; solo stays on the same player |
| Draw pile empty | minimum becomes 1; players with empty hands skipped (with `PlayerSkipped` events); auto end turn when the hand empties |
| Win | the last card played gives won, score 0 |
| Stuck (§11.1) | lost at turn start only when NO hand card is playable; NOT lost while a single legal card exists even if the minimum is unreachable: the player keeps playing and is lost once no card fits and the minimum is unmet (score counts cards left then); sequence enabled by backwards trick works; not stuck once the minimum is met (can end turn) |
| Atomicity | for every error case, `reflect.DeepEqual(before, g)` where `before := g.Clone()` |
| ViewFor | contains only the viewer's hand; other players show counts only; `LegalPiles` only for the active viewer; Revision increments once per command |

---

## 2. Ports (`backend/internal/core/ports`)

```go
// errors.go
var (
    ErrGameNotFound    = errors.New("game not found")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrVersionConflict = errors.New("version conflict")
    ErrDuplicateGameID = errors.New("duplicate game id")
)

// repository.go (driven)
type GameRecord struct {
    Game        *domain.Game
    TokenHashes map[domain.PlayerID]string // hex(sha256(token)); never the raw token
    Version     int64                      // OCC version as loaded
}
type GameRepository interface {
    // Create inserts a new game with Version=1 (sets rec.Version=1). ErrDuplicateGameID if the id exists.
    Create(ctx context.Context, rec *GameRecord) error
    // Load returns a fresh, independent *domain.Game each call. ErrGameNotFound.
    Load(ctx context.Context, id domain.GameID) (*GameRecord, error)
    // Update persists rec if the stored version == rec.Version, then sets rec.Version++.
    // Inserts token hashes not stored yet. ErrVersionConflict / ErrGameNotFound.
    Update(ctx context.Context, rec *GameRecord) error
}

// publisher.go (driven)
type GameUpdate struct {
    Game   *domain.Game   // read-only; must not be retained after Publish returns
    Events []domain.Event
}
type EventPublisher interface {
    // Publish must not block on network I/O: compute per-player payloads synchronously, enqueue, return.
    Publish(ctx context.Context, u GameUpdate)
}
type NopPublisher struct{} // Publish does nothing (tests, CLI)

// service.go (driving)
type Session struct {
    GameID   domain.GameID
    PlayerID domain.PlayerID
    Token    string // raw token, returned only here
    View     domain.PlayerView
}
type GameService interface {
    CreateGame(ctx context.Context, hostName string, opts domain.Options) (Session, error)
    JoinGame(ctx context.Context, id domain.GameID, name string) (Session, error)
    Authenticate(ctx context.Context, id domain.GameID, token string) (domain.PlayerID, error) // ErrUnauthorized / ErrGameNotFound
    Summary(ctx context.Context, id domain.GameID) (domain.GameSummary, error)
    View(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
    SetOptions(ctx context.Context, id domain.GameID, player domain.PlayerID, opts domain.Options) (domain.PlayerView, error)
    StartGame(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
    ChooseFirstPlayer(ctx context.Context, id domain.GameID, player, first domain.PlayerID) (domain.PlayerView, error)
    PlayCard(ctx context.Context, id domain.GameID, player domain.PlayerID, card, pile int) (domain.PlayerView, error)
    EndTurn(ctx context.Context, id domain.GameID, player domain.PlayerID) (domain.PlayerView, error)
}
```

All mutating calls return the **actor's** fresh view. They also publish to every connected client.

---

## 3. Service layer (`backend/internal/core/service`)

```go
type Config struct {
    NewGameID   func() domain.GameID   // default: 5 chars from the alphabet in §0, crypto/rand
    NewPlayerID func() domain.PlayerID // default: "p_" + 10 lowercase base32 chars, crypto/rand
    NewToken    func() string          // default: 32 random bytes, base64.RawURLEncoding
    NewRNG      func() domain.RNG      // default: rand.New(rand.NewChaCha8(seed from crypto/rand))
    Logger      *slog.Logger
}
func New(repo ports.GameRepository, pub ports.EventPublisher, cfg Config) *Service // nil fields get defaults
var _ ports.GameService = (*Service)(nil)
```

**Concurrency (decided): a per-game mutex in the service, plus OCC in the repo as a safety net.**
- `locks.go` holds a keyed mutex: `map[GameID]*entry{mu sync.Mutex; refs int}` guarded by a `sync.Mutex`, removed when `refs` drops to 0.
- Every mutation runs through one helper:
  ```
  mutate(ctx, id, fn func(*ports.GameRecord) error) (*ports.GameRecord, error):
    lock(id); defer unlock(id)
    for attempt in 0..2:
      rec := repo.Load            // ErrGameNotFound passes through
      if err := fn(rec); err != nil → return err   (domain error, nothing saved)
      events := rec.Game.PullEvents()
      err := repo.Update(rec); if ErrVersionConflict → continue
      pub.Publish(ctx, GameUpdate{rec.Game, events})   // still under the lock ⇒ per-game publish order = commit order
      return rec
    return ports.ErrVersionConflict
  ```
- **CreateGame:** generate the host ID and token, then up to 5 attempts of `NewGameID` → `domain.NewGame` → `repo.Create`, retrying on `ErrDuplicateGameID`. Domain validation errors return immediately. No publish.
- **JoinGame:** `mutate` with `g.AddPlayer(newID, name)` and `rec.TokenHashes[newID] = hash(token)`.
- **Authenticate:** Load, hash the token, compare with `subtle.ConstantTimeCompare` against each entry, return the matching `PlayerID` or `ErrUnauthorized`.
- **View / Summary:** Load, then `ViewFor` / `Summary`.
- **StartGame:** `g.Start(player, cfg.NewRNG())`.
- **Tests** use the memory repo (WP4) plus a recording publisher:
  - create, join, start, choose, play and end turn with an RNG that does not shuffle
  - token auth success and failure
  - ID collision retry
  - events are published once per command, in order
  - domain errors are not saved and not published
  - concurrent `PlayCard` from two goroutines is serialised (run with `-race` if the toolchain allows)

---

## 4. REST API (`backend/internal/adapters/httpapi`)

Auth: `Authorization: Bearer <token>`. A missing or invalid token gives 401 `unauthorized`. `{code}` is case-insensitive. Bodies are capped at 4 KiB (`http.MaxBytesReader`). Responses are JSON.

| Method & path | Auth | Request body | 2xx response |
|---|---|---|---|
| `GET /api/health` | – | – | existing |
| `POST /api/games` | – | `CreateGameRequest` | 201 `SessionResponse` |
| `GET /api/games/{code}` | – | – | 200 `GameSummary` |
| `POST /api/games/{code}/players` | – | `JoinGameRequest` | 201 `SessionResponse` |
| `GET /api/games/{code}/view` | yes | – | 200 `PlayerView` |
| `PUT /api/games/{code}/options` | yes (host) | `GameOptions` | 200 `PlayerView` |
| `POST /api/games/{code}/start` | yes (host) | – | 200 `PlayerView` |
| `POST /api/games/{code}/first-player` | yes (host) | `{ "playerId": string }` | 200 `PlayerView` |
| `POST /api/games/{code}/plays` | yes | `{ "card": number, "pile": number }` | 200 `PlayerView` |
| `POST /api/games/{code}/end-turn` | yes | – | 200 `PlayerView` |
| `POST /api/games/{code}/end-game` | yes | – | 200 `PlayerView` (host only; outcome `aborted`) |
| `/api/` (anything else) | – | – | 404 `{error:{code:"not_found"}}`. Register `mux.HandleFunc("/api/", ...)` so the SPA fallback never swallows API typos. |

Error body for every non-2xx response, on REST and on WS: `{"error":{"code":"<code>","message":"<human text>"}}`.

**Error mapping**, in `apidto.ErrorFor(err) (status int, ErrorBody)`, used by both REST and WS:

| Error | code | HTTP |
|---|---|---|
| malformed JSON or missing field | `bad_request` | 400 |
| `ports.ErrUnauthorized` | `unauthorized` | 401 |
| `domain.ErrNotHost` | `not_host` | 403 |
| `ports.ErrGameNotFound` | `game_not_found` | 404 |
| `domain.ErrWrongPhase` | `wrong_phase` | 409 |
| `domain.ErrGameFull` | `game_full` | 409 |
| `domain.ErrNameTaken` | `name_taken` | 409 |
| `domain.ErrNotYourTurn` | `not_your_turn` | 409 |
| `ports.ErrVersionConflict` | `conflict` | 409 |
| `domain.ErrInvalidName` | `invalid_name` | 422 |
| `domain.ErrInvalidOptions` | `invalid_options` | 422 |
| `domain.ErrUnknownPlayer` | `unknown_player` | 422 |
| `domain.ErrCardNotInHand` | `card_not_in_hand` | 422 |
| `domain.ErrInvalidPile` | `invalid_pile` | 422 |
| `domain.ErrIllegalPlacement` | `illegal_placement` | 422 |
| `domain.ErrMinimumNotReached` | `minimum_not_reached` | 422 |
| anything else, incl. `ErrDuplicatePlayer` | `internal` | 500 (message is `"internal error"`; log the real error) |

`Deps` gains `Games ports.GameService` and `WS http.Handler`. The router registers `GET /ws/games/{code}` → `d.WS` when it is non-nil.

---

## 5. WebSocket protocol (`backend/internal/adapters/ws`, library `github.com/coder/websocket`)

**Connection**
- URL: `GET /ws/games/{code}?token=<token>`. Browsers cannot set headers on a WebSocket, so the token goes in the query. The request logger logs only `URL.Path`, never the query.
- The server always runs `Accept` first, then authenticates through `svc.Authenticate`. On failure it closes with **4401** (`unauthorized`) or **4404** (`game_not_found`) so the browser can see the reason.
- Origin: default same-host check. `AcceptOptions.OriginPatterns` comes from optional env `WS_ORIGIN_PATTERNS` (comma-separated). The Vite proxy preserves Host, so dev works with the default.
- `SetReadLimit(4096)`. The server pings every 25 s (`c.Ping` with a 10 s timeout). If a ping fails, the connection is closed.

**Server → client messages**
```jsonc
{ "type": "view", "view": PlayerView, "events": GameEvent[] }   // sent on connect (events: []) and after every committed command
{ "type": "presence", "online": ["p_…", ...] }                  // on connect (to the new conn) and to all on any connect/disconnect change
{ "type": "ack", "requestId": "7" }
{ "type": "error", "requestId": "7", "error": { "code": "illegal_placement", "message": "…" } } // requestId omitted for malformed messages
```

**Client → server messages.** Every message carries `requestId: string`, which the client generates (a counter).
```jsonc
{ "type": "setOptions",        "requestId": "1", "options": { "variant": "classic", "pro": true, "reducedHandSize": false } }
{ "type": "startGame",         "requestId": "2" }
{ "type": "chooseFirstPlayer", "requestId": "3", "playerId": "p_…" }
{ "type": "playCard",          "requestId": "4", "card": 42, "pile": 0 }
{ "type": "endTurn",           "requestId": "5" }
{ "type": "endGame",           "requestId": "6" }
```
An unknown `type` gets `error` with `bad_request`.

**Hub design**
- `ws.NewHub(logger) *Hub` implements `ports.EventPublisher`.
- `ws.NewHandler(hub *Hub, svc ports.GameService, opts Options) http.Handler`.
- Rooms: `map[GameID]map[*conn]struct{}`, guarded by a mutex.
- Each conn has `playerID` and `send chan []byte` (capacity 32), plus a writer goroutine (write timeout 5 s).
- `Publish`:
  1. Snapshot the room's conns.
  2. For each distinct playerID, build `u.Game.ViewFor(pid)`, map it with `apidto`, and marshal once.
  3. Enqueue non-blockingly. If a conn's buffer is full, close it with 1008 "slow consumer"; the client reconnects and gets a fresh view.
- Connect sequence:
  1. Authenticate.
  2. Register in the room.
  3. `svc.View`, then enqueue `view` (with `events:[]`).
  4. Enqueue `presence` to the new conn and broadcast `presence` to the room.
- Read loop: decode, call the matching `svc` method, enqueue `ack` or `error` to this conn. The service publishes before returning, so the broadcast `view` is enqueued **before** the `ack`.
- On read error or close: unregister and broadcast presence. Multiple tabs per player are allowed. A player is online if they have at least one conn.
- Ordering caveat: the initial view can race with a concurrent publish. The client must drop any `view` whose `revision` is lower than the one it holds.

**Client reconnection**
- Backoff 500 ms × 2ⁿ, max 10 s, ±20 % jitter.
- Reconnect immediately on `visibilitychange` to visible and on the `online` event.
- On 4401: clear the stored session and go to `/join/{code}`. On 4404: show "game not found". Any other close: reconnect.
- On every (re)connect the server sends the full view, so the client has no delta logic.

**Decision: always push full per-player views.** They are small (< 3 KB) and avoid all sync bugs. `events` is extra, for toasts and animations only.

---

## 6. Wire contract: DTOs (Go `internal/adapters/apidto`) and TS (`frontend/src/lib/types.ts`)

All JSON is camelCase. **Slices are never `null`**: the mappers always `make([]T, 0)`.

### 6.1 Go DTOs (`apidto/dto.go`)
```go
type Card struct {
    Value   int    `json:"value"`
    Kind    string `json:"kind"`              // "number"
    Command string `json:"command,omitempty"` // Phase 2
}
type HandCard struct {
    Card
    LegalPiles []int `json:"legalPiles"`
}
type GameOptions struct {
    Variant         string `json:"variant"` // "classic"; empty on input means classic
    Pro             bool   `json:"pro"`
    ReducedHandSize bool   `json:"reducedHandSize"`
}
type Rules struct {
    MinCardsPerTurn       int `json:"minCardsPerTurn"`
    MinCardsDrawPileEmpty int `json:"minCardsDrawPileEmpty"`
    HandSize              int `json:"handSize"`
    MinPlayers            int `json:"minPlayers"`
    MaxPlayers            int `json:"maxPlayers"`
}
type PlayerSummary struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Seat      int    `json:"seat"`
    HandCount int    `json:"handCount"`
    IsHost    bool   `json:"isHost"`
    IsYou     bool   `json:"isYou"`
    IsActive  bool   `json:"isActive"`
}
type Pile struct {
    Index      int    `json:"index"`
    Direction  string `json:"direction"` // "up" | "down"
    StartValue int    `json:"startValue"`
    TopValue   int    `json:"topValue"`
    TopCard    *Card  `json:"topCard"` // null when only the start card shows
    CardCount  int    `json:"cardCount"`
}
type Play struct {
    Card Card `json:"card"`
    Pile int  `json:"pile"`
}
type Turn struct {
    Number          int    `json:"number"`
    ActivePlayerID  string `json:"activePlayerId"`
    CardsPlayed     int    `json:"cardsPlayed"`
    RequiredMinimum int    `json:"requiredMinimum"`
    CanEndTurn      bool   `json:"canEndTurn"`
    Plays           []Play `json:"plays"`
}
type Result struct {
    Outcome             string `json:"outcome"` // "won" | "lost" | "aborted"
    Score               int    `json:"score"`
    Excellent           bool   `json:"excellent"`
    CardsLeftInHands    int    `json:"cardsLeftInHands"`
    CardsLeftInDrawPile int    `json:"cardsLeftInDrawPile"`
    StuckPlayerID       string `json:"stuckPlayerId,omitempty"`
}
type PlayerView struct {
    GameID                  string          `json:"gameId"`
    Revision                int64           `json:"revision"`
    Phase                   string          `json:"phase"`
    You                     string          `json:"you"`
    HostID                  string          `json:"hostId"`
    Options                 GameOptions     `json:"options"`
    Rules                   Rules           `json:"rules"`
    Players                 []PlayerSummary `json:"players"`
    Hand                    []HandCard      `json:"hand"`
    Piles                   []Pile          `json:"piles"`
    DrawPileCount           int             `json:"drawPileCount"`
    Turn                    *Turn           `json:"turn"`
    CommunicationRestricted bool            `json:"communicationRestricted"`
    Result                  *Result         `json:"result"`
}
type GameSummary struct {
    GameID      string `json:"gameId"`
    Phase       string `json:"phase"`
    HostName    string `json:"hostName"`
    PlayerCount int    `json:"playerCount"`
    MaxPlayers  int    `json:"maxPlayers"`
    Joinable    bool   `json:"joinable"`
}
type Event struct { // flattened union; Type is one of the event names below
    Type          string `json:"type"`
    PlayerID      string `json:"playerId,omitempty"`
    Name          string `json:"name,omitempty"`
    Card          *Card  `json:"card,omitempty"`
    Pile          *int   `json:"pile,omitempty"`
    Number        *int   `json:"number,omitempty"`
    CardsDrawn    *int   `json:"cardsDrawn,omitempty"`
    Auto          bool   `json:"auto,omitempty"`
    Outcome       string `json:"outcome,omitempty"`
    Score         *int   `json:"score,omitempty"`
    StuckPlayerID string `json:"stuckPlayerId,omitempty"`
}
// requests / responses
type CreateGameRequest struct {
    Name    string       `json:"name"`
    Options *GameOptions `json:"options,omitempty"`
}
type JoinGameRequest struct {
    Name string `json:"name"`
}
type ChooseFirstPlayerRequest struct {
    PlayerID string `json:"playerId"`
}
type PlayCardRequest struct {
    Card *int `json:"card"`
    Pile *int `json:"pile"`
} // nil means bad_request
type SessionResponse struct {
    GameID   string     `json:"gameId"`
    PlayerID string     `json:"playerId"`
    Token    string     `json:"token"`
    View     PlayerView `json:"view"`
}
type ErrorBody struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
type ErrorResponse struct {
    Error ErrorBody `json:"error"`
}
// ws.go
type ServerMessage struct {
    Type      string      `json:"type"` // view|presence|ack|error
    RequestID string      `json:"requestId,omitempty"`
    View      *PlayerView `json:"view,omitempty"`
    Events    []Event     `json:"events,omitempty"` // non-nil for "view" (may be empty: send [] — use a separate struct or custom marshal so view always has "events")
    Online    []string    `json:"online,omitempty"`
    Error     *ErrorBody  `json:"error,omitempty"`
}
type ClientMessage struct {
    Type      string       `json:"type"`
    RequestID string       `json:"requestId"`
    Options   *GameOptions `json:"options,omitempty"`
    PlayerID  string       `json:"playerId,omitempty"`
    Card      *int         `json:"card,omitempty"`
    Pile      *int         `json:"pile,omitempty"`
}
// map.go
func FromView(v domain.PlayerView) PlayerView
func FromSummary(s domain.GameSummary) GameSummary
func FromEvents(evs []domain.Event) []Event // never nil
func (o GameOptions) ToDomain() domain.Options
func ErrorFor(err error) (int, ErrorBody)
```
Additive helpers (WP1, use them instead of hand-rolling): `FromCard`, `FromOptions`, `FromEvent`; `Code*` error-code constants, `Msg*` / `Cmd*` / `Event*` type-string constants; `ErrBadRequest` sentinel + `BadRequest(format, args...)` (maps to `bad_request`; `ErrorFor` also maps `*json.SyntaxError`, `*json.UnmarshalTypeError`, `*http.MaxBytesError` to it); `ViewMessage(v, events)`, `PresenceMessage(online)`, `AckMessage(id)`, `ErrorMessage(id, body)`. `ServerMessage.MarshalJSON` always emits `"events"` for `view` and `"online"` for `presence` (as `[]` when empty). In the domain, `(*Game).emit(Event)` queues an event and `(*Game).commit()` does `Revision++`.

Event `type` strings: `playerJoined`, `optionsChanged`, `gameStarted`, `firstPlayerChosen`, `turnStarted`, `cardPlayed`, `turnEnded`, `playerSkipped`, `gameFinished`.

### 6.2 TypeScript (`frontend/src/lib/types.ts`), mirror exactly
```ts
export type Phase = 'lobby' | 'choosingFirstPlayer' | 'playing' | 'finished';
export type Direction = 'up' | 'down';
export type Outcome = 'won' | 'lost' | 'aborted';
export interface Card { value: number; kind: 'number'; command?: string }
export interface HandCard extends Card { legalPiles: number[] }
export interface GameOptions { variant: 'classic'; pro: boolean; reducedHandSize: boolean }
export interface Rules { minCardsPerTurn: number; minCardsDrawPileEmpty: number; handSize: number; minPlayers: number; maxPlayers: number }
export interface PlayerSummary { id: string; name: string; seat: number; handCount: number; isHost: boolean; isYou: boolean; isActive: boolean }
export interface Pile { index: number; direction: Direction; startValue: number; topValue: number; topCard: Card | null; cardCount: number }
export interface Play { card: Card; pile: number }
export interface Turn { number: number; activePlayerId: string; cardsPlayed: number; requiredMinimum: number; canEndTurn: boolean; plays: Play[] }
export interface Result { outcome: Outcome; score: number; excellent: boolean; cardsLeftInHands: number; cardsLeftInDrawPile: number; stuckPlayerId?: string }
export interface PlayerView {
	gameId: string; revision: number; phase: Phase; you: string; hostId: string;
	options: GameOptions; rules: Rules; players: PlayerSummary[]; hand: HandCard[]; piles: Pile[];
	drawPileCount: number; turn: Turn | null; communicationRestricted: boolean; result: Result | null;
}
export interface GameSummary { gameId: string; phase: Phase; hostName: string; playerCount: number; maxPlayers: number; joinable: boolean }
export interface SessionResponse { gameId: string; playerId: string; token: string; view: PlayerView }
export type GameEvent =
	| { type: 'playerJoined'; playerId: string; name: string }
	| { type: 'optionsChanged' }
	| { type: 'gameStarted' }
	| { type: 'firstPlayerChosen'; playerId: string }
	| { type: 'turnStarted'; playerId: string; number: number }
	| { type: 'cardPlayed'; playerId: string; card: Card; pile: number }
	| { type: 'turnEnded'; playerId: string; cardsDrawn: number; auto?: boolean }
	| { type: 'playerSkipped'; playerId: string }
	| { type: 'gameFinished'; outcome: Outcome; score: number; stuckPlayerId?: string };
export type ApiErrorCode =
	| 'bad_request' | 'unauthorized' | 'not_host' | 'game_not_found' | 'not_found' | 'wrong_phase' | 'game_full'
	| 'name_taken' | 'not_your_turn' | 'conflict' | 'invalid_name' | 'invalid_options' | 'unknown_player'
	| 'card_not_in_hand' | 'invalid_pile' | 'illegal_placement' | 'minimum_not_reached' | 'internal'
	| 'network' | 'disconnected' | 'timeout'; // last three are client-side only
export interface ErrorBody { code: ApiErrorCode; message: string }
export type ServerMessage =
	| { type: 'view'; view: PlayerView; events: GameEvent[] }
	| { type: 'presence'; online: string[] }
	| { type: 'ack'; requestId: string }
	| { type: 'error'; requestId?: string; error: ErrorBody };
export type ClientCommand =
	| { type: 'setOptions'; options: GameOptions }
	| { type: 'startGame' }
	| { type: 'chooseFirstPlayer'; playerId: string }
	| { type: 'playCard'; card: number; pile: number }
	| { type: 'endTurn' };
// sent as ClientCommand & { requestId: string }
```

---

## 7. Persistence (`backend/internal/adapters/sqlite` and `backend/internal/adapters/memory`)

**Migration `0002_player_tokens.sql`.** A separate table keeps secrets out of the game JSON snapshot and enables future lookups by token. Game ID = join code, so no `join_code` column is needed.
```sql
CREATE TABLE player_tokens (
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id  TEXT NOT NULL,
    token_hash TEXT NOT NULL,          -- NOT unique: rematch games copy the token hashes (§11.2)
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, player_id)
);
```

**`games` columns**
- `variant` = `Options.Variant`.
- `status`: lobby → `lobby`; choosingFirstPlayer or playing → `running`; finished → the outcome (`won` / `lost`).
- `state` = JSON of `gameDTO`.

**`sqlite/dto.go`** holds private tagged structs mirroring every domain field:
- `gameDTO{SchemaVersion int "schemaVersion" (=1), ID, Options, Rules, Phase, HostID, Players, DrawPile, Piles, Turn, Result, Revision}`, plus `cardDTO` and the others.
- `toDTO(*domain.Game) gameDTO` and `fromDTO(gameDTO) (*domain.Game, error)`. `fromDTO` rejects an unknown `SchemaVersion`.
- Pending events are not persisted.

**`sqlite/games.go`**
- `func NewGameRepository(db *DB) *GameRepository`, which implements `ports.GameRepository`.
- Create: tx `INSERT games(..., version=1)` + `INSERT player_tokens`. A primary-key constraint error (check the `modernc.org/sqlite` error code, or fall back to `strings.Contains(err.Error(), "UNIQUE constraint failed")`) gives `ErrDuplicateGameID`.
- Update: tx `UPDATE games SET state=?, status=?, variant=?, version=version+1, updated_at=CURRENT_TIMESTAMP WHERE id=? AND version=?`. If 0 rows are affected, check whether the game exists and return `ErrGameNotFound` or `ErrVersionConflict`. Then `INSERT OR IGNORE INTO player_tokens`. Commit, then `rec.Version++`.
- Load: two queries.

**`memory/games.go`**
- A `sync.Mutex` map. It stores `Clone()`s and returns `Clone()`s (fresh objects each time) and implements the same version semantics.
- Used by service tests, transport tests and WP9.

**Tests**
- Round trip for a game in each phase (lobby, choosing, playing mid-turn, finished). Build domain structs directly. Compare `json.Marshal(toDTO(orig))` with the same for the loaded game, so nil and empty slices compare equal.
- Version conflict, not found, duplicate ID, tokens persisted and loaded, migrations idempotent.
- Run the same contract test suite against memory and sqlite. Put a shared `repotest` helper in `backend/internal/adapters/repotest/repotest.go`, exporting `func Run(t *testing.T, newRepo func(t *testing.T) ports.GameRepository)`.

---

## 8. Frontend

### 8.1 Routes
| Route | Purpose |
|---|---|
| `/` | Landing page. Nickname field, prefilled from `localStorage['the-game:nickname']`. **Host a game** button: `POST /api/games`, save the session, go to `/g/{code}`. **Join a game**: 5-char code input (uppercased, filtered to the alphabet) and Join: `POST .../players`, save, go to `/g/{code}`. Optional: "Continue" links for stored sessions. |
| `/join/[code]` | Share-link target. If a session exists for this code, replace-navigate to `/g/{code}`. Otherwise `GET /api/games/{code}`: 404 shows "Game not found" and a home link. `!joinable` shows "This game has already started" (phase ≠ lobby) or "Game is full". Otherwise show "Join {hostName}'s game (n/5)", the nickname field and Join. |
| `/g/[code]` | The game. No session: go to `/join/{code}`. Otherwise create a `GameConnection` (connect on mount, close on destroy). Render by `view.phase`: `Lobby`, `ChooseFirstPlayer`, or `Board` (playing and finished, with `GameOver` overlaid when finished). |

The share link is `${location.origin}/join/${code}`. The Go SPA fallback already serves `index.html` for these paths.

### 8.2 Modules
- `src/lib/api.ts`: `class ApiError extends Error { status; code: ApiErrorCode }`. Functions: `createGame(name, options?)`, `joinGame(code, name)`, `getSummary(code)`, `getView(code, token)`, plus REST action wrappers. Network failure gives `code: 'network'`. Keep `getHealth`.
- `src/lib/session.ts`:
  - `createSessionStore(storage: Pick<Storage,'getItem'|'setItem'|'removeItem'>)` returns `{ get(code), save(code, {token, playerId, name}), clear(code), list(), getNickname(), setNickname() }`.
  - Keys: `the-game:session:{CODE}` and `the-game:nickname`.
  - Default export `sessions` uses `globalThis.localStorage`. The test uses an in-memory storage, because the node test environment has no localStorage.
- `src/lib/game/protocol.ts` (pure, unit-tested): `wsUrl(loc: {protocol, host}, code, token)`, `parseServerMessage(data: string): ServerMessage | null`, `nextBackoff(attempt, rand=Math.random)`, `shouldAcceptView(current: PlayerView|null, next: PlayerView)` (next.revision ≥ current.revision).
- `src/lib/game/connection.svelte.ts`:
  ```ts
  export type ConnectionStatus = 'connecting' | 'open' | 'reconnecting' | 'closed';
  export class ActionError extends Error { code: ApiErrorCode }
  export class GameConnection {
    view = $state<PlayerView | null>(null);
    online = $state<string[]>([]);
    status = $state<ConnectionStatus>('connecting');
    fatal = $state<'unauthorized' | 'game_not_found' | null>(null);
    lastEvents = $state<GameEvent[]>([]);
    constructor(code: string, token: string);
    connect(): void;  close(): void;
    setOptions(o: GameOptions): Promise<void>;  startGame(): Promise<void>;
    chooseFirstPlayer(playerId: string): Promise<void>;
    playCard(card: number, pile: number): Promise<void>;  endTurn(): Promise<void>;
    get me(): PlayerSummary | undefined;  get isHost(): boolean;  get isMyTurn(): boolean;
  }
  ```
  - `requestId` comes from an incrementing counter (no `crypto.randomUUID`).
  - Pending requests resolve on `ack` and reject on `error` with `ActionError`. They time out after 10 s (`timeout`) and are all rejected with `disconnected` on close.
  - Sending while not open rejects with `disconnected`.
  - Listens to `visibilitychange` and `online`, and removes the listeners in `close()`.
- `src/lib/errors.ts`: `errorText(code: ApiErrorCode): string`, short English messages, e.g. `illegal_placement` → "That card can't go on this pile."
- `src/lib/game/format.ts`: `resultHeadline(result)`:
  - score 0: "You won! All 98 cards played."
  - score < 10: "Excellent! Only N cards left."
  - otherwise: "Game over: N cards left."
  - Also `stuckLine(view)`: "{name} had no playable card." (§11.1)

### 8.3 Components (`src/lib/components/game/`)
- `ShareInvite.svelte`:
  - Large monospace code. A read-only input with the link.
  - "Copy link" uses `navigator.clipboard.writeText` when available. Otherwise select the input and show "long-press to copy".
  - "Share" is shown only if `navigator.share` exists.
- `PlayerStrip.svelte`: horizontal chips with name, `handCount`, an online dot (from `conn.online`), a crown icon for the host, a highlighted ring for the active player, and "(you)".
- `Lobby.svelte`:
  - `ShareInvite`, a live player list, and the options.
  - Options are a `Switch` for "Pro (min 3 cards per turn)" and a `Switch` for "Reduced hand size (−1)". The second is disabled unless Pro is on, and turning Pro off also sends `reducedHandSize:false`. Non-hosts see the switches disabled.
  - Text line: "Hand size {rules.handSize} · min {rules.minCardsPerTurn} cards per turn".
  - Host sees "Start game ({n} player(s))". Others see "Waiting for {host} to start…".
- `PileCard.svelte`: direction icon (lucide `ArrowUp`/`ArrowDown`), start label ("1" / "100"), big `topValue`, `cardCount`. Props: `highlight` (legal target for the selected card) and `dimmed`. It is a button and emits a click.
- `Hand.svelte`: the sorted hand. Props: `cards`, `selected`, `interactive`, and an `onselect` callback.
  - Layout: CSS grid `grid-template-columns: repeat(N, minmax(0,1fr))`, card `aspect-[2/3] max-w-14`, `text-lg font-bold tabular-nums`, `gap-1`. Eight cards fit in 360 px (about 40 px each).
  - When interactive, cards with empty `legalPiles` are dimmed but still tappable. Tapping one shows the toast "No legal pile for this card".
- `TurnBar.svelte`:
  - On your turn: "Your turn · {cardsPlayed}/{requiredMinimum} played" and the **End turn** button (enabled iff `turn.canEndTurn` and not pending).
  - Otherwise: "{name}'s turn · {cardsPlayed} played".
- `ChooseFirstPlayer.svelte`:
  - Shows the read-only Board (piles and own hand).
  - Host gets the banner "Look at your cards, discuss, then choose who starts" and one button per player. Tapping one opens an `AlertDialog` "{name} starts?", then `chooseFirstPlayer`.
  - Others see "Waiting for {host} to choose who starts".
- `Board.svelte`: header (code, connection indicator, "Deck {drawPileCount}"), `PlayerStrip`, 2×2 pile grid (up piles top row, down piles bottom row), `TurnBar`, `Hand`.
  - **Interaction:**
    - Local `selected = $state<number|null>(null)`.
    - Tapping a hand card toggles selection. While selected, the piles in its `legalPiles` are highlighted and the others dimmed.
    - Tapping a highlighted pile calls `conn.playCard(selected, pile)`, then clears the selection. On error, show a toast with `errorText`.
    - Clear the selection whenever `view.revision` changes and the card is no longer in hand, or it is no longer your turn.
    - No undo.
    - Controls are disabled when `conn.status !== 'open'`.
- `GameOver.svelte`: an overlay card over the board with `resultHeadline`, the stuck line, the breakdown (hands / draw pile), and "Back to home".
- `ConnectionBanner.svelte`: thin top banner when reconnecting.

### 8.4 Portrait phone layout (target width 360 px; `h-svh` flex column, the board takes the remaining height)
```
+------------------------------------+
| K7MQ2   (o) online      Deck 64    |  header
| [*Anna 6] [Ben 6] [You 6] [Cleo 6] |  player strip (scroll-x if needed)
| +-------------+  +-------------+   |
| |  ^      1   |  |  ^      1   |   |  ascending piles
| |     34      |  |     12      |   |
| +-------------+  +-------------+   |
| +-------------+  +-------------+   |
| |  v    100   |  |  v    100   |   |  descending piles
| |     71      |  |     88      |   |
| +-------------+  +-------------+   |
| Your turn - 1/2 played  [End turn] |  turn bar
+------------------------------------+
| [7][15][23][41][56][62][80][93]    |  hand (sticky bottom, safe-area padding)
+------------------------------------+
```
- Desktop: the same column, centred, `max-w-md`.
- In `app.html`, add `viewport-fit=cover` to the viewport meta. Use `pb-[env(safe-area-inset-bottom)]` on the hand and `touch-action: manipulation` on tappables.
- Change the root layout `<main>` to `mx-auto flex min-h-svh w-full max-w-5xl flex-col px-3 py-3 sm:p-6`.

### 8.5 shadcn-svelte components to add (WP2)
`npx shadcn-svelte@latest add switch separator alert-dialog skeleton -y` (run `npx svelte-kit sync` first if needed). The repo already has badge, button, card, input, label and sonner.

### 8.6 Frontend tests (vitest, node env, `src/**/*.test.ts`)
- `session.test.ts`
- `protocol.test.ts`: backoff bounds and jitter, stale-view rejection, URL building for http and https, message parsing of garbage
- `format.test.ts`

---

## 9. Work breakdown

Waves: **W1** WP1 ∥ WP2. **W2** WP3 ∥ WP4 ∥ WP6 ∥ WP7. **W3** WP5 ∥ WP8. **W4** WP9.

### WP1: Backend contracts (W1, **strong model / medium effort**)
- **Goal:** compile-ready skeleton of §1–§2 and §6.1.
- **Files:**
  - `domain/*` per §1.1 (types, constants, errors, events, view types; implement `NewRuleset`, `Options.Validate`, `HandSize`, `MinCardsToPlay`, `NewDeck`, `NewPiles`, `Pile` helpers, `Turn.PilesTouched`, `Clone`, `PullEvents`, `Score`, `PlayerByID`, `IsHost`, `CommunicationRestricted`, `Summary`; every other method is a stub)
  - `ports/{errors,repository,publisher,service}.go`
  - `adapters/apidto/{dto,ws,map,errors}.go` with full mappers and `ErrorFor`
- **Acceptance:**
  - `go build ./... && go vet ./...`
  - unit tests for the implemented helpers: ruleset table, deck, Clone independence
  - `apidto` tests: empty slices marshal as `[]`, mapping of every field, `ErrorFor` table incl. wrapped errors
- **Depends on:** nothing.

### WP2: Frontend contracts and UI kit (W1, **Haiku or Sonnet / low**)
- **Files:** `src/lib/types.ts` (verbatim §6.2), `src/lib/errors.ts`, the shadcn adds (§8.5), the root layout `<main>` class change, the `app.html` viewport.
- **Acceptance:** `npm run check && npm run lint && npm test`.
- **Depends on:** nothing.

### WP3: Domain rules engine (W2, **strong model / high effort**)
- **Goal:** implement §1.3–§1.5 bodies, `placement.go`, `stuck.go`, `ViewFor`, and the full §1.9 test suite.
- **Files:** `domain/{lifecycle,play,placement,stuck,view}.go` and `*_test.go`. Do not change exported signatures.
- **Acceptance:** `go test ./internal/core/domain/... -count=1` passes with high coverage of the rules (aim ≥ 90 % statements: `go test -cover`); `go vet`.
- **Also:** draft the Interpretation entries for `docs/rules.md` from §10 (items 1–4) and list them in the handback. Do not commit them until the owner confirms.
- **Depends on:** WP1.

### WP4: Persistence (W2, **Sonnet / low–medium**)
- **Files:**
  - `sqlite/migrations/0002_player_tokens.sql`, `sqlite/dto.go`, `sqlite/games.go`, `sqlite/games_test.go`
  - `adapters/memory/games.go` and its test
  - `adapters/repotest/repotest.go`
- **Acceptance:** §7 tests pass for both implementations; `go vet`.
- **Depends on:** WP1. It constructs domain structs directly and needs no WP3 behaviour.

### WP6: Transport adapters (W2, **Sonnet / medium-high**, or a strong model / medium, because of the WS concurrency)
- **Files:**
  - `httpapi/router.go` (extend `Deps`, `/api/` 404, `/ws` mount), `httpapi/games.go`, `httpapi/auth.go`
  - `ws/{hub,conn,handler}.go`
  - `go get github.com/coder/websocket@latest`
- **Must not touch:** `main.go`.
- **Tests** use a hand-written fake `ports.GameService` in `_test.go` (canned views and errors).
  - REST: status codes and bodies for every route, bearer auth, the error table, body limit.
  - WS (`httptest.NewServer` plus `websocket.Dial`): initial `view` then `presence`; command produces `ack`; domain error produces `error` with the right code; bad token closes with 4401, unknown game with 4404; `Publish` reaches two players with different views; presence updates on disconnect; slow consumer does not block `Publish`.
- **Acceptance:** `go vet ./... && go test ./internal/adapters/...`.
- **Depends on:** WP1.

### WP7: Frontend core (W2, **Sonnet / medium**)
- **Files:** `src/lib/api.ts`, `src/lib/session.ts`, `src/lib/game/protocol.ts`, `src/lib/game/connection.svelte.ts`, `src/routes/+page.svelte` (landing; replace the health card), `src/routes/join/[code]/+page.svelte`, and tests §8.6 (`session`, `protocol`).
- **Acceptance:** `npm run check && npm run lint && npm test`. Manual check with the backend once WP5 lands is optional.
- **Depends on:** WP2.

### WP5: Service and wiring (W3, **Sonnet / medium**)
- **Files:**
  - `service/{service,locks,ids,tokens}.go` and `service_test.go` (§3)
  - `cmd/server/main.go` wiring:
    1. `repo := sqlite.NewGameRepository(db)`
    2. `hub := ws.NewHub(logger)`
    3. `svc := service.New(repo, hub, service.Config{Logger: logger})`
    4. `httpapi.Deps{Games: svc, WS: ws.NewHandler(hub, svc, ws.Options{OriginPatterns: ...})}`
    5. env `WS_ORIGIN_PATTERNS`
- **Acceptance:** service tests (with `-race` if the toolchain allows); `go vet ./... && go test ./...`; `go run ./cmd/server` starts, and the create/join/view `curl` flow works.
- **Depends on:** WP3, WP4 (memory repo), WP6 (constructors).

### WP8: Frontend game screen (W3, **strong model / medium**, or **Sonnet / high**, because the UX matters)
- **Files:**
  - `src/routes/g/[code]/+page.svelte`
  - `src/lib/components/game/*.svelte` (§8.3)
  - `src/lib/game/format.ts` and its test
  - optional `src/lib/game/fixtures.ts` with sample `PlayerView`s for each phase, for manual layout work
- **Acceptance:**
  - `npm run check && npm run lint && npm test`
  - layout verified at 360×740 and on desktop: 8 cards with no horizontal scroll, piles prominent, End turn reachable with a thumb
  - every phase renders from fixtures
- **Depends on:** WP2, WP7.

### WP9: Integration, E2E and polish (W4, **strong model / medium**)
- **Go end-to-end test** `backend/internal/e2e/e2e_test.go`:
  - real sqlite (temp file), service with a deterministic RNG, `httpapi` and `ws` on `httptest.NewServer`
  - flow: create, join (2 players), both WS connected, start, choose first player, legal plays taken from `legalPiles`, end turn, other client sees the update, restart persistence (reopen the DB, rejoin by token)
- **Manual run:** `npm run dev -- --host` against `go run ./cmd/server`, with one phone browser and one desktop browser. Play a solo game to the end and a 2-player Pro game.
- Fix contract drift. Update the `CLAUDE.md` commands and env list (`WS_ORIGIN_PATTERNS`).
- **Acceptance:** all backend and frontend checks green; the manual checklist (create, share link, join, lobby live, options, start, choose first, play, reload mid-turn restores the seat, game over screen) passes.
- **Depends on:** everything.

---

## 10. Decisions taken, plus questions for the owner

Default decisions. These are implemented unless the owner overrides them, and are to be recorded in `docs/rules.md` → *Interpretations* once confirmed.

1. ~~**Stuck detection is a lookahead.**~~ **The owner decided otherwise, see §11.1.** At every turn start and after every play, the engine checks by brute force whether the active player can still reach a legal end of turn. If not, the game is lost immediately, before any further card is played. This follows rules.md "ends immediately when the active player cannot play the required minimum" and CLAUDE.md's search requirement. It refines the literal "no legal action left" wording in *Enforcement*. **Please confirm.**
2. **Solo games skip the "choose first player" step.** There is only one choice, so it is not a random default.
3. **The turn ends automatically when the active player's hand becomes empty** and the minimum is met. Nothing else is possible at that point.
4. **"Reduced hand size" is only available together with Pro**, per rules.md. The UI disables it otherwise.
5. **Game ID = join code.** 5 chars give 32⁵ ≈ 33.5 M codes. A code alone only allows joining a lobby. Seeing any game state needs a token.
6. **Not in Phase 1:**
   - leave or kick in the lobby
   - turn timeouts and abandoned-game cleanup. If the host disconnects during first-player choice, the game waits until they return.
   - revealing other players' hands at game end
   - "play again with the same group"

   Owner: say if any of these are wanted.
7. The token travels in the WS query string. This is acceptable for this app, but avoid logging full URLs at any reverse proxy.

---

### Critical Files for Implementation
- C:\Users\marti\github\The-Game\backend\internal\core\domain\ (new: game.go, play.go, placement.go, stuck.go, view.go)
- C:\Users\marti\github\The-Game\backend\internal\core\ports\ (new: repository.go, publisher.go, service.go, errors.go)
- C:\Users\marti\github\The-Game\backend\internal\adapters\apidto\dto.go (new; the wire contract) and C:\Users\marti\github\The-Game\frontend\src\lib\types.ts (new; its TS mirror)
- C:\Users\marti\github\The-Game\backend\internal\adapters\httpapi\router.go and C:\Users\marti\github\The-Game\backend\internal\adapters\ws\ (hub.go, handler.go)
- C:\Users\marti\github\The-Game\backend\internal\adapters\sqlite\sqlite.go plus migrations\0002_player_tokens.sql, and C:\Users\marti\github\The-Game\frontend\src\lib\game\connection.svelte.ts


---

## 11. Amendments: owner decisions (2026-10-10). These SUPERSEDE earlier sections where they conflict.

### 11.1 Stuck rule: no lookahead
The game is lost **only when the active player has no legal card play AND cannot end the turn**. There is no brute-force search for "can the minimum still be reached". A player who can play one more card but can never reach the minimum keeps playing until nothing fits. Then the game is lost, and the score counts the cards left at that moment.
- `ActivePlayerStuck() bool` = `len(LegalMoves(activeID)) == 0 && CanEndTurn() != nil`. It replaces `ActivePlayerCanFinishTurn` (rename the WP1 stub).
- `checkStuck()` runs at `beginTurn` and after every successful play that does not auto-end the turn, exactly as before.
- `LegalMoves` still enumerates hand × piles through `checkPlacement`, so the Phase 2 constraints stay composable.

### 11.2 Play again (rematch), same group
- **Domain:**
  - `Game` gets `NextGameID GameID`. It is empty until a rematch is created, and it is persisted. `Clone` copies it.
  - New method `func (g *Game) Rematch(by PlayerID, newID GameID) (*Game, error)`:
    - Errors: `ErrUnknownPlayer`, `ErrWrongPhase` (not finished), and the new sentinel `ErrRematchExists` (NextGameID is already set).
    - **Any** player may call it.
    - On success: set `g.NextGameID = newID`, emit `RematchCreated{GameID: newID}`, `commit()`, and return the new game.
    - The new game: `Phase=lobby`, same `Options` (Rules re-derived), the same players in the same seat order with the **same PlayerIDs and names** and empty hands, `HostID = by` (the caller becomes host), `Piles=NewPiles`, `Revision=1`, no pending events.
- **Ports:** add `Rematch(ctx, id GameID, player PlayerID) (domain.PlayerView, error)` to `GameService`. It returns the actor's view of the **old** game, which now carries `NextGameID`.
- **Service:**
  1. Under the old game's lock, load the old record.
  2. If `NextGameID` is already set, return the actor's view of the old game (idempotent, no error).
  3. Otherwise, up to 5 tries: `NewGameID`, then `old.Rematch` on a fresh clone, then `repo.Create(new record with TokenHashes copied from old)`. Retry on `ErrDuplicateGameID`.
  4. Then `repo.Update(old)` and publish the old game's update.
  - Players keep their raw tokens, which is valid because the token hashes are copied.
- **REST:** `POST /api/games/{code}/rematch` (auth) → 200 `PlayerView`. **WS:** client message `{ "type": "rematch", "requestId": "…" }`.
- **Error map:** `domain.ErrRematchExists` → `rematch_exists` / 409. The service swallows it, but map it anyway. Add `'rematch_exists'` to the TS `ApiErrorCode` and to `errorText`.
- **DTO / TS:**
  - `PlayerView.nextGameId`: Go `NextGameID string json:"nextGameId,omitempty"`, TS `nextGameId?: string`.
  - Event `rematchCreated` with `gameId`: Go Event DTO gains `GameID string json:"gameId,omitempty"`. TS: `| { type: 'rematchCreated'; gameId: string }`.
  - `ClientCommand` gains `| { type: 'rematch' }`.
- **Frontend:**
  - `session.copy(fromCode, toCode)` copies `{token, playerId, name}`.
  - `GameConnection.rematch()`, and an `api.ts` wrapper.
  - `GameOver` shows "Play again" when `nextGameId` is absent: it calls `rematch()`, then copies the session and navigates.
  - When `nextGameId` is present it shows "Join next game": copy the session to the new code, then `goto('/g/{next}')`.
  - Never auto-navigate. Players may be looking at the revealed hands.

### 11.3 Reveal hands at game end
- Domain `PlayerSummary` gets `RevealedHand []Card`. It is filled (sorted) for **every** player only when `Phase == finished`, and nil otherwise.
- DTO: `RevealedHand []Card json:"revealedHand,omitempty"`. TS: `revealedHand?: Card[]`. `undefined` means "not revealed" or "empty".
- `GameOver` lists each player's remaining cards.
- ViewFor test: hands are never revealed before the game is finished.

### 11.4 Abandoned-game cleanup
- `ports.GameRepository` gains `DeleteInactive(ctx context.Context, olderThan time.Time) (int, error)`. It deletes games whose `updated_at < olderThan` (tokens cascade) and returns the count.
  - sqlite: `updated_at` is stored as `CURRENT_TIMESTAMP` (UTC, `YYYY-MM-DD HH:MM:SS`), so compare against `olderThan.UTC().Format("2006-01-02 15:04:05")`.
  - Add `CREATE INDEX games_updated_at ON games(updated_at);` in migration 0002.
  - memory repo: track `updatedAt` per game (set on Create/Update via an injectable `now func() time.Time`, default `time.Now`).
- Add it to the repotest contract suite.
- Service: `func (s *Service) PurgeInactive(ctx context.Context, maxAge time.Duration) (int, error)`. This is not part of `GameService`; it is a maintenance entry point. Log the count when it is > 0.
- `main.go`:
  - Env `GAME_RETENTION` (Go duration, default `168h`) and `CLEANUP_INTERVAL` (default `1h`).
  - A goroutine runs `PurgeInactive` once at startup and then on a ticker, and stops on ctx cancel.
  - Document both env vars in CLAUDE.md.
- A WS client of a deleted game gets 4404 on reconnect. The client shows "game not found" and clears the stored session.
