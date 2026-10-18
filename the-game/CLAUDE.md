# The Game

**The Game** (Steffen Benndorf, NSV), cooperative for 1–5 players, with the Pro variant and **The Game Extreme** (command cards; the host's *Extreme* lobby toggle combines freely with Pro and Reduced hand size). Common rules: root [CLAUDE.md](../CLAUDE.md). Rules: [docs/rules.md](docs/rules.md) (rules summary, Extreme card→command mapping, App decisions).

## Intents and per-player view

- **Intents:** set options (host, lobby; the host's last options are remembered per device in `the-game:lobby-options` and sent with `POST /api/games`), start, choose first player (host, after the deal), play a card on a pile, undo the last play, end turn, give up.
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

The Game Extreme, made calm. Tokens in `routes/layout.css`, background mirrored in `app.html` and `+layout.svelte`.

- **Palette:** a plain pale blue-grey table (no page glows). The box's blue and orange are the pile colours: `--up-*` and `--down-*` (mat edge, soft tint, label text, solid direction pill and its ink). Green (`--go*`) means "your turn" and "played this turn". `--extreme` is the orange tape of the EXTREME tag.
- **Type:** Inter for all UI, League Gothic (`font-display`) only for card values and the home page's start values.
- **Cards (`GameCard`):** navy face (`--card-face`) with a cream numeral, sized by an inherited `--w`. The value appears in the top left and bottom right corners, with the kind mark (Infinity or lightning) below it. Commands use their printed badges (`CommandBadge`): Infinity as a blue square, lightning as an orange triangle pointing down. `mode="start"` shows an empty pile's start value as a dashed outline. `mode="mini"` (28 px) is used in the turn bar and on the results screen.
- **Piles:** 2 × 2, each a mat that is the whole tap target. A solid arrow pill sits top left and a green "+N" (cards played there this turn) sits top right. There is no card count. `Board` computes one card size for all piles from the grid, so the card clears the pills; pile size never changes during a game. Bound pile: dashed green outline. Legal pile: solid green outline. Skull: red edge and "Cover!".
- **Below the piles, all at fixed heights:**
  - Turn bar (48 px): whose turn; then your count with dots, or the other player's plays as mini cards; then the rules in force as badges (`turnStatus().rules`).
  - Button row (48 px): Undo and End turn, or "Waiting for X to play."
  - One-row hand: card width comes from the game's hand size, cards overlap so the corner value and half the numeral stay visible, and covered cards drop their bottom corner value. New cards get `NewCardMark dot="corner-left"`.
  - There is no hint line under the hand.
- **Header:** on phones the EXTREME tag shortens to "EXT", so Rules and Give up fit at 360 px.
- **Player strip:** one row. Names truncate, your own pill says "You", the active player's pill is green, and there is no crown in game.
- **No communication:** a blue Infinity glow around the screen (`NoCommunicationFrame`).
- **Card motion:** plays fly from the hand onto the pile (undo flies back), other players' plays come in from the top edge, and refills drop into the hand. Cards are keyed by value (`lib/game/motion.ts`), so a card keeps its key between the hand and the pile.
