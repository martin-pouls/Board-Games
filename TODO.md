# Games

## All Card Games

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



## Fox in the Forest: Duet

## The Crew

# General
- maybe drop test from ci, so we have faster pipelines?
