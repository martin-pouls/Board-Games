# Start page

The landing page at `/board-games/`: one tile per game, plus imprint and privacy policy (German). Frontend only, fully prerendered with `csr = false`, so no framework JS ships. The game list is `src/lib/games.ts`.

## Look

- **Slate table:** quiet slate background tokens in `src/routes/layout.css`, red meeple as the accent. The gate's login page uses the same base tokens: change both together, including `theme-color` in `src/app.html` and `static/favicon.svg`.
- **One title face for the whole page:** every tile title is Inter extra bold, like the page heading, in its game's colour. The games' own fonts (League Gothic, Alegreya) appear only inside the tile art. Don't give titles a per-game font.
- **Each tile borrows its game's look through colour and art only:** tile, text and ring colours come from `--<game>-*` tokens in layout.css (copied from the game's theme, since apps can't import each other), the art from `src/lib/components/GameArt.svelte`. A game's theme change updates its tokens and art here too (the `game-theme` skill covers it).
  - The Game: its two printed start cards, fanned −5°/+5° (mirrors `StartCard` in the-game).
  - Regicide: a K on the three ridges with a K♠ index.
  - Fox Duet: a slice of the board (path, stepping stones, gem tiles, the fox disc).
  - The Crew, Deep Sea: icon art until their themes are redone.
- **Layout:** phone is a shelf of rows (art square on the left, title, player count, one-line `tagline`, chevron). Desktop is a grid of fixed-size tiles (2 per row from `md`, 3 from `lg`), art band on top and the longer `description`; a short last row is centred. No feature tags: they added nothing and cost space.
- The art band is a size container: everything in it scales with its height (`cqh`), and SVG scenes use `slice`, so the phone's square shows the middle of the desktop band.
