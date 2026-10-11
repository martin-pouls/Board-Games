# Ideas for future games

Short list of games we might add. Not a plan: no rules have been checked against the publishers' PDFs, and facts below come from web summaries (retailer, library and review pages), so verify before building. Per [CLAUDE.md](CLAUDE.md), each game would get its own app and its own `docs/rules.md` first.

## Overview

| Game | Players | Time | Type | Fit with current stack | Effort |
|---|---|---|---|---|---|
| Regicide Inferno | 1–4 | not stated | Cooperative, Regicide with classes and a growing enemy | Excellent (same lobby, session, hand and enemy-flow plumbing as Regicide) | Medium |
| The Crew: Quest for Planet Nine | 2–5 | ~20 min/mission | Cooperative trick-taking, no talking | Very good (same hidden-hand and per-player-view model) | Low |
| The Crew: Mission Deep Sea | 2–5 | ~10–20 min/mission | Same, bigger task deck | Very good; share the engine with Planet Nine | Low (after Planet Nine) |
| Sail | 2 | ~20 min | Cooperative trick-taking + track | Good | Low–medium |
| Sky Team | 2 | ~15–20 min | Cooperative dice placement, no talking | Needs a simultaneous, secret dice-placement UI | Medium |
| LOTR Trick-Taking (Fellowship, Two Towers) | 1–4 | ~15–30 min | Cooperative trick-taking, character goals | Good; shares a trick-taking core with The Crew | Medium |
| Aeon's End | 1–4 | ~60 min | Cooperative deckbuilder | Large card pool, many effects | High |
| Marvel Champions | 1–4 | 45–90 min | Cooperative Living Card Game | Huge card pool and effect system | Very high |

Cheap wins first: the four trick-taking games (The Crew ×2, Sail, LOTR ×2) can share one trick-taking module of the domain (follow suit, trick winner, trump, goals checked per trick).

## Regicide Inferno

- **What:** A standalone Regicide follow-up by the same publisher (Badgers From Mars, art by Sketchgoblin). Rulebook: A4 PDF, 4 pages, in Downloads (`A4-Rulebook-RegicideInferno.pdf`, image-only, no text layer). Facts below are from my reading of it; the next step is to turn it into `docs/rules.md` and settle the open questions at the end.
- **Aim:** Defeat 15 enemies: 4 Justicars (10/20), 4 Queens (15/30), 4 Kings (20/40) and 3 Mage enemies (strength/health below). Same loss condition as Regicide: you must discard enough cards to cover the enemy's strength, or everyone loses.
- **Differences from Regicide** (the real work):
  - **Classes instead of suits.** Warrior (double damage), Mage (combine cards from the tavern), Cleric (heal from the discard pile), Bard (draw cards), Paladin (reduce enemy strength). Each enemy has a class and is immune to that class's ability, but the card's strength still counts.
  - **Castle order (top to bottom):** Justicars, Kalannar (Mage, strength 10, health 30), Queens, Malice (Mage, 15/40), Kings, Tathzaer (Mage, 20/50, always last). Each group is shuffled; the Mages are fixed. A Mage enemy is also immune to the classes of every card in its enemy zone.
  - **Enemy zone (the big new mechanic).** After a player suffers damage, the top card of the discard pile goes into the enemy zone. The enemy gains strength equal to the total strength of the cards there. On defeat the zone is cleared into the discard pile. Discarding is a real cost, not just a payment.
  - **Kinfolk attacks.** Any number of same-strength cards whose total is 12 or less (not 10 as in Regicide combos). Other players may add cards to the attack after it is formed, without talking.
  - **Animal companions.** Strength 1, may pair with exactly one other card, never in a kinfolk attack.
  - **Multi-class attacks.** All class abilities in the attack resolve, in the rulebook's order (Mage, Cleric, Bard, then Warrior, then Paladin).
  - **Jester.** Played alone; any player can claim it and take over the turn. The attack formed with it ignores immunity and its enemy does not lose its immunity afterwards. A Paladin effect that ignored immunity keeps working on later turns.
  - **Defeated enemies in hand.** You can draw a defeated enemy; it counts as a card with the enemy's strength (Justicar 10, Queen 15, King 20) and has no health.
  - **Feign death:** you may always discard your whole hand to survive an enemy strike, even if it would not cover the damage. Hand sizes are 8/7/6/5 for 1/2/3/4 players, as in Regicide.
  - **Yielding:** same as Regicide, with the rule that you cannot yield if every other player yielded in a row. If you can neither play nor yield, everyone loses.
  - **Solo:** two Jesters and five side cards (strengths 2 to 6, one per class: gnome, halfling, dwarf, avion, human) are set aside. Hand size 8. After forming an attack you may add one side card if it makes a valid kinfolk attack. Twice per game a Jester can be played on its own: attack strength 8, with one chosen class ability (Cleric, Bard, Warrior or Paladin), ignoring immunity. A Jester always returns to the side, and is removed for good when the enemy is defeated.
- **Scale:** 1–4 players. The game is one run of 15 enemies, no campaign.
- **Credits:** Badgers From Mars, Sketchgoblin. Check the designer credit and licensing before building (see *Licensing*).
- **Toggle in Regicide or separate app?** Two sensible options:
  1. **A ruleset choice inside Regicide** (like The Game's Pro and Extreme). Pros: reuses lobby, join, sessions, WS hub, persistence, the hand/discard/tavern views and the whole frontend shell with one deploy and one tile. The domain already has a `Ruleset` type, which would grow from a player-count table into a variant-plus-player-count table.
  2. **A sibling app (`regicide-inferno/`)** that starts from a copy of the Regicide plumbing. Pros: no risk to the working Regicide domain, own look and own `rules.md`; cons: more duplication, and a second tile and image.
  - **Lean:** option 1 is cheaper at the plumbing level but the domain differs a lot (classes instead of suits, enemy zone, kinfolk attacks, different deck and castle order, a larger immunity model). If it goes in as a toggle, plan on splitting the rule engine behind an interface (variant picks the deck, the attack validator and the damage pipeline) and keep one set of lobby and view code. If that interface looks leaky once drawn up, fall back to a sibling app.
- **Build notes:**
  - Per-player views and server-authoritative moves are unchanged. The hidden information is the same (hands, tavern order) plus the Mage ability, which lets the active player look at tavern cards secretly: send that only to the acting player, and decide how to persist the look.
  - Kinfolk collaboration (other players adding cards after the attack is formed) needs a short open "attack in progress" phase on the wire, which Regicide does not have.
  - Cleric puts discard cards face down under the tavern deck: the hidden tavern order has to be modelled, not just a count.
- **Open questions (check on the physical cards or ask before building):**
  1. Which numbered cards belong to which class: the rulebook has no table, only the card art does. The Tavern is 2 to 10 plus 5 Animal Companions plus 1 Jester, which suggests 9 values for each of 5 classes.
  2. The enemy table on page 1 is small and icon-labelled: verify 10/20, 15/30, 20/40 (Justicar, Queen, King) and 10/30, 15/40, 20/50 (Mages) against the source before writing them into `rules.md`.
  3. The Jester count for 2 to 4 players: setup shows one Jester in the Tavern and no per-player-count table.
  4. The PDF was read from low-resolution page images, so recheck the Mage ability and solo Jester wording on the next read.
- **Risk:** Same publisher as Regicide, so the licensing thought is the same. See *Licensing*.

## The Crew: The Quest for Planet Nine

**Implemented as `the-crew/`.**

- **What:** Cooperative trick-taking. 40 cards: four colours 1–9 plus rockets 1–4 as permanent trump. Each mission deals task cards ("you must win this card in a trick"), and the team wins only if every task is fulfilled.
- **Hook:** Players may not talk about their hands. They have one limited signal (a card played face up in front of them, saying it's their highest, lowest or only card of that colour), usable once per mission.
- **Scale:** 50 missions, escalating (task order, "must win last trick", communication changes). Best at 4 players, 2-player variant exists with a virtual third player.
- **Credits:** Thomas Sing, Kosmos (2019). BGG weight about 2.0/5, rating about 7.8.
- **Build notes:**
  - Easy fit: hidden hands, per-player views and server-authoritative moves already match the existing design.
  - Mission data (50 entries, each with task rules and communication rules) lives in a data table, not in code.
  - Campaign progress (current mission, attempts) must persist in the game aggregate, since sessions are paused between missions.
  - The communication token is a UI-and-rule feature; cards are never revealed beyond that.
- **Risk:** Rights holder is Kosmos. A public hosting of a licensed game needs a licensing thought (see *Licensing* below).

## The Crew: Mission Deep Sea

- **What:** Standalone sequel using the same core trick-taking rules.
- **Differences:** Task deck grows from 36 to 96 cards, and each card carries a difficulty value for 3, 4 or 5 players. A mission sets a total difficulty; tasks are drawn and revealed until it is reached. Some tasks are general rules ("win exactly X tricks") instead of a specific card. Players can pass on tasks if enough remain for the others. Some missions change the communication rules or put all goals on one player.
- **Scale:** About 32 missions (from memory, verify). Plays well at 3, 4 and 5.
- **Credits:** Thomas Sing, Kosmos (2021).
- **Build notes:** Build as a second "ruleset" over the Planet Nine engine in the same app or a sibling app. The task-draw-until-difficulty step and the new task types are the real new work. Decide early whether both Crew games live in one app (`the-crew/` with a mode choice) or two. The plan for building it in the same app is §11 ("Deep Sea outlook") of `the-crew/docs/implementation-plan.md`.

## Sail

- **What:** Two-player cooperative trick-taking on the sea. Players can't talk. The goal is to sail the ship to the final token before the Kraken reaches the Death tile or the Kraken deck runs out.
- **Mechanics (BGG tags):** Cooperative, variable player powers, communication limits, track movement, trick-taking.
- **Credits:** Akiyama Koryo and Kozu Yusei, Allplay (2023). Rated about 7.3. Suggested age 11+. There is a Seafarers expansion.
- **Build notes:**
  - 2 players only, like Sky Team, so lobby and start flow are the simple case.
  - Needs an on-screen track (ship vs. Kraken) alongside the trick area.
  - Component details (deck contents, power cards, scenarios) still to be researched.

## Sky Team

- **What:** Two-player cooperative landing game. Pilot and co-pilot each roll dice and secretly place them on cockpit controls (axis, speed, flaps, landing gear, radio, coffee) to land the plane without crashing.
- **Hook:** Open talking about dice is forbidden. Only the radio and coffee actions allow limited communication.
- **Scale:** 11 airports, 21 scenarios, a campaign with new modules (box contents: 8 dice, control panel, altitude and approach tracks, switches).
- **Credits:** Luc Rémond, Le Scorpion Masqué (2023), Hachette Boardgames in the US. Won the Jokers Prix (duo) and Gold'n Gob (2-player) in 2024. Age 12+/14+ depending on listing.
- **Build notes:**
  - First non-card game: dice rolling must be server-side (authoritative), and the board is a custom SVG/CSS control panel.
  - Both players place dice at the same time; the server must reveal or lock placements per round while each player's roll is private until placed.
  - Probably the most "visual design" work of the list, but the rules are small.

## LOTR Trick-Taking Games (Fellowship of the Ring and The Two Towers)

- **What:** Cooperative trick-taking set in Middle-earth. Players take character roles per chapter. Each character has a goal; the team wins a chapter only if every goal is met. Failing any required objective ends the quest.
- **Fellowship of the Ring:** Chapters 1–18 of the story, 1–4 players, about 15–30 min. Cooperative card adventure with a narrative and special quests.
- **The Two Towers:** Standalone sequel covering chapters 19–36 (18 scenarios per Tolkien Gateway, 17 per another listing). 1–4 players, age 10+, about 20 min per hand, 144 cards. Release about Jan 2026. Chapters can be played in any order. Reviewers find the rules tighter than in the first game.
- **Special cards (Two Towers):**
  - The One Ring is the only trump card and cannot be led until someone plays one off-suit.
  - Tower cards (White and Black) can only be played when you can't follow suit. A Tower normally wins its trick, but two Towers in one trick cancel out. The player receives the matching token.
  - Orc cards can only be played when you can't follow suit, never win a trick, and leading one loses the chapter.
  - Some characters set up special rules (exchange cards, tuck cards to create a void).
- **Credits:** Office Dog, designer Bryan Bornmueller (Two Towers). Publisher listings via Asmodee.
- **Build notes:**
  - Reuses the trick-taking core of The Crew, but characters, items and per-chapter setup rules need their own effect model.
  - The solo and 2-player modes need virtual players or extra hands.
  - Campaign mode ("The Road Goes Ever On") adds progress across chapters.
- **Risk:** Tolkien licence. See *Licensing*.

## Aeon's End

- **What:** Cooperative deckbuilder. Players are breach mages defending the city of Gravehold against a nemesis. Win by defeating the nemesis; lose if Gravehold falls or all mages are exhausted.
- **Hook:** Decks are never shuffled. The discard order is preserved and becomes the next draw, so you plan your discards. Turn order is random per round (a turn-order deck mixing player and nemesis cards).
- **Core pieces:**
  - Each mage has breaches (open or closed) holding prepared spells, and a unique ability charged with Aether.
  - Supply market of gems, relics and spells (chosen per game, so every game differs).
  - Nemesis deck of attacks, minions and powers, in three tiers of increasing difficulty.
  - Phases: cast, main (buy, focus/open breaches, use ability), draw.
- **Scale:** 1–4 players, about 60 minutes, best at 2. Second Edition is BGG rank around #109, rating around 7.9. Many standalone boxes and expansions (Legacy, The New Age, The Ancients, Return to Gravehold, Into the Wild, The Void) with a large pool of mages and nemeses.
- **Credits:** Kevin Riley, Indie Boards & Cards (from memory, verify). Digital version exists.
- **Build notes:**
  - The rule engine is the work: card effects need a data-driven model (triggers, targets, costs) to avoid hundreds of special cases.
  - Random turn order and persistent discard order must survive snapshotting.
  - Start with the base game, a handful of mages, and 1–2 nemeses.

## Marvel Champions: The Card Game

- **What:** Cooperative Living Card Game (Fantasy Flight Games). Each player controls a hero and an aspect deck. The team faces a villain and a scheme. Heroes lose if the scheme advances to completion, so pressure comes from both villain attacks and scheme threat.
- **Hook:** Players switch between hero form (attack, defend) and alter-ego form (recover, build resources). Conflicts resolve with cards, not dice.
- **Scale:** 1–4 players, 45–90 min, age 14+. BGG weight about 2.9/5. Continuous releases: hero packs, scenario packs, campaign boxes. A huge card pool.
- **Build notes:**
  - The largest scope by far: every hero card and encounter card has its own text, timing windows, keywords and interrupts.
  - A realistic start: core set only (a few heroes, a few villains), data-driven effects, accept that the rest is a long tail.
  - Deckbuilding (aspect decks) is a feature of its own; first versions would use pre-built decks.
- **Risk:** Marvel and FFG licence; highest risk for a public site. See *Licensing*.

## Licensing and rights (all of the above)

The app would reimplement rules and use its own artwork and text, not scans of the publishers' cards. Still, names and thematic IP (Marvel, Tolkien) are protected, and some publishers actively police digital clones. Decide per game before public deployment: ideal is a private/friends deployment, generic art and renamed content, or asking the publisher. A note to settle this belongs in each game's `docs/rules.md`.

## Sources

- The Crew / Deep Sea: [Dale Yu's review of Deep Sea](https://opinionatedgamers.com/2021/09/28/dale-yu-review-of-the-crew-mission-deep-sea/), [Board Game Oracle](https://boardgameoracle.com/en-GB/boardgame/price/AfIwbkw6qf/the-crew-the-quest-for-planet-nine)
- Regicide Inferno: the publisher's A4 rulebook PDF (local copy in Downloads)
- Aeon's End: [Aeon's End Wiki rules](https://aeonsend.wiki.gg/wiki/Rules), [Board Game Oracle](https://www.boardgameoracle.com/en-NZ/boardgame/price/VGJVRnItEw/aeons-end-past-and-future)
- Marvel Champions: [library listing](https://umu.libguides.com/boardgames/marvel_champions), [Wayland Games](https://waylandgames.co.uk/marvel-champions-the-card-game-core-set-ffgmc01)
- Sail: [Tlama games](https://www.tlamagames.com/en/board-games/sail/), [Board Match](https://boardmatch.app/en/game/377470)
- Sky Team: [The Fandomentals review](https://www.thefandomentals.com/sky-team-game-review/), [Hachette Boardgames](https://www.hachetteboardgames.com/blogs/infos/sky-team-touching-down-at-gen-con)
- LOTR: [Dale Yu's Two Towers review](https://opinionatedgamers.com/2026/02/05/dale-yu-review-of-the-lord-of-the-rings-the-two-towers-trick-taking-game/), [Tolkien Gateway](https://tolkiengateway.net/wiki/The_Lord_of_the_Rings:_The_Two_Towers_%E2%80%93_Trick-Taking_Game), [Asmodee Fellowship listing](https://shop.asmodee.ca/en/family-games/odlo01en)
