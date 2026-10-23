---
name: game-theme
description: Create or rework one game's visual theme (palette, typography, card faces, its start-page tile) from its physical box. Use only when the user asks for a new or reworked look, or when visual-review or implement-game reaches its theme step. Not for layout fixes or small UI changes.
argument-hint: <game>
---

# Theme for $ARGUMENTS

Follow the design-guidelines skill throughout; this adds the theme process.

1. **Ground it in the physical game.** Look at the box and card images in `<game>/docs/reference/` (gitignored, the repo is public; none there? ask the user for some, don't invent from the name). Note palette, typefaces or their feel, card layout, materials (linen, wood, foil), iconography. Read the current `src/routes/layout.css` and card components.
2. **Plan before code**, shown to the user as a plain HTML preview artifact (see Previews in design-guidelines). Inside `/implement-game` the preview is the brief's `<game-docs>/mockup.html` instead, and the orchestrator publishes it:
   - 4–6 named colours with their role, as light and dark tokens; contrast checked (text ≥ 4.5:1, marks ≥ 3:1).
   - Typefaces: one sans for UI and numbers, at most one display face, each with its roles. Fontsource variable packages only, nothing from a CDN.
   - Card faces: what carries value, suit and colour, at the size the game screen gives them on a 375 px phone.
   - The game screen at 375×629 (13 mini in Safari, the tightest) and 393×852 (iPhone 16) as a mockup, plus a few card faces at real size. Take the header, footer and other shared chrome at the heights measured in the current fixtures, so the card sizes the mockup promises hold once built.
   - **The game's start-page tile** (`start-page/frontend/src/routes/+page.svelte`, tokens in its `layout.css`), next to two neighbouring tiles as they are today, light and dark: tile background, art strip, name in the game's display face, accent colours, all taken from the new tokens so the tile reads as the same game. The start page keeps its own table and chrome; only the tile changes.
   - Then review the plan against the generic-defaults list in design-guidelines and revise anything that would come out the same for any game.
   - Decide in rounds: keep what the user has locked fixed, and show two or three options only for what is still open. Revisions republish the same artifact.
3. **Boldness in one place**, usually the card faces. Background, panels and shared components stay quiet and unchanged.
4. **Wait for the user's approval**, then implement: tokens in `src/routes/layout.css` (`:root` and `.dark`), fonts in `package.json` and layout.css, card components. Mirror changed background colours in `src/app.html` (theme-color and the first-paint script). Deep Sea: tokens go in `the-crew/frontend/src/routes/theme-deep-sea.css` under `[data-edition='deepSea']`.
   Then the tile in `start-page/frontend/`: copy the changed values into its game tokens (`:root` and `.dark` in `layout.css`, plus the `@theme` mapping) and update the comment above them; rework the tile's art strip and classes in `+page.svelte`. A display face the tile uses goes into the start page's `package.json` and `layout.css` too; drop a Fontsource package no tile uses any more. Shorten the tagline in `games.ts` if the new tile needs it.
5. **Screenshot loop:** start the game and the start page (`run` skill), shoot the main fixtures, the home page, the lobby and the start page (tiles) at the design-guidelines targets in light and dark, critique against the plan and the guidelines, fix, repeat. Show the user the final shots (in the chat; no artifact needed). Inside `/implement-game` skip this step: its UI checkpoint reviews the theme on the real table.
6. Run the `check` skill for the game and `start-page`.
