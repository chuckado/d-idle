# postfix-soak1: first soak on the fixed bot, Gate 1 shape (2026-10-31)

> status: current. Read-only analysis of one finished hidden d2d soak (exit 0, 120 wall min, one injected d2d kill). n=1, not a statistical claim. Pre-fix baseline: [navsoak3-summary.md](navsoak3-summary.md). Tags: V = counted from the log.

## Setup
- Bot build: origin/main f8c1b7f (fix-belt, fix-pickup, fix-loot, fix-portal, fix-portal2, enterportal, fix-wp merged). Not in the build: fix-bank, fix-obj, fix-price, waits-audit, fix-missile and CastGapMs 650 (merged later).
- Lab binary: the fork-port build (upstream #148 base plus the click-modifier patch). navsoak3 ran the older lab6 build, so the d2d side differs; treat the comparison as indicative.
- Recipe of navsoak3 with a fresh uncheated Sorceress (character Post1, seed 5, class 4), `--backend d2d --managed --goals campaign --d2d-args --lockstep`, no debug verbs, nice 10, hidden slot 0, no stream. Supervisor limit 7200 s. Capacity check on the worker: WARN (idle+nice 65%, iowait 23%, PSI io 40), not NO.
- Length: 119 wall min, 75111 s = 20.9 virtual h (ratio 10.5). navsoak3: 45.3 virtual h. Rates below are per virtual hour (/vh).

## Comparison (V)
| | navsoak3 raw (/vh) | postfix1 raw (/vh) |
|---|---|---|
| clvl / XP / kills / deaths | 13 / 131798 / 9406 / 0 | 13 / 119024 / 6331 / 0 |
| XP per vh, kills per vh | 2909, 208 | 5705, 304 |
| `no portal after the click` | 1627 (35.9) | 0 |
| `no portal opened` | 418 (9.2) | 11 (0.53) |
| `portal did not lead` | 103 (2.3) | 8 (0.38) |
| `belt: moved 0 potions` of `belt: moved` | 14 of 207 (6.8%) | 47 of 190 (24.7%, 2.3/vh) |
| `potion already in the pack` skips | 663 (14.6) | 650 (31.2) |
| `recovery: loot` non-zero | 0 of 828 | 156 of 249 |
| minutes with belt potions 0 | 29.8% of 2726 | 2.2% of 1253 (27) |
| `no healing potion in the belt` | 1791 (39.5) | 92 (4.4) |
| `nav waypoint failed` (0 successes) | 295 (6.5) | 0 |
| `nav stalled by monsters` / `failsafe ladder` | 1586 / 32 | 635 (30.4) / 17 (0.8) |
| `no eligible reachable` / `campaign stalled` / `no route` | 7 / 0 / 0 | 48 / 0 / 0 |
New diagnostics: `belt: put` 349; `loot: potion skip` 412; `d2d town portal: no portal source (empty)` 143, `attempt failed` 13, `no portal opened, falling back` 11, `click without a portal` 3. Town arrivals through an open portal 88; `no town portal: taking the waypoint to town` 33 (16 ended `level changed during the walk 6 -> 20`, 2 `none on level N`); `no town portal and no waypoint: walking back` 40; `wp: activated` 6. packHP above 0 in 288 of 1253 minutes. Chicken lines 64 (d2dsoak counts 55 chickens, 19 fresh games). Furthest areas: Burial Grounds, Mausoleum, Crypt, Forgotten Tower, Black Marsh; gold 345 at the end, 634 potions used; final `restarts` counter 19 (see below).

## Fault injection (V, ~60 wall min as specified)
At 14:57:56 (59 min) the slot's d2d was killed with `kill -9` (pid found through the slot's registry process group; nothing else touched). The bot logged `d2d process exited ... signal: killed` after 1 s, `game lost, restarting it` and was back in game (Rogue Encampment, then Blood Moor) 2 s after the kill; kills and XP resumed within 5 s. Minute-line `restarts` 8 to 9 (+1). Saved XP: the previous save point was 72638; after the restart it was 72638 (not lower); the unsaved 1263 XP since that save was lost (`xp lost`, 46 s of play). No leftover of the killed pid; the registry held only the run's two entries, and after the end `--list` showed none of soakpost and no d2d of the slot ran. Note: the counter also counts routine d2d recycles (8 before the kill with no game-lost event), so "exactly +1" is only readable as a delta around the kill. A herdr box crash at 15:40 cut the analyst session; the soak on the worker was unaffected.

## Gate 1 verdict (soak-gates spec)
| criterion | result |
|---|---|
| Stability: no exit but the injected one, 0 panics, alive at 2 h | PASS (exit 0, 0 panics, 0 bot restarts). `saveXp == charXp` after the graceful stop: UNVERIFIED (the end is a d2d kill by the bot, no save-and-exit line) |
| Deaths: death rule | PASS (0 deaths; `INVARIANTS: PASS`, hang 0, stall 0, wedge max 1 min, no-progress max 5 min). Chickens 2.6 to 3.1/vh but 27/wall h (criterion text is 4/h; lockstep unit unclear) |
| True XP/h at or above the d2d baseline | INDICATIVE FAIL: 5719 vs 9172 (-37.6%), verdict INCONCLUSIVE (1 seed); clvl 13 baseline comparison is not like for like |
| Belt: no window over 5 min with 0 healing outside town | FAIL on virtual minutes (longest run 7 minute lines, about 0.7 wall min); PASS if read in wall minutes |
| Portal: at least one trip by town portal | PASS (88 arrivals through a portal; 33 waypoint-to-town fallbacks) |
| Route guard | PASS (0 `blocked: not ready`, no boss run) |
| Injected fault recovery | PASS (2 s, +1, saved XP equal, no orphan) |
`d2dsoak compare` OVERALL INCONCLUSIVE (exit 2, one seed), invariants PASS including the primitives checks it covers.

## What improved, what did not
- Improved (V): the portal verb (3 failure families from 2100 to 19), empty-belt minutes 29.8% to 2.2%, `recovery: loot` now recovers (156 of 249 non-zero), waypoints activate (6) and the waypoint-to-town verb is attempted; XP/vh doubled.
- Not improved: `potion already in the pack` is a pickit filter message and did not drop (31/vh); belt fill failures got worse in share.
- NEW family: `belt: move failed ... town: no belt column for Minor Mana Potion (kind mp)` marks the belt "full" and the fill moves 0 (12 of the 47 zero-move fills show hp 6->6 with 1 rv in the pack); the cause (columns held by other kinds) is UNVERIFIED. Needs a drill (D1 with a Rejuvenation or mixed belt).
- NEW family: waypoint-to-town fails with `the level changed during the walk 6 -> 20` (16), i.e. the bot is walked into a cave mouth (Black Marsh to Tower area) instead of the waypoint; `no activated waypoint` hides Dark Wood and Black Marsh from the leveling picker (60 `leveling target not reachable` lines).
- Poverty: gold under the kit cost with no town portal (`poverty mode`, 143 empty-source portal attempts) is the remaining belt driver.

## Measure next
Rerun on the current main (fix-bank, fix-obj, fix-price, fix-missile in) with 3 seeds for a real XP/h verdict; log wall and virtual minutes in the belt-window check; add a save-and-exit at the end for the `saveXp == charXp` line; drill the mixed-kind belt and the cave-mouth waypoint walk.
