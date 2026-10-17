# Mission Deep Sea — rules reference

Status: from the rulebook, the fan mission sheet v1d and the community task-card list (see
[plan.md](plan.md) §0); the user's answers of 2026-10-11 are in. This is the source of truth for Deep
Sea. Interpretations we chose are marked **App decision**. Where nothing is said, Planet Nine's [../rules.md](../rules.md) applies (tricks
§5, communication §8, distress signal §9, two players §12, attempts §15, app decisions §16).

Terms: submarine = trump (Planet Nine: rocket), captain = commander, Tonoja = the dummy (JARVIS),
sonar token = radio token.

## 1. Components

40 playing cards: pink, blue, green, yellow 1–9 and submarines 1–4 (trump). 96 task cards, each with
a difficulty for 3, 4 and 5 crew members. 5 sonar tokens, 1 distress signal token, captain token.
No task tokens.

## 2. Setting up an attempt

1. Deal all 40 cards. With 3 players one member gets 14 cards, so one card is left over (unplayed)
   after the last trick, as in Planet Nine. 2 players: §7.
2. Draw tasks: take cards from the shuffled task deck; a card whose difficulty (column for the number
   of crew members; 2 players use the 3 column) would overshoot the mission's total is skipped; stop
   when the total is reached exactly. Reveal them.
3. The holder of submarine 4 is the captain.
4. Re-deal check (§6, task cards with a re-deal note): if an alignment makes a displayed task
   impossible, re-deal the playing cards; this is not an attempt.
5. Selection (§3), then predictions (cards 90/91), then the distress signal may be sent (before any
   communication), then the captain leads the first trick.

**App decision:** prefer unused task cards across attempts and missions of a game; reshuffle the used
ones when the deck runs out.

## 3. Task selection

- The captain picks first, then clockwise, round after round, until every task is taken. Members may
  hold several tasks; not everyone needs one.
- **Passing:** allowed only if, at the start of the selection, there are fewer tasks than members
  (this includes the captain). After one round every task must be assigned, so a member may not pass
  when the remaining members could not take the remaining tasks.
- **Captain comparison tasks** (cards 4–6): the captain may never take them. **App decision:** a
  member who may take none of the remaining tasks (the captain facing only cards 4–6) passes, even
  where passing is otherwise not allowed; the draft then goes on round after round until every task
  is taken. The captain is never chosen to take all tasks (6, 10, 13, 14–16) when one of them is a
  captain comparison task.
- **Mission 19 App decision:** the captain's first task must be one of the highest difficulty among
  the tasks the captain may take.
- Mission rules may replace this (§8).
- **Impossible combinations** (rulebook): if two displayed tasks can't both be met by different
  members and there are too few tasks for one member to take both, replace the most recently
  revealed one with another of the same difficulty. **App decision:** no automatic check. Before the
  first pick, the captain may **replace** any displayed task: it goes back into the deck and the next
  card of the same difficulty (for this player count) takes its place. Any number of times, not an
  attempt. Not in missions with fixed tasks (32) or face-down tasks (Q2).
- **After a failed attempt:** the crew may keep the same tasks or draw new ones. **App decision:**
  the host chooses in the debrief.
- **Unavoidable failures** (rulebook rule of thumb: a loss you could not have avoided is not a failed
  attempt). **App decision:** after a loss the host may choose **don't count this attempt**; the
  logbook's attempt count stays as it was.
- **Predictions (90, 91) App decision:** after the selection the owners predict one after the
  other, in task order (the captain for Tonoja), 0 to the number of tricks. A hidden prediction is
  seen only by its owner (and the captain for Tonoja) until the debrief.

## 4. Communication

As Planet Nine §8 (one sonar communication per member per attempt; only after selection; only before
a trick; a colour card that is your highest / only / lowest of its colour; never a submarine).
Mission modes:

- **Normal.**
- **Currents** (`?`): communicate without placing the token on the card (= Planet Nine dead zone).
- **Rapture of the deep** (`-2`): put n−2 sonar tokens in the middle; whoever wants to communicate
  takes one without discussion. A member may communicate several times while tokens are left. No
  personal tokens. With 2 players Tonoja counts as a member: 3 − 2 = 1 shared token (Tonoja never
  communicates).
- **Unfamiliar terrain** (grid 1-2-3 / 4-5-6 / 7-8-9): before dealing, draw a random colour card
  (redraw on a submarine); 1–3 normal, 4–6 currents, 7–9 rapture; mix it back in. Per attempt.
- **No communication** (16 untimed).
- Mission 23: nobody communicates before trick 2. Missions 10/13 when the tasks are passed on: all
  communication must happen before the first trick.

## 5. Distress signal

As Planet Nine §9: before the first trick and before any communication, each member passes one
non-submarine card to the same side; the token stays active for all further attempts of the mission;
the mission's logbook attempts +1.

## 6. Objectives: general rules

- A task applies only to its owner. A task is fulfilled once met and no longer able to fail; the
  mission fails as soon as any task fails. All tasks fulfilled → mission won.
- "Win a card" means the card is in a trick the owner won.
- "Exactly" counts at the end of the attempt; "at least" may be exceeded.
- Tasks about card values of a trick (sums, all below/above, odd/even): no submarine may be in the
  trick.
- Multi-coloured pictures on cards mean the four colours, never submarines.
- "Don't open a trick with X": the owner never leads a trick with a card of the colours shown.
- "More X than Y cards": among the cards won at the end; zero Y allowed (then one X suffices).
- "Same number of X and Y": at least one of each.
- Predictions: after selection the owner notes how many tricks they will win (open: told to all;
  hidden: secret until the end); fulfilled if exactly right.
- Re-deal notes (submarine tasks): if the named alignment is in one hand, re-deal before play.

## 7. Two players: Tonoja

As Planet Nine §12 (JARVIS): set aside submarine 4; 7 cards face down in a row, 7 face up on top;
shuffle submarine 4 back in, deal 13 each. Tonoja is a third crew member (3-player difficulty
column); the captain picks her tasks and plays her face-up cards, without discussion. A face-down
card is turned up after a trick once the card on top of it is played.

## 8. Missions

Source: fan mission sheet v1d. *D* = total difficulty. Every mission allows the distress signal.
Free selection = the crew discusses who takes what (no talk about hands).
**App decision:** the captain makes every "decide together" call in the app and assigns the tasks in free selection (as in Planet Nine); the crew coordinates outside the game.
**App decision (numbering):** the side quests come after the epilogue in the app's mission list: Q1 is mission 42, Q2 mission 43 (shown as "Q1", "Q2").
**App decision (Q1, confirmed at WP5):** stage B is a new deal after stage A is won, the captain (submarine 1) again has to win the tricks; a lost stage B is retried as stage B. With 2 players submarine 1 is set aside for a human (as submarine 4 otherwise), so Tonoja is never captain.

| # | D | Comm | Selection / special rules |
|---|---|---|---|
| 1 | 1 | normal | |
| 2 | 2 | normal | |
| 3 | 3 | normal | |
| 4 | 4 | normal | |
| 5 | 5 | normal | |
| 6 | 5 | normal | All tasks to one crew member, decided together (no talk about hands). |
| 7 | 6 | normal | |
| 8 | — | normal | No tasks. Crew objective: at no point may a member have won two more 9s than any other member. |
| 9 | 7 | currents | |
| 10 | 4 | normal | The captain takes all tasks or passes them to a willing member; if passed on, all communication must happen before the first trick. |
| 11 | 8 | rapture | |
| 12 | — | normal | No tasks. Crew objective: no trick may be led with a pink card or a submarine. |
| 13 | 5 | normal | As 10. |
| 14 | 6 | normal (untimed: currents) | One volunteer takes all tasks. Timed 3:30; untimed → currents. |
| 15 | 6 | normal (untimed: rapture) | One volunteer. Timed 3:00; untimed → rapture. |
| 16 | 6 | normal (untimed: none) | One volunteer. Timed 2:30; untimed → no communication. |
| 17 | 9 | normal | Free selection. |
| 18 | 9 | normal | |
| 19 | 9 | normal | The captain must take the displayed task with the highest difficulty as the first task. |
| 20 | 10 | unfamiliar | |
| 21 | — | unfamiliar | No tasks. Crew objective: at no point may a member have won two more 1s than any other member. |
| 22 | 11 | unfamiliar | |
| 23 | — | unfamiliar | No tasks. Crew objective: whoever wins the first trick must have won more tricks than every other member at every point. No communication before trick 2. |
| 24 | 12 | unfamiliar | |
| 25 | 12 | unfamiliar | The captain is skipped in the selection (receives no tasks). |
| 26 | 10 (untimed: 12) | unfamiliar | Two volunteers; they split the tasks, each takes at least one. Timed 5:00; untimed → D 12. |
| 27 | 5 | unfamiliar | Crew objective: the yellow 5 must be the last card played in the last trick. |
| 28 | 14 | normal | Free selection. |
| 29 | 15 | normal | Free selection. |
| 30 | 16 | normal | Free selection. |
| 31 | 17 | normal | Free selection. |
| 32 | — | normal | Fixed tasks, selected as usual: "I will win 0 tricks", "I will win exactly 3 tricks, in a row", "I will win 2 tricks in a row", "I will win the first & last trick". |
| Q1 | — | normal | Side quest, no task cards. The holder of **submarine 1** is the captain (not submarine 4). Re-deal if the captain has all four submarines (not an attempt). Stage A: the captain must win exactly 2 tricks and none of them may contain a 7, 8 or 9. Once A is won, stage B (a new deal): the same with exactly 3 tricks. **App decision:** the two stages are played one after the other; Q1 is complete in the logbook when B is won; attempts of both stages add up. |
| Q2 | 3p 13 / 4p 10 / 5p 8 | normal | Side quest. Task cards are drawn as usual but stay **face down**: only their difficulty (the back) is visible. Selection as usual. Each member must win exactly as many tricks as the sum of the difficulties of the cards they selected (no cards: no tricks). The totals equal the number of tricks (13 / 10 / 8). 2 players: the 3-player value. |
| 33–41 | 18…26 | normal | Epilogue: free selection, attempts are not counted, the distress signal is tracked. |

**Volunteer missions (14, 15, 16, 26):** the captain asks each member clockwise once; only "yes" or
"no". The first "yes" (26: the first two) takes all tasks. The tasks must be assigned after one round:
if too few said yes, the last members asked must take them (rulebook example for 26: with 4 players,
if the first two say no, member three and the captain take them).

**App decision (as Planet Nine's poll, its rules.md §10):** every human except the captain answers
yes/no, in any order, visible to all; then the captain chooses the member(s), themselves and Tonoja
included: one for 14–16, two for 26 (then the captain assigns the tasks between the two, each at
least one). The app does not enforce "first yes in clockwise order"; the crew coordinates outside the
game. Missions 10/13 work the same way: poll, then the captain chooses any member, themselves
included; if it is not the captain, communication is allowed only before the first trick.

**Real-time missions:** **App decision:** no timers; always the untimed alternative.

## 9. Task cards

Difficulty = 3/4/5 crew members. *Kind* is the proposed objective kind (plan.md §3); existing Planet
Nine kinds in `code`, new ones in **bold**. Card numbers are our own (list order of the source).

### Trick counts and comparisons

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 1 | 2/3/3 | Win more tricks than everyone else | **cmpTricks** strictly more than each other member |
| 2 | 3/4/5 | Win more tricks than everyone else together | **cmpTricks** more than the sum of all others |
| 3 | 2/2/3 | Win fewer tricks than everyone else | **cmpTricks** fewer than each other member |
| 4 | 2/2/3 | Win more tricks than the captain (not for the captain) | **cmpCaptain** more |
| 5 | 2/2/2 | Win fewer tricks than the captain (not for the captain) | **cmpCaptain** fewer |
| 6 | 4/3/3 | Win the same number of tricks as the captain (not for the captain) | **cmpCaptain** equal |
| 74 | 4/3/3 | Don't win any tricks | `winNoTricks` |
| 83 | 3/2/2 | Win exactly one trick | `winExactlyTricks` 1 |
| 84 | 2/2/2 | Win exactly two tricks | `winExactlyTricks` 2 |
| 87 | 2/3/5 | Win exactly four tricks | `winExactlyTricks` 4 |
| 90 | 3/2/2 | Win X tricks (predict the exact number and show it) | **predict** open |
| 91 | 4/3/3 | Win X tricks (predict the exact number, keep it hidden) | **predict** hidden |

### Specific tricks and sequences

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 71 | 1/2/3 | Don't win any of the first four tricks | **avoidTricks** 1–4 |
| 72 | 1/2/2 | Don't win any of the first three tricks | **avoidTricks** 1–3 |
| 73 | 2/3/3 | Don't win any of the first five tricks | **avoidTricks** 1–5 |
| 75 | 3/2/2 | Don't win two consecutive tricks | **noConsecutive** |
| 76 | 2/3/3 | Win the last trick | `winTricks` [last] |
| 77 | 2/3/4 | Win the first three tricks | `winTricks` [1,2,3] |
| 78 | 1/1/2 | Win the first two tricks | `winTricks` [1,2] |
| 79 | 1/1/1 | Win the first trick | `winTricks` [1] |
| 80 | 3/4/4 | Win the first and the last trick | `winTricks` [1,last] |
| 81 | 4/4/4 | Win only the last trick | `winOnlyTricks` [last] |
| 82 | 4/3/3 | Win only the first trick | `winOnlyTricks` [1] |
| 85 | 1/1/1 | Win two consecutive tricks | **consecutive** ≥2 in a row (more tricks allowed) |
| 86 | 2/3/4 | Win three consecutive tricks | **consecutive** ≥3 in a row |
| 88 | 3/3/4 | Win exactly three consecutive tricks | **consecutive** exact: exactly 3 tricks in total, all in a row (as mission 32: "exactly 3 tricks, in a row") |
| 89 | 3/3/3 | Win exactly two consecutive tricks | **consecutive** exact: exactly 2 tricks in total, in a row |

### Winning a trick in a certain way

"Win a trick with a 6" = the card that wins the trick is a 6 (rulebook: "I will win a trick using
a 2"). For 12–14 the winning card has the second rank and the first rank is another card in the
same trick, any colours ("any 6 with another 6": two 6s in the trick, one of them wins).

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 9 | 2/3/3 | Win a trick with a 6 | **winWith** using rank 6 |
| 10 | 2/3/4 | Win a trick with a 5 | **winWith** using rank 5 |
| 11 | 3/4/5 | Win a trick with a 3 | **winWith** using rank 3 |
| 15 | 3/4/5 | Win a trick with a 2 | **winWith** using rank 2 |
| 12 | 1/2/2 | Win a 5 with a 7 | **winWith** using rank 7, target rank 5 |
| 13 | 3/4/5 | Win an 8 with a 4 | **winWith** using rank 4, target rank 8 |
| 14 | 2/3/4 | Win any 6 with another 6 | **winWith** using rank 6, target rank 6 |
| 58 | 3/3/3 | Win the pink 7 with a submarine | **winWith** using trump, target pink 7 |
| 59 | 3/3/3 | Win the green 9 with a submarine | **winWith** using trump, target green 9 |

### Trick values (no submarine in the trick)

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 7 | 2/3/3 | Win a trick where all cards are lower than 7 | **winTrickWhere** all < 7 |
| 8 | 2/3/4 | Win a trick where all cards are greater than 5 | **winTrickWhere** all > 5 |
| 46 | 2/5/6 | Win a trick that has only even numbers (2, 4, 6, 8) | **winTrickWhere** all even |
| 47 | 2/4/5 | Win a trick that has only odd numbers (1, 3, 5, 7, 9) | **winTrickWhere** all odd |
| 48 | 3/3/4 | Win a trick with a total value of at least 23/28/31 (3/4/5 players) | **winTrickWhere** sum ≥ (by players; 2 players: 23) |
| 49 | 3/3/4 | Win a trick with a total value lower than 8/12/16 (3/4/5 players) | **winTrickWhere** sum < (2 players: 8) |
| 50 | 3/3/4 | Win a trick with a total value of 22 or 23 | **winTrickWhere** sum ∈ {22, 23} |

### Specific cards

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 16 | 1/1/1 | Win the pink 3 | `winCards` |
| 17 | 1/1/1 | Win the yellow 1 | `winCards` |
| 18 | 1/1/1 | Win the blue 4 | `winCards` |
| 19 | 1/1/1 | Win the green 6 | `winCards` |
| 27 | 2/3/3 | Win the blue 1, 2 and 3 | `winCards` |
| 28 | 2/2/3 | Win the blue 6 and the yellow 7 | `winCards` |
| 29 | 2/2/3 | Win the pink 5 and the yellow 6 | `winCards` |
| 30 | 2/2/3 | Win the green 5 and the blue 8 | `winCards` |
| 31 | 2/2/3 | Win the blue 5 and the pink 8 | `winCards` |
| 32 | 2/2/3 | Win the pink 9 and the yellow 8 | `winCards` |
| 33 | 2/2/2 | Win the pink 1 and the green 7 | `winCards` |
| 34 | 2/3/3 | Win the yellow 9 and the blue 7 | `winCards` |
| 35 | 3/4/4 | Win the green 3, the yellow 4 and the yellow 5 | `winCards` |
| 36 | 3/4/5 | Win the green 2 in the last trick | `cardInLastTrick` owned by the member: the owner wins the last trick and it contains green 2 |
| 54 | 1/1/1 | Win submarine 3 | `winCards` |

### Counts of a kind (won cards at the end)

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 20 | 3/4/5 | Win all four 3s | **winCount** rank 3 = 4 |
| 21 | 3/4/5 | Win at least three 5s | **winCount** rank 5 ≥ 3 |
| 22 | 3/4/5 | Win at least three 9s | **winCount** rank 9 ≥ 3 |
| 23 | 2/2/2 | Win at least two 7s | **winCount** rank 7 ≥ 2 |
| 24 | 4/5/6 | Win all four 9s | **winCount** rank 9 = 4 |
| 25 | 3/4/4 | Win exactly three 6s | **winCount** rank 6 = 3 |
| 26 | 2/3/3 | Win exactly two 9s | **winCount** rank 9 = 2 |
| 37 | 4/4/4 | Win exactly one pink and one green card | **winCount** pink = 1 and green = 1 |
| 38 | 3/3/3 | Win at least seven yellow cards | **winCount** yellow ≥ 7 |
| 39 | 2/3/3 | Win at least five pink cards | **winCount** pink ≥ 5 |
| 40 | 3/4/4 | Win exactly two green cards | **winCount** green = 2 |
| 41 | 3/4/4 | Win exactly two blue cards | **winCount** blue = 2 |
| 42 | 3/3/4 | Win exactly one pink card | **winCount** pink = 1 |
| 43 | 2/2/2 | Win no pink cards | **winCount** pink = 0 |
| 62 | 2/2/2 | Don't win any green cards | **winCount** green = 0 |
| 63 | 2/2/2 | Don't win any yellow cards | **winCount** yellow = 0 |
| 64 | 3/3/3 | Don't win any pink or blue cards | **winCount** pink+blue = 0 |
| 65 | 3/3/3 | Don't win any yellow or green cards | **winCount** yellow+green = 0 |
| 66 | 3/3/2 | Don't win any 8s or 9s | **winCount** ranks 8,9 = 0 |
| 67 | 1/1/1 | Don't win any 9s | **winCount** rank 9 = 0 |
| 68 | 1/2/2 | Don't win any 5s | **winCount** rank 5 = 0 |
| 69 | 2/2/2 | Don't win any 1s | **winCount** rank 1 = 0 |
| 70 | 3/3/3 | Don't win any 1s, 2s or 3s | **winCount** ranks 1–3 = 0 |
| 44 | 2/3/4 | Win at least one card of each colour (not submarines) | **winEachColour** |
| 45 | 3/4/5 | Win all cards of at least one colour (not submarines) | **winAllOfAnyColour** |

Ranks count colour cards only (a "9" is never a submarine).

### Submarines

| # | 3/4/5 | Task | Kind | Re-deal if one hand holds |
|---|---|---|---|---|
| 51 | 3/3/3 | Win exactly one submarine | **winCount** trump = 1 | all four submarines |
| 52 | 3/3/3 | Win submarine 1 and no other | **winCount** trump = 1, must include sub 1 | subs 1 and 4, or 1, 2 and 3 |
| 53 | 3/3/3 | Win submarine 2 and no other | **winCount** trump = 1, must include sub 2 | subs 2 and 4, or 1, 2 and 3 |
| 55 | 3/3/4 | Win exactly two submarines | **winCount** trump = 2 | subs 2, 3 and 4 |
| 56 | 3/4/4 | Win exactly three submarines | **winCount** trump = 3 | all four submarines |
| 57 | 1/1/1 | Win no submarines | **winCount** trump = 0 | |

### Leading

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 60 | 4/3/3 | Don't open a trick with a pink, yellow or blue card | **neverLead** pink, yellow, blue |
| 61 | 2/1/1 | Don't open a trick with a pink or green card | **neverLead** pink, green |

### Colour comparisons

| # | 3/4/5 | Task | Kind |
|---|---|---|---|
| 92 | 4/4/4 | Win the same number of pink and yellow cards (more than 0) | **equalSuits** pink, yellow, overall |
| 93 | 2/3/3 | Win the same number of green and yellow cards in one trick (more than 0) | **equalSuits** green, yellow, in one trick |
| 94 | 2/3/3 | Win the same number of pink and blue cards in one trick (more than 0) | **equalSuits** pink, blue, in one trick |
| 95 | 1/1/1 | Win more yellow cards than blue cards (0 blue allowed) | **moreSuit** yellow > blue |
| 96 | 1/1/1 | Win more pink cards than green cards (0 green allowed) | **moreSuit** pink > green |

Count check: 12 + 15 + 9 + 7 + 15 + 25 + 6 + 2 + 5 = 96.

### Mission-level crew objectives (not task cards)

| Mission | Objective | Kind |
|---|---|---|
| 8 | Never two 9s more than another member | `crewBalance` generalised with a card filter (rank 9), max lead 1 |
| 21 | Never two 1s more than another member | as 8, rank 1 |
| 12 | No trick led with pink or a submarine | **neverLead** owned by the crew |
| 23 | First-trick winner always has strictly more tricks than every other member | **leaderStaysAhead** |
| 27 | Yellow 5 is the last card played in the last trick | **lastCardPlayed** (with 3 players, fails if it is the left-over card) |

8/21: a lead of 2 cards fails, a lead of 1 is allowed (check how Planet Nine's `crewBalance` counts).
