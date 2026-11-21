# Survival A/B holdout 3: starved caster yield, risk route at the k limit (2026-11-18, PARTIAL)

> status: partial (card survhold3). Tags: V verified (builder logs, `abreport -death-slack 1 -chicken-slack 1 -xp-floor 0.95 -horizon 90`), I inferred, U unverified. `bot.zoneModel` stays 0; no default changed; recommendation only. Follows [holdout 2](survival-holdout2-2026-11-18.md). Heroes CHEATED for provenance as in holdout 2 (SETUP clvl 12, 500 gold; Sorceress no weapon): a relative A/B, not evidence about game behaviour. Writes after start: 0 in all 4 logs of this record (V, last line of each log).

## 1. Fix (data, `internal/survival`; V by unit tests, must-fail checked against the old params.json)
1. **Risk route closed at the k limit.** New dial field `riskMaxK` (balanced 2.5, bold 2.9, cautious has no route; 0 = no condition). When the zone's `k` (zone x hero) reaches it, the pay-for-risk route is refused (why `margin_low`, "model wrong about this zone"). Cause (holdout 2 seed 103): Dark Wood picked for risk with k 3.00 = the clamp, then a chicken. Test `TestRiskRouteClosedAtKLimit` (Paladin and Sorceress); fails with the old params.json (V).
2. **Caster trusts the realized yield sooner.** New archetype fields `yieldMinMinutes` (caster 1.5, global 3) and `starveValueMul` (caster 0.25, global 0.5); a starved zone's xp is also capped at what it realized (`internal/survival/controller.go judge`). Test `TestCasterTrustsRealizedYieldSooner`; fails with the old params.json (V). `TestStarvedZoneLosesValue` now pins the multiplier per archetype.
3. **Paladin seed 167, double chicken in Underground Passage Level 1 (not fixed, I).** Two chickens in 40 s with 4 and 2 potions after a leave-game and the re-planned walk, the model having just chosen Cave Level 1: a plan/walk mismatch after a chicken, not a zone-pick or potion defect. I found no small safe rule in the zone model that explains it (the logs of holdout 2 are not in this branch); a new Paladin seed set (below) is the test of whether it recurs.

## 2. Sorceress screen (class 4, lockstep speed 3, horizon 90, both arms on the new binary, 4 runs at once on the long lane)
| seed | legacy xp@88 / chick / gold | margin xp@88 / chick / gold |
|---|---|---|
| 101 | 16099 / 1 / 11756 | 12018 / 1 / 6449 |
| 103 | 16880 / 0 / 1850 | 22883 / 0 / 7945 |

Pooled over 2 pairs: deaths 0 / 0, chickens 1 / 1, XP ratio 1.058, per-seed difference mean +961 (sd 7130). **Verdict: INCONCLUSIVE (abreport needs 3 pairs).** Seed 103 (holdout 2: margin 2.5k against legacy 23.6k) now reads margin 22.9k against legacy 16.9k, with no chicken: the catastrophic seed is gone (V; one pair, so U that the fix is the cause, the legacy arm also moved with the lockstep run). Seed 101 margin is still 25 percent under legacy (12.0k against 16.1k): the model sat in lower-paying zones for part of the run; not analysed further.

## 3. Not run (stopped on the coordinator's request)
A first wave was lost to a builder stall (load above 500, bot watchdog exits at about 17 virtual minutes, no data kept). The second wave finished (table above). The third wave (Sorceress seed 107 and a fresh seed 181, both arms) was cancelled 10 minutes in because the shared builder had to serve merge gates; the Paladin waves (two fresh seeds) were not started. Ready commands (from the worktree, one wave at a time, lane long, speed 3 so every run reaches minute 90 with 4 at once):
- `tools/survival/wavelane.sh --lane long survhold3 1.5 3 4 legacy:107 margin:107 legacy:181 margin:181`
- `tools/survival/wavelane.sh --lane long survhold3 1.5 3 2 legacy:191 margin:191 legacy:193 margin:193` (Paladin, `AB_CLASS` 2 is the class argument; heroes as in holdout 2)
- report: `abreport -death-slack 1 -chicken-slack 1 -xp-floor 0.95 -horizon 90 DIR` over the logs of all seeds; `wavelane.sh` passes `--notify`, which needs the push webhook config.
Work dir on the builder `d2d-lab/survhold3` (bin/ built from this branch).

## 4. Recommendation (the user decides; nothing flipped)
1. **Sorceress: INCONCLUSIVE, no flip.** One screen seed recovered, one still under legacy. The screen needs seed 107 (and a fresh one) to reach 3 pairs; then 8 fresh seeds for the flip (docs/ab-design.md), at 4 runs per wave about 40 minutes each.
2. **Paladin: PASS at 6 seeds (holdout 2); 2 more fresh seeds are still needed for the 8**, not run here.
3. Keep the data fields merged: with the knob at 0 they change nothing.
