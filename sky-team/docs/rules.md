# Rules reference

English summary of the official rules of **Sky Team** (Luc Rémond, Le Scorpion Masqué 2023; Hachette Boardgames in the US). The code follows this file. Where the official rules are ambiguous or the app must decide something, the choice is recorded under **App decision**. The app is titled "Sky Team" and lives at `/sky-team/`.

Sources:
- The English rulebook *Landing Procedure* (12 pages) and a web copy of the *Flight Log* booklet (8 pages: modules, special abilities, 21 scenarios). Where the user's printed flight log differs, the printed one wins (HND red, §9).
- Photos of all approach tracks (both sides), both sides of the altitude track, the table, the six special-ability cards (German edition) and the wind ring. §10 transcribes the tracks, §12.3 the wind ring and §13 the cards from these photos.
- [Dized rules for Sky Team](https://rules.dized.com/game/HPLDQm89RB-4e9yi2MQQGw/rfj57n1yQP-jSbISfRmOoA/mastery-special-ability) (Le Scorpion Masqué's official rules app), only for the Mastery clarification. The traffic die faces were confirmed by the user.

**Licensing:** the app runs behind the site's password gate for friends only. It uses its own art, icons and wording, with no Scorpion Masqué imagery or flight-log story text. The name "Sky Team" appears only as the game's title. Airport codes and names are real-world names, not publisher material.

---

## 1. Overview

Cooperative dice-placement game for **exactly 2 players**: the **Pilot** (blue) and the **Co-Pilot** (orange). Together they land an airliner in **7 rounds**. Each round both players roll 4 dice in secret and place them one at a time, alternating, on the control panel. The dice steer the plane (axis), set its speed (engines), clear traffic (radio), and deploy landing gear, flaps and brakes. Players may not talk about their dice during a round.

The team wins if, at the end of round 7, the plane is at the airport, level, with all gear and flaps deployed, no traffic left, and a speed below the brakes. Many things lose the game at once (§7.2).

The base game is the Montréal scenario. The flight log adds 20 more scenarios built from approach-track effects (§11), six modules (§12) and special-ability cards (§13).

## 2. Components

| ID | Component | Count | Notes |
|---|---|---|---|
| `bluedice` | Pilot dice | 4 | standard d6 |
| `orangedice` | Co-Pilot dice | 4 | standard d6 |
| `controlPanel` | Control panel | 1 | spaces in §3 |
| `axis` | Airplane axis disc | 1 | §3.2 |
| `aeroBlue`, `aeroOrange` | Aerodynamics markers | 2 | on the speed gauge, §3.3 |
| `brakeMarker` | Brake marker (red) | 1 | §3.4 |
| `switch` | Switches | 10 | 3 landing gear, 4 flaps, 3 brakes; show red or green |
| `planeToken` | Airplane (traffic) tokens | 12 | supply is limited to 12 |
| `coffee` | Coffee tokens | 3 | max 3 on the panel |
| `reroll` | Reroll tokens | 2 | §5.3 |
| `altitudeTrack` | Altitude track | 1 | two sides, §4.2 |
| `approachTrack` | Approach tracks | 11 double-sided (+1 promo) | §10 |
| `screen` | Player screens | 2 | hide the dice |
| `trafficDie` | Traffic die (black) | 1 | faces **2, 3, 3, 4, 4, 5** |
| `keroseneTrack` | Kerosene track + marker | 1 | §12.1, §12.5 |
| `internBoard` | Intern board + 6 intern tokens (values 1–6) | 1 | §12.2 |
| `windRing` | Wind ring + blue airplane token | 1 | §12.3 |
| `timer` | 60-second timer (players' phone) | – | real-time module, **not implemented** (§12.4) |
| `iceBrakes` | Ice brakes board | 1 | §12.6 |
| `ability` | Special-ability cards | 6 | §13 |

## 3. The control panel

### 3.1 Spaces

The Pilot sits on the left (blue side), the Co-Pilot on the right (orange side). A die may only go on a space of its owner's colour; two-coloured spaces take either colour. A space holds one die per round. Dice are removed at the end of the round; switches and markers stay.

| ID | Colour | Value | Effect | Section |
|---|---|---|---|---|
| `axisPilot` | blue | any | axis, **mandatory** | §6.1 |
| `axisCopilot` | orange | any | axis, **mandatory** | §6.1 |
| `enginesPilot` | blue | any | speed, **mandatory** | §6.2 |
| `enginesCopilot` | orange | any | speed, **mandatory** | §6.2 |
| `radioPilot` | blue | any | remove a plane | §6.3 |
| `radioCopilot1`, `radioCopilot2` | orange | any | remove a plane | §6.3 |
| `gear12` | blue | 1 or 2 | landing gear | §6.4 |
| `gear34` | blue | 3 or 4 | landing gear | §6.4 |
| `gear56` | blue | 5 or 6 | landing gear | §6.4 |
| `flaps12` | orange | 1 or 2 | flaps, 1st | §6.5 |
| `flaps23` | orange | 2 or 3 | flaps, 2nd | §6.5 |
| `flaps45` | orange | 4 or 5 | flaps, 3rd | §6.5 |
| `flaps56` | orange | 5 or 6 | flaps, 4th | §6.5 |
| `brakes2` | blue | 2 | brakes, 1st | §6.6 |
| `brakes4` | blue | 4 | brakes, 2nd | §6.6 |
| `brakes6` | blue | 6 | brakes, 3rd | §6.6 |
| `concentration1..3` | both | any | gain a coffee | §6.7 |

Per player: the Pilot has 1 radio space, the Co-Pilot 2. Gear and brakes are Pilot only, flaps Co-Pilot only.

Module spaces (only with that module): `kerosene` (both colours, any value, §12.1), `internPilot` (blue) and `internCopilot` (orange) (§12.2), the ice-brake spaces (§12.6), and the Working Together card (§13).

### 3.2 Axis

The axis has 7 positions: **−3 … +3**. Negative means tilted toward the Pilot, positive toward the Co-Pilot. It starts at **0** (level). Positions **±3** are the red crosses: reaching one loses the game (spin). Playable positions are therefore −2 … +2. The axis is never reset between rounds.

### 3.3 Speed gauge

Two aerodynamics markers sit between numbers on a gauge from 2 to 12. Code stores each as a threshold `T` meaning "the marker sits between T and T+1".

| Marker | Start | +1 per | Max |
|---|---|---|---|
| Blue (`aeroBlue`) | 4 | landing gear deployed | 7 (all 3 gear) |
| Orange (`aeroOrange`) | 8 | flap deployed | 12 (all 4 flaps: "just past the 12") |

Movement for a speed `s` (§6.2): `s ≤ blue` → 0 spaces; `blue < s ≤ orange` → 1 space; `s > orange` → 2 spaces.

### 3.4 Brake track

The red marker starts left of the 2 and moves one slot per brake deployed. The final-round speed must be **lower** than the marker.

| Brakes deployed | Marker position | Highest speed that lands |
|---|---|---|
| 0 | left of 2 | 1 |
| 1 (`brakes2`) | between 2 and 3 | 2 |
| 2 (`brakes4`) | between 4 and 5 | 4 |
| 3 (`brakes6`) | right of 6 | 6 |

Speed below 2 is only possible with the wind module (§12.3).

### 3.5 Coffee and reroll supply

Up to **3 coffee tokens** sit on the panel. Reroll tokens taken from the altitude track go to the reroll supply (top left). Both are shared by the team.

## 4. Setup

### 4.1 Base setup (Montréal)

1. Axis at 0. All 10 switches show red.
2. Blue aerodynamics marker at 4, orange at 8, brake marker left of 2.
3. Pilot takes the 4 blue dice, Co-Pilot the 4 orange dice.
4. Altitude track on the side for the scenario's colour (§4.2), at 6000.
5. Approach track of the scenario, at its start space (§10). Put one plane token on each airplane icon of the track.
6. Put a reroll token on each reroll icon of the altitude track.
7. Coffee tokens and the remaining plane tokens form the supply.
8. Set up the scenario's modules (§12) and special abilities (§13).

- **App decision:** the host picks the scenario in the lobby (default `yul-green`) and, if it has abilities, the ability cards (§13). Both players see the chosen scenario and abilities in the lobby and during the game.
- **App decision:** seat 1 (the host) is the Pilot and seat 2 the Co-Pilot by default. The host can swap the roles in the lobby.

### 4.2 Altitude track

7 spaces, one per round. The arrow on each space names the **first player** of that round. The two sides differ only in reroll icons.

| Round | Altitude | First player | Reroll icon, green/yellow side | Reroll icon, red/black side |
|---|---|---|---|---|
| 1 | 6000 | Pilot | yes | yes |
| 2 | 5000 | Co-Pilot | – | – |
| 3 | 4000 | Pilot | – | – |
| 4 | 3000 | Co-Pilot | – | – |
| 5 | 2000 | Pilot | yes | – |
| 6 | 1000 | Co-Pilot | – | – |
| 7 | landing | Pilot | – | – |

Green and yellow scenarios use the green/yellow side (2 reroll tokens); red and black scenarios use the red/black side (1 reroll token). The side is identified by the colour bars on the track.

## 5. A round

Each round has three phases: **discussion and roll**, **placement**, **end of round**.

### 5.1 Discussion and roll

1. **Reroll pickup:** if the current altitude space has a reroll token, move it to the reroll supply.
2. **Traffic** (only on tracks with traffic-die icons, §11.1).
3. **Discussion:** players talk about strategy ("we must clear that plane", "let's advance two"). They may not talk about dice values or which die goes where.
4. **Roll:** each player rolls their 4 dice behind their screen. From now until the end of the round, players are silent (correcting a rules mistake is allowed).

- **App decision:** the server rolls all dice. A player sees the values of their own unplaced dice only; of the partner they see only how many unplaced dice remain. Placed dice are public.
- **App decision:** the discussion phase ends when both players press **Ready**; then the server rolls. The app shows the reroll pickup and traffic results before the discussion.

### 5.2 Placement

- Players **alternate**, starting with the first player of the altitude space (§4.2). On a turn, a player places exactly **one** of their unplaced dice on a free space they may use (colour and value, §3.1). Effects resolve immediately (§6).
- The round continues until all 8 dice are placed.
- Axis and engines need not be filled first, but each must have a die of each colour by the end of the round (§6.1, §6.2).
- **App decision:** placement is strictly alternating, enforced by the server; it is never simultaneous. A placement is one intent: die, space, coffee modifier (§5.4).
- **App decision:** if the player whose turn it is has no legal placement for any of their remaining dice (possible only when intern tokens or the traffic die fill spaces, or concentration is full), their remaining dice are discarded for the round and the partner places the rest of theirs in a row.

### 5.3 Rerolls

At any time during placement, either player may spend one reroll token from the supply. Then **both** players may reroll any number of their unplaced dice, once each.

- **App decision:** spending a reroll opens a reroll step. Each player selects zero or more of their unplaced dice and confirms; the server rerolls the selected dice when a player confirms. Placement pauses until both have confirmed. A player without unplaced dice confirms automatically.

### 5.4 Coffee

When placing a die, a player may spend any number of coffee tokens. Each token changes the die's value by +1 or −1. The result must stay within 1–6 (no wrap from 1 to 6). Either player may spend any token. Unspent tokens carry over to later rounds.

- **App decision:** the intent carries a net modifier `d`; it costs `|d|` tokens. Only the final value counts for the space's constraint.

### 5.5 End of round

When all 8 dice are placed (or discarded, §5.2):

1. **Mandatory check:** if either axis space or either engine space is empty, the game is lost.
   - **App decision:** a space counts as covered by anything on it: a player's die, an intern token (§12.2) or the traffic die (§13). Without the real-time module (§12.4) this loss can only happen when dice were discarded (§5.2).
2. Module steps (kerosene, §12.1).
3. Advance the altitude track one space (the plane always descends).
4. Players take their dice back.
5. If the altitude track now shows the **landing** space:
   - the plane is at the airport → play the **final round** (§7.1);
   - otherwise → the game is lost (crash before the airport).
6. Otherwise start the next round.

The game always has exactly 7 rounds.

## 6. Actions

In this section a "die" on a space is anything placed there: a player's die, an intern token (§12.2) or the traffic die (§13); its value is what counts.

### 6.1 Axis (mandatory)

When the **second** axis die is placed, compare the two values. Equal: nothing happens. Otherwise the axis turns **toward the player with the higher die** by the difference. If it reaches ±3, the game is lost (spin).

Example: Pilot 5, Co-Pilot 3 → the axis moves 2 toward the Pilot (from 0 to −2).

### 6.2 Engines (mandatory)

When the **second** engine die is placed, the speed is the sum of both dice (plus wind, §12.3).

**Rounds 1–6:** move the plane on the approach track by 0, 1 or 2 spaces (§3.3). Each single step:
1. If the plane is on the airport: the game is lost (**overshoot**).
2. If plane tokens are on the current space: the game is lost (**collision**).
3. Move one space toward the airport.
4. Turns (§11.2): if the new space has a turn icon and the axis is not in a permitted position, the game is lost.

Plane tokens on the current space are harmless as long as the plane does not advance. **Holding pattern:** at the airport before round 7, the team must keep its speed at 0 spaces.

**Final round:** do not move. Compare the speed with the brake marker instead (§7.1, condition D).

### 6.3 Radio

A die with value `v` removes one plane token from the space `v − 1` steps ahead of the current space (`v = 1`: the current space). If that space has no token, or lies beyond the airport, nothing happens.

### 6.4 Landing gear (Pilot)

Place a die that fits the space's values. If the space's switch is red, turn it green and move the blue aerodynamics marker +1. Gear can be deployed in any order. A die on a space whose switch is already green has no effect (allowed).

### 6.5 Flaps (Co-Pilot)

Place a die that fits the space's values. Flaps deploy **in order**: `flaps12`, `flaps23`, `flaps45`, `flaps56`. A die on the next undeployed flap turns its switch green and moves the orange marker +1. A space after the next undeployed one is not playable.

- **App decision:** a die may go on an already green flap (matching its values); it has no effect, as for landing gear.

### 6.6 Brakes (Pilot)

Place a die of exactly the space's value. Brakes deploy **in order**: 2, then 4, then 6. Each deployment moves the brake marker one slot (§3.4). Deploying brakes is optional; they only matter in the final round.

- **App decision:** only the next undeployed brake is playable. A deployed brake takes no more dice (as with ice brakes, §12.6); after all three are deployed the brake spaces are closed.

### 6.7 Concentration

Any die of either player. Gain one coffee token if fewer than 3 are on the panel.

## 7. Final round and end of game

### 7.1 Final round (round 7)

The final round starts when the plane is at the airport and the altitude track shows landing. It is played like any round, except that engines compare speed with the brake marker (§6.2). At the end of the final round the team **wins** if all of these hold:

| ID | Condition |
|---|---|
| A | No plane tokens left on the approach track. |
| B | All 3 landing-gear and all 4 flaps switches are green. |
| C | The axis is at 0. |
| D | The final-round speed (engines + wind) is lower than the brake marker (§3.4). |

Plus module conditions: all intern tokens used (§12.2), ice-brake marker past 5 (§12.6), kerosene not below 0.

- **App decision:** a condition that can no longer change ends the game at once: C after the second axis die of round 7, D after the second engine die of round 7, against the brake (or ice-brake) marker at that moment; brakes placed later in round 7 don't help. A and B are checked at the end of the round.
- **App decision:** a `landing` loss lists every failed condition, including the module conditions. When several losses apply at the end of the game, the reported reason is the first of `landing`, `intern`, `iceBrakes`; the others are listed too.

### 7.2 Losses

The game is lost immediately when any of these happens:

| ID | Loss |
|---|---|
| `spin` | Axis reaches ±3 (§6.1). |
| `collision` | Advancing while plane tokens are on the current space (§6.2). |
| `overshoot` | Advancing from the airport (§6.2). |
| `turn` | Entering a turn space with the axis in a forbidden position (§11.2). |
| `missedAirport` | Altitude reaches landing while the plane is not at the airport (§5.5). |
| `mandatory` | End of round with an empty axis or engine space (§5.5). |
| `kerosene` | Kerosene below 0 (§12.1, §12.5). |
| `landing` | Final round ends with A, B, C or D failed (§7.1). |
| `intern` | Intern tokens left at the end of the game (§12.2). |
| `iceBrakes` | Ice-brake marker not past 5 at the end of the game (§12.6). |

- **App decision:** the game-over screen names the loss (and for `landing` the failed conditions). Unplaced dice of both players are revealed.
- **App decision:** any player may **give up** at any time, with confirmation (→ lost). The host can end a game in the lobby, and finished games can be rematched as in the other games. A rematch keeps the scenario, the abilities and the roles.

There is no score. The flight log has a check box per scenario for victories.

- **App decision:** scenario progress (those check boxes) is not tracked in v1.

## 8. Communication

Before rolling, players may discuss strategy freely but never dice values or "put your 6 there". During placement they are silent. Placing a die is the intended way to communicate.

- **App decision:** the app has no chat and cannot enforce the talking rules (players usually sit together or on a call). The discussion phase (§5.1) is the only place the app reminds players that talking is allowed. All public information (placed dice, tracks, tokens, switches, markers, partner's remaining-die count) is on screen.

## 9. Scenarios

Colours give the difficulty: **green** routine, **yellow** exceptional conditions, **red** elite pilots only, **black** heroic. "Abilities" is the number of special-ability cards (§13). Traffic dice and turns (§11) are part of the approach track and apply whenever the track shows them.

| ID | Airport | Colour | Track (§10) | Modules | Abilities |
|---|---|---|---|---|---|
| `yul-green` | YUL Montréal-Trudeau | green | YUL-G | – | 0 |
| `lhr-green` | LHR Heathrow | green | LHR-G | – | 0 |
| `hnd-green` | HND Haneda | green | HND-G | – | 0 |
| `osl-green` | OSL Gardermoen | green | OSL-G | kerosene | 0 |
| `atl-green` | ATL Hartsfield-Jackson | green | ATL-G | intern | 0 |
| `prg-green` | PRG Václav Havel | green | PRG-G | kerosene | 2 |
| `lhr-yellow` | LHR Heathrow | yellow | LHR-Y | intern | 0 |
| `tgu-yellow` | TGU Toncontín | yellow | TGU-Y | kerosene | 2 |
| `gig-yellow` | GIG Galeão | yellow | GIG-Y | wind | 1 |
| `kef-yellow` | KEF Keflavík | yellow | KEF-Y | ice brakes | 1 |
| `prg-yellow` | PRG Václav Havel | yellow | PRG-Y | kerosene leak | 2 |
| `kul-yellow` | KUL Kuala Lumpur | yellow | KUL-Y | kerosene | 1 |
| `atl-yellow` | ATL Hartsfield-Jackson | yellow | ATL-Y | kerosene leak | 1 |
| `pbh-red` | PBH Paro | red | PBH-R | kerosene (~~real-time~~) | 2 |
| `hnd-red` | HND Haneda | red | HND-R | kerosene, intern | 1 |
| `gig-red` | GIG Galeão | red | GIG-R | wind, kerosene leak | 2 |
| `osl-red` | OSL Gardermoen | red | OSL-R | kerosene leak, ice brakes | 2 |
| `tgu-red` | TGU Toncontín | red | TGU-R | kerosene, wind | 2 |
| `kef-black` | KEF Keflavík | black | KEF-B | wind, ice brakes | 2 |
| `kul-black` | KUL Kuala Lumpur | black | KUL-B | kerosene (~~real-time~~) | 2 |
| `pbh-black` | PBH Paro | black | PBH-B | kerosene (~~real-time~~) | 2 |

- `hnd-red` follows the user's printed flight log (kerosene + intern, 1 ability); the web copy lists only the intern. The track header agrees with the printed log.
- **App decision:** the real-time module is not implemented (§12.4). The four scenarios that use it (`pbh-red`, `kul-black`, `pbh-black`, `dus-black`) are played without it; their other modules stay.
- **App decision:** both Düsseldorf promo faces are included as scenarios.

Promo tracks (not in the flight log; modules and abilities from the track headers):

| ID | Airport | Colour | Track | Modules | Abilities |
|---|---|---|---|---|---|
| `dus-yellow` | DUS Düsseldorf | yellow | DUS-Y | kerosene leak | 2 |
| `dus-black` | DUS Düsseldorf | black | DUS-B | kerosene leak (~~real-time~~) | 1 |

## 10. Approach tracks

Spaces are numbered from the **start** (1, the cloud space the plane begins on) to the **airport** (N, the last space). "Planes" is the number of plane tokens placed at setup. "Traffic" is the number of traffic-die icons (§11.1). "Turn" lists the axis positions of the turn icon (§11.2) as 5 characters for positions −2, −1, 0, +1, +2 (left = toward the Pilot): `o` permitted, `x` forbidden. Empty cells are 0 or no icon.

**YUL-G** (N = 7; printed the same on both sides of the YUL track)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|
| Planes | | | 1 | 2 | 1 | 3 | 2 |

**LHR-G** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | | 1 | 1 | 2 | 2 | 2 |
| Traffic | 1 | | 1 | | 1 | |

**LHR-Y** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | 1 | 1 | 1 | 2 | 1 | 2 |
| Traffic | 2 | | 1 | 1 | 1 | |

**HND-G** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | 1 | 1 | 2 | 1 | | 2 | 1 |
| Traffic | 2 | | | | | | | |
| Turn | | | xooxx | | ooxxx | oooxx | | |

**HND-R** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | 1 | | 1 | 1 | 1 | 2 | 1 | 1 |
| Traffic | 3 | | | 1 | 1 | 1 | | |
| Turn | | xooxx | | ooxxx | xooxx | | ooxxx | |

**OSL-G** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | 1 | | 1 | 1 | 1 | |
| Traffic | 2 | | | 1 | | | | |

**OSL-R** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | 1 | | 1 | | | 1 | | |
| Traffic | 3 | | 1 | | 1 | | 1 | |

**ATL-G** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | | | 1 | 2 | 1 | 2 |
| Traffic | 4 | | 1 | | 1 | 1 | | |

**ATL-Y** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | 1 | | 2 | 2 | 1 | 3 | 1 |
| Traffic | 1 | 1 | | 2 | 1 | | | |
| Turn | | | xooox | | | | xxxoo | |

**PRG-G** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | 1 | 1 | | 1 | 1 | 1 |
| Traffic | 1 | | 1 | | | 1 | | |
| Turn | | xooxx | | xxxoo | | | | |

**PRG-Y** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | 1 | 3 | | 3 | 2 | 3 |
| Traffic | | | | 1 | | 1 | | |

**TGU-Y** (N = 5)

| Space | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Planes | | 1 | 1 | 1 | 1 |
| Traffic | 3 | | | | |
| Turn | | ooxxx | xooxx | ooxxx | |

**TGU-R** (N = 5)

| Space | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Planes | | 1 | 1 | 1 | 2 |
| Traffic | 3 | | 2 | | |
| Turn | | ooxxx | xooxx | ooxxx | |

**GIG-Y** (N = 7)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|
| Planes | | | 2 | 1 | 2 | 3 | 1 |
| Traffic | 2 | | | 1 | 1 | | |

**GIG-R** (N = 7)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|
| Planes | | 1 | 2 | 2 | 1 | 1 | 2 |
| Traffic | 3 | | 1 | | 1 | | |

**KEF-Y** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | | | 1 | 1 | 1 | |
| Traffic | 2 | | 1 | | 1 | |

**KEF-B** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | | | 2 | 1 | 1 | |
| Traffic | 2 | | 1 | | | |
| Turn | | oooxx | | xxooo | xooox | |

**KUL-Y** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | 1 | 1 | | 1 | 1 | 1 |
| Traffic | 2 | | | 1 | | | | |
| Turn | | xxxoo | | xxoox | xxxoo | | xxxoo | |

**KUL-B** (N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | | 1 | | 1 | 1 | 1 | 1 |
| Traffic | 3 | | 1 | | | 1 | | |
| Turn | | xxoox | | oxxxx | xooxx | xxoox | xxxoo | |

**PBH-R** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | | 1 | 1 | 1 | 1 | 1 |
| Traffic | 3 | | | 1 | | |
| Turn | | | ooxxx | oooxx | xxxoo | |

**PBH-B** (N = 6)

| Space | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| Planes | 1 | | 1 | 1 | 1 | 1 |
| Traffic | 3 | | 1 | 1 | 1 | |
| Turn | | | ooxxx | ooxxx | xxxxo | |

**DUS-Y** (promo, N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | | 1 | 1 | 2 | 1 | 2 | 1 | 1 |
| Traffic | 3 | | | | | 2 | | |
| Turn | | | | xooox | | xooox | | |

**DUS-B** (promo, N = 8)

| Space | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |
|---|---|---|---|---|---|---|---|---|
| Planes | 1 | | | 1 | 2 | 1 | 3 | 2 |
| Traffic | 3 | | 1 | 1 | | | | |
| Turn | | xxoxx | | ooxxx | | xxxoo | | |

Highest setup count is PRG-Y with 12 plane tokens, the whole supply.

## 11. Approach-track effects

### 11.1 Traffic die

At the start of a round (§5.1 step 2), if the current space has traffic-die icons, roll the traffic die once per icon. For each roll `r`, add a plane token from the supply to the space `r − 1` steps ahead (the current space counts as the first). A space beyond the airport means the airport.

- Spaces passed during a 2-space advance do not trigger. Staying on a traffic space triggers it again every round.
- If the supply is empty, no token is added.
- **App decision:** the server rolls; the result and the new tokens are shown to both players.

### 11.2 Turns

Whenever the plane **moves onto** a space with a turn icon (§6.2 step 4), the axis must be in one of that icon's permitted positions, or the game is lost. On a 2-space advance both entered spaces are checked, in order, with the axis as it is at that moment. If the plane does not move, turn icons have no effect, including the one on the current space.

The start space never has a turn icon; the airport never has one either.

## 12. Modules

### 12.1 Kerosene

- **Setup:** kerosene track left of the panel, marker at **20**. The track runs 20 … 0, then the red cross.
- **Action:** `kerosene` space, either player, any value. Move the marker down by the die's value.
- **End of round** (§5.5 step 2): if no die is on the `kerosene` space this round, lose **6** kerosene.
- Kerosene below 0 (on the cross) loses the game, at any time, including the final round.

### 12.2 Intern

- **Setup:** intern board below the panel: blue die space, a row of 6 token spaces, orange die space. Place the 6 intern tokens (values 1–6) on the row in random order, face up.
- **Training** (on your turn, as your placement): place a die of any value on your colour's intern space. Its value must differ from the **next available token on your side** (the Pilot takes from the left end of the row, the Co-Pilot from the right end). Take that token and place it on any free space you could place a die on, except concentration. It acts there as a die with the token's value. Coffee cannot modify it.
- Each player has one intern space, so each player trains at most once per round.
- If tokens are left on the intern board at the end of the game, the game is lost.
- **App decision:** training and placing the token are one intent in the same turn; the token is never kept for later. The token is shown as its own (green) piece, but it counts as a die for all purposes: it covers its space (mandatory check, §5.5) and counts for Control, Mastery and Synchronisation (§13). Only two exceptions: it cannot go on concentration, and coffee cannot modify it. At the end of the round used tokens are removed from the game.

### 12.3 Wind

- **Setup:** wind ring right of the panel, blue airplane token pointing at the top space (+3).
- **Axis:** right after each axis resolution (§6.1), turn the airplane token by the current axis position, even if the axis did not move this round: toward the Pilot = counter-clockwise, toward the Co-Pilot = clockwise.
- **Engines:** add the wind value the token points at to the engine sum, every round including the final one.
- The ring has 20 spaces (checked against a photo of the physical ring). Store the token as an offset from the top space; with `p = offset mod 20` (0–19), the value depends on `min(p, 20 − p)`:

| `min(p, 20 − p)` | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Wind | +3 | +3 | +2 | +2 | +1 | 0 | −1 | −2 | −2 | −3 | −3 |

### 12.4 Real-time (not implemented)

In the box: a 60-second timer starts right after the roll; when it runs out the round ends and unplaced dice are ignored.

- **App decision:** the app does not implement the real-time module. Scenarios that list it are played without it (§9).

### 12.5 Kerosene leak

- **Setup:** as kerosene (§12.1), but the `kerosene` space is closed.
- Instead of the action and the −6 rule: each round, lose **(difference of the two engine dice) + 1** kerosene. Example: 6 and 3 → lose 4.
- **App decision:** the loss is applied when the second engine die is placed (wind does not count; the dice values after coffee do).

### 12.6 Ice brakes

- **Setup:** the ice-brake board covers the normal brakes (those spaces are gone).
- The board has 4 columns with values **2, 3, 4, 5**, left to right. Each column has a top space (blue: Pilot only) and a bottom space (blue and orange: either player), both of exactly the column's value. Between the rows runs the marker track: left of 2, then one slot after each column, the last one "past 5".
- **Open columns:** any column to the right of the marker is open; both its spaces take dice. Columns the marker has passed are closed.
- **Pair:** a column is *paired* when both its spaces hold a die in the same round.
- **Deployment, left to right:** after every placement, while the column directly right of the marker is paired, the marker moves past it (one slot per column; several columns in one round are allowed).
- **A die on a column that is not next in line** stays there. If that column becomes paired, it waits: it deploys later this round only if every column between the marker and it gets paired too, and then the marker moves through all of them in order.
- **End of round:** all dice on columns the marker has not passed are wasted, whether alone or paired; the marker does not move for them. Those columns are open and empty again next round.
- The marker **must be past 5** at the end of the game, or the game is lost. The final speed must be lower than the marker, i.e. **5 or less**.
- Example: the marker is left of 2. The Pilot puts a 4 on column 4 (top), the Co-Pilot a 4 below it: column 4 is paired but waits. Later the Pilot places a 2 and the Co-Pilot a 2 on column 2: the marker moves past 2. Column 3 is empty, so column 4 still waits; at the end of the round its two dice are wasted.

## 13. Special abilities

A scenario gives 0, 1 or 2 cards (§9). The players choose which cards from the 6. Texts transcribed from the cards (German edition; German name in brackets).

- **App decision:** the host picks the cards in the lobby, exactly the scenario's number. The lobby accepts a partial selection while the host chooses; *Start* requires the full number. Both players see the chosen cards in the lobby and during the game.

| ID | Name | Effect |
|---|---|---|
| `adaptation` | Adaptation (Flexibilität) | Once per game, each player may turn one of their unplaced dice to its opposite face (`v → 7 − v`). |
| `anticipation` | Anticipation (Vorahnung) | Once per round, before the first die of the round is placed, the first player of the round may reroll one of their dice. |
| `control` | Control (Ruhiges Händchen) | When both axis spaces hold dice of the same value, gain a coffee token immediately (if fewer than 3 are on the panel). |
| `mastery` | Mastery (Präzision) | When both engine spaces hold dice of the same value, gain a reroll token immediately if one is available: not on the altitude track and not in the supply (2 tokens exist). (Dized clarification) |
| `synchronisation` | Synchronisation (Teamwork) | When, in the same round, at least one die is on the landing gear and one on the flaps, roll the traffic die immediately. The Co-Pilot places it on an empty space of their choice, of any colour, respecting the space's value constraint, and it resolves like a normal die. It is an extra action. |
| `workingTogether` | Working Together (Hilfsbereitschaft) | Once per round, at any time, one player puts one of their unplaced dice on the card. The other must then put one of theirs there too (if they have any left). The two dice swap values and go back behind their owners' screens. |

- **App decision (Control, Mastery):** everything on the space counts: players' dice, intern tokens (§12.2) and the Synchronisation traffic die.
- **App decision (Synchronisation):** it triggers at most once per round, the moment the round has both a gear die and a flaps die. The traffic die follows all normal rules for its space (value constraints, ordered flaps and brakes, switches already green); concentration is allowed. Two exceptions: coffee cannot modify it, and it cannot go on an intern space. If no space is legal, the die is discarded. An intern token on the gear or flaps counts toward the trigger. The Co-Pilot places it right away, before the next turn; it covers its space like any die (§5.5).
- **App decision:** Adaptation and Working Together are intents outside the turn order; Anticipation is offered to the first player before their first placement. Working Together pauses placement until the partner has chosen their die. The partner sees the value of the die the initiator put on the card (it is public, as on the physical card). It cannot be started when the partner has no unplaced dice.
- **App decision:** the side intents (spend a reroll, Adaptation, Working Together, Anticipation) are allowed only during placement (step `place`), not during discussion, a reroll or a swap.

## 14. Player count

Exactly 2 players. There is no solo or 3+ variant.

## 15. App decisions without a rule counterpart

- **Physical actions:** sliding the tracks is a position index; the screens are the per-player view; the server rolls every die (player dice, rerolls, traffic die) and shuffles the intern tokens.
- **Automatic resolution:** axis, engines, approach movement, radio, switches, markers, coffee gain, kerosene and wind resolve automatically when the triggering die is placed. A reroll pickup and traffic rolls run at round start.
- **No undo.** Every placement is final.
- **Hidden information:** a player sees only their own unplaced dice. Their partner's unplaced dice and reroll selection stay hidden until revealed by the rules. The die a player puts on the Working Together card is public; the partner's answer stays hidden (it goes back behind the initiator's screen with a new value).
- **No timers:** nothing in the app is timed (the real-time module is left out, §12.4).
