# Fox Duet — online multiplayer

A web implementation of the cooperative trick-taking game **The Fox in the Forest: Duet** (Foxtrot Games), played live by exactly 2 players in the browser, with the three difficulty levels. Served at `/fox-duet/`.

The shared stack, architecture, conventions, commands and deployment are in the [root CLAUDE.md](../CLAUDE.md). This file covers what is specific to Fox Duet. Its lobby, session, WebSocket and persistence plumbing comes from [shared/](../shared/README.md). Everything that isn't game logic follows [regicide/](../regicide/); when in doubt about a convention, look there.

**Rules:** [docs/rules.md](docs/rules.md) is the source of truth. It contains the English rules summary, the board tables and the interpretations we chose (marked **App decision**). Read it before touching game logic. Do not re-read the PDF or the board photos. If a rule question is not answered there, ask, then record the answer in that file.

## App specifics

- Go module `github.com/martin-pouls/fox-duet/backend`. Default DB `data/fox-duet.db`. Local ports: Vite `:5176`, backend `:8082`.
- **Intents:** "start at level N", "play this card", "choose the Foxes player", "swap the decree / keep it", "give this gift card", "choose this move", "cover this end" (host, round end), "give up".
- **Per-player view:** own hand (with the playable cards), the teammate's hand *count*, the board (gems per location, plus, covered spaces, traveller, forest tokens left), round, trick number, dealer, decree card, the current trick, the last completed trick, the pending decision with its options for the viewer, the viewer's own gift card, and the result with the score. Never send the teammate's hand, their gift choice or the deck.
- **Deployment:** container `fox-duet`, SQLite in the volume `fox-duet_data`, Caddy snippet [deploy/fox-duet.caddy](deploy/fox-duet.caddy).

## Domain modelling rules

- **Cards are structs** (`Card{Suit, Rank}`). Movement and ability are functions of the rank, never stored.
- **Boards and levels are tables** (`boards`, `levels` in `board.go`): the topology of sides A and B, starting gems, pluses, start space, forest tokens and base score. No scattered constants.
- **Board state is minimal:** gems per location, covered count per end, traveller position, forest tokens left. Positions are indices from the left end (seat 1's end). A negative delta moves toward the left end.
- **The RNG is injected** (seedable), so shuffles are deterministic in tests. Every mutator that can end a round takes it, because the next round is dealt right away.
- **Phases and steps are explicit:** `lobby → playing{step: lead | follow | foxesChoose | foxesSwap | gift | move | forestEnd} → won | lost`, plus `lobby → ended` when the host closes the lobby. Every action checks the phase, the step and the acting player first.
- **The pending ability belongs to the card played last:** with no follow card yet it was the lead card, otherwise the follow card. Who acts is derived from that, not stored.
- **Derived values are computed, not cached:** the trick winner, the legal cards and the move options come from the trick, the decree and the board.
- **Automatic choices** (rules §10) are resolved inside the domain, in the same mutator that reaches them.
- **The game is lost only by the rules in docs/rules.md §8** (or by give-up).

## Look and feel

A forest theme, distinct from The Game and Regicide: moss green, fox orange and parchment in light mode, deep forest night in dark mode. Cards are rendered with their suit glyph, rank, paw-print movement and ability name. The **board** is an SVG with the real topology (path spaces, locations with gems and pluses, forest tokens, traveller), rotated so that each player's own end is on the left. Below it sit the trick row (decree, lead, follow), a status line or decision panel, and an overlapping hand where illegal cards are dimmed. Mobile-first: the board is wide and short, so it fits in portrait above the hand. Card motion (deal, play, gift) uses the headless hand/card-motion layer in `shared/frontend`. Light/dark toggle via mode-watcher; theme tokens live in `src/routes/layout.css` (`:root` light, `.dark` dark), so components use tokens, not raw colours.
