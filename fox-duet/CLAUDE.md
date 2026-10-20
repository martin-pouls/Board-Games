# Fox Duet

**The Fox in the Forest: Duet** (Foxtrot Games), cooperative trick-taking for exactly 2 players, with the three difficulty levels. The **reference game** for conventions (root [CLAUDE.md](../CLAUDE.md)). Rules: [docs/rules.md](docs/rules.md) (rules summary, board tables, App decisions).

## Intents and per-player view

- **Intents:** start at level N (the host picks it in the lobby; the last pick is remembered per device in `fox-duet:level`), play a card, choose the Foxes player, swap the decree / keep it, give a gift card, choose a move, cover an end (host, round end), give up.
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

**Storybook meadow:** a pale meadow page with cream cards by day, a forest night by night (the cards stay cream). Alegreya for card ranks and headings, Nunito for everything else. Suits as printed: green doves (a dove), amber roses, night-blue stars (a four-pointed star).

- **Cards** (`PlayingCard`): a suit-coloured tab in the top left holds rank and suit, the paws (movement) and on small cards the ability picture sit below it; that column stays readable in the overlapped hand. Odd ranks have an oval with the ability picture running behind the tab; trick cards name the ability on a ribbon. Sizes: hand 62 px, decree 54 px, trick 84 px (90 px from 390 px wide), results 40 px.
- **Game screen, top to bottom:** header with the round; the board row (ends, gems left, forest tokens, teammate online); the **board**, an SVG of the real topology meandering once, each player's own end on the left, tiles above the path or below it with a plus, wooden discs for forest tokens, the traveller as a fox on an orange disc; the decree with trump and trick count, then the trick (the completed trick stays until the next lead, crown on the winner); the **turn bar** and one button row as in The Game and Regicide; one overlapping row of cards, illegal cards dimmed.
- **No narration line:** what changed without a move (round-end gems, gems moved by a forest token, a new forest token) is marked on the board ("+1", a ring) until the next card is led (`boardMarks.svelte.ts`). Card motion for deal, play and gift; received cards get the new-card mark.
