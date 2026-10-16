# Fox Duet

**The Fox in the Forest: Duet** (Foxtrot Games), cooperative trick-taking for exactly 2 players, with the three difficulty levels. The **reference game** for conventions (root [CLAUDE.md](../CLAUDE.md)). Rules: [docs/rules.md](docs/rules.md) (rules summary, board tables, App decisions).

## Intents and per-player view

- **Intents:** start at level N, play a card, choose the Foxes player, swap the decree / keep it, give a gift card, choose a move, cover an end (host, round end), give up.
- **Per-player view:** own hand (with the playable cards), the teammate's hand *count*, the board (gems per location, pluses, covered spaces, traveller, forest tokens left), round, trick number, dealer, decree card, the current and the last completed trick, the pending decision with its options for the viewer, the viewer's own gift card, and the result with the score. Never the teammate's hand, their gift choice or the deck.

## Domain modelling

- **Cards are structs** (`Card{Suit, Rank}`). Movement and ability are functions of the rank, never stored.
- **Boards and levels are tables** (`boards`, `levels` in `board.go`): topology of sides A and B, starting gems, pluses, start space, forest tokens, base score.
- **Board state is minimal:** gems per location, covered count per end, traveller position, forest tokens left. Positions are indices from the left end (seat 1's end); a negative delta moves left.
- **The RNG is injected** (seedable). Every mutator that can end a round takes it, because the next round is dealt right away.
- **Steps while playing:** `lead | follow | foxesChoose | foxesSwap | gift | move | forestEnd`. Every action checks phase, step and actor.
- **The pending ability belongs to the card played last** (the lead card until there is a follow card). Who acts is derived, not stored.
- **Derived, not cached:** trick winner, legal cards and move options come from the trick, the decree and the board. **Automatic choices** (rules §10) resolve inside the mutator that reaches them.
- Lost only by rules §8 (or give-up).

## Look and feel

A forest theme: moss green, fox orange and parchment in light mode, deep forest night in dark mode. Cards show suit glyph, rank, paw-print movement and ability name. The **board** is an SVG with the real topology, rotated so each player's own end is on the left. Below it: trick row (decree, lead, follow), status line or decision panel, and an overlapping hand with illegal cards dimmed. The board is wide and short so it fits in portrait above the hand. Card motion for deal, play and gift.
