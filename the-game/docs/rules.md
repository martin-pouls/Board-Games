# Rules reference

English summary of the official German rules (Nürnberger-Spielkarten-Verlag, Steffen Benndorf / Reinhard Staupe). Code should follow this file. Wherever the official rules are ambiguous, we note it here along with the interpretation we chose.

Sources:
- The Game, 10th-anniversary edition (contains base game, Extreme and mini-expansions): https://www.nsv.de/wp-content/uploads/2025/09/TheGame_10th_D.pdf
- The Game Extreme: https://www.nsv.de/wp-content/uploads/2024/04/TheGame_EX_D.pdf
- Extreme card → command mapping: photos of the physical cards supplied by the project owner (see table below).

---

## 1. The Game (base game)

Cooperative game for 1–5 players. The team wins together or loses together.

### Components
- **4 start cards** (pile markers, never in hand):
  - 2× **ascending** piles, start value **1** (play upwards 1→99)
  - 2× **descending** piles, start value **100** (play downwards 100→2)
- **98 number cards**, values **2–99**, one of each.

### Setup
- Lay out the 4 start cards. Shuffle the 98 number cards.
- Hand size by player count:

| Players | Hand size | (Pro variant, harder) |
|---|---|---|
| 1 (solo) | 8 | 7 |
| 2 | 7 | 6 |
| 3–5 | 6 | 5 |

- The remaining cards form the face-down **draw pile**.
- The players agree who starts. Play then proceeds clockwise.

### Placement rules
- Each start card has exactly one pile; only its top card counts.
- **Ascending pile:** the new card must be **greater** than the top card. Gaps may be any size.
- **Descending pile:** the new card must be **smaller** than the top card.
- **Backwards trick (Rückwärts-Trick):**
  - Ascending pile: you may play a card that is **exactly 10 lower** than the top card.
  - Descending pile: you may play a card that is **exactly 10 higher** than the top card.
  - You may use it any number of times per turn, on any piles, mixed with normal plays.

### Turn
1. The active player **must play at least 2 cards** from hand, one at a time, on one or more piles. They may then play any number of additional cards.
2. At the end of the turn they **draw as many cards as they played**, which refills the hand to the hand size.
3. The next player clockwise takes their turn.

### Draw pile exhausted
- Once the draw pile is empty, play continues without drawing.
- The minimum is now **1 card** per turn instead of 2.
- A player who has emptied their hand is **skipped** for the rest of the game.

### End of game
- The game ends **immediately** when the active player **cannot play the required minimum**: 2 cards while the draw pile has cards, 1 card once it is empty.
  - *Interpretation (decided by the project owner):* there is no lookahead. The game is lost only once the active player has **no legal card to play and cannot end the turn**. A player who can still play a card keeps playing, even if the minimum is out of reach. The score counts the cards left at that moment.
- *App decisions:* the host picks the starting player after the cards are dealt (there is no random default; a solo game starts immediately). The turn ends automatically when the active player's hand becomes empty and the minimum is met. Once the game is over, every player's remaining hand is revealed.
- **Score** = cards not played = all cards left in hands + cards left in the draw pile. Lower is better. Fewer than 10 is "excellent". **0 = won** (all 98 cards played).

### Communication
- Players must **never reveal or ask for concrete numbers**, in any form.
- Anything else is allowed, e.g. "please don't play on this pile" or "only a small jump here, please".
- In this app, chat and in-game hints must not leak exact hand values.

### Pro variant
- The minimum per turn is **3** instead of 2. Optionally the hand size is also reduced by 1 (see table above).
- *Interpretation:* once the draw pile is empty, the minimum is still 1.

---

## 2. The Game Extreme

All base rules apply unchanged. Hand sizes and minimums are the same.

### Components / setup
- **28 of the 98 number cards are replaced** by "extreme" cards with the **same values** that carry a **command**: 7 commands × 4 cards each.
- The deck therefore still has 98 cards with values 2–99, each exactly once. 28 of them have a command.
- In the physical game, the replaced base cards have blue skulls at the bottom edge.

### Command → card values (from photos of the physical cards)

| Kind | Command | Card values |
|---|---|---|
| ⚡ Lightning | **STOP** | 6, 35, 57, 95 |
| ⚡ Lightning | **Skull** (must be covered) | 17, 41, 61, 87 |
| ⚡ Lightning | **3!** (exactly three cards) | 19, 28, 53, 81 |
| ∞ Infinity | **No communication** | 26, 65, 78, 92 |
| ∞ Infinity | **No backwards trick** | 21, 32, 59, 70 |
| ∞ Infinity | **Single pile** (all cards on one pile) | 9, 47, 68, 84 |
| ∞ Infinity | **Draw only one** | 13, 38, 43, 76 |

(The examples in the rulebook agree with this table: 6 = STOP, 41 = Skull, 81 = 3!, 9 / 13 / 32 / 65 = Infinity cards.)

### General command rules
- When the active player plays one or more command cards, **all** of those commands must be obeyed. Otherwise **the game is lost immediately**.
  - **In this app** (decided by the project owner): the engine **blocks** any action that would break a command, so a command can never actually be broken. See *Enforcement* below.
- Players may talk about commands, e.g. which ones they hold or which should not be played now. Concrete numbers stay taboo.

### ⚡ Lightning commands
They take effect immediately, apply **only to the player who played them**, and last only for the rest of that turn. A lightning symbol still visible on a pile afterwards has no meaning.

- **STOP:** the turn **ends immediately**. It may be the **first** card of the turn; the turn then ends after that single card.
  - *Interpretation:* this overrides the usual minimum of 2 cards.
- **Skull:** must be **covered by another card in the same turn**. No skull may be visible on top of a pile at the end of the turn.
  - If all 98 cards are played but a skull from the very last turn stays visible, the game is still **lost**.
- **3!:** the player must play **exactly 3 cards** this turn, counting this one.
  - A STOP may be played **only as the third card**.
  - If the player had only 1 or 2 cards at the start of the turn, they may still play 3!, but the game is **lost** because the command cannot be fulfilled. *In this app that play is blocked instead (see Enforcement).*

### ∞ Infinity commands
They take effect from the moment the card is played and **apply to everyone** for as long as the card is **visible on top of a pile**. Several can be active at once on different piles. Covering the card cancels its effect.

- **No communication:** players may not talk about the game or give any signals. In this app, chat and hints are disabled while it is active.
- **No backwards trick:** the active player may not use the backwards trick on any pile.
- **Single pile:** the active player must put **all** cards they play this turn on **one** pile of their choice.
- **Draw only one:** the active player draws only **1 card** at the end of the turn, however many they played. Hands can therefore shrink below the hand size. Once the command is covered, the active player refills to the full hand size at the end of their turn again.

### Interpretations (decided)
- **Single pile / no backwards trick played mid-turn:** "valid from the moment it is played". Only cards played *after* it in the same turn are restricted. Cards already played earlier in the turn on other piles stay where they are.
- **Single pile played mid-turn binds to its own pile:** all remaining cards of that turn must go on the pile where the Single pile card was played. *(Decided by the project owner.)*
  - If a Single pile card is already visible at the start of a turn, the active player's first card picks the pile, and all further cards that turn must go on it.
- **Draw only one:** we evaluate whether it is visible at the end of the turn, when drawing.

### Enforcement: illegal moves are blocked (decided by the project owner)
The engine rejects any action that would break a command, returning a domain error; game state stays unchanged. Concretely:

| Situation | Blocked action |
|---|---|
| Placement breaks base rules, No backwards trick, or Single pile | the card play |
| A 4th card after 3! | the card play |
| STOP after 3! as the 1st or 2nd card | the card play |
| 3! when the player started the turn with fewer than 3 cards | the card play |
| Turn would end with a Skull visible on top of a pile | ending the turn |
| Turn would end with fewer than 3 cards after 3! | ending the turn |
| Turn would end below the minimum (unless STOP was played) | ending the turn |

- STOP ends the turn automatically once it is played; no further card can be played that turn.
- **Stuck player = lost game.** The engine does not look ahead to warn players. If the active player has no legal action left (no playable card *and* ending the turn is blocked), the game is **lost**, just like failing the minimum in the base game. Example: a Skull is played and no hand card can cover it.
- This also covers the "skull visible after the last card" rule: playing the final card as an uncovered Skull leaves the player stuck, so the game is lost.

---

## 3. Mini-expansions (10th-anniversary edition): not in scope yet

The 10th-anniversary booklet also contains three mini-expansions. We record them here so the card model stays general enough (card *kinds* beyond plain numbers):

- **On Fire:** cards 22, 33, 44, 55, 66, 77 are replaced by On Fire cards. An On Fire card must be covered by the same player in the same turn, or at the latest by the **next** player. Otherwise the game is lost immediately. In solo play, cover it at the latest in your next turn.
- **Joker** (10 cards, shuffled in): may be played on any pile. It must be covered in the same turn by a number card that is legal against the last *number* card underneath it. Several jokers may be stacked. A joker that is never covered means the game is lost.
- **Zehner** (10 tens cards, "0–9" … "90–99", shuffled in): a tens card may be played on any card of the same tens range, or by normal ascending/descending rules as if it were its whole range. On top of a tens card you may play any card of the same range, or follow normal direction rules. No backwards trick is possible on a tens card.
