---
name: design-guidelines
description: Look, layout and UX rules for every app's frontend (games, start page, gate). Use whenever you build or change UI in this repo: screens, components, styles, copy.
user-invocable: false
---

# Design guidelines

When rules pull apart: **playability > clarity > familiarity across games > looks.** Never trade gameplay for aesthetics.

## Look: from the physical game
- Each game's palette, type, card faces and icons come from its box (reference images in `<game>/docs/reference/`, local only). New or reworked look: the `game-theme` skill.
- Colours only via the tokens in `src/routes/layout.css` (`:root` light, `.dark` dark). No hex or Tailwind palette colours (`bg-amber-500`) in components; card and suit colours are tokens too.
- **Fonts, per game:** `--font-sans` for all UI text and numbers; `--font-display` (optional) only for the game title, h1–h3 and card values. Buttons, labels and hints use sans. `font-mono` only for game codes (shared). No `font-family` in components: glyph fallbacks (suit symbols) get a class in layout.css.
- **Text:** Tailwind's scale, at most three sizes on the game screen. Nothing below `text-xs`, and that only for secondary labels. Inputs ≥ 16 px (iOS zooms otherwise).
- Boldness in one place (usually the card faces). Everything else stays quiet and uses the shared components unchanged.
- **Avoid generic AI defaults:** cream + terracotta, near-black + one acid accent, identical rounded cards with the same soft shadow, all-caps eyebrow labels, `A · B · C` meta strings, `→` on buttons, gradients and glows for their own sake.

## Familiar across games
- **Shared and identical:** PageFrame, home and join pages, HeaderButton (X close lobby, DoorOpen leave, Flag give up/end; text in the lobby, icon + tooltip in game), LobbyActions and its confirmations, ShareInvite, ConnectionBanner, card motion, NewCardMark, theme toggle, RulesHelp in the table header. Need something different? Add a prop or snippet in `shared/`, don't fork.
- **Same concept, same treatment:**
  - *Whose turn:* always visible on the game screen. Your turn is unmistakable (seat or hand highlight **plus** text), never colour alone.
  - *Waiting:* plain text in the action area, "Waiting for Alice to play." No spinner: a spinner means the app is loading.
  - *Moves:* tap selects, a clear button commits. Illegal options are disabled with the reason shown, not rejected after the tap.
  - *Server errors:* a toast saying what happened and how to fix it.
  - *Lobby:* title, seats, settings, primary action in a sticky bottom bar. The title fits on one line at 375 px.
  - *Game over:* a sheet over the table (the table stays visible), outcome first, rematch as the primary button.
  - A game may deviate where its flow needs it; say why in its CLAUDE.md. Example: The Crew's per-mission debrief is an inline panel in the table, since it leads straight into the next mission.
- **Copy:** English, sentence case, plain verbs. An action keeps its name through the flow ("Leave" → "Left"). No apologies.

## Phone first, desktop works
- **Targets:** iPhone 13 mini (375×812) and iPhone 16 (393×852), each also in Safari with its toolbars showing (375×629 and 393×659: the height a player really gets, so check the game screen there; the mini is the tight one), desktop (1280×800). A desktop-only game says so in its CLAUDE.md and shows phones a notice instead of a broken table.
- **Game screen: no scrolling** at the targets, neither page nor panels (dialogs like RulesHelp may scroll). Use `dvh`/`svh`, never `vh`. Pinned edges respect `env(safe-area-inset-*)`.
- **No layout jumps:** reserve space for hints, banners and changing counts; fixed-size card slots; tabular numbers.
- **Thumb zone:** your hand and the action button at the bottom, shared state above. Touch targets around 44 px: a guideline, not a floor. Dense boards may go a few px below when that buys better spacing, never drastically. Overlapping cards still need a clear tap area. No hover-only information. No gestures without a visible alternative.
- Landscape on phones must not break; it need not be optimal. Setup screens (home, join, lobby) may scroll.
- Desktop: centred with a max width; cards don't grow without limit; no stretched phone layout.

## Information
- Each piece of information once, clearly. Deliberate exceptions for rules that need emphasis (The Game's "no talking").
- On the game screen, card size beats labels: a tab, heading, scale or hint line that makes the cards smaller has to earn it.
- The game decides the hierarchy: what the player must decide **now** is visible without tapping (The Crew: tasks and communication; Regicide: enemy health and attack).
- What just happened stays readable for a player who looked away: last play, new-card marks, card motion.
- Colour is never the only carrier: suits also differ by symbol. Text contrast ≥ 4.5:1, icons and marks ≥ 3:1, in light **and** dark.
- Every tap answers at once (pressed, selected or disabled state). Motion ≤ 300 ms, only to explain a change, and off under `prefers-reduced-motion`.
- Confirm only what ends or leaves the game (give up, leave, end). Moves use select + commit instead.

## Process
- Work inside the theme tokens and the shadcn components.
- Check through the dev fixtures (`run` skill), not by playing: targets × light and dark. On the game screen assert `document.documentElement.scrollHeight <= innerHeight`.
- A new UI state gets a fixture in the same change.
- Look and UX audit of a whole game: the user runs `/visual-review <game>`.

## Previews (artifacts)
Show a look before building it only when the user has to decide between looks before any code exists. Pick the cheapest form that fits:
- **No preview:** small fixes, layout tweaks, anything the fixtures can show. Fixture screenshots are enough.
- **Plain HTML artifact (default):** theme plans, a new screen, a reworked component. One self-contained file in the scratchpad, using the game's real or planned tokens as CSS variables and its fonts, in light and dark, at real sizes, with a 375 px game screen mockup. Each revision edits and republishes the same file, so the URL stays the same.
- **Claude Design artifact:** only when the user asks for one, or wants several variants side by side on a canvas to compare and comment on. Start it with `design_systems: false`: the game's tokens are the design system.
- Previews are throwaway. Once the code exists, the fixtures are the truth: don't keep artifacts in sync, and never copy markup from them into the app (it gets rebuilt from Svelte and shadcn components).
- Only the main session publishes. Subagents (`visual-reviewer`) return files and findings.
