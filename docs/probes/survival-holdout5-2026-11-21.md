# Survival A/B holdout 5: Sorceress 8 fresh seeds, Paladin 2 (2026-11-21)

> status: current (card survhold8). Tags: V verified (abreport and bot logs of this run), I inferred. `bot.zoneModel` stays 0; no default changed; recommendation only. Follows [holdout 4](survival-holdout4-2026-11-19.md) sections 5 and 6 (wand hero, starved-yield fix).

## 1. Setup (V)
Second worker (the worker host), current origin/main bot (33e6ea8aa) built into its own dir, lab d2d `d2d.survtune-base-fabae935-linuxbuild` (same as the earlier second-worker waves), `tools/survival/wave.sh` with `abrun.sh`: speed 3, horizon 1.5 virtual hours, hero SETUP clvl 12, 500 gold, wand (class 4) or hand axe (class 2), `resource.pointsDecisions` 1, `bot.portalGuard` 0, only `bot.zoneModel` differs. Both arms of every pair ran on this host, in waves of 4 runs (2 pairs), each about 20 real minutes. All 20 logs: `exit 0` and `writes after start: 0`. Heroes are CHEATED for provenance (SETUP): a relative A/B only. Fresh seeds, none of 101 103 107 181 211 213 241 243. Runs are not bit-identical between repeats and cross-host numbers are unverified, so these pairs are not mixed with builder runs.

## 2. Sorceress (class 4), 8 pairs
`abreport -death-slack 1 -chicken-slack 1 -xp-floor 0.95 -horizon 90` (xp at minute 88 to 89):

| seed | legacy xp / chick | margin xp / chick | margin/legacy | deaths L/M |
|---|---|---|---|---|
| 301 | 13356 / 0 | 11056 / 0 | 0.83 | 0 / 0 |
| 303 | 24908 / 0 | 11004 / 0 | 0.44 | 0 / 0 |
| 305 | 26942 / 1 | 20333 / 0 | 0.75 | 0 / 0 |
| 307 | 31025 / 1 | 17047 / 0 | 0.55 | 0 / 0 |
| 309 | 11322 / 4 | 6128 / 0 | 0.54 | 0 / 0 |
| 311 | 23183 / 0 | 5234 / 0 | 0.23 | 0 / 0 |
| 313 | 24985 / 0 | 3538 / 0 | 0.14 | 0 / 0 |
| 315 | 20076 / 0 | 14741 / 1 | 0.73 | 0 / 0 |

Pooled: difference mean -11041, sd 6757, 95% interval [-16691, -5391], ratio 0.497 (floor 0.95), deaths 0 / 0, chickens 6 / 1. **Verdict: FAIL** (the whole interval is under the floor; every one of the 8 seeds is below legacy). Combined with holdout 4 section 6 (4 builder seeds, ratio 0.615) the picture is the same: 12 of 12 fresh or reused seeds below legacy.

## 3. Paladin (class 2), 2 fresh pairs
| seed | legacy xp / chick | margin xp / chick | margin/legacy |
|---|---|---|---|
| 321 | 24088 / 0 | 37462 / 1 | 1.56 |
| 323 | 26615 / 0 | 31181 / 1 | 1.17 |

abreport: ratio 1.364, deaths 0 / 0, chickens 0 / 2: **FAIL by the chicken rule** (2 > 0 with slack 1), although XP is well above legacy; below the 3-pair minimum, so treat as **INCONCLUSIVE** (I). Holdout 4 section 3 recorded 8 earlier Paladin seeds as PASS; two new chickens in two pairs is a signal to watch (the margin arm played more minutes in areas 4 and 10 at higher intake: area 4 30.8 against 16.2 life/min, area 10 75.7 against 34.1), not a verdict. Needs 3 or more further Paladin pairs before the flip decision.

## 4. Analysis of the Sorceress FAIL (I, from `zone choice`, `zone revive` and `zone shadow` lines of the margin logs)
- **Two modes.** Seeds 301, 305, 307, 315, 303 (ratio 0.44 to 0.83): the model kept k at 1.0 to 2.4 and moved to Stony Field and Cave Level 1 and 2 (realized 200 to 950 xp/min against model 57 to 257): fine, just a bit behind. Seeds 309, 311, 313 (ratio 0.14 to 0.54): k hit the 3.00 clamp inside the first minutes (k 3.00 in 18 of 19, 36 of 38 and 33 of 35 zone-choice lines) and the model never came back down, so it lived in Cold Plains and Den of Evil (realized 7 to 80 xp/min) for the rest of the run.
- **Zone: Burial Grounds (area 17) and Mausoleum.** The first choice on every seed is Burial Grounds (model 150 to 160 xp/min, margin 11.2); where it is tried at k 3 the realized yield is 4 to 11 (seeds 309, 313, 307: "realized 4 (model 160)", "realized 5 (model 152)"). The #997 starve fix does move the model off it after about 3 minutes, but not to a productive zone: Cold Plains stays at model 54 against realized 40 to 85.
- **Decision: the merc revive.** The margin arm logs `zone revive: merc_left_dead ... the kit comes first and the purse cannot cover both` 15 to 26 times in the three worst seeds (and 8 to 15 in the others), against 2 to 3 per legacy run. The model keeps the mercenary dead to save gold, then plays the caster without it, which lifts intake and the k escalation (k 3.00) that holds it in the low-yield zones. This is inferred from the log counts, not isolated by an experiment (U).
- Safety does not pay for it: deaths 0 / 0, chickens 6 against 1 (legacy 4 of the 6 on seed 309 alone).

## 5. Recommendation (the user decides; nothing flipped)
1. Sorceress: **FAIL at 8 fresh seeds** (ratio 0.497, interval under the floor); do not flip, and do not run more of the same screen. Next card: look at the merc revive decision (`zone revive: merc_left_dead` under k 3.00 on the caster archetype) and at why k, once at the clamp, has no way down; then a new 3-seed screen.
2. Paladin: earlier PASS at 8 seeds stands; these 2 pairs read FAIL on chickens only and below the minimum; run 3 more pairs before any per-class flip.
3. `bot.zoneModel` default stays 0.
