# Rules reference

English summary of the official rules of **The Fox in the Forest: Duet** (Foxtrot Games, 2020; original *Fox in the Forest* design by Joshua Buergel). The code follows this file. Where the official rules are ambiguous or the app must decide something, the choice is recorded under **App decision**.

Sources: the English rulebook, both sides of the board and the German reference card. The board tables in §3 were transcribed from photos of the board.

---

## 1. Overview

Cooperative trick-taking game for **exactly 2 players**. Each round has 11 tricks. After each trick the **traveller** (the "team tracker") moves along the forest path toward the trick winner, and the team collects a gem where it stops. The team wins by collecting every gem on the board within 3 rounds without getting lost in the forest.

## 2. Cards

30 cards in 3 suits: **Doves, Roses, Stars**, ranks **1–10** each. Every card also has a **movement value** (paw prints), which depends only on its rank:

| Rank | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 |
|---|---|---|---|---|---|---|---|---|---|---|
| Movement | 0 | 3 | 2 | 1 | 1 | 2 | 0 | 2 | 0 | 3 |
| Ability | Musician | – | Foxes | – | Gazelle | – | Gift | – | Royal Heir | – |

The odd ranks have abilities, which trigger when the card is played (§6).

## 3. The board

The forest path is a line of path spaces between two **ends**. Each path space is connected to one **location**. A location has a number of hanging **squares** and **diamonds** (the starting gems) and may have a **plus** (gems added at round end). The **start** space is circled.

- **App decision:** seat 1 (the host, the first player to join) faces the **left** end and seat 2 the **right** end. "Toward a player" means toward that player's end. On screen, each player sees their own end on the left.

### Side A (level 1)

11 spaces from the left end to the right end; the start is **C**.

| Space | L5 | L4 | L3 | L2 | L1 | **C** | R1 | R2 | R3 | R4 | R5 |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Squares | 1 | 1 | 2 | 1 | 1 | 0 | 1 | 1 | 2 | 1 | 1 |
| Plus | | + | | + | | + | | + | | + | |

12 gems at the start, 5 pluses.

### Side B (levels 2 and 3)

10 spaces from the left end to the right end; the start is **C**, so the left side is one space shorter than the right.

| Space | L4 | L3 | L2 | L1 | **C** | R1 | R2 | R3 | R4 | R5 |
|---|---|---|---|---|---|---|---|---|---|---|
| Squares | 1 | 2 | 2 | 0 | 2 | 2 | 0 | 2 | 1 | 1 |
| Diamonds | 1 | | | | 1 | | | | 1 | |
| Plus | | + | + | + | | + | + | + | | |

6 pluses. Level 2 uses the squares only (13 gems); level 3 uses squares and diamonds (16 gems).

## 4. Difficulty levels

| Level | Board | Starting gems | Forest tokens | Base score |
|---|---|---|---|---|
| 1 | Side A | squares (12) | 4 | 10 |
| 2 | Side B | squares (13) | 3 | 20 |
| 3 | Side B | squares + diamonds (16) | 3 | 30 |

- **App decision:** the host picks the level when starting the game (default level 1).

## 5. Setup and dealing

1. Put the starting gems on the locations (§3, §4). The rest form an unlimited supply.
2. Put the traveller on the start space.
3. Pick the dealer of the first round.
   - **App decision:** the first dealer is chosen at random.
4. Shuffle all 30 cards and deal **11 cards** to each player. The remaining 8 cards form the deck. Reveal the top card of the deck: the **decree card**. Its suit is the **trump suit**. The other 7 cards stay hidden for the round.

A game has **3 rounds** of **11 tricks**.

## 6. Tricks

### Playing

- **Leading:** in the first trick of a round the player who did **not** deal leads. Afterwards the winner of the previous trick leads. The leader may play any card.
- **Following:** the other player must play a card of the **lead suit** if they have one (any rank). Otherwise they may play any card.
  - **Royal Heir (9):** when a 9 is **led**, the follower need not follow suit (they still may). A 9 played second has no effect.

### Abilities that act when played

Each resolves immediately after its card is played, before the next card is played or the trick is resolved.

- **Foxes (3):** the player who played it chooses a player (themselves or their teammate). That player may **exchange the decree card** with a card from their hand, or keep it. A new decree changes the trump suit right away, which may change who wins this trick. The lead suit does not change: it is always the suit of the led card.
  - **App decision:** if the chosen player's hand is empty, nothing happens. If both hands are empty (a 3 played second in the 11th trick), the choice is skipped.
- **Gift (7):** both players choose one card from their own hand and give it to their teammate at the same time, without knowing what they will receive. This is mandatory as long as both players have at least one card. Otherwise nothing happens.
  - **App decision:** each player submits their card secretly. The exchange happens once both have chosen. A player with exactly one card gives it automatically.

### Winning the trick

- If one or both cards are in the **trump suit**, the higher trump wins.
- Otherwise the higher card of the **lead suit** wins (a card of a third suit never wins).

### Moving the traveller

Add the movement values of both cards and move the traveller that many spaces toward the **winner's** end.

- **Musician (1):** the winner may instead move the traveller the same distance in the opposite direction (toward their teammate).
- **Gazelle (5):** the winner may ignore the movement value of one card of the trick (either card). With two Gazelles in the trick the winner may ignore one, none or both cards.
- Both abilities belong to the winner, no matter who played the card. They combine: e.g. a 1 and a 5 let the winner choose the direction and whether to ignore a card.
- **App decision:** the winner picks one move from the list of possible results, and only the result counts (e.g. "ignore the 1" and "ignore nothing" are one option when the 1 moves 0). If all choices lead to the same result, the move happens automatically.

### Beyond the path

If the traveller would move past the last uncovered space of an end:

1. If no forest token is left, the game is **lost** (lost in the forest).
2. Otherwise put the traveller back on the **start** space and cover the last uncovered space of that end with a forest token. Gems on the location of the covered space move to the location of the next space toward the centre.
3. No gem is collected this trick.

### Collecting gems

When the traveller ends its move on a space whose location has gems (**even if it did not move**), and it was not just reset, the team collects **one** gem from that location. Collecting the last gem on the board **wins** the game immediately.

### Clean up

The two cards are set aside face down. Players may not look at earlier tricks.

- **App decision:** the last completed trick stays visible on the table until the next card is led, so both players can see how the trick ended.

## 7. Round end

After the 11th trick of round 3 the game ends (§8). After the 11th trick of rounds 1 and 2:

1. **Add gems:** add one gem from the supply to each location with a plus. There is no limit per location.
   - **App decision:** a plus location whose path space is covered gets its gem on the outermost uncovered location of the same end. (The rulebook doesn't cover this; the gem would otherwise be lost.)
2. **Place a forest token** (if one is left) on the last uncovered space of one end. The players choose the end, but if the traveller stands on the last uncovered space of an end, the token goes on the other end. Gems on the covered location move one location toward the centre (as in §6). Without a forest token, nothing happens.
   - **App decision:** the **host** chooses the end. When only one end is allowed, the token is placed automatically.
3. **Deal the next round:** the player who did not deal last round deals. Reshuffle all 30 cards, deal 11 each and reveal a new decree card. The traveller stays where it is.

The start space can never be covered: each end has more spaces than there are forest tokens.

## 8. End of the game

The game ends immediately when one of these happens:

- **Victory:** the last gem on the board is collected. Score:
  - 10 / 20 / 30 for level 1 / 2 / 3,
  - +1 for each card left in either player's hand,
  - +10 if the game was won before round 3 (in practice in round 2: one round has only 11 tricks and every level starts with at least 12 gems),
  - +3 for each forest token left.
- **Lost in the forest:** the traveller would move beyond the path and no forest token is left.
- **Out of time:** gems remain on the board after the 11th trick of round 3.

- **App decision:** any player may **give up** at any time (the rulebook calls it resigning, for when the gems can no longer all be collected). It ends the game as a loss for both, after a confirmation.
- **App decision:** the host can end a game in the lobby (phase `ended`), and finished games can be rematched, as in the other games.

## 9. Communication

During a round players must not talk about their cards, ask revealing questions or discuss strategy. They may talk freely before the first deal and between rounds. The app has no chat; all public information (decree, traveller, gems, forest tokens, hand sizes) is on screen.

## 10. App decisions without a rule counterpart

- **Automatic choices:** a decision with only one distinct result is taken automatically (single-card gift, a move with one result, a forced forest-token end, Foxes with both hands empty).
- **No undo.** Every play is final.
- **Hidden information:** a player sees only their own hand and their own pending gift card, never the teammate's hand, their gift choice or the deck.
