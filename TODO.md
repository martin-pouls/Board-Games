# Games

## All Card Games

## The Game

## Regicide

## Fox in the Forest: Duet

## The Crew

## Mission Deep Sea

## Sky Team
- overall look and feel is the right direction
- Home Page more inspired by game box, give me options here
- Game start screen (host / join) missing. also needs image or something like the other start screens
- leave / close lobby consistent with other games
- same for light / dark mode icons. why is the header so inconsistent from other games in the design?
- probably remove the narration line - does it really add much? we removed such stuff on other games. maybe then the whole line can go and free up space?
- i dont really udnerstand the top bar above the die placement area? what does it show when? does it need to show the checkmarks on landing?
- axis icon is kind of ugly. same for the speed area. also the land <= x. dice and placement areas look great.
- where do you see the reroll tokens? bottom right?
- the slight misalignment, when a height field gets the reroll is annoying. maybe abbreviate as 1k, 2k, etc. or something else?
- the text "green" on the mission select is kindof redundant, also the "kerosene", etc.
- line wrap in "Speed 3, brakes allow 4" in the end screen is annoying, just drop some text.
- scenario screen instead of a bunch of text under the picked scenario "7 rounds. Kerosene: put a die on fuel each round or lose 6. Wind: the axis turns the wind ring, its value adds to the speed." maybe allow to tap the icons for the special stuff and see niceliy formatted rules there. also info about the track (7 rounds, etc.) viewable somewhere instead of text. maybe even a preview of the track?
- abilities in the abilities select need more info, e.g. working together is unclear with the simultaneous die placement. again maybe tap and then show a nice preview card. could also be reused in game to get a refresh.
- generally dont take my advice literally, if you have better ideas to solve the problems feel free.
- Progress tracking over several flights (campaign) is not in v1 (rules.md §7.2).
- Real-time module (§12.4) is not built; its four scenarios are played without it. Notes in `sky-team/docs/implementation-plan.md` §5.8.
- Plumbing copied from Fox Duet a fifth time (service `mutate/act`, ws dispatch skeleton, sqlite codec): candidates for `shared/` once the pattern settles.

# General