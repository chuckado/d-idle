# Survival A/B 8-seed confirmation, Paladin, bot.zoneModel 0 vs 1, 2026-11-16

> status: current (TASK-568). Follows [the 3-seed rescreen](survival-ab-rescreen-2026-11-16.md). Harness: [ab-harness](../survival/ab-harness.md). Tags: V verified (run logs, `abreport` output, ledger files of this run), I inferred, U unverified. Lockstep d2d only (no game.exe); heroes are CHEATED for provenance (SETUP: clvl 12, 500 gold, free Rogue), so this is a relative A/B, not evidence about game behaviour.

## Setup (V)
- Binary: `d2bot` built from origin/main at the survledger merge (realized yield, mana floor, fresh_game guard, revive-loop breaker, merc deaths and fight kills in the ledger, `k` on every `zone choice:` line); one binary for both arms; arms differ only in `bot.zoneModel` 0 vs 1, `bot.portalGuard` 0 in both.
- Seeds 101 103 107 (the rescreen seeds) and 109 113 127 131 137, class Paladin, 2 virtual hours, 16 runs: `AB_CLASS=2 AB_MAX=4 AB_SPEED=5 tools/survival/ab.sh 2 101 103 107 109 113 127 131 137` under `tools/agentrun.sh --ttl`, about 1 h 35 min wall on four slots (host load 20 to 28 on 16 cores, the user was playing another game on the host, so my processes were reniced to 19; the user's sandbox container and volume untouched).
- Every log ends `writes after start: 0` (16 of 16); none CHEATED, none excluded, no run short.
- Report: `abreport -base legacy -arm margin -death-slack 1 -chicken-slack 1`.

## Result (V)
| seed | arm | deaths | chickens | merc deaths | potions | xp@30 / 60 / 90 / last |
|---|---|---|---|---|---|---|
| 101 | legacy | 0 | 5 | 1 | 38 | 4376 / 10139 / 14561 / 17853 |
| 101 | margin | 0 | 3 | 1 | 32 | 4376 / 7341 / 9977 / 14804 |
| 103 | legacy | 0 | 3 | 3 | 59 | 5748 / 16017 / 26592 / 31707 |
| 103 | margin | 0 | 3 | 3 | 53 | 6907 / 11865 / 14313 / 22629 |
| 107 | legacy | 0 | 4 | 5 | 69 | 10836 / 24619 / 30279 / 40466 |
| 107 | margin | 0 | 1 | 5 | 57 | 10836 / 25336 / 29366 / 36386 |
| 109 | legacy | 0 | 2 | 5 | 89 | 5939 / 14402 / 28414 / 39257 |
| 109 | margin | 0 | 0 | 5 | 47 | 5342 / 10394 / 16287 / 27405 |
| 113 | legacy | 0 | 8 | 5 | 48 | 6095 / 10273 / 21412 / 27785 |
| 113 | margin | 0 | 1 | 5 | 48 | 2671 / 5093 / 11403 / 23885 |
| 127 | legacy | 1 | 6 | 1 | 46 | 1029 / 2163 / 3617 / 5086 |
| 127 | margin | 1 | 1 | 1 | 35 | 1187 / 2115 / 2234 / 4281 |
| 131 | legacy | 0 | 3 | 5 | 73 | 5271 / 19055 / 29062 / 32435 |
| 131 | margin | 0 | 0 | 5 | 45 | 5271 / 11127 / 16254 / 22527 |
| 137 | legacy | 1 | 5 | 5 | 66 | 7006 / 13034 / 18830 / 20659 |
| 137 | margin | 0 | 1 | 5 | 64 | 7006 / 13034 / 18830 / 24738 |

Pooled (sums of the table): deaths legacy 2, margin 1; chickens legacy 36, margin 10; merc deaths 30 vs 30; potions 488 vs 381; gold change +2831 vs +2866. Clvl end: both arms 12 to 13.
XP at the horizon (last minute line of each pair, 116 to 120), margin minus legacy: mean -4744, sd 5372, 95 percent interval [-9236, -252], **ratio 0.823**. Per seed margin/legacy: 0.83, 0.71, 0.90, 0.70, 0.86, 0.84, 0.69, 1.20 (seed 137 is identical for the first 60 minutes, then margin pulls ahead).

**`abreport` verdict: FAIL, exit 1** ("XP at 120 min is below the floor with the 95 percent interval"). The deaths and chickens parts of the gate pass with large margin (chickens 10 vs 36, deaths 1 vs 2); the XP part fails: the whole interval is below zero in 7 of 8 seeds' direction.

## Intake per zone (V: abreport, life lost per minute, minutes in brackets, a lower bound)
Legacy vs margin: area 1 38.8 (151) vs 29.8 (135); area 3 30.1 (166) vs 17.0 (233); area 5 55.7 (112) vs 44.6 (81); area 10 62.3 (57) vs 49.6 (54); area 2 7.2 (64) vs 2.1 (113); area 4 20.7 vs 20.9; area 9 28.7 vs 26.1. The margin arm takes less damage per minute in every zone it visits and spends time in the quieter ones (areas 2 and 3, 113 and 233 minutes against 64 and 166).

## Zone choice and k (V: `zone choice:` lines and ledger.json of the margin instances)
- The margin arm logs 4 to 33 `zone choice:` lines per seed (161 in all), the legacy arm 2 to 11. Reasons over all margin lines: `best_pay_eligible` 98, `depleted_next` 59, `stay_current` 5.
- Where the arms differ the margin arm trades pay for safety: it sits in Stony Field, Cold Plains, Blood Moor and Cave Level 1 (value 54 to 64 xp/min on the lines) where legacy goes to Dark Wood, Black Marsh and Burial Grounds (the legacy arm's own picks: Burial Grounds 15, Stony Field 9, Dark Wood 8, Black Marsh 8). That is the whole XP gap: about 18 percent less XP for about 70 percent fewer chickens.
- **k now moves** (the rescreen asked): `k` leaves 1.00 and is visibly learned in six of eight seeds (Cold Plains 3.45, 2.65, 1.98 in seed 101, i.e. rises after an event and decays; Dark Wood 2.2 to 2.5 in seed 109; Cave Level 1 2.00 to 1.94, 1.75). Most lines of the zones that were never hurt stay at 1.00. The ledger now has kills and seconds per zone (survledger works; mercDeaths per zone not audited here).
- Merc deaths are equal (30 vs 30 pooled, 5 per run in most seeds), so the margin controller does nothing for the mercenary (it only raises the hero's own safety).

## Risks and caveats
- The harness Paladin still reads `src=default` (dps=30) on all 63 margin-arm `zone shadow` lines: the `debug equip hax 4` verb replied `quantity=-1` and the weapon never reached the model (U whether it was equipped). The margin model therefore runs with a default offence and may be too cautious for a real hero with a weapon; this alone could explain part of the XP gap (I, not tested). Fix this before the next screen.
- One hero build and one class (Paladin clvl 12 late start, 2 h). Sorceress was NOT run: the brief gated it on the Paladin gate passing, and the gate failed. The first Sorceress screen was inconclusive (xp ratio 0.76).
- Seed 127 is a bad seed for both arms (both died once, clvl 12, 5000 xp); the result without it is ratio 0.82 as well (V from the table).
- Lockstep polls once per tick, so intake reads higher than in realtime; d2d-only (no game.exe); the heroes are CHEATED (declared SETUP).

## Recommendation (what the user must decide)
1. **Do not flip `bot.zoneModel` to 1 for any class yet.** The agreed gate (no more deaths or chickens, XP at least equal within noise) fails on XP: margin earns about 18 percent less (interval -9236 to -252 xp on a mean legacy value of about 27000). It is safer (chickens 10 vs 36, deaths 1 vs 2, less intake per minute), but the gate is explicit and the XP interval excludes zero.
2. The user can choose to accept a safety-for-speed trade instead: the numbers say "about 18 percent slower levelling for a sharply lower chicken rate", at the same merc-death count. That is a product decision, not an A/B pass; I would only take it for friends' first-hour characters, where deaths cost more than speed, and only after item 3.
3. Before another screen: (a) make the harness Paladin actually wield a weapon (the equip verb reply is `quantity=-1`; check slot code and read the damage stat back), because the model reads default offence; (b) retune: the margin picks Cave Level 1 and Cold Plains at 54 to 64 xp/min where Dark Wood pays 442, which is the XP gap, so the pay-versus-margin threshold (`need 1.5`) is the knob to look at; (c) then rerun 8 seeds on the same seeds (the deaths and chickens side is already won) and only then, if XP is within the floor, the Sorceress 3-seed rescreen.
4. Not recommended: flipping for Paladin builds only, or for all classes. Verdict: **not yet**.

## Provenance
Heroes `SaHero-<seed>` for the eight seeds: [test-characters](test-characters.md). Logs: worker host `d2d-lab/survab/ab` and a copy in `d2d-lab/survab-pal8` (previous results in `d2d-lab/survab/old2`).
