# Claude Code setup: plan and notes

## Done: frontend-design plugin removed

Uninstalled. It is written for landing pages (hero, page-load reveals, "take aesthetic risk") and triggered on almost any UI task, pulling towards "distinctive" where we want consistency inside existing themes and shared components. Copying it with a "stay inside the tokens" header would give mixed signals (the body still says "take risks") and leave us maintaining a fork of something not meant for game UIs.

Its useful ideas go into our own skills instead:
- **Into design-guidelines:** the list of generic AI-design defaults to avoid (cream + terracotta, near-black + one acid accent, identical rounded cards with the same soft shadow, all-caps eyebrow labels, `A · B · C` meta strings, `→` on buttons), and the copywriting rules (an action keeps its name through the flow: "Leave" → "Left"; errors say what happened and how to fix it, no apologies; sentence case, plain verbs).
- **Into a new `/game-theme <game>` skill** (below).

## Done: design system (2026-10-18)

- **`design-guidelines`** (`.claude/skills/`, Claude-only): the look, familiarity, phone-first and information rules. Loaded on any UI work, so keep it short; add rules there, not here.
- **`/game-theme <game>`** (model-invocable, narrow description): theme from the physical box, plan → approval → tokens → screenshot loop.
- **`/visual-review <game>`** (user-only) + agent **`visual-reviewer`** (read-only, Playwright): fixture screenshot matrix → findings → user picks → implement → before/after.

Open:
- **Reference images:** no game has box/card images in `<game>/docs/reference/` (local, gitignored) yet; game-theme and visual-reviewer need them.
- **Fonts:** the guidelines now fix the roles (sans for UI, display for title/h1–h3/card values, mono for codes). Today The Game has no display face, The Crew uses Space Grotesk in 17 places, Regicide has inline `font-family` for suit glyphs. A `/visual-review` per game will flag the rest.
- **Conventions picked in the guidelines, to confirm:** waiting = text, no spinner (TODO.md); game over = sheet over the table; lobby title on one line at 390 px (Deep Sea).
- **Idea:** Screen Wake Lock during a game (phones lying on the table go dark mid-game).

## Done: `/implement-game` (2026-10-18, not yet run)

- **Skill** `.claude/skills/implement-game/` (user-only): `SKILL.md` (orchestration), `plan-template.md`, `scaffold.sh` (copies Fox Duet's committed files and renames the app identity; tested: the copy passes all checks).
- **Agents:** `rules-writer` (opus), `game-planner` (opus), `game-implementer` (sonnet; the skill passes `model: "opus"` for domain, views, table UI, theme), `rules-test-writer` (sonnet, black-box: reads only rules.md and `go doc`). Existing agents now fixed too: `rules-auditor` opus, `visual-reviewer` sonnet.
- **Model choice:** frontmatter = default for standalone use; the plan's WP table picks the model per package (decided by the Opus planner, as in the Crew plans); the orchestrator may raise a WP to opus, never lower one.
- **The main session only orchestrates:** it reads reports, check summaries and diffs at review gates, not code. Follow-ups go to the same agent via `SendMessage`.
- **Handover:** rules.md + implementation plan (§12 status and open items) are committed after every WP, so `/clear` at a checkpoint and `/implement-game <game>` resumes.
- **Checkpoints:** A rules questions + placement + scope; B plan + theme + static HTML mockup (screenshots via `serve-mockup.mjs`, Playwright blocks `file:`; the same file published as a private artifact for a look on a real phone); C the real UI on fixtures, after the table UI WP and a visual-reviewer pass, while the backend track continues; D play-test + push. Autonomous in between.
- **Placement:** toggle (The Game's Extreme) / edition (Deep Sea) / new app, criteria in the skill.
- **`/visual-review` stays user-only:** implement-game spawns the `visual-reviewer` agent directly and fixes blocking/clarity findings itself; the skill's point is the user picking.
- **No driver script** (`claude -p` per phase): subagents already give fresh contexts, and a script would lose the checkpoints and permission prompts. Revisit if the main context still fills up.
- **Publisher material** goes to `<game>/docs/reference/`, gitignored (public repo).

Open:
- First run will be Regicide Inferno: expect to tune briefs and the skill afterwards (its Retro step).
- Fox Duet as scaffold source: the stripping step costs one Sonnet run per game. If a third game copies the same plumbing unchanged, move more of it to `shared/` or keep a lobby-only template.

## Later / maybe

- `fewer-permission-prompts` run once a few sessions have accumulated.
- A Stop hook that reminds to run `check` when app code changed and no check ran (risk: noise).
