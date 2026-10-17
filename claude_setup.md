# Claude Code setup: plan and notes

## Done: frontend-design plugin removed

Uninstalled. It is written for landing pages (hero, page-load reveals, "take aesthetic risk") and triggered on almost any UI task, pulling towards "distinctive" where we want consistency inside existing themes and shared components. Copying it with a "stay inside the tokens" header would give mixed signals (the body still says "take risks") and leave us maintaining a fork of something not meant for game UIs.

Its useful ideas go into our own skills instead:
- **Into design-guidelines:** the list of generic AI-design defaults to avoid (cream + terracotta, near-black + one acid accent, identical rounded cards with the same soft shadow, all-caps eyebrow labels, `A · B · C` meta strings, `→` on buttons), and the copywriting rules (an action keeps its name through the flow: "Leave" → "Left"; errors say what happened and how to fix it, no apologies; sentence case, plain verbs).
- **Into a new `/game-theme <game>` skill** (below).

## Next: `/game-theme <game>` skill (model-invocable, narrow description)

Run only when a game gets or reworks its look: by the user directly, or by `/visual-review` and `/implement-game`, which call it via the Skill tool. So **no** `disable-model-invocation` (that would hide it from Claude and block those calls). The description keeps it from firing on ordinary UI work: "Create or rework a game's visual theme (palette, typography, card faces) from its physical box. Use only when the user asks for a new or reworked look, or when visual-review or implement-game reaches its theme step. Not for layout fixes or small UI changes." Short (~30 lines):

1. **Ground it in the physical game:** box and card images from `<game>/docs/`: palette, typography, card faces, materials.
2. **Token plan first:** 4–6 named colours (light and dark), typefaces and their roles, the game screen as an ASCII wireframe. Review it against the generic-defaults list in design-guidelines and revise anything that would come out the same for any game.
3. **Spend boldness in one place** (usually the card faces); everything else stays quiet and uses the shared components unchanged.
4. Wait for the user's approval of the plan, then implement in `src/routes/layout.css` and the card components.
5. Screenshot via the dev fixtures in light and dark on phone viewports, critique, fix, show the result.

The real quality lever is the screenshot loop (fixtures + visual-review), not prose: Claude follows design text loosely but reacts reliably to a screenshot that overflows.

## Next: design-guidelines skill (Claude-invocable, `user-invocable: false`)

Loaded automatically whenever Claude touches a frontend. Content:

**Distinct per game, coherent across games**
- Each game's look comes from its physical box: palette, typography, card faces, iconography (tokens in `src/routes/layout.css`, `:root` light and `.dark` dark).
- Shared and identical everywhere: PageFrame, header buttons (X / DoorOpen / Flag, text in the lobby, icon + tooltip in game), LobbyActions and confirmations, ConnectionBanner, card motion, new-card marks, RulesHelp placement, the home/join pages. Add to this list rather than diverging. When a game needs something different, make it a prop or snippet in `shared/`.
- Still inconsistent today (TODO.md): waiting for another player's turn (spinner or not), lobby title length (the Deep Sea name doesn't fit).
- *Missing:* a table of which concepts are shared and which are game-specific: turn indicator, "your turn" cue, error toasts, game-over screen layout, player list and seat display.

**Mobile first**
- Phones in portrait first. Must fit iPhone 13 (390×844) and iPhone 16 (393×852). Also check a small Android (360×800) and desktop.
- Game screen: never scroll, and no layout jumps when state changes (reserve space for banners, hints and changing counts; fixed-size slots for cards).
- Lobby, home and other setup screens may scroll.
- *Missing:* safe-area insets (`env(safe-area-inset-*)`, notch and home indicator), `dvh` instead of `vh` (Safari's toolbars), touch targets ≥ 44 px, landscape behaviour (allowed to be merely usable?), on-screen keyboard on the join page.

**Information**
- Screen space is valuable. Show each piece of information once, clearly. Deliberate exceptions are allowed when a rule needs emphasis (The Game's "no talking").
- The game decides the hierarchy: what's important differs per game (The Crew: tasks and communication; Regicide: the enemy's health/attack).
- *Missing:* text and suits legible in both light and dark mode, the copy language and tone (one language per app, consistent action names: "Leave" button → "Left" toast).

**Process**
- Work inside the theme tokens and shadcn components. No ad-hoc colours.
- Before calling UI work done: Playwright screenshots on iPhone 13/16 viewports in light **and** dark mode.

## Next: visual review (skill + `visual-reviewer` agent)

Goal: look at one game, judge its look and UX against design-guidelines and the physical game, and propose improvements **before** changing code (screenshots + suggestions up front). After approval, implement and show before/after.

- **Agent `visual-reviewer`** (read-only: Playwright, Read, Grep, Glob). Shoots every screen in a fixed matrix: {home, lobby (host + guest), game screen in its key states, game over (won/lost)} × {iPhone 13, iPhone 16, desktop} × {light, dark}. Reports issues ranked by severity: scroll or jump on the game screen, overflow and truncation, redundancy, inconsistency with the other games' shared elements, theme fit, legibility in both modes.
- **Skill `/visual-review <game>`** (user-only). Runs the agent, presents its findings with screenshots, waits for the user's pick, implements (invoking the `game-theme` skill for theme work), reshoots the same matrix and shows before/after pairs.
- **Prerequisites done (2026-10-17):** the `run` skill (starts/stops apps on their fixed ports) and dev fixtures in every game (`/dev/fixtures/<name>` renders a fixture on the real game screen; see the root CLAUDE.md). The agent drives those instead of playing games into states.
- Screenshots: the Playwright MCP may only write inside the repo, so they go to `.playwright-mcp/` (gitignored), not the scratchpad.
- Comparison against the physical game: the agent gets the box/card images from `<game>/docs/` (add them there) as reference.

## Idea: `/implement-game` skill (user-only)

Implements a new game end to end from reference material (rules PDF, images, links), building on add-app.

1. **Intake** (main model): read the material and write `<game>/docs/rules.md`. Collect every ambiguity as a question for the user. Answers become **App decision**s. Nothing is built until the rules doc is approved.
2. **Plan** (main model): a phased implementation plan like `the-crew/docs/implementation-plan.md` (domain → ports/service → adapters/e2e → frontend plumbing → table UI → theme), each phase with an owner model, inputs and acceptance commands.
3. **Scaffold:** add-app checklist, copying Fox Duet.
4. **Build:** delegate mechanical phases (adapters, DTO mirror, e2e, plumbing, fixtures) to Sonnet agents with tight briefs ("follow fox-duet file by file"). Keep the domain and the per-player views on the main model, or at least review them there. Run `check` after every phase.
5. **Verify:** rules-auditor → fix → code review (`/code-review`) → design-guidelines + visual-review once they exist → `/check --all`.
6. **Finish:** start-page tile, privacy policy check, CLAUDE.md for the game.

Each phase ends with a short status for the user. The user approves the rules doc and the theme direction before the work goes on.

## Later / maybe

- `fewer-permission-prompts` run once a few sessions have accumulated.
- A Stop hook that reminds to run `check` when app code changed and no check ran (risk: noise).
