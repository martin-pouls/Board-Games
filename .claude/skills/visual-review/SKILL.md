---
name: visual-review
description: Review one game's look and UX through fixture screenshots, propose improvements, implement the chosen ones and show before/after.
argument-hint: <game>
disable-model-invocation: true
---

# Visual review of $ARGUMENTS

1. **Start** the game and `start-page` with the `run` skill (Deep Sea: `deep-sea`). Note which apps were already running; stop only the ones you started, at the end.
2. **Review:** spawn the `visual-reviewer` agent with the game, its frontend URL, the start page URL (`http://localhost:5175/`) and `.playwright-mcp/review-<game>/before/` as screenshot directory.
3. **Present** its findings to the user grouped by severity, each with its screenshot (Read the file to show it) and the proposed fix. Drop findings you can't confirm on the screenshot. Ask which to implement. **Change nothing before the user picks.**
4. **Implement** the picked fixes, following design-guidelines. Picked *theme work* goes through the `game-theme` skill. A tile fix changes only that game's tile in `start-page/` (tokens copied from the game's `layout.css`). A fix that belongs in `shared/` changes every game: say so before doing it.
5. **Reshoot** the same screens, viewports and modes into `.playwright-mcp/review-<game>/after/` (yourself, or the agent again for larger changes) and show before/after pairs for every change. Show the pairs in the chat. Publish a gallery artifact (screenshots uploaded as assets, not embedded) only if the user wants to share them.
6. Run the `check` skill (plus `start-page` if the tile changed), stop the apps you started, delete `.playwright-mcp/review-<game>/`.
