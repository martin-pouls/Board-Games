# Rules reference

English summary of the official Regicide rules (Paul Abrahams, Luke Badger, Andy Richdale; Badgers From Mars). Code follows this file. Where the official rules are ambiguous, the interpretation we chose is recorded under **App decision**.

Source: https://www.regicidegame.com/site_files/33132/upload_files/RegicideRulesA4.pdf

---

## 1. Overview

Cooperative game for **1–4 players**. Players defeat 12 enemies (4 Jacks, 4 Queens, 4 Kings) by playing cards from hand. Each turn the enemy strikes back and the active player must discard cards to absorb the damage. If they cannot, everyone loses. Players win when the last King is defeated.

## 2. Components (one standard 54-card deck)

| Card | Count | Role | Value |
|---|---|---|---|
| 2–10 of ♥ ♦ ♣ ♠ | 36 | Tavern deck | face value |
| Ace ("Animal Companion") of each suit | 4 | Tavern deck | 1 |
| Jester (Joker) | 2 | Tavern deck (depends on player count) or solo power | 0 |
| J / Q / K of each suit | 12 | Castle deck (enemies) | see enemy table |

### Enemy table

| Enemy | Attack | Health | Value when in a player's hand |
|---|---|---|---|
| Jack | 10 | 20 | 10 |
| Queen | 15 | 30 | 15 |
| King | 20 | 40 | 20 |

## 3. Setup

- **Castle deck:** shuffle the 4 Kings, place facedown; shuffle the 4 Queens, place on top; shuffle the 4 Jacks, place on top. Reveal the top card: the first enemy (always a Jack).
- **Tavern deck:** shuffle all 2–10s, the 4 Animal Companions and the Jesters for the player count.
- **Discard pile:** shared, starts empty, face up.
- Deal each player cards up to their **max hand size**.

| Players | Jesters in Tavern | Max hand size |
|---|---|---|
| 1 | 0 (2 set aside, see Solo) | 8 |
| 2 | 0 | 7 |
| 3 | 1 | 6 |
| 4 | 2 | 5 |

- First player: "the player to have most recently committed regicide".
  - **App decision:** the host picks the first player in the lobby, or chooses "random".

## 4. Turn structure

Each turn has four steps.

### Step 1: Play a card or yield

The active player either plays a **legal play** from hand (see §5) or **yields** (see §8).

### Step 2: Activate the suit power

Each distinct suit among the played cards activates its power, using the play's **total attack value** (sum of the card values). The power does not activate if the enemy is immune to that suit (§6). Powers are mandatory.

- **♥ Hearts – heal from the discard:** shuffle the discard pile, take *N* cards (N = attack value; fewer if the discard pile is smaller) facedown and put them **under** the Tavern deck.
- **♦ Diamonds – draw cards:** starting with the active player and going clockwise, players draw one card at a time until *N* cards in total have been drawn. Players at their max hand size are skipped. Drawing stops early when every player is at max hand size or the Tavern deck is empty (no penalty).
- **♣ Clubs – double damage:** in Step 3 the play's damage is doubled.
- **♠ Spades – shield:** the enemy's attack is reduced by *N*. Shields are cumulative over all plays against the current enemy, by any player, and last until that enemy is defeated.
- If Hearts and Diamonds resolve in the same play, **Hearts resolves first**.
- If the play contains several cards of the same suit, that suit's power still applies **once**, at the total value.

Example: 3♦ + 3♠ + 3♣ (total 9): draw 9 cards, shield 9, deal 18 damage.
Example: 8♦ + A♣ (total 9): draw 9 cards, deal 18 damage.

### Step 3: Deal damage and check

- Damage = total attack value, ×2 if Clubs power applies.
- Damage accumulates against the current enemy across all turns. If accumulated damage ≥ enemy health, the enemy is **defeated**:
  1. If accumulated damage is **exactly** equal to health, the enemy card goes **facedown on top of the Tavern deck**. Otherwise it goes to the discard pile.
  2. All cards played against that enemy go to the discard pile.
  3. Reveal the next Castle card as the new enemy (damage and shield reset to 0, Jester immunity cancellation resets).
  4. The **same player** skips Step 4 and starts a new turn at Step 1 against the new enemy.
- If the Castle deck is empty after a defeat, the players **win**.

### Step 4: Suffer damage

- Effective attack = enemy attack − total shield, never below 0.
  - **App decision:** if effective attack is 0, Step 4 completes automatically with no discard.
- The active player must discard cards from hand with total value ≥ effective attack. Values when discarding: number cards face value, Animal Companion 1, Jester 0, J 10, Q 15, K 20.
- Cards are discarded one at a time and the player stops once the damage is covered.
  - **App decision:** the player submits a set of cards at once. The set is legal iff `sum ≥ attack` and `sum − (smallest value in the set) < attack` (no superfluous card; equivalently, some discard order reaches the threshold only with its last card). Jesters (value 0) can therefore never be part of a discard set unless attack is 0, which never needs a discard.
- If the total value of the player's whole hand is less than the effective attack, the player dies and **all players lose**.
  - **App decision:** this is detected automatically when Step 4 begins (no player action needed).
- An empty hand is allowed. After Step 4, the next player clockwise begins Step 1.

## 5. Legal plays (Step 1)

A play is exactly one of:

1. **Single card:** any one card from hand (number card, Animal Companion, J/Q/K in hand, or Jester).
2. **Combo:** 2, 3 or 4 number cards of the **same rank**, total value ≤ 10. So only: pairs of 2s/3s/4s/5s, triples of 2s/3s, quadruple 2s.
3. **Animal Companion pair:** one Animal Companion plus exactly one other card that is not a Jester. The other card may be any number card, another Animal Companion, or a J/Q/K from hand. Value = 1 + other card's value.

Not allowed: Animal Companions inside a combo; a combo of Animal Companions alone beyond a pair; a Jester together with any other card; J/Q/K in a combo (they have no "same rank ≤ 10" combination).

## 6. Enemy immunity

- An enemy is **immune to the suit power of its own suit**: e.g. playing ♦ against the J♦ draws no cards. The card's value still counts towards damage (and ♣ against a ♣ enemy is not doubled; ♠ against a ♠ enemy gives no shield).
- Playing a **Jester** against an enemy cancels that enemy's immunity for the rest of the fight against it.
  - Spades played against a ♠ enemy **before** the Jester start reducing its attack once the Jester has been played (retroactive).
  - Clubs played against a ♣ enemy before the Jester do **not** become doubled retroactively. Hearts/Diamonds before the Jester are not retroactively resolved either.
- **App decision (derivation):** shield = sum of the attack values of all plays against the current enemy that contained a ♠, counted only if spades are not immune *now* (enemy suit ≠ ♠, or a Jester has been played against this enemy). Clubs doubling and red powers are evaluated at the moment of the play.

## 7. The Jester (multiplayer)

- Played in Step 1, always **on its own**. Attack value 0.
- Effect: cancels the current enemy's immunity (§6).
- After playing it, **skip Steps 3 and 4**. The player who played it chooses **any player** to take the next turn.
  - **App decision:** the Jester player may choose themselves. This matches the designers' FAQ.
- Communication relaxation: until the chosen player starts, players may say in a general way whether they would like to go next, but not reveal their hand.
  - **App decision:** while the Jester player is choosing, every other player can set a public signal: *"I have a good play"* / *"I'd rather not go next"* / no signal. Signals are cleared when the next turn starts. There is no free chat.
- A Jester in hand may be discarded in Step 4 with value 0.

## 8. Yielding

- In Step 1 the player may say "Yield" instead of playing: skip Steps 2 and 3 and go directly to Step 4.
- A player **may not yield if every other player yielded on their most recent turn**. Example: in a 3-player game, if two players have yielded in a row, the third may not yield.
  - **App decision:** each player has a flag "yielded on their last turn", set when they yield and cleared when they take any other Step 1 action (card play or Jester). A player may yield iff at least one *other* player's flag is false. Flags carry across enemies.
  - **App decision (solo):** with one player the condition is vacuous, so in solo play **yielding is always allowed**.

## 9. Game end

- **Win:** the last King is defeated.
- **Loss:**
  - In Step 4 the active player cannot cover the damage (§4, Step 4).
  - A player cannot play a card or yield in Step 1 (their hand is empty and yielding is forbidden).
    - **App decision:** detected automatically at the start of Step 1. In solo, an unused Jester power counts as an available action, so the loss triggers only when no Jester power remains.
    - **App decision (solo):** although solo yielding is always allowed (§8), an empty hand at the start of Step 1 with no Jester power left is a loss. Otherwise a fully shielded enemy (effective attack 0) would let the player yield forever without progress.
- **App decision:** any player may **give up**, which ends the game as a loss for everyone, after a confirmation.

## 10. Communication

Players may not reveal or hint at the contents of their hand. They may share public information (hand counts, Tavern size). The app has no chat; all public information is shown in the UI. See §7 for the Jester relaxation.

## 11. Solo play

- Set up as usual, but put the **2 Jesters aside** (not in the Tavern). Max hand size **8**.
- **Jester power** (each Jester once, so at most 2 per game): "Discard your hand and refill to 8 cards." This does **not** count as drawing for the purpose of enemy Diamond immunity. Using it does **not** cancel enemy immunity.
- The power may be used:
  - a) at the start of Step 1, before playing a card or yielding, or
  - b) at the start of Step 4, before discarding to take damage.
  - **App decision:** in Step 4 the automatic "hand too weak → loss" check (§4) is deferred while a solo Jester power is unused. The player then either discards normally, uses a Jester, or, when the hand cannot cover the attack and no Jester remains, the loss triggers.
  - **App decision:** after defeating an enemy, the player is in Step 1 again and may use a Jester power there.
- **Victory grade:** win with 0 Jesters used = **Gold**, 1 = **Silver**, 2 = **Bronze**.
