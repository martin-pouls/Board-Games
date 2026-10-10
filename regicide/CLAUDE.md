# Regicide — online multiplayer

A web implementation of the cooperative card game **Regicide** (Badgers From Mars), played live by 1–4 players in the browser. Includes solo mode with Jester powers and Gold/Silver/Bronze victories. Served at `/regicide/`.

The shared stack, architecture, conventions, commands and deployment are in the [root CLAUDE.md](../CLAUDE.md). This file covers what is specific to Regicide. Its lobby, session, WebSocket and persistence plumbing was ported from [the-game/](../the-game/); when in doubt about a convention, look at how The Game does it.

**Rules:** [docs/rules.md](docs/rules.md) is the source of truth. It contains the English rules summary and the interpretations we chose for ambiguities (marked **App decision**). Read it before touching game logic. Do not re-read the PDF. If a rule question is not answered there, ask, then record the answer in that file.

## App specifics

- Go module `github.com/martin-pouls/regicide/backend`. Default DB `data/regicide.db`.
- `frontend/src/lib/game/` also holds the play/discard helpers. Game screen components: Lobby, Table, Hand, EnemyCard, ...
- **Intents:** "play these cards", "yield", "discard these cards", "choose next player", "use solo Jester".
- **Per-player view:** own hand, other players' hand *counts*, current enemy (suit, rank, health, accumulated damage, shield, effective attack, immunity status), cards played against the enemy, Tavern/Castle/discard counts, phase, whose turn, yield flags, Jester signals, solo Jesters left. Never send the order of the Tavern or Castle deck.
- **Deployment:** container `regicide`, SQLite in the volume `regicide_data`, Caddy snippet [deploy/regicide.caddy](deploy/regicide.caddy).

## Domain modelling rules

- **Cards are structs** (`Card{Suit, Rank}`), with Jester as its own rank/kind. Values are derived by functions (attack value in hand, discard value, enemy attack/health), never stored redundantly.
- **The RNG is injected** (seedable), so shuffles are deterministic in tests.
- **Phases are explicit:** `lobby → playing{step: play | discard | chooseNext} → won | lost`. Every action checks the phase and the active player first.
- **Derived values are computed, not cached:** shield, effective attack, immunity cancellation and damage are derived from the cards played against the current enemy plus a "Jester played against this enemy" flag, so the snapshot stays minimal and consistent.
- **The game is lost only by the rules in docs/rules.md §9** (or by give-up).
- Player count 1–4; ruleset values (hand size, Jester count) come from a table keyed by player count, not scattered constants.

## Look and feel

Same stack and component library as The Game, but a distinct identity: a dark-only, medieval "castle" theme (deep colours, gold accents, Cinzel as the serif display font for headings and card ranks, EB Garamond for text), rendered cards with suit glyphs (♥ ♦ ♣ ♠; hearts/diamonds red), and a prominent **enemy card** showing health bar, accumulated damage, attack, shield and immunity. The hand supports multi-select so combos, Animal Companion pairs and discard sets can be composed, with a live preview of attack value and which suit powers will trigger.

The bits-ui `AlertDialog.Action` gotcha (root CLAUDE.md) bites the solo "Use Jester" confirm: without closing it explicitly, the overlay blocks the hand.
