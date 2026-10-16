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

A medieval "castle" theme: parchment and ink in light mode, deep colours in dark mode, both with gold accents; Cinzel for headings and card ranks, EB Garamond for text; cards with suit glyphs (hearts/diamonds red). The hand supports multi-select to compose combos, Animal Companion pairs and discard sets, with a live preview of attack value and triggered suit powers. At the table the enemy panel is compact (small card, HP on the bar, attack − shield = damage as symbols, one overlapping stack per player), a Played/Discarded row shows the last turn with icons, and the action area is one slim bar above the hand (suit-power badges, no damage arithmetic). The hand sort menu sits in the table header. Card motion: plays and discards fly from the hand to the Played/Discarded piles (other players' from the top edge), a defeated enemy to its slot, drawn cards into the hand.

The `AlertDialog.Action` gotcha bites the solo "Use Jester" confirm: without closing it explicitly, the overlay blocks the hand.
