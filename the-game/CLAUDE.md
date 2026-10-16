# The Game

**The Game** (Steffen Benndorf, NSV), cooperative for 1–5 players, with the Pro variant and **The Game Extreme** (command cards; the host's *Extreme* lobby toggle combines freely with Pro and Reduced hand size). Common rules: root [CLAUDE.md](../CLAUDE.md). Rules: [docs/rules.md](docs/rules.md) (rules summary, Extreme card→command mapping, App decisions).

## Intents and per-player view

- **Intents:** set options (host, lobby), start, choose first player (host, after the deal), play a card on a pile, undo the last play, end turn, give up.
- **Phases:** the extra `choosingFirstPlayer` sits between `lobby` and `playing`. A lost game records why (`stuck` at turn start, or `gaveUp`) and who; the score is the cards left.
- **Per-player view:** own hand, piles (top card and count), draw-pile count, other players' hand *counts*, whose turn, the current turn's plays, active commands, stuck flag; after the game all hands. Never other hands or the deck order.
- `frontend/src/lib/game/commands.ts` holds the Extreme command metadata (also used by `RulesHelp`).

## Domain modelling (open for more card kinds)

Keep every concept open for further card kinds (the mini-expansions Joker, Zehner, On Fire), the way Extreme slotted in:

- **Ruleset config on the game:** variant (`classic` | `extreme`), min cards per turn, hand-size table, Pro flag. Rules never hard-code 2/6/7/8.
- **Cards are structs** (`Card{Value, Command}`), and the deck comes from a variant-specific factory (Extreme swaps in the 28 command cards).
- **The rules follow the cards, not the variant.** No rule code checks the variant; a classic deck has no command cards, so every command rule is a no-op there.
- **Turn state is explicit** (cards played this turn, piles touched), because commands constrain the current turn: STOP, 3!, Single pile, Skull must be covered. Command state is derived from the piles and `Turn.Plays`, never cached, so `UndoPlay` just pops the last play.
- **Placement validation is composable:** base ascending/descending/backwards-trick rules plus modifiers from the visible Infinity commands and this turn's lightning commands. End-of-turn hooks check Skull and 3! and pick the draw policy (refill vs. Draw only one).
- **Communication restriction is game state.** The UI disables hints while "No communication" is visible; a future chat must enforce it server-side.
- **Turns end only by *End turn*** (owner decision); plays of the current turn can be undone.
- **Stuck detection has no lookahead** (owner decision): stuck = no legal card play *and* the turn cannot end. Stuck at turn start loses automatically; stuck mid-turn the player undoes or gives up.

## Look and feel

shadcn base colour `zinc` tinted toward blue, Inter. The accents are the blue and orange of The Game Extreme, which are also the ascending (sky) and descending (amber) pile colours: a blue primary and two soft page glows (`--glow-up`, `--glow-down` in `routes/layout.css`). Violet is reserved for Infinity commands. The start page tile uses the same colours. On phones the in-game header shortens the EXTREME badge to "EXT" so Rules and Give up fit at 360 px. Card motion: plays fly from the hand onto the pile (undo flies back), other players' plays come from the top edge, refills drop into the hand; cards are keyed by value (`lib/game/motion.ts`), so a card keeps its key between hand and pile.
