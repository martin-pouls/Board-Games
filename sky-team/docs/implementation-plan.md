# Sky Team: implementation plan

Read before working on a package: root `CLAUDE.md`, `sky-team/docs/rules.md` (§ numbers below refer to it),
`shared/README.md`, and the files each package names. Don't re-read the rulebook. A rule question
rules.md doesn't answer goes back to the orchestrator, not into the code.

## 1. Scope

- **Version 1:** exactly 2 players (Pilot, Co-Pilot). All 21 flight-log scenarios and both Düsseldorf promo faces (23 scenarios on 23 track faces). All modules except real-time: kerosene, kerosene leak, intern, wind, ice brakes. All six special abilities. Host picks scenario and abilities, and can swap roles.
- **Not in the app at all:** the real-time module (§12.4). Its four scenarios are played without it.
- **Later, not in v1:** scenario progress (the flight log's check boxes, §7.2); seam in §5.8.

## 2. Placement and identity

- **Placement:** new app. Sky Team is the repo's first dice game, with its own board, rules engine and look; nothing in it is an edition or a variant of an existing game.

| | |
|---|---|
| Directory / URL | `sky-team/` / `/sky-team/` |
| Go module | `github.com/martin-pouls/sky-team/backend` |
| DB file | `data/sky-team.db` |
| Ports | Vite `:5180`, backend `:8086` |
| Container / volume | `sky-team` / `data` (network `web`) |
| Start-page tile | yes (`players: '2 players'`) |
| Game retention | default `168h` |
| UI language | English |

- **Template sources:** all plumbing from Fox Duet via `scaffold.sh` (also 2 players, so the lobby, seats, rematch and lobby actions carry over unchanged). The domain follows Fox Duet's patterns: injected RNG, a step machine with derived actors, automatic resolutions inside the mutator, `Clone` + `reflect.DeepEqual` tests. From The Crew: the RNG with `IntN`, the data tables checked against independent expectation tables, `rules_*_test.go` black-box files, lobby options sent with `POST /api/games`, fixture presets.

## 3. Decisions

User decisions (2026-10-22):
- New app `sky-team/`, title "Sky Team", ports 5180/8086, scaffolded from Fox Duet.
- v1: all 21 scenarios + DUS yellow and black, all modules, all six abilities. No real-time (§12.4). No progress tracking.
- Host picks the scenario and the abilities and can swap roles; both players see the abilities (§4.1, §13).
- Layout: concept A "whole cockpit" (`sky-team/docs/layout-concepts.html`), worst cases TGU red, KEF black and HND red fit at 375×630. 44 px is a guideline. Theme: `sky-team/docs/mockup.html` (made in parallel).
- Look (2026-10-23, mockup round 3, checkpoint B passed): start-page tile B "final approach"; no edge marks on the panel; the axis as a dial (positions 30° apart, crosses at ±90°, the plane's fin as the needle); speed as three ranges with the markers as dividers ("≤ 5 stay | 6–10 +1 | 11+ +2"; round 7: land / crash); green turn masks on the white strip; one icon per ability chip, dashed with a check when used. The UX is final; the UI gets a polish pass at checkpoint C (WP20).
- Dice live in the game (no `shared/` changes).

App decisions in rules.md: host picks scenario (default `yul-green`) and abilities (§4.1, §13); seat 1 is Pilot, swappable (§4.1); server rolls, own unplaced dice only, partner count (§5.1); Ready ends discussion, pickup and traffic shown before it (§5.1); strictly alternating, one intent per placement (§5.2); no legal placement → remaining dice discarded (§5.2); reroll step with per-player confirmation (§5.3); coffee as a net modifier (§5.4); anything covers the mandatory spaces (§5.5); green flaps take dice (§6.5); only the next brake is playable (§6.6); C and D end the game early (§7.1); loss named, dice revealed, give up, rematch keeps scenario, abilities and roles (§7.2); all failed conditions listed, D at the second engine die (§7.1); partial ability selection in the lobby (§13); no progress tracking (§7.2); no chat (§8); real-time dropped, DUS included (§9); traffic rolled by the server (§11.1); intern training is one intent, the token counts as a die except for concentration and coffee (§12.2); leak applied at the second engine die (§12.5); Control/Mastery count everything on the space incl. intern tokens, Synchronisation rules (intern tokens trigger it; its die takes no coffee and no intern space), Working Together die public, side intents only in `place`, Adaptation and Working Together outside the turn order (§13); automatic resolution, no undo, no timers (§15).

- **Lobby view** carries the scenario catalogue (no REST route), so fixtures are self-contained.

## 4. Shared code

Nothing moves. Dice, the die component and the RNG helpers are game-specific until a second dice game exists; then `Die.svelte` and a dice-rolling helper are candidates. The scaffold copies Fox Duet's remaining plumbing (service `mutate/act`, ws dispatch skeleton, sqlite codec) a fifth time; name it in TODO.md, don't move it.

## 5. Domain design (`backend/internal/core/domain`)

### 5.1 Files

| File | Content |
|---|---|
| `doc.go`, `ids.go`, `errors.go`, `events.go`, `rng.go` | As in Fox Duet. `RNG` = The Crew's (`Shuffle`, `IntN`). |
| `role.go` | `Role` (`pilot`, `copilot`; index 0/1), `Other()`, `PieceKind` (`pilot`, `copilot`, `intern`, `traffic`), `Colour` (`blue`, `orange`, `both`). |
| `panel.go` | `SpaceID`, `SpaceKind`, the `spaces` table (§3.1 plus module spaces), `activeSpaces(modules)`, `fits(space, value)`. |
| `scenario.go` | Types and lookups: `ScenarioID`, `ScenarioColour`, `Module`, `ScenarioDef`, `TrackID`, `TrackDef`, `TrackSpace`, `TurnMask`; `ScenarioFor`, `Scenarios()`, `TrackFor`; the altitude table (§4.2), traffic-die faces (§2), `WindValue(offset)` (§12.3), `AbilityID` and `Abilities()` (§13). |
| `scenarios_data.go` | The 23 `ScenarioDef` and 23 `TrackDef` rows (§9, §10). **Data only.** WP3 writes YUL only; WP4 the rest. |
| `game.go` | Aggregate `Game`, `Player`, `Actors()`, `Clone`, `PullEvents`, `PlayerByID`, `RoleOf(seat)`. |
| `flight.go` | `Flight` (state across rounds) and derived values: aero thresholds, brake limit, plane supply, rerolls on the track, Mastery availability. |
| `round.go` | `Round`, steps, `Ready`, the roll, `advance()` (turn passing, Synchronisation trigger, discards, end of round), round start (pickup, traffic) and end (§5.5). |
| `place.go` | `legalSpaces(role, piece, value)`, `Place` incl. intern training, the effects of every space kind (§6), approach movement (§6.2, §11.2). |
| `modules.go` | Kerosene, leak, wind turn, ice-brake deployment, intern token supply (§12). |
| `abilities.go` | Adaptation, Anticipation, Working Together (`OfferSwap`, `AnswerSwap`), Synchronisation (`PlaceSync`), Control and Mastery hooks (§13). |
| `reroll.go` | `SpendReroll`, `ConfirmReroll` (§5.3). |
| `result.go` | `LossReason`, `Condition`, `Result`, final evaluation (§7). |
| `lifecycle.go` | Fox Duet's lobby code plus `Configure`, `SwapRoles`, `Start`, `GiveUp`, `EndGame`, `Rematch`, `Summary`. |
| `view.go` | `PlayerView` and `ViewFor` (§5.5). |
| `export_test.go` | Test builders for `domain_test` (§9). |

### 5.2 Aggregate and phases

Phases: `lobby → playing → won | lost`, `lobby → ended`. No in-between phase; rounds are inside `playing`.

```go
type Game struct {
    ID; Phase; HostID; Players []Player   // seat order, max 2
    PilotSeat int                          // 0 or 1; SwapRoles flips it (lobby)
    Scenario  ScenarioID                   // default "yul-green"
    Abilities []AbilityID                  // chosen cards, ≤ scenario's count in the lobby
    Flight    *Flight                      // nil in the lobby
    Result    *Result
    NextGameID; Revision; CreatedAt; events
}
type Flight struct {                       // persists across rounds
    Round      int                         // 1..7 = altitude index (§4.2)
    Position   int                         // approach space, 1..N
    Planes     []int                       // tokens per approach space
    Axis       int                         // -3..3
    Gear       [3]bool                     // any order (§6.4)
    Flaps, Brakes int                      // deployed count, in order (§6.5, §6.6)
    Ice        int                         // ice marker slot 0..4 (4 = past 5)
    Coffee, Rerolls int                    // panel coffee 0..3, reroll supply
    Fuel       int                         // kerosene, if the module is in play
    Wind       int                         // ring offset from the top space
    Intern     []int                       // token values in row order; InternLeft/InternRight = next indices
    InternLeft, InternRight int
    AdaptationUsed [2]bool                 // by role
    R          Round
}
type Round struct {
    Step       Step                        // discussion | place | reroll | swap | sync
    Turn       Role
    Ready      [2]bool                     // discussion
    Dice       [2][4]int                   // by role and slot; 0 = not in hand (placed, discarded, not rolled)
    Discarded  [2]int
    Placements []Placement                 // in order: {Space, Piece, Value, Coffee, Role, Slot}
    Speed      *Speed                      // set at the second engine piece: {Sum, Wind, Move}
    RerollConfirmed [2]bool
    Swap       *SwapOffer                  // {By Role, Slot int}
    SyncValue  int                         // traffic die waiting for the Co-Pilot (step sync)
    AnticipationUsed, TogetherUsed, SyncUsed bool
    Start      RoundStart                  // {RerollPicked bool, Traffic []TrafficRoll{Roll, Space}}
}
```

**Derived, never stored:** occupancy of spaces (from `Placements`), aero thresholds (blue `4 + gear`, orange `8 + flaps`), brake limit, plane supply (`12 − Σ Planes`), rerolls still on the altitude track, the first player (odd round → Pilot), legal spaces, ice pairs, actors, Control/Mastery "triggered this round", wind value. Hand slots are stable (slot `i` stays empty after its die is placed), matching the concept's dice tray.

**Space effects resolve by space, not by piece owner:** the axis compares the `axisPilot` piece with the `axisCopilot` piece, whatever lies there (die, intern token, traffic die).

### 5.3 Steps and actors

| Step | Actor | Ends with |
|---|---|---|
| `discussion` | each role not yet `Ready` | both ready → roll all 8 dice → `place` |
| `place` | `Turn` | a placement → `advance()` |
| `reroll` | each role that has not confirmed | both confirmed → `place` |
| `swap` | the partner of `Swap.By` | `AnswerSwap` → `place` |
| `sync` | Co-Pilot | `PlaceSync` → `advance()` |

**Side intents in `place`** (not steps, no actor change): `SpendReroll` (either role, supply > 0) → `reroll`; `Adapt` (either role, ability in play, once per game per role); `OfferSwap` (either role, once per round, both have unplaced dice) → `swap`; `Anticipate` (first player of the round, on their turn, no placement yet this round, once per round).

**Resolution order inside `Place`** (also `PlaceSync`, and the token half of training):
1. Validate everything (§5.6). 2. Spend `|coffee|` tokens, take the die out of the hand, append the placement (training: the die on the intern space, then the token on its target; the token is taken from the player's end of the row). 3. Resolve the space:
   - axis, second piece: move toward the higher side; `spin` at ±3; round 7 and axis ≠ 0 → `landing` (C); then wind turn (§12.3); then Control.
   - engines, second piece: `Speed` = sum + wind; leak loss (§12.5); round 7: speed ≥ brake limit → `landing` (D); else move 0/1/2 with the per-step checks overshoot → collision → move → turn (§6.2); then Mastery.
   - radio, gear, flaps, brakes, concentration, kerosene: §6.3–§6.7, §12.1.
   - after every placement: ice deployment loop (§12.6).
4. A loss sets `lost` and stops. 5. `advance()`.

**`advance()`** (the one place for automatic resolutions, run after every gameplay mutator):
1. After a placement: `Turn` passes to the partner if they hold dice, else stays.
2. Synchronisation: in play, not used this round, and the round now has a piece (die or intern token) on gear and one on flaps → roll the traffic die (no coffee, never on an intern space) (`SyncUsed = true`). If the Co-Pilot has a legal space for it, step `sync`; else `SyncDiscarded`.
3. In step `place`: if the turn player holds dice but has no legal placement (with coffee and training), discard them (`DiceDiscarded`); the turn passes; repeat for the partner.
4. Both hands empty and step `place` → end of round (§5.5): mandatory check; kerosene −6 (§12.1); round 7 → final evaluation (§7.1, §5.4 below); else `Round++`, `missedAirport` if round 7 and not at the airport; new round: pickup, traffic rolls, step `discussion`.

### 5.4 Data tables

- **Panel spaces** (`panel.go`, §3.1, §12): 20 base spaces plus `kerosene`, `internPilot`, `internCopilot`, `ice2Top…ice5Top` (blue), `ice2Bottom…ice5Bottom` (both). Each: kind, colour, allowed values, order index (flaps, brakes, ice columns). `activeSpaces`: ice brakes remove `brakes2/4/6`; the leak closes `kerosene`.
- **Tracks** (`scenarios_data.go`, §10): 23 `TrackDef{ID, Spaces []TrackSpace{Planes, Traffic, Turn TurnMask}}`; `TurnMask` parsed from the 5-character strings, `""` = no icon.
- **Scenarios** (§9): 23 `ScenarioDef{ID, Airport, City, Colour, Track, Modules []Module, Abilities int, RealTimeDropped bool}`. Modules: `kerosene`, `keroseneLeak`, `intern`, `wind`, `iceBrakes`. Altitude side follows the colour.
- **Altitude** (§4.2): 7 rows with first player and the reroll icon per side.
- **Wind** (§12.3): the 11-value table and `WindValue(offset)`; **traffic die** faces `2,3,3,4,4,5`; **abilities**: the 6 IDs.
- Accessors for the expectation tests: `Scenarios() []ScenarioDef`, `TrackFor(TrackID) (TrackDef, bool)`, `AltitudeRow(round int) AltitudeRow`, `WindValue(int) int`, `TrafficFaces() [6]int`, `Abilities() []AbilityID`, `Spaces() []SpaceDef`.

### 5.5 Views and hidden information

| Field | Viewer | Partner |
|---|---|---|
| Own unplaced dice (values by slot) | yes | count only (`diceLeft`) |
| Placed pieces (die, intern token, traffic die) with value and coffee | yes | yes |
| Reroll selection | never stored (rolled at confirm) | only "confirmed" |
| Working Together: initiator's die value | yes | yes (public, as on the card) |
| Working Together: initiator's slot, partner's answer | own only | hidden |
| Adaptation / Anticipation results | own dice only | only "used" |
| Discarded dice | count | count |
| Legal spaces, intern target spaces, sync targets | the acting viewer only | – |
| Intern token row, traffic rolls, Sync die value, ice marker, fuel, wind | yes | yes |
| RNG state, future rolls | never | never |
| Game over | both players' unplaced dice revealed | same |

**Events are public and value-free for private moves:** `DiceRolled{}`, `RerollConfirmed{Role}`, `AdaptationUsed{Role}`, `AnticipationUsed{Role}`, `SwapOffered{Role, Value}` (the offered value is public), `DiceSwapped{}`, `DiceDiscarded{Role, Count}` carry no die values. Placements, traffic and Sync rolls carry values (public).

### 5.6 Intents

| Domain method | WS command | Errors (besides `ErrUnknownPlayer`, `ErrWrongPhase`) |
|---|---|---|
| `Configure(by, scenario, abilities)` | `configure` | `ErrNotHost`, `ErrUnknownScenario`, `ErrInvalidAbilities` (unknown, duplicate, too many) |
| `SwapRoles(by)` | `swapRoles` | `ErrNotHost` |
| `Start(by, rng)` | `startGame` | `ErrNotHost`, `ErrNotEnoughPlayers`, `ErrAbilitiesIncomplete` |
| `Ready(by, rng)` | `ready` | `ErrWrongStep`, `ErrAlreadyReady` |
| `Place(by, PlaceIntent{Slot, Space, Coffee, TokenSpace}, rng)` | `place` | `ErrWrongStep`, `ErrNotYourTurn`, `ErrNoSuchDie`, `ErrNotEnoughCoffee`, `ErrValueOutOfRange`, `ErrIllegalSpace`, `ErrIllegalTokenSpace` |
| `SpendReroll(by, rng)` | `spendReroll` | `ErrWrongStep`, `ErrNoReroll` |
| `ConfirmReroll(by, slots, rng)` | `confirmReroll` | `ErrWrongStep`, `ErrAlreadyConfirmed`, `ErrNoSuchDie` |
| `Adapt(by, slot, rng)` | `adapt` | `ErrWrongStep`, `ErrAbilityNotInPlay`, `ErrAbilityUsed`, `ErrNoSuchDie` |
| `Anticipate(by, slot, rng)` | `anticipate` | as `Adapt`, plus `ErrNotYourTurn` |
| `OfferSwap(by, slot, rng)` | `offerSwap` | as `Adapt`, plus `ErrPartnerHasNoDice` |
| `AnswerSwap(by, slot, rng)` | `answerSwap` | `ErrWrongStep`, `ErrNotYourTurn`, `ErrNoSuchDie` |
| `PlaceSync(by, space, rng)` | `placeSync` | `ErrWrongStep`, `ErrNotYourTurn`, `ErrIllegalSpace` |
| `GiveUp(by)`, `EndGame(by)`, `Rematch(by, newID)`, `AddPlayer`, `Leave` | as Fox Duet | as Fox Duet |

`Coffee` is the net modifier; `TokenSpace` is required exactly when `Space` is an intern space. Every gameplay mutator takes the RNG (the round end can roll traffic). RNG order is documented on each method (e.g. `Ready`: Pilot slots 0–3, then Co-Pilot slots 0–3).

### 5.7 Randomness and persistence

- **RNG injected:** a die is `1 + IntN(6)`, the traffic die `faces[IntN(6)]`, the intern row `Shuffle(6)` at `Start`. The service gets it from `Config.NewRNG` (Fox Duet). Tests use a scripted RNG (§9).
- **Snapshot:** `sqlite/dto.go` maps `Game`, `Flight`, `Round` explicitly; enums as their string values; `Swap` and `Speed` nullable. Scenario and track data are not persisted (looked up by ID). Migration `0001_init.sql` copied from Fox Duet. First version, so nothing old to keep loading; from the first deploy on, renamed values are mapped in the DTO.

### 5.8 Seams for later

- **Real-time:** `ScenarioDef.RealTimeDropped` records it. A timer would end the round through the same discard path as §5.2 (`discardHand(role)`), plus a deadline in `Round`. Not built.
- **Progress tracking:** a per-device list in localStorage (a `createPref`) or a logbook in `Game` as in The Crew. Not built.

## 6. Service, ports, adapters

Mirror Fox Duet file by file. Deviations:
- **ports/service:** one method per intent of §5.6 (no level). `CreateGame(ctx, hostName, opts *FlightOptions)` takes the host's remembered scenario and abilities (as The Crew's options); invalid options fall back to defaults.
- **httpapi:** `POST /api/games` body may carry `{"flight": {"scenario", "abilities"}}`. No other new route.
- **cmd/server:** `serverkit.LoadConfig(":8086", "data/sky-team.db")`.

**Wire contract** (camelCase; `types.ts` mirrors it; the UI track builds from this section):

```ts
type Role = 'pilot' | 'copilot';
type Piece = 'pilot' | 'copilot' | 'intern' | 'traffic';
type Step = 'discussion' | 'place' | 'reroll' | 'swap' | 'sync';
type SpaceId = 'axisPilot' | 'axisCopilot' | 'enginesPilot' | 'enginesCopilot' | 'radioPilot' | 'radioCopilot1'
  | 'radioCopilot2' | 'gear12' | 'gear34' | 'gear56' | 'flaps12' | 'flaps23' | 'flaps45' | 'flaps56'
  | 'brakes2' | 'brakes4' | 'brakes6' | 'concentration1' | 'concentration2' | 'concentration3' | 'kerosene'
  | 'internPilot' | 'internCopilot' | 'ice2Top' | 'ice3Top' | 'ice4Top' | 'ice5Top'
  | 'ice2Bottom' | 'ice3Bottom' | 'ice4Bottom' | 'ice5Bottom';
type Module = 'kerosene' | 'keroseneLeak' | 'intern' | 'wind' | 'iceBrakes';
type AbilityId = 'adaptation' | 'anticipation' | 'control' | 'mastery' | 'synchronisation' | 'workingTogether';
type LossReason = 'spin' | 'collision' | 'overshoot' | 'turn' | 'missedAirport' | 'mandatory' | 'kerosene'
  | 'landing' | 'intern' | 'iceBrakes' | 'gaveUp';
type Condition = 'A' | 'B' | 'C' | 'D' | 'intern' | 'iceBrakes';

interface ScenarioInfo { id: string; airport: string; city: string; colour: 'green'|'yellow'|'red'|'black';
  modules: Module[]; abilities: number; realTimeDropped: boolean; spaces: number }

interface PlayerView {
  gameId: string; revision: number; phase: 'lobby'|'playing'|'won'|'lost'|'ended';
  you: string; hostId: string; nextGameId: string;
  players: { id: string; name: string; seat: number; role: Role; isHost: boolean; isYou: boolean;
    mustAct: boolean; diceLeft: number; ready: boolean; revealedDice: number[] | null }[];
  scenario: ScenarioInfo; abilities: AbilityId[];
  catalogue: ScenarioInfo[] | null;           // lobby only, the 23 scenarios in §9 order
  flight: FlightView | null;                  // null in the lobby
  result: { won: boolean; reason: LossReason | null; failed: Condition[] } | null;
}
interface FlightView {
  round: number; firstPlayer: Role; step: Step; turn: Role;
  approach: { spaces: { planes: number; traffic: number; turn: string | null }[]; position: number };
  rerollRounds: number[];                     // altitude rounds still holding a reroll token
  axis: number; aero: { blue: number; orange: number };
  gear: [boolean, boolean, boolean]; flaps: number; brakes: number; brakeLimit: number;
  coffee: number; rerolls: number; planeSupply: number;
  fuel: number | null; leak: boolean; wind: { offset: number; value: number } | null;
  ice: { marker: number } | null;
  intern: { tokens: { value: number; taken: boolean }[]; nextPilot: number | null; nextCopilot: number | null } | null;
  placements: { space: SpaceId; piece: Piece; value: number; coffee: number; role: Role | null }[];  // in order
  speed: { sum: number; wind: number; total: number; move: number | null } | null;
  roundStart: { rerollPicked: boolean; traffic: { roll: number; space: number }[] };
  yourDice: (number | null)[];                // 4 slots; null = placed or discarded
  abilityState: { adaptationUsed: Record<Role, boolean>; anticipationUsed: boolean; togetherUsed: boolean;
    syncUsed: boolean; controlTriggered: boolean; masteryTriggered: boolean };
  can: { spendReroll: boolean; adapt: boolean; anticipate: boolean; offerSwap: boolean };
  legal: { byValue: SpaceId[][]; internToken: number | null; internTargets: SpaceId[] } | null; // your turn in place; byValue[v-1]
  reroll: { confirmed: Record<Role, boolean> } | null;
  swap: { by: Role; value: number; yourSlot: number | null } | null;   // value public; yourSlot only for the initiator
  sync: { value: number; targets: SpaceId[] } | null;   // targets only for the Co-Pilot
}
```

**Events** (`{type, ...}`): `playerJoined`, `playerLeft`, `flightConfigured`, `rolesSwapped`, `gameStarted`, `roundStarted{round, rerollPicked, traffic}`, `playerReady{role}`, `diceRolled`, `diePlaced{role, piece, space, value, coffee}`, `internTrained{role, token, space}`, `axisMoved{from, to}`, `windTurned{offset, value}`, `planeMoved{from, to}`, `planeCleared{space}`, `switchDeployed{space}`, `coffeeGained`, `fuelChanged{from, to}`, `iceDeployed{marker}`, `rerollSpent{role}`, `rerollConfirmed{role}`, `rerollGained`, `adaptationUsed{role}`, `anticipationUsed{role}`, `swapOffered{role, value}`, `diceSwapped`, `syncRolled{value}`, `syncDiscarded`, `diceDiscarded{role, count}`, `roundEnded{round}`, `gameWon`, `gameLost{reason, failed}`, `gameEnded`, `rematchCreated{gameId}`.

**WS commands** (`ClientMessage`: `type`, `requestId`, plus): `configure{scenario, abilities}`, `swapRoles`, `startGame`, `ready`, `place{slot, space, coffee, tokenSpace?}`, `spendReroll`, `confirmReroll{slots}`, `adapt{slot}`, `anticipate{slot}`, `offerSwap{slot}`, `answerSwap{slot}`, `placeSync{space}`, `giveUp`, `endGame`, `rematch`.

**Error codes** (`apidto/errors.go`): one snake_case code per sentinel of §5.6 (`illegal_space`, `not_enough_coffee`, `ability_used`, …).

## 7. Frontend

Approved layout: concept A in `sky-team/docs/layout-concepts.html` (text sections and `conceptA`, `panel`, `approach`, `hand`, `action`); look: `sky-team/docs/mockup.html`.

- **Lobby** (`Lobby.svelte`): both seats with their role and the host's *Swap roles*; the scenario picker grouped by colour (airport, city, module icons, ability count, "played without real-time" note for the four); the ability picker (exactly n of the 6 cards, with their texts) when the scenario has abilities; Start (disabled with a reason until n are picked); `LobbyActions`. Host's last pick per device: `sky-team:flight` (`createPref`), sent with `POST /api/games`.
- **Table, top to bottom (phone):** header (home, game code, connection dot, scenario chip, RulesHelp, give-up flag, theme toggle; no round counter, since the altitude track shows the round and the first player); track strip: approach (cells with planes, traffic icons, turn masks with the current axis column outlined, the plane's space, airport code) and altitude (7 cells, first-player arrow, reroll tokens); one chip row (fuel bar, wind needle and value, ability chips with their own icon, dashed with a check when used, that open a sheet with the text); the **control panel** (blue column: radio, gear ×3, fuel; centre: axis dial, engines with the three speed ranges or the round-7 land / crash ranges, brakes or ice brakes with the marker, concentration ×3, intern board; orange column: radios ×2, flaps ×4); the status line; the dice tray (4 fixed slots, partner's remaining count); the action bar (coffee stepper, Place, reroll button with the supply; Adapt / Anticipate / Together when `can` says so).
- **Placement flow:** tap a die → its legal spaces (from `legal.byValue` at the current coffee modifier) get a ring, the rest dims; tap a space → the status line previews the effect (`preview.ts`); Place commits. A dimmed space explains itself ("needs 5–6", "deploy flaps 2–3 first", `illegal.ts`). Training: choose the intern space, then the token's target among `internTargets`. The partner's last placement gets a "new" mark.
- **Steps:** discussion (round-start results: reroll picked up, traffic rolls and new planes; "talk strategy now, not dice"; Ready, partner's ready state); reroll (tap dice to mark, "Keep all" / "Reroll n", partner confirming); swap (initiator waits; partner picks the die to give); sync (the black die takes an empty tray slot, Co-Pilot taps a target; Pilot waits); waiting (status line names the partner).
- **Desktop (1280×800):** tracks vertical on the left as on the table, ability texts on the right, 60 px spaces (concept A's desktop frame).
- **GameOver:** won or the named loss with the failed conditions in words, revealed dice, Rematch; **RulesHelp:** round flow, each space, the scenario's modules and abilities.
- **Helpers with vitest:** `format.ts` (roles, scenarios, modules, abilities, loss reasons, conditions), `preview.ts` (effect preview per space and value: speed with wind and bands, movement with collision/turn warnings, radio target, switch and marker changes, coffee, fuel, ice), `illegal.ts`. They are UX only; legality comes from the server.
- **Fixtures** (`fixtures.ts`, presets select a die and a coffee modifier): `lobbyHost`, `lobbyHostAbilities` (TGU red, 2 picked), `lobbyGuest`, `lobbyAlone`; `discussion`, `discussionTraffic`, `discussionPartnerReady`; `yulTurn`, `yulWaiting`, `yulFinalRound`; `anticipation` (KEF black); `tguTurn`, `tguWaiting`, `tguReroll`, `kefTurn`, `hndTurn` (Co-Pilot, coffee −1), `hndTraining`; `rerollConfirmed`; `swapOffered`, `swapAnswer`; `syncPlace` (Co-Pilot), `syncWaiting` (Pilot); `won`, `lostSpin`, `lostCollision`, `lostLanding` (A, B, D failed), `lostKerosene`, `lostIntern`, `ended`. `fixtures.test.ts` checks every phase **and every step** is covered.

## 8. Integration checklist

- [ ] `sky-team/Dockerfile`, `Dockerfile.dockerignore`, `deploy/compose.yaml`, `deploy/sky-team.caddy` (`import gate`): from the scaffold, checked.
- [ ] `start-page/frontend/src/lib/games.ts`: id `'sky-team'`, path `/sky-team/`, devPort 5180, `'2 players'`, tagline and description (own wording); placeholder tile in `GameArt.svelte` and `+page.svelte`; tile tokens in `layout.css` (the theme WP draws the art); port comment in `vite.config.ts`.
- [ ] `.claude/skills/run/run.sh`: `[sky-team]="sky-team 5180 8086 -"` and the two `--all` lists.
- [ ] Root `CLAUDE.md`: app table, local ports table. `gate/backend/cmd/server/main.go`: port comment (`:8086` Sky Team's).
- [ ] `IDEAS.md`: mark Sky Team implemented. `TODO.md`: a Sky Team section (progress tracking, real-time, plumbing copies).
- [ ] `sky-team/CLAUDE.md`: modelled on `fox-duet/CLAUDE.md` (WP19).
- [ ] CI and deploy scripts pick the app up on their own; no `setup-server.sh` re-run.

## 9. Testing

- **`export_test.go`** (WP3): `ScriptedRNG(ints ...int)` (`IntN` returns them in order, `Shuffle` is the identity), `DieRoll(v)` and `TrafficRoll(r)` (the `IntN` result for a die value / traffic roll), `NewFlightForTest(FlightSetup) *Game` (a started game in step `place` with given scenario, abilities, round, position, planes, axis, switches, coffee, rerolls, fuel, wind, ice, intern row, dice, turn and placements), `MustPlace`, `LegalSpaces(g, role, value)`, `FailUnchanged(t, g, fn, err)`.
- **Domain (table-driven, per §):** every space's constraints and effect; coffee bounds; alternation and the discard rule; axis incl. spin and C; engines with 0/1/2 moves, holding pattern, overshoot, collision, each turn mask, D; radio reach and beyond the airport; gear, flaps in order and green flaps, brakes in order and closed; concentration cap; reroll pickup per side; traffic rolls and supply limit; end of round and mandatory; missed airport; win; every module (§12 examples verbatim, ice "waits" case); every ability incl. Mastery availability on a red track, Sync discard, Working Together with values; every error leaves the state unchanged.
- **Black-box** (`rules_*_test.go`, package `domain_test`, rules-test-writer): §3–§7, §11 (WP7) and §12–§13 (WP10).
- **Data:** independent expectation tables for the 23 tracks, 23 scenarios, altitude, wind, traffic faces, panel spaces (WP5), compared through the §5.4 accessors.
- **Simulation** (`sim_test.go`): every scenario × a random legal ability set × ≥ 20 seeds; random legal intents incl. side intents. Invariants after every action: no panic; per round each role's dice in hand + placed + discarded = 4; plane tokens on the track ≤ 12; coffee 0..3; supply + track + spent rerolls = 2; axis in −2..2 while playing; actors never empty while playing; every game ends by round 7; the partner's view never carries the viewer's unplaced values or swap slot.
- **Adapters:** repotest (copied), sqlite round trip of a mid-round KEF black (ice, wind, swap pending) and HND red (intern, kerosene) game, apidto mapping of every field, ws dispatch for every command, httpapi create with options.
- **e2e:** two players over HTTP and WS, YUL with a scripted RNG: discussion, roll, a full round, a reroll step; assert the partner's views and all events never contain the other's unplaced values.
- **Frontend:** vitest for `format.ts`, `preview.ts`, `illegal.ts`, the pref parser, fixtures.

## 10. Work packages

All acceptance runs from the repo root; "check" = `bash .claude/skills/check/check.sh sky-team`. Domain WPs also run `go test ./internal/core/domain/...` from `sky-team/backend`.

| WP | Work (rules.md §) | Agent (model) | Depends on | May touch | Acceptance |
|---|---|---|---|---|---|
| 1 | **Scaffold.** `bash .claude/skills/implement-game/scaffold.sh sky-team "Sky Team" 5180 8086`, then strip Fox Duet's game to a lobby-only 2-seat game (no level, no cards) in domain, service, adapters and frontend; keep the fixtures route with lobby/ended fixtures. | game-implementer (sonnet) | – | `sky-team/**` | check; create/join/lobby at :5180 |
| 2 | **Integration** (§8 except `sky-team/CLAUDE.md`), plain placeholder tile. | game-implementer (sonnet), worktree | – (∥1) | `start-page/frontend/**`, `.claude/skills/run/run.sh`, `CLAUDE.md`, `gate/backend/cmd/server/main.go`, `IDEAS.md`, `TODO.md` | `bash .claude/skills/check/check.sh start-page gate` |
| 3 | **Domain foundation** (§1–§4, §14): files of §5.1 with the types, `panel.go` complete, `scenario.go` types and accessors, YUL track and scenario, `Game`/`Flight`/`Round`, steps and `Actors()`, all §5.6 signatures (stubs → `ErrNotImplemented`), all events, `PlayerView` types, lobby lifecycle incl. `Configure`, `SwapRoles`, `Start` (setup §4.1, intern shuffle, round 1 start), `export_test.go` of §9. | game-implementer (**opus**) | 1 | `sky-team/backend/internal/core/domain/**` | domain tests green; `go doc` shows every §5.6 method |
| 4 | **Data tables** (§4.2, §9, §10, §12.3, §13): the 22 remaining tracks and scenarios, altitude, wind, faces. | game-implementer (sonnet) | 3 | `domain/scenarios_data.go`, `domain/scenario.go` | domain tests green |
| 5 | **Data expectation tables**, from rules.md only (§3.1, §4.2, §9, §10, §12.3, §12.6, §13). | rules-test-writer (sonnet), worktree | 3 (∥4) | `domain/data_expect_test.go` | merged and green after reconciliation |
| 6 | **Base round** (§5, §6, §7, §11, §15): discussion and roll, placement, coffee, alternation, discard, every base space, approach with turns, traffic die, reroll pickup, reroll step, end of round, final round, results; `ViewFor` base fields with hidden-info tests. | game-implementer (**opus**) | 3 (YUL from 3; ∥4) | `domain/**` except the data files | domain tests green |
| 7 | **Black-box base rules** (§3–§7, §11). | rules-test-writer (sonnet), worktree | 3 (∥6) | `domain/rules_base_*_test.go` | merged and green after reconciliation |
| 8 | **Modules** (§12.1–§12.3, §12.5, §12.6; §12.4 stays out): kerosene, leak, intern training with token targets, wind, ice brakes; their view fields. | game-implementer (**opus**) | 4, 6 | `domain/**` except data files | domain tests green |
| 9 | **Abilities** (§13): Adaptation, Anticipation, Control, Mastery, Synchronisation step, Working Together step; `can`, `abilityState`, `swap`, `sync` view fields; hidden-info tests for swap, adapt, anticipate, reroll. | game-implementer (**opus**) | 8 | `domain/**` except data files | domain tests green |
| 10 | **Black-box modules and abilities** (§12, §13). | rules-test-writer (sonnet), worktree | 3 (∥8, 9) | `domain/rules_modules_*_test.go`, `domain/rules_abilities_*_test.go` | merged and green after reconciliation |
| 11 | **Simulation test** (§9). | game-implementer (sonnet) | 9 | `domain/sim_test.go` | green, all scenarios × ≥ 20 seeds |
| 12 | **Ports, service, sqlite** (§6): ports, service with `CreateGame` options, sqlite DTO and round-trip tests. | game-implementer (sonnet) | 9 | `sky-team/backend/internal/core/{ports,service}/**`, `internal/adapters/{sqlite,memory,repotest}/**` | check |
| 13 | **apidto, ws, httpapi, e2e** (§6 wire contract, error codes). | game-implementer (sonnet) | 12 | `internal/adapters/{apidto,ws,httpapi}/**`, `internal/e2e/**`, `cmd/server/**` | check |
| 14 | **Frontend plumbing and fixtures** (§6, §7): `types.ts` from §6, client commands, connection, `format.ts` + tests, `flightPref.ts`, `fixtures.ts` with presets (all §7 fixtures, worst cases matching the concept's TGU/KEF/HND data), `fixtures.test.ts` incl. step coverage. | game-implementer (sonnet) | 1 | `sky-team/frontend/src/lib/{types.ts,game/**}` | check |
| 15 | **Theme** (from `mockup.html`): tokens in `layout.css` light/dark, Fontsource fonts, `HomeBand`, start-page tile art and tokens. | game-implementer (**opus**) | 1, mockup approved | `sky-team/frontend/src/routes/layout.css`, `src/lib/components/HomeBand.svelte`, `package.json`, `start-page/frontend/src/lib/components/GameArt.svelte`, `start-page/frontend/src/routes/layout.css` | check + start-page; lobby fixture at 375×812 light and dark |
| 16 | **Table: cockpit and placement** (§3, §5.2, §5.4, §6, §10, §11, §12 readouts): `Table`, `ControlPanel`, `Space`, `Die`, `ApproachTrack`, `AltitudeTrack`, `ModuleChips`, `DiceTray`, `StatusLine`, `ActionBar`, `preview.ts` and `illegal.ts` with tests; the place step incl. training. | game-implementer (**opus**) | 14, 15 | `sky-team/frontend/src/lib/components/game/**`, `src/lib/game/{preview,illegal}*.ts` | check; worst-case fixtures fit 375×630 without scrolling (Playwright) |
| 17 | **Table: the other steps and desktop** (§5.1, §5.3, §8, §13): discussion, reroll, swap, sync, Adapt/Anticipate actions, waiting states, the "new" mark, card motion for dice, the 1280×800 layout. | game-implementer (**opus**) | 16 | same as 16 | check; every step fixture at 375×812 and 1280×800 |
| 18 | **Lobby, GameOver, RulesHelp** (§4.1, §7.2, §9, §13): scenario and ability pickers, swap roles, real-time note, results with conditions, revealed dice, rules summary. | game-implementer (sonnet) | 14, 15 (∥16) | `components/game/{Lobby,GameOver,RulesHelp}.svelte`, `components/game/lobby/**` | check; lobby and game-over fixtures |
| 19 | **Hook-up and docs:** diff `apidto` against `types.ts`, fix drift; play YUL and TGU red end to end with two browser profiles; `sky-team/CLAUDE.md`; TODO.md entries. | game-implementer (sonnet) | 13, 17, 18 | `sky-team/**`, `TODO.md` | check; a 2-player round played locally |
| 20 | **UI polish pass** at checkpoint C, on the real components through fixtures (`/visual-review sky-team` first). Known items: (1) the speed strip reads uneven in some states, worst in round 7, where "✓ land" overlaps its narrow segment (`docs/reference/mockup3-speed-land-overlap.png`); copy uses "<" with the marker. (2) A second pass on the axis dial towards the physical gauge (`docs/reference/panel-axis-speed.jpg`): black bezel with white triangle ticks and a hollow zero mark, red crosses at the ends, sky and ground disc behind the banking plane; the box arcs the speed numbers around the dial, worth a look. (3) The altitude track reads unclear in dark mode: make it clearer. Look only, no UX changes. | main session with the user | 17, 18 | `sky-team/frontend/src/lib/components/**`, `src/routes/layout.css` | check; the touched fixtures at 375×630, 375×812 and 1280×800, light and dark |

**Critical path:** 1 → 3 → 6 → 8 → 9 → 12 → 13 → 19.
**Parallel tracks:** 2 with 1; 4 and 5 with 6; 7 with 6; 10 with 8–9; 11 with 12–13; UI: 14 and 15 right after 1, then 16 → 17, with 18 alongside 16. The UI checkpoint (C) follows 17 and 18 and needs no backend; the polish pass (20) runs there, alongside 19.

**Review gates:**
- **G1 after WP3:** domain shape, step machine, signatures, `export_test.go` (before 5, 7, 10 start).
- **G2 after WP4 + WP5:** the 23 tracks and scenarios against rules.md §9–§10 row by row.
- **G3 after WP6 + WP7:** base semantics; `rules-auditor` early on §5–§7, §11 (resolution order, discards, round 7).
- **G4 after WP9 + WP10:** modules, abilities and hidden information in the view and events.
- **G5 after WP13:** no leaks on the wire (view → DTO mapping, events, e2e).
- **G6 after WP2:** the start-page integration.
- **G7 after WP17 + WP18:** the table UI on fixtures (checkpoint C), then the polish pass WP20.

## 11. Risks

| Risk | Handling |
|---|---|
| The panel does not fit 375×630 in the real theme | Concept A measured 14 px spare; keep its sizes, worst-case fixtures from the concept's data, Playwright check in WP16's acceptance. |
| Resolution order bugs (axis → wind → engines, Sync after the turn passes, discards, round 7 early ends) | One `advance()`; order written in §5.3; black-box tests; `rules-auditor` at G3 and G4. |
| Leaks through events or view fields (reroll, swap, adapt, anticipate) | Value-free events (§5.5); view tests in WP6/WP9; sim invariant; e2e; gate G5. |
| A stuck round (no actor, dice left) | `advance()` discards; sim checks actors never empty and every game ends. |
| Fixtures drift from the real wire | `types.ts` from §6; apidto tests; contract diff in WP19. |
| TS preview duplicating rules | Preview and illegal texts are UX only; legality comes from `legal.byValue`; vitest cases. |
| Transcription errors in 23 tracks | Independent expectation table (WP5) and gate G2. |

## 12. Status

**Run state (2026-10-23):** step 3 (build). Plan, rules and look are approved (checkpoint B passed 2026-10-23 with mockup round 3). Branch `game/sky-team`, not pushed.

| WP | Status | Commit | Notes |
|---|---|---|---|
| 1 | done | 98f6613 | Lobby-only skeleton. `fixtures.test.ts` has a per-phase flag: flip playing/won/lost when their fixtures arrive. |
| 2 | done | d2ba725 | Placeholder tile with neutral `--sky-*` tokens (the real tile comes in WP15); dependabot entry added. |
| 3 | done | c077dd0 | Gate G1 passed (shape as §5). Gameplay intents are stubs (`ErrNotImplemented`), `legalSpaces` empty (WP6). Already filled: altitude, wind, traffic faces, abilities tables, so WP4 has only the 22 tracks and scenarios left. Names: `AltitudeSpace` type, `TrafficDie(r)` test helper; `Actors()` returns `[]Role`. `brakeLimit` on the wire is the marker value: the round-7 speed must be lower than it (2/3/5/7; ice 2 + slot), confirmed by the user; the fixtures follow. UI copy uses "<" with the marker ("Speed 3 < 5"), not the mockup's "≤ 4". `TrafficRoll.Added` (false when the supply was empty): WP13 decides if it goes on the wire. Lobby leave: the remaining player keeps their role, and a newcomer takes the free one. |
| 14 | done | d24720d | All 29 fixtures plus `presets`. **WP13 must use the error codes in `frontend/src/lib/errors.ts` verbatim** (snake_case of the §5.6 sentinels). Wire details beyond §6: `POST /api/games` body `{name, flight}`; `diePlaced.role` is null for the traffic die; `approach.position` 1-based; reroll supply + track + spent = 2. WP16: add a `preset` prop to `GameScreen` and pass `data.preset` (marked in `dev/fixtures/[name]/+page.svelte`); `fixtureLegal.ts` is for fixtures only. WP18/19: wire `flightPref.load()` into the home page's create call. |
| 4–13 | open | | Next: 4 ∥ 5 ∥ 7 ∥ 6 may start (G1 passed). Backend track: 4 ∥ 5 after 3, 6 ∥ 7 after 3, then 8 → 9 → 12 → 13; 10 and 11 per §10. |
| 15–18 | open | | Unblocked: build from `docs/mockup.html` with the locked picks (§3). |
| 19 | open | | |
| 20 | open | | UI polish pass at checkpoint C (§10). |

**Resuming in a fresh session** (`/implement-game sky-team`):
1. Check for leftovers of the session that ran out: `git status` (uncommitted WP3 domain files in the main checkout, an edited `sky-team/docs/mockup.html`) and `git worktree list` (the WP14 worktree). Run `bash .claude/skills/check/check.sh sky-team` on what is there. If it's green and complete for its WP, review and commit it; else re-run that WP with a fresh agent. Remove stale worktrees.
2. **Mockup:** approved. Round 3 (the locked version) is `sky-team/docs/mockup.html`, published as version 4 of https://claude.ai/artifact/WZeWdkJR13z1DA6jXTaqra. WP15–18 may start; the look's leftovers are WP20. No "Round r of 7" in the table header (user, 2026-10-23).
3. Continue the backend track from the first open WP.

**Open items:** none. The checkpoint B answers are App decisions in rules.md: the intern token counts as a die, the Sync die takes no coffee and no intern space, the Working Together die is public, and the planner's decisions. Discards stay as written in §5.2, even with a reroll left.
