# The Game — online multiplayer

A web implementation of the cooperative card game **"The Game"** (Steffen Benndorf, NSV), played live by 1–5 players in the browser. Served at `/the-game/`.

The shared stack, architecture, conventions, commands and deployment are in the [root CLAUDE.md](../CLAUDE.md). This file covers what is specific to The Game.

- **Phase 1:** the regular game, including the Pro variant. Implemented.
- **Phase 2:** **The Game Extreme**, which adds command cards. Implemented. The host enables it with the *Extreme* lobby toggle, which combines freely with Pro and Reduced hand size.

Keep every domain concept open for further card kinds (the mini-expansions Joker, Zehner, On Fire), the way Extreme slotted in (see *Extensibility* below).

**Rules:** [docs/rules.md](docs/rules.md) is the source of truth. It contains the English rules summary, the Extreme card→command mapping, and the interpretations we chose for ambiguities. Read it before touching game logic. Do not re-read the PDFs or card photos. If a rule question is not answered there, ask, then record the answer in that file.

## App specifics

- Go module `github.com/martin-pouls/the-game/backend`. Default DB `data/the-game.db`.
- `frontend/src/lib/game/commands.ts` holds the Extreme command metadata. Game screen components: Lobby, Board, GameOver, ...
- **Per-player view:** own hand, piles, draw-pile count, other players' hand *counts*, whose turn, active commands, etc.
- **UI:** shadcn base color `zinc` tinted toward blue, Inter, light/dark theme toggle (mode-watcher). The accents are the blue and orange of The Game Extreme, which are also the ascending (sky) and descending (amber) pile colours: a blue primary and two soft page glows (`--glow-up`, `--glow-down` in `routes/layout.css`). Violet stays reserved for Infinity commands. The start page tile uses the same colours.
- **Deployment:** container `the-game`, SQLite in the volume `the-game_data`, Caddy snippet [deploy/the-game.caddy](deploy/the-game.caddy).

## Extensibility for The Game Extreme

These are built in and Extreme relies on them; keep them when adding card kinds:

- **Variant / ruleset config on the game:** variant (`classic` | `extreme`), min cards per turn, hand-size table, Pro flag. Rules read from the config and never hard-code 2/6/7/8.
- **Cards are structs, not ints.** For example `Card{Value int; Command *Command}` (or a kind field), because Extreme cards carry commands and the mini-expansions (Joker, Zehner, On Fire) would add other card kinds.
- **The deck is built by a variant-specific factory.** Extreme swaps in the 28 command cards listed in docs/rules.md.
- **The rules follow the cards, not the variant.** No rule code checks the variant; a classic deck simply has no command cards, so every command rule is a no-op there.
- **Turn state is explicit.** Track cards played this turn and the piles touched, because commands constrain the *current turn*: STOP blocks further cards, 3! needs an exact count, Single pile, Skull must be covered.
- **Placement validation is a composable rule check.** Base ascending/descending/backwards-trick rules, plus modifiers derived from the active (visible) Infinity commands and the lightning commands played this turn.
- **Hooks at turn end:** end-of-turn checks (Skull covered, 3! satisfied) and a draw-count policy (normal refill vs. Draw only one).
- **Communication restriction is game state.** The UI must be able to disable chat or hints while a "No communication" card is visible. There is no chat yet; a future chat must enforce the flag server-side.
- **Illegal moves are blocked, never punished.** Every action that would break a rule or command is rejected with a typed domain error, and state stays unchanged. The game is lost only when the active player is *stuck*: no legal card play and ending the turn is blocked. See *Enforcement* in docs/rules.md.
- **Stuck detection has no lookahead** (owner decision): the game is lost only when the active player has no legal card play *and* cannot end the turn. A player who can still play a card but can no longer reach the minimum keeps playing until nothing fits.
- **Turns end only by End turn, and plays of the current turn can be undone** (owner decision). Command state is derived from the piles and `Turn.Plays`, never cached, so `UndoPlay` just pops the last play. Being stuck loses automatically only at turn start; mid-turn the player undoes or gives up (`GiveUp`).
