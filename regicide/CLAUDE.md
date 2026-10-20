# Regicide

**Regicide** (Badgers From Mars), cooperative for 1–4 players, incl. solo mode with Jester powers and Gold/Silver/Bronze victories. Common rules: root [CLAUDE.md](../CLAUDE.md). Rules: [docs/rules.md](docs/rules.md) (rules summary, App decisions).

## Intents and per-player view

- **Intents:** play these cards, yield, discard these cards, choose next player, use solo Jester, set Jester signal, give up.
- **Per-player view:** own hand, other players' hand *counts*, current enemy (suit, rank, health, accumulated damage, shield, effective attack, immunity status), cards played against the enemy, the last turn (who, the cards played or a yield, the enemy it defeated and whether exactly, the cards discarded), Tavern/Castle/discard counts, whose turn, yield flags, Jester signals, solo Jesters left. Never the order of the Tavern or Castle deck.

## Domain modelling

- **Cards are structs** (`Card{Suit, Rank}`), with Jester as its own rank. Values are derived by functions (attack value, discard value, enemy attack/health), never stored.
- **The RNG is injected** (seedable).
- **Steps while playing:** `play | discard | chooseNext`. Every action checks phase, step and active player.
- **Derived, not cached:** shield, effective attack, immunity cancellation and damage come from the cards played against the current enemy plus a "Jester played against this enemy" flag. The exception is `LastTurn`, the record behind the table's Played and Discarded piles: every play or yield replaces it and the discard fills it in, so it survives a defeat (the killing play stays visible, its discard empty).
- Lost only by rules §9 (or give-up). Ruleset values (hand size, Jester count) come from a table keyed by player count.

## Look and feel

"Mist", after the blue Regicide box: a pale blue-green table (dark: deep blue-green) with misty ridges at the top of every page (`MistRidges`), slate ink and primary, off-white card stock, muted card red and black, a red for the enemy's health and a green Jester. Cabin for all UI text, Alegreya for headings, card ranks and the enemy's name. Tokens in `src/routes/layout.css`; suit symbols use the `.suit-glyph` class.

- **Cards** (`PlayingCard`, sizes in card widths): number cards show rank and suit in the top left and one big suit in the middle, no bottom index. J, Q and K show the letter where the suit sits on number cards, standing on three ridges (the darker ridges stay below the tails of J and Q). The Jester is a green three-pointed cap (`JesterHat`, also the UI icon for Jesters). The 34 px size shows the corner only.
- **Decks:** a 40 px row of their own above the enemy panel: the Tavern deck face down and the discard pile face up as small stacks, the count on the card and the name beside it; an empty deck is a dashed outline. Not in the header.
- **Enemy panel:** the Castle as a track of J, Q, K slots (defeated filled, current ringed red, "N left"), the 84 px enemy card with a Jester badge on its corner while its immunity is cancelled, the name on one line, the health bar with the selected play's damage hatched, attack − shield = damage, and one overlapping stack per player (initials move onto the stacks when four don't fit side by side).
- **Last turn:** Played (or Yielded), the defeated enemy and Discarded, each with a text label below; "Exact kill" in red when it went onto the Tavern deck. Bigger piles overlap closer, so the row never wraps.
- **Action area, split like The Game's turn bar:** a 40 px bar (whose turn, outlined when it is yours, and on the right the selected cards' effect badges on one line, blocked powers struck through), then a 44 px row with the buttons (solo Jester, Yield, a wide Play) or "Waiting for X to play." in plain text. No hand value: as with the physical game, players add it up themselves.
- **Hand:** one row of 72 px cards, as in The Game: side by side while they fit, overlapping only as much as the width needs. The players sit in the header's second row (compact chips from four players). The hand sort menu sits in the table header.
- **Start-page tile** (`start-page/frontend/src/lib/components/GameArt.svelte`): a king's K standing on the three ridges with a K♠ card index in the corner, the tile in the Mist tokens (the title is in the page's sans like every tile). Keep it in step with the game's tokens.
- **Card motion:** plays and discards fly from the hand to the Played/Discarded piles (other players' from the top edge), a defeated enemy to its slot, drawn cards into the hand.

The `AlertDialog.Action` gotcha bites the solo "Use Jester" confirm: without closing it explicitly, the overlay blocks the hand.
