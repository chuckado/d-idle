# navsoak3: final numbers and comparison (2026-10-31)

> status: current. Read-only analysis of the finished hidden d2d soak on the worker host (single log `bot-0.log`, exit 0, supervisor restarts 0). Handoff: [navsoak3/README.md](navsoak3/README.md). Tags: V = counted from the logs. n=1 per soak, not a statistical claim.

**Pre-fix caveat.** navsoak3 ran a bot built from main 35ce4c2, BEFORE the primitive fixes (fix-belt, fix-pickup, fix-portal, fix-portal2, fix-loot). It is the pre-fix baseline for the primitives; the drill suite (docs/probes/drills) and the next soak are the post-fix measures. It does include TOWNWEDGE and NOWALL, so it is the first soak that ran the whole 240 minutes.

## Final state (V)
clvl 13, XP 131798, 9406 kills, 0 deaths, 4 h wall (14340 s at the last pace line) = 45.3 virtual h (ratio 11.4), 15 in-process restarts, furthest area id 19 (minute-line area ids: 4 Stony Field 690, Rogue Encampment 663, Cold Plains 550, Blood Moor 391, then 9, 17, 18, 8, 5, 19, 10, 13, 6). Final gold 247, median 131, max 1041, under 100 gold in 39% of minutes. 1075 potions used. Quests: Den of Evil and Sisters' Burial Grounds progress; route guard kept monastery and tower blocked (no merc, clvl < 20).

## Against the earlier soaks (V, same greps on each log)
| | navsoak | navsoak2 | navsoak3 |
|---|---|---|---|
| final clvl / kills / deaths | 13 / 7216 / 1 | 9 / 1784 / 0 | 13 / 9406 / 0 |
| progress ends | wedge at wall 179 | wedge at wall 45 | none (ran to the limit) |
| `campaign stalled` / `no route` / `edge marked bad` | 0 / many / yes | 0 / many / yes | 0 / 0 / 0 |
| `no eligible reachable` | 51.9k | 173k | 7 |
| `portal did not lead` | 66 | 29 | 103 |
| `no portal after the click` | 1808 | 170 | 1627 |
| `no portal opened` | 373 | 45 | 418 |
| `belt: moved 0 potions` (of all `belt: moved`) | 11 | 2 | 14 of 207 |
| `potion already in the pack` skips | 409 | 117 | 663 |
| `no healing potion in the belt` | n/a | n/a | 1791 |
| `recovery: loot` lines with items or gold | 0 of 4238 | 0 of 13563 | 0 of 828 |
| minutes with potions 0 (field `potions` on the minute line) | 10.8% of 6965 | 0.8% of 18874 | 29.8% of 2726 (813) |
The empty-belt row is not comparable across soaks (navsoak and navsoak2 minute lines include the 60 and 167 wall minutes of wedge waiting in town, and their share was reported as 33% and 25% by navsoak-compare.md with a different count). The primitives review's 31% (664 of 2148) was an earlier snapshot of this same soak; the final value is 29.8%. Stall signals: 1586 `nav stalled by monsters, yielding`, 5 `failsafe: skipped the stalled target area`, 32 `failsafe ladder` lines, 0 wedge lines, 0 stall-ladder restarts.

## What it shows about the primitives
- The town wedge and the wall are gone (0 `no route`, 7 `no eligible`); the campaign now runs a full 240 min. Policy blockers are no longer the limit; the primitives are.
- Portal: 1627 `no portal after the click` and 418 `no portal opened` against 103 recoveries by walking: the portal verb is the biggest failure count (fix-portal, fix-portal2 target it).
- Belt and pickup: 29.8% of minutes with no potion, 663 pickup skips because a potion was "already in the pack", 14 of 207 belt moves moved 0 (fix-belt, fix-pickup target these).
- Loot recovery never recovered anything (0 of 828 non-zero), consistent with the earlier soaks (fix-loot).
- Gold is still thin (39% of minutes under 100), which feeds the empty belt.

## Measure in the next soak (post-fix build, same recipe: Sorceress, seed 5, 240 min, campaign)
Same greps as above, in particular: `no portal after the click` and `no portal opened` per virtual h (expect near 0), `belt: moved 0 potions` count, `potion already in the pack` skips, empty-belt minute share with one shared definition (potions 0 on the minute line), non-zero `recovery: loot` lines, plus clvl/XP/kills/deaths and `campaign stalled` / `no route` staying at 0.
