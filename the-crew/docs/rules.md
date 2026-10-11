# Rules reference

English summary of the official rules of **The Crew: The Quest for Planet Nine** (Thomas Sing, Kosmos 2019). The code follows this file. Where the official rules are ambiguous or the app must decide something, the choice is recorded under **App decision**.

Sources:
- The English rulebook (Thames & Kosmos 2019, 22 pages). The logbook with the mission texts is not part of it.
- The missions come from the community **mission sheet v1c** by David Fox (10 July 2020), which condenses the logbook into icons and one line of text per mission. §14 transcribes it.
- For the outlook only: the rulebook of the sequel *Mission Deep Sea* (Kosmos 2021). It is not implemented yet. The differences are in [implementation-plan.md](implementation-plan.md#11-deep-sea-outlook).

**Licensing:** the app runs behind the site's password gate for friends only. It uses its own card art, icons and wording, with no Kosmos imagery or logbook story text. The name "The Crew" appears only as the game's title.

---

## 1. Overview

A cooperative trick-taking game for **2–5 players** (2 players use the dummy crew member JARVIS, §12). The crew plays **missions 1–50** in order or freely. Each mission sets **tasks**: a task names a playing card that its owner must win in a trick, sometimes with an order. Some missions add **special rules**. The crew wins a mission only if every task and special rule is fulfilled. If one fails, the attempt is lost and the crew tries again with a new deal. A logbook counts the attempts per mission.

Players may not talk about their cards. The only information channel is one **communication** per player per attempt (§8).

## 2. Components

- **40 playing cards:** four colours, **pink, blue, green, yellow**, each with values **1–9** (36 colour cards), plus **4 rocket cards** with values **1–4**. Rockets are trump.
- **36 task cards:** one per colour card (pink 1 … yellow 9). There are no task cards for rockets.
- **10 task tokens:** `1` `2` `3` `4` `5` (fixed position), `Ω` (last), and the arrows `›` `››` `›››` `››››` (relative order).
- One communication token per player, a distress signal token and a commander token. In the app these are state, not objects.

**App decision:** each colour has its own glyph next to its colour (the rulebook also gives each colour a symbol for colour-blind players). The glyphs are our own design.

## 3. Seats and directions

- **App decision:** the seat order is the join order. Play goes **clockwise**, which is the next seat. "Left neighbour" means the next seat, "right neighbour" the previous seat. Both wrap around.
- With 2 players JARVIS takes the third seat, after the second player (§12).

## 4. Setting up an attempt

Every attempt of a mission starts fresh:

1. Shuffle the 40 cards and deal them all:

   | Players | Cards per player | Tricks | Unplayed cards |
   |---|---|---|---|
   | 2 (+ JARVIS) | 13 each; JARVIS 14 (§12) | 13 | 1 (JARVIS's last card) |
   | 3 | 14, 13, 13 | 13 | 1 |
   | 4 | 10 | 10 | 0 |
   | 5 | 8 | 8 | 0 |

   With 3 players, one player gets an extra card, and the card left after the last trick stays unplayed. **App decision:** the extra card goes to a random seat.
2. Whoever holds the **rocket 4** is the **commander** for this attempt. The commander is still a crew member. Their duties: start the task selection, lead the first trick, and carry out the mission's special rules (§10, §13).
3. Draw the mission's task cards, place its task tokens, and distribute the tasks (§6, §7, §10).
4. Before the first trick: the distress signal (§9), the five-player handover (§11), and communication (§8) become possible.
5. The commander leads the first trick. The winner of each trick leads the next one.

## 5. Tricks

- Each crew member plays one card, in seat order, starting with the leader.
- **Follow suit:** you must play a card of the led suit if you have one. Rockets are a suit too: if a rocket is led, you must follow with a rocket if you can. If you cannot follow, play any card, including a rocket.
- **Winner:** the highest rocket in the trick, or if there is none, the highest card of the led colour. A card of another colour never wins.
- You are never forced to win a trick (no obligation to play high).
- Won tricks are set aside. Only the **last completed trick** may be looked at again. How many tricks each crew member has won is public.

**App decision:** the screen shows the cards of the last completed trick until the next trick's first card is played, and the number of tricks each member has won. It never shows the cards of earlier tricks.

## 6. Tasks

- **Drawing:** the mission says how many task cards to draw (the number in the blue box on the mission sheet). They are drawn from the shuffled 36 and laid out face up from left to right.
- **Open distribution (default):** the commander takes one task card first. Then, clockwise, each crew member takes one, round after round, until all are taken. Members may end up with different numbers of tasks. Passing is not allowed.
- **Fulfilling:** a task is fulfilled when its owner wins a trick containing the task's card. One trick can fulfil several tasks, even of different members.
- **Failing:** if anyone else wins a trick containing a task's card, the attempt is lost immediately. If the last trick has been played and a task is still open, the attempt is lost. This includes a task card that stays unplayed (§4).
- Missions with a task count and no other distribution symbol use open distribution. Other distribution modes are in §10.

## 7. Task tokens

When a mission shows tokens, place them first. The drawn task cards are then laid out from left to right starting at the first token, so the first drawn cards carry the tokens in the order shown on the sheet. Whoever takes a task card also takes its token.

| Token | Condition |
|---|---|
| `1` … `5` | This task must be the 1st … 5th task fulfilled, counting all tasks of the mission (also those without a token). |
| `Ω` | This task must be the last task fulfilled. |
| `›` `››` `›››` `››››` | Relative order among arrow tasks only: `›` before `››` before `›››` before `››››`. Tasks without an arrow may be fulfilled at any time. |

- Token `1` does not mean "in the first trick". It means "first of all tasks".
- Breaking an order condition loses the attempt immediately.
- **Same trick:** several tasks fulfilled in one trick count as fulfilled **at the same time**. The rulebook allows this for consecutive tokens (`1`+`2`, `›`+`››`).
- **App decision:** we generalise the same-trick rule. The tasks fulfilled in one trick may be put in any order among themselves, and the order conditions hold if at least one such order satisfies all tokens. Formally, for a task fulfilled in trick *t*, let *b* be the number of tasks fulfilled before trick *t* and *s* the number fulfilled in trick *t*:
  - **Numbered `k`:** valid iff *b* < *k* ≤ *b* + *s*.
  - **`Ω`:** valid iff no task is still open after trick *t*.
  - **Arrows:** valid iff the trick of `›` ≤ the trick of `››` ≤ the trick of `›››` ≤ the trick of `››››`.
  - An open numbered task fails as soon as *k* or more other tasks are fulfilled. An open `Ω` task fails as soon as it is fulfilled while other tasks are still open. An open arrow task fails as soon as a later arrow is fulfilled before it.

## 8. Communication

- Each **human** crew member may communicate **once per attempt**. JARVIS never communicates, since its face-up cards are public anyway.
- **When:** only after all tasks are distributed, and only **before a trick**, never during one.
  - **App decision:** "before a trick" means no card of the current trick has been played yet, and the game is not waiting for another decision (task selection, distress pass, …). Whoever acts first wins: once the leader plays, the window closes for that trick.
- **What:** put one **colour card** from your hand face up in front of you (never a rocket) and say which of these it is:
  - **highest:** your highest card of that colour (token at the top),
  - **only:** your only card of that colour (token in the middle),
  - **lowest:** your lowest card of that colour (token at the bottom).
  - The statement must be true when it is made. **App decision:** a card that is your only card of its colour must be communicated as **only**; "highest" and "lowest" require at least two cards of that colour.
- The communicated card stays part of your hand and is played normally later. The token is never moved, even if the statement stops being true. Once the card is played, the display disappears, and you have still used your communication.
- **Dead zone** (mission symbol `?`): you communicate a card as usual, and it must still meet one of the three conditions, but you do not say which one. The others must guess. **App decision:** the server checks that the card qualifies and shows the card without a position.
- **Disruption** (lightning symbol with a number *N*): nobody may communicate before trick *N*. Communicating directly before trick *N* is allowed.
- **Silenced member** (mission 11): the member chosen by the commander may not communicate this attempt.

## 9. Distress signal

- After the playing cards and tasks are distributed and before **anyone has communicated** in the attempt, the crew may send a distress signal.
  - **App decision:** the **commander** decides, and chooses the direction, left or right. The players coordinate outside the app.
- Then **every** human crew member passes exactly **one card** to their neighbour in that direction, all in the same direction. **Rockets may never be passed.**
  - **App decision:** each member chooses their card secretly. When all have chosen, the cards move at once. Each player sees which card they gave and which they received. The others see only that cards were passed.
- The signal stays active until the mission is completed. At the start of each later attempt of that mission, the commander may again choose to pass, in either direction, or not to pass, at no extra cost.
- **Logbook:** if the signal was used for a mission, its attempt count is increased by one when the mission is completed.
- **App decision (2 players):** JARVIS does not take part. The two humans each give one non-rocket card to the other, so direction does not matter.

## 10. Commander's special distributions and polls

Some missions replace open distribution, or add a special objective whose owner the commander chooses. These steps use a **poll**.

- **Poll:** the commander asks each crew member a question that may only be answered with yes/no. Mission 5 asks "well" or "ill".
  - **App decision:** every human crew member except the commander answers, in any order. Answers are visible to everyone as they come in. JARVIS does not answer. Then the commander makes the choice.
- **Commander's decision** (commander → *n* "face down"; missions 20, 27, 37): *n* task cards are drawn face down. After the poll, the commander chooses **one other crew member**, never themselves, JARVIS allowed. That member receives **all** tasks, which are then revealed.
- **Commander's distribution** (commander with three arrows; missions 24, 32, 36, 43): task cards are revealed **one at a time**. For each one, poll, then the commander gives it to any crew member, themselves and JARVIS included.
  - At the end, no member may have **two tasks more** than another. **App decision:** the server rejects an assignment after which an even split is no longer reachable.
  - With tokens (mission 36), the first revealed card carries the first token, and so on.
- **Chosen member** (missions 5, 33, 41): after the poll, the commander chooses **another** crew member, JARVIS allowed, who receives the special objective (§13).

## 11. Five-player handover

With **5 players**, in missions with a **golden frame** (§14, column "5p"), one crew member may hand **one** of their task cards (with its token) to another crew member. This happens after task distribution and before the first card of the first trick, at most once per attempt. It does not count as an extra attempt.

**App decision:** the commander carries it out (choose a task, choose the receiver) or skips it. The players coordinate outside the app.

## 12. Two players: JARVIS

1. Put the rocket 4 aside. Shuffle the other 39 cards.
2. Deal JARVIS **7 face-down cards** in a row and **7 face-up cards** on top of them (7 stacks of 2).
3. Shuffle the rocket 4 into the remaining 25 cards and deal **13 to each player**. The commander is therefore always a human.

- JARVIS is a full crew member: it takes tasks, wins tricks and owns objectives. It sits in the third seat (§3).
- The **commander controls JARVIS**: on JARVIS's turn the commander chooses which card JARVIS plays, and in task selection which task JARVIS takes. The other player may not comment.
- JARVIS may play **only its face-up cards**. **App decision:** follow suit applies to JARVIS's face-up cards. If none of them matches the led suit, JARVIS may play any face-up card, even if a hidden card would match.
- When a face-up card has been played, the card below it is turned face up **after the trick ends**, never during it.
- JARVIS's face-up cards are visible to both players. Its hidden cards are not.
- After the last trick, one JARVIS card stays unplayed (13 tricks, 14 cards).

## 13. Special objectives

Special objectives come from the mission (§14). They belong either to **the crew** or to **one member**. They are fulfilled and fail like tasks: an objective is **fulfilled** once its condition is met **and can no longer fail**, and it **fails** as soon as it can no longer be met. At the latest after the last trick, every open objective fails.

| Kind | Used in | Fulfilled when | Fails when |
|---|---|---|---|
| Win no trick (member) | 5 | the last trick is played and the member won none | the member wins a trick |
| Win exactly one trick, not with a rocket (member) | 33 | the last trick is played, the member won exactly 1, and it was not won by a rocket | the member wins a 2nd trick, or wins one with a rocket, or after the last trick has 0 |
| Win only the first and the last trick, not with rockets (member) | 41 | the member won tricks 1 and last and no other, neither by a rocket | the member does not win trick 1, wins any trick in between, wins with a rocket, or does not win the last trick |
| Win the first and the last trick (commander) | 34 | the commander won trick 1 and the last trick | another member wins trick 1 or the last trick |
| Win only the first four tricks (member A) and only the last trick (member B) | 50 | A won exactly tricks 1–4, B won exactly the last trick | A misses one of tricks 1–4 or wins a later one; B wins any trick before the last or misses the last |
| Win all pink cards (member) | 46 | the member has won every pink card (and none stays unplayed) | anyone else wins a trick containing a pink card |
| A trick is won by a colour **1** (crew), once / twice | 9 / 26 | the needed number of tricks have a colour 1 as the winning card | too few colour 1s are left unplayed to still reach the number |
| Each rocket wins a trick (crew) | 13 | rockets 1, 2, 3 and 4 have each been the winning card of a trick | a rocket is played and does not win its trick, or stays unplayed |
| …in the order 1, 2, 3, 4 (crew) | 44 | as above, and rocket *k*'s trick comes before rocket *k*+1's | as above, or rocket *k*+1 wins while rocket *k* has not yet won |
| No trick is won by a 9 (crew) | 16, 17 | the last trick is played and no trick was won by a colour 9 | a trick is won by a colour 9 |
| Balance (crew) | 29, 34 | the last trick is played and the condition always held | after any trick, a member has won **two tricks more** than another member |
| The `Ω` task is won in the last trick (task modifier) | 48 | its task is fulfilled in the last trick | its task is fulfilled earlier |

Notes:
- "The last trick" is trick 13 (2–3 players), 10 (4) or 8 (5).
- "Won by a rocket" or "won by a colour 1/9" refers to the **winning card** of the trick.
- In mission 46 the owner is the **left neighbour** (next seat) of whoever holds the **pink 9** after the deal. **App decision:** this is fixed when the attempt is dealt, before any distress pass, and shown to everyone.
- In mission 50 the commander chooses A and B, two different members, after the deal. Themselves and JARVIS are allowed.
- JARVIS counts as a member for balance (29, 34).

**App decision: when a mission is won.** The mission is won as soon as every task and every special objective is fulfilled. Objectives that can still fail until the end (win no trick, no trick won by a 9, balance, "only"/"exactly" conditions, the last trick) are fulfilled only after the last trick. So a mission with such an objective is always played to the end. Mission 17, for example, continues after both tasks are done, because a 9 could still win a trick. This follows Deep Sea's definition: "a task is completed when its conditions have been met and can no longer fail".

## 14. Missions

Columns:
- **Tasks:** number of task cards (open distribution unless stated).
- **Tokens:** in placement order.
- **Comm:** communication restrictions (§8): `dead zone` or `disruption N`.
- **5p:** golden frame, so the five-player handover is allowed (§11).
- **Special:** extra setup or objectives (§10, §13).

| # | Tasks | Tokens | Comm | 5p | Special |
|---|---|---|---|---|---|
| 1 | 1 | | | | |
| 2 | 2 | | | | |
| 3 | 2 | 1 2 | | | |
| 4 | 3 | | | | |
| 5 | – | | | | Poll "well/ill". The commander chooses another member, who must **win no trick**. |
| 6 | 3 | › ›› | dead zone | | |
| 7 | 3 | Ω | | | |
| 8 | 3 | 1 2 3 | | | |
| 9 | – | | | | Crew: **one trick must be won by a colour 1**. |
| 10 | 4 | | | | |
| 11 | 4 | 1 | | | The commander chooses a member who **must not communicate** this attempt. **App decision:** any human member, the commander included. |
| 12 | 4 | Ω | | | **After the first trick**, each member takes a random card from their right neighbour (§16). |
| 13 | – | | | | Crew: **each rocket must win a trick**. |
| 14 | 4 | › ›› ››› | dead zone | | |
| 15 | 4 | 1 2 3 4 | | | |
| 16 | – | | | | Crew: **no trick may be won by a 9**. |
| 17 | 2 | | | | Crew: **no trick may be won by a 9**. |
| 18 | 5 | | disruption 2 | | |
| 19 | 5 | 1 | disruption 3 | | |
| 20 | 2, face down | | | | **Commander's decision** (§10). |
| 21 | 5 | 1 2 | dead zone | | |
| 22 | 5 | › ›› ››› ›››› | | | |
| 23 | 5 | 1 2 3 4 5 | | | Before distribution, the commander may **swap the tokens of two task cards** (§16). |
| 24 | 6 | | | | **Commander's distribution** (§10). |
| 25 | 6 | › ›› | dead zone | ✓ | |
| 26 | – | | | | Crew: **two tricks must be won by a colour 1**. |
| 27 | 3, face down | | | ✓ | **Commander's decision**. |
| 28 | 6 | 1 Ω | disruption 3 | ✓ | |
| 29 | – | | dead zone | | Crew: **balance** (nobody ever has two tricks more than another). |
| 30 | 6 | › ›› ››› | disruption 2 | ✓ | |
| 31 | 6 | 1 2 3 | | ✓ | |
| 32 | 7 | | | ✓ | **Commander's distribution**. |
| 33 | – | | | | Poll "yes/no". The commander chooses another member, who must **win exactly one trick, not with a rocket**. |
| 34 | – | | | | Crew: **balance**. The **commander must win the first and the last trick**. |
| 35 | 7 | › ›› ››› | | ✓ | |
| 36 | 7 | 1 2 | | ✓ | **Commander's distribution**; the tokens go with the first two revealed cards. |
| 37 | 4, face down | | | ✓ | **Commander's decision**. |
| 38 | 8 | | disruption 3 | ✓ | |
| 39 | 8 | › ›› ››› | dead zone | ✓ | |
| 40 | 8 | 1 2 3 | | ✓ | Before distribution, the commander may **move one token to a task card without a token** (§16). |
| 41 | – | | | | Poll "yes/no". The commander chooses another member, who must **win only the first and the last trick, not with rockets**. |
| 42 | 9 | | | ✓ | |
| 43 | 9 | | | ✓ | **Commander's distribution**. |
| 44 | – | | | | Crew: **each rocket must win a trick, in the order 1, 2, 3, 4**. |
| 45 | 9 | › ›› ››› | | ✓ | |
| 46 | – | | | | The **left neighbour of the pink 9's holder** must **win all pink cards**. |
| 47 | 10 | | | ✓ | |
| 48 | 3 | Ω | | ✓ | The `Ω` task must be won **in the last trick**. |
| 49 | 10 | › ›› ››› | | ✓ | |
| 50 | – | | | | The commander assigns two roles: member A must **win only the first four tricks**, member B must **win only the last trick**. The other members win the tricks in between. |

Reading notes for the sheet:
- On the sheet, a dark-helmet icon is **the commander** (with one arrow: commander's decision; with three arrows: commander's distribution; alone in mission 34: the commander must win the first and last trick). A floating astronaut is **a member the commander chooses**.
- **App decision (mission 34):** the sheet shows the commander icon next to "first & last trick", so the commander owns that objective.

## 15. Attempts, success, the logbook

- **Lost attempt:** as soon as a task or objective fails (§6, §7, §13), or when a player gives up the attempt. The crew starts a new attempt of the same mission with a new deal and new tasks.
- **Mission completed:** as soon as all tasks and objectives are fulfilled (§13). Tricks that have not been played do not matter.
- **Logbook** (per mission): the number of attempts needed, whether the distress signal was used, and the total (attempts + 1 with the distress signal). The logbook also keeps the attempts of missions not yet completed.

## 16. App decisions without a rule counterpart

- **Session model:** a game (lobby code) is one sitting of a crew. In the lobby the host picks any mission 1–50; the next unfinished mission is preselected. After a mission, the crew returns to the lobby or goes straight to the next mission. Players may join or leave only in the lobby. The logbook lives in the game and is shown in the lobby. Games idle for 7 days are deleted like all games, so to continue later, the host picks the mission in a new game.
- **Starting:** 2–5 players. The host starts the selected mission. With 2 players, JARVIS joins automatically.
- **Commander makes the group calls:** whenever the rules say "decide together", the commander decides in the app: the distress signal and its direction, the five-player handover, the token swap (23) and token move (40), and the roles in mission 50. The players may talk about these outside the app.
- **Token swap / move (missions 23, 40):** after the tasks are drawn and the tokens placed, and before distribution, the commander may swap the tokens of two task cards (23), or move one token to a task card without one (40). Either may be skipped. One swap or move per attempt.
- **Mission 12 card draw:** right after the first trick, before the next lead, each human member gets a random card from their right neighbour's hand. All draws are decided first, then the cards move at once.
  - A **communicated** card still waiting on the table cannot be drawn.
  - With 2 players JARVIS does not take part: the two humans each draw from the other.
  - Each player sees the card they gave and the card they got. The others see only that cards were exchanged.
- **After an attempt** (won or lost): the game shows the reason and **reveals all hands**, including JARVIS's hidden cards and the unplayed card. Then the host chooses **next attempt** (after a loss), **next mission** (after a win), or **back to the lobby**. Back in the lobby after a win, the next unfinished mission is selected; after a loss the mission stays selected.
- **Replaying a completed mission:** allowed from the lobby. Its attempts count from 1 again for the display, but they do not change its logbook line. An unfinished mission continues its attempt count, also across a return to the lobby.
- **Last trick:** the last completed trick can always be looked at, also while the next one is played. Played cards are public.
- **Give up:** any player may give up the current attempt. It counts as a lost attempt.
- **Ending:** the host may end the session from the lobby or after an attempt (phase `ended`).
- **Not implemented:** the *Challenge for three* variant from the rulebook (no green cards, no rocket 1). It may come later as a lobby option.
