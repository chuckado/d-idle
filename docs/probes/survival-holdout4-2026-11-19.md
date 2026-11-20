# Survival A/B holdout 4: Sorceress screen, Paladin seeds 8 (2026-11-19/20)

> status: current (card survhold4). Tags: V verified (builder logs, `abreport -death-slack 1 -chicken-slack 1 -xp-floor 0.95 -horizon 90`), I inferred, U unverified. `bot.zoneModel` stays 0; no default changed; recommendation only. Follows [holdout 3](survival-holdout3-2026-11-18.md). Heroes are CHEATED for provenance as before (SETUP clvl 12, 500 gold; Sorceress no weapon): a relative A/B only. Host: the builder, long lane, for every run listed here (both arms of a pair on the same host); no second worker was used. Writes after start: 0 in all 16 logs (V, last line of each log). Binary: origin/main cb2c98156 plus the harness change below.

## 1. Harness defect found (V) and fixed
The first Sorceress wave of this card (seeds 107, 181) was INVALID and is not counted: the heroes never spent their points. Since agencyp0 (4af9dc62a) the bot banks stat and skill points unless `resource.pointsDecisions` is 1 (docs/points.md); the A/B hero is made at clvl 12 with unspent points, so it had 1 skill, life 40, "no bound attack skill" and about 5k XP (seed 181 legacy: 123 chickens). `tools/survival/abrun.sh` now writes `resource.pointsDecisions` 1 into every run's tuning.json (both arms). Holdout 3 and earlier runs predate the gate and are unaffected. All numbers below are from runs with the fix.

## 2. Sorceress (class 4, speed 3, horizon 90, 4 pairs)
| seed | legacy xp / chick | margin xp / chick | deaths L/M |
|---|---|---|---|
| 107 | 24868 / 0 | 26317 / 0 | 0 / 0 |
| 181 | 21970 / 1 | 19940 / 1 | 0 / 0 |
| 211 | 25239 / 1 | 1983 / 0 | 0 / 0 |
| 213 | 16097 / 2 | 10023 / 0 | 0 / 0 |

Pooled: deaths 0 / 0, chickens 4 / 1, XP ratio 0.669 (floor 0.95), difference mean -7213 (sd 10999), 95% interval [-24713, 10286]. **Verdict: INCONCLUSIVE by abreport (interval spans the floor) but a practical FAIL on yield: two of four seeds are far under legacy, so wave 4 (5 more seeds) was not run.**

Analysis of the two low seeds (I; minute lines of the logs):
- Seed 211 margin: xp 90183 at start, 91907 after 75 minutes (legacy +14000): it sat in Cold Plains (area 3) and Rogue Encampment/Burial Grounds (areas 3, 1, 3, 3, 17) with little yield. Intake in area 3 was low (12.1 life/min over 90 minutes in the pooled table), so the model judged it safe but the realized yield stayed near zero for the whole run; the starved-zone fix (`yieldMinMinutes` 1.5, `starveValueMul` 0.25) did not move it out. Finding: **a safe zone with near-zero realized xp is held, the yield-starve route not firing for this hero (no weapon, 1 Charged Bolt-class skill line)**. Not a clean data defect (the cause is unproven, U; the zone-pick lines of the log were not found by the grep I ran), so not fixed.
- Seed 213 margin: `potion_crisis` re-ask at k 3.00 (the clamp) early, then Burial Grounds (17) and low areas; ended at 10.0k against 16.1k, 0 chickens against legacy 2: a safety-for-xp trade, the model working as designed but expensive.
- Seeds 107 and 181 (margin 26.3k / 19.9k against 24.9k / 22.0k) are within noise.

## 3. Paladin (class 2, 2 fresh pairs, the post-gate harness)
| seed | legacy xp / chick | margin xp / chick | deaths L/M |
|---|---|---|---|
| 221 | 25946 / 0 | 34864 / 0 | 0 / 0 |
| 223 | 26940 / 0 | 23063 / 0 | 0 / 0 |

Pooled 2 pairs: ratio 1.085, no deaths or chickens. **Verdict: no regression; with the 6 holdout-2/3 seeds (PASS, pre-gate harness) that is 8 Paladin seeds, PASS (I: the 6 earlier seeds were run without the points gate, so the heroes were as here; the fix only restores what those runs did).**

## 4. Recommendation (the user decides; nothing flipped)
1. **Paladin: PASS at 8 seeds, ready for the flip decision** (6 seeds pre-gate harness plus 2 post-gate).
2. **Sorceress: do not flip.** INCONCLUSIVE by the stats rule, with a repeatable low-yield mode (211, 213) and ratio 0.669. Next step: a task that reads the zone-pick lines of seed 211 margin and finds why a zero-yield safe zone is kept (suspected yield judge for a hero with no damage skill), then a new 3-seed screen. A per-class flip (Paladin only) needs the user's call.
3. Harness note for other A/B work: any run that uses `abrun.sh` needs `resource.pointsDecisions` 1 now; fixed here.

## 5. Addendum 2026-11-20 (card survhold6): Sorceress rerun with a wand
> status: the weaponless-hero explanation in section 2 was incomplete. The harness now gives the Sorceress hero a wand (`wnd`, SETUP, both arms; `tools/survival/abrun.sh`); the Paladin heroes always had `hax` (V: test-characters log, holdout 2), so they were never weaponless. The hero line now reads `dmg 2-4` instead of `1-5` (V: zone-entry lines of the new runs). Every run here: writes after start 0 (V, last log line). Both arms of a pair on the same host; the pair 241 and 243 was run on the second worker (no other pair crossed hosts), cleaned up afterwards.

| seed | host | legacy xp@~89 / chick | margin xp@~88 / chick | deaths L/M |
|---|---|---|---|---|
| 211 | builder | 19629 / 0 | 3066 / 0 | 0 / 0 |
| 213 | builder | 11463 / 0 | 16872 / 0 | 0 / 0 |
| 241 | second worker | 17443 / 1 | 4419 / 0 | 0 / 0 |
| 243 | second worker | 24739 / 0 | 4766 / 0 | 0 / 0 |

abreport per host: builder ratio 0.643 (interval spans the floor), second worker ratio 0.218; each INCONCLUSIVE at 2 pairs. Pooled by hand over the 4 pairs: margin 29123 against legacy 73274 (ratio 0.397), per-seed difference mean -11038, sd 11060, 95% interval about [-28600, +6600] (t, 3 df). Chickens 1 / 0, deaths 0 / 0.

**Verdict: FAIL on yield (practical); the statistics alone read INCONCLUSIVE.** Three of four seeds sit at 16 to 27 percent of legacy; seed 213 recovered (margin above legacy). The safety numbers are not better enough to pay for it (chickens 1 against 0, deaths 0 against 0).

**Cause, from the model's own log lines (V, seed 241 margin, 39 `zone choice` lines; seed 211 margin: Burial Grounds 16, Cold Plains 8, Mausoleum 5):** the weapon was not the main cause. The model keeps choosing Burial Grounds (and Cold Plains) as `best_pay_eligible` with a model yield of about 150 xp/min, while the hero realizes 4 to 27 (log: "realized 4 (model 152)", "realized 6 (model 152)", "realized 27 (model 165)"), and k sits at the clamp 3.00 from the second decision on. `weakest=offence_weapon_class_default dps=11.5 src=skill`: the offence input is still the class default, not a measured value. The realized-yield cap (survhold3) lowers the value to about 54 to 67, but the other zones are then excluded by `margin_low` or lower value, so the same zone is picked again for 60 to 99 minutes. Legacy leaves for Stony Field, Dark Wood and the Cave (areas 4, 5, 6, 9, 10) and earns 3 to 5 times as much. This is a model defect (yield estimate for a starved caster, and no escape from a zone that realizes under 10 percent of the model), not a data-only defect I can set in `internal/survival` with a must-fail test in this card: finding filed (U on the right fix: cap the model yield of a zone by the realized yield with no floor, or count a zone as depleted after N minutes under X percent of the model).

**Recommendation (user decides; nothing flipped).** Sorceress: do not flip; the zone model needs a fix to its caster yield estimate before another screen (a new card; wave 4 and the 8-seed run are not worth running on this build). Paladin: unchanged, PASS at 8 seeds (section 3).
