# Survival A/B screen, bot.zoneModel 0 vs 1 (3 seeds, Paladin and Sorceress), 2026-11-16

> status: current (TASK-554). Harness: [ab-harness](../survival/ab-harness.md). Tags: V verified (the run logs and `abreport` output of this run), I inferred. Lockstep d2d only (no game.exe); the heroes are CHEATED for provenance (SETUP: clvl 12, 500 gold, free Rogue), so this is a relative A/B, not evidence about game behaviour.

## Setup (V)
- Binary: `d2bot` built from main at the S-D merge (`bot.zoneModel` present, checked by `abrun.sh`), one binary for both arms; arms differ only in `bot.zoneModel` 0 (legacy) vs 1 (margin), `bot.portalGuard` 0 in both.
- `AB_CLASS=2 AB_MAX=4 tools/survival/ab.sh 2 101 103 107` (Paladin) then the same with `AB_CLASS=4` (Sorceress): 3 seeds x 2 arms x 2 virtual hours, four lockstep slots, niced. Wall: about 40 minutes per class, about 2.2 slot-hours per class (the memo said about 2).
- Capacity: idle plus nice 50 to 66 percent throughout (load 5 to 19 on 16 cores); no pause was needed. The user's sandbox container and its volume were not touched.
- Every log ends `writes after start: 0` (12 of 12): no CHEATED run, none excluded.
- Report: `abreport -death-slack 1 -chicken-slack 1`.

## Result, Paladin (V)
| seed | arm | deaths | chickens | merc deaths | XP gained at 30 / 60 / 90 / end (min 119) | gold change |
|---|---|---|---|---|---|---|
| 101 | legacy | 0 | 0 | 1 | 6462 / 11801 / 16891 / 20678 | +101 |
| 101 | margin | 1 | 2 | 1 | 6462 / 11801 / 16891 / 18021 | -45 |
| 103 | legacy | 0 | 0 | 1 | 3460 / 19026 / 23347 / 30060 | -39 |
| 103 | margin | 0 | 1 | 5 | 10347 / 24360 / 33912 / 41023 | +2335 |
| 107 | legacy | 0 | 1 | 5 | 9568 / 18288 / 24245 / 32853 | +272 |
| 107 | margin | 0 | 1 | 5 | identical to legacy (same trajectory the whole run) | +272 |

Pooled: deaths 0 vs 1 (slack 1: ok); chickens legacy 1, margin 4 (limit 1 + 1 slack = 2: FAIL). `abreport` verdict: **FAIL (chickens 4 > 1 plus slack 1)**.
- Seed 107: the margin controller made the same picks as the planner for the whole run; the two runs are bit-identical (no `zone choice:` line differs). Seed 101 diverged only at minute 106; the extra death and both extra chickens happen in minutes 106 to 120 (the hero chickened twice in a row, then died in area 1 on the rejoin; the zone choice lines at that point are Stony Field margin 3.9 and Cave Level 1 margin 11.9). Seed 103 diverged at minute 13 and the margin arm earned 36 percent more XP (41023 vs 30060) with one chicken, but 5 merc deaths against 1 and more potions drunk.
- Intake per zone (lifeLost per minute, lower bound): the margin arm took more in area 4 (23.2 vs 16.7), area 5 (44.5 vs 37.5) and area 1 (41.6 vs 36.0); less in area 17 (5.0 vs 30.2). Different zones are visited, so the per-zone figures are not paired.

## Result, Sorceress (V)
| seed | arm | deaths | chickens | merc deaths | XP gained at 30 / 60 / 90 / end |
|---|---|---|---|---|---|
| 101 | legacy | 0 | 1 | 1 | 7487 / 8981 / 14158 / 18008 |
| 101 | margin | 0 | 0 | 1 | 7487 / 8831 / 10594 / 12181 |
| 103 | legacy | 0 | 0 | 2 | 5436 / 7099 / 10799 / 13181 |
| 103 | margin | 0 | 3 | 1 | 5436 / 7166 / 8673 / 10033 |
| 107 | legacy | 0 | 4 | 1 | 2674 / 4108 / 6003 / 6400 |
| 107 | margin | 0 | 1 | 1 | 2674 / 3711 / 4820 / 6316 |

Pooled: deaths 0 vs 0; chickens legacy 5, margin 4 (ok). XP at the end of the run: margin 28530 vs legacy 37589, ratio 0.76 (per seed 0.68, 0.76, 0.99); the margin arm is at or below legacy on all three seeds. `abreport`: **INCONCLUSIVE (interval spans the floor)**. The arms diverge at minutes 61, 58 and 39 and the margin arm spends more time in areas 17 to 19 (about 47 minutes vs 4).

## Verdict
**The screen FAILS the agreed rule** (Paladin: 4 chickens vs 1, limit 2 with the slack; Sorceress: inconclusive, XP leaning 24 percent lower). It is thin evidence and says so: the Paladin failure comes from 3 extra chickens in 2 seeds, 2 of them in the last 14 minutes of seed 101; one seed (107) shows no difference at all; the one seed with a large divergence (103) is better for XP and worse for merc deaths. The default stays 0.

## Recommendation for the 8 seeds
Do not run the 8-seed confirmation yet: it is a pass-to-confirm step and the screen failed. First read seed 101 margin minutes 100 to 120 (why two chickens and a death in town after the rejoin; `ab/margin/101.log` on the worker host) and the Sorceress XP loss (zone choice lines of seeds 101 and 103 from minute 58). If a cause is fixed, then 8 seeds per class cost 16 runs x 2 arms, about 6 slot-hours per class (about 1.7 hours wall on four slots at today's load, scaled from 2.2 slot-hours per 3 seeds), 12 for both classes.

## Harness defect found (V)
`abreport` reads the XP at minute 120 from `XPAt[120]`, but a 2 h run has 119 to 120 minute lines, so `Minutes` is at most 119 and the 120 checkpoint is absent; `xpAtHorizon` then silently falls back to the 90-minute value (the `xp120=` column equals `xp90` above, and the printed XP ratio and interval are for minute 90). The XP columns in this doc come from the last minute line instead. Fix for the harness owner: record the last minute as the horizon value when `Minutes >= 0.95 * horizon`, and say the minute used. The chicken counts used here are `abreport`'s (my own grep missed them).

## Provenance
Heroes `SaHero-101`, `SaHero-103`, `SaHero-107` (Paladin and Sorceress, 12 runs): logged in [test-characters](test-characters.md). Logs: the worker host survab run folder (`ab-pal` Paladin, `ab` Sorceress).
