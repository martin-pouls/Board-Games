# Games

## All Card Games
- Some highlight for drawn cards would be nice. so when i draw i want to see which cards i just drew. maybe highlight them in hand for a few seconds or until i select something or something along those lines. think of a good idea, its also an option to give me multiple ui designs and let me choose.

## The Game

## Regicide

- **Hand sort option** (per player, display only). Today the backend sorts by rank (Jester, A, 2–10, J, Q, K), then suit (♥ ♦ ♣ ♠) (`SortCards` in `domain/card.go`). Options:
  - rank, then suit, ace low (current default)
  - rank, then suit, ace high (Jester stays first; only the sort position of the Ace changes, not its value)
  - suit, then rank
  - **Frontend only, no backend change:** commands send cards, not indices. Re-sort the hand once, in a single `$derived` in the table, and pass that array to everything (selection indices refer to it). Clear the selection when the order changes. New-card marks are keyed by card, so re-sorting is safe for them.
  - **Placement:** a small icon button (lucide `ArrowDownUp`) at the hand that cycles through the three orders; the hand updates right away. Not in the lobby (lobby options apply to the whole table), and no burger menu needed. Check where it fits on an iPhone 13: a corner above the cards, or the action bar.
  - **Storage:** `localStorage`, key e.g. `regicide-hand-order` (not a cookie: the server doesn't need it). Per browser/device, so each player keeps their own. Wrap reads in try/catch, and fall back to the default.
  - Add half a sentence on this setting to the privacy policy's local storage section.
  - Comparators next to `compareCards` in `frontend/src/lib/game/rules.ts`, with tests.
  - Other games: no option. Fox Duet and The Crew sort by suit, then rank (following suit makes that the natural order), and The Game by value.
- we went a little over board when saving space yesterday, so we can make some stuff a little larger now. one specific thing is that i would put the attack, block, etc in a separate line from the health bar again. see also the screenshot where we can see a bug where it wraps depending on whether a card is selected. so just make that two rows. probably make this enemy area than slightly larger again. but dont overdo it, we need to support smaller screens like the iphone 13 and there should be some wiggleroom. dont make stuf larger just for the sake of it. also we found today that we have all of the info we need, so that was fine. if you have smart suggestions for the space, feel free to add them. give me ui designs before implementing anything bigger though.



## Fox in the Forest: Duet

## The Crew
- when you select a task you should have to confirm it similar like when you play a card. 
- when you have a mission where you can swap tokens, the final ordering of the tasks should also swap. so when playing the tasks should be ordered appropriately, e.g. 1 - 5 or >, >>, >>>.
- the distress signal stuff is annoying. give me an option when starting a lobby to en / disable the distress signal. this should be default to disabled, because we never use it. what i mean is this question at the start o the game whether i want to use a distress beacon or how its named. you know what i mean.
- the biggest stuff is the layout. currently we sometimes need to scroll which is annoying so we need so save some height. the tasks section can probably be made more compact. we optimize for two player mode with jarvis here, because thats what we regularly play. obviously other modes should still be functional. give me layout suggestions here.

# General
- maybe drop test from ci, so we have faster pipelines?
