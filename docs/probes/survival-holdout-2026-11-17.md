# Survival A/B holdout, seed 149: cause, fix, rerun (2026-11-17)

> status: current (card survhold, user decision 2026-11-17: `bot.zoneModel` stays 0 until the holdout is fixed). Follows [the re-ask record](survival-ab-reask-2026-11-17.md). Tags: V verified (run logs, `abreport -horizon 90 -death-slack 1 -chicken-slack 1 -xp-floor 0.95`), I inferred, U unverified. Lockstep d2d on the Tower builder, heroes CHEATED for provenance (SETUP clvl 12, 500 gold, free Rogue, Paladin weapon `hax`; Sorceress none): a relative A/B, not evidence about game behaviour. Writes after start: 0 for every run of this card (V: the final line of each log, and for the runs the 45 minute builder lane cut short the count of `err lab verbs off` lines, 0 in all).

## 1. Why seed 149 fails (V: `margin/149.log` of three runs)
- **Not an artefact.** The margin arm of seed 149 was run three times (the Orion batch of the re-ask record, then again on the Tower with the same binary): the same decisions at the same virtual minutes, the same chickens and the same death (2 to 3 chickens, 1 death, one or two recoveries). The legacy arm is 0 chickens, 0 deaths, no recovery in both of its runs. Lockstep is deterministic per seed, so there is no variance to average: what 149 shows, it shows every time (V).
- **The chain.** The hero (clvl 12, 154 life, weapon damage 3-6, 500 gold) buys a tome and a scroll (400 gold) in the first minute, then the model sends it to Black Marsh and Dark Wood on the pay-for-risk route (`why: pays_for_risk`, margin 0.7 to 0.9 against a need of 1.5; k read 1.0, then 1.35, 1.93, 2.91, 3.31 as the ledger learned the zone hurt more than predicted). The gold on hand was 100 to 290. The hero used 25 potions in about 6 minutes, ran out of gold (5 gold, `recovery: start (gold 105 < kit 500)`), and recovery then farmed with no potion: chickens at 83/154 life in Stony Field and Cave Level 1, a death on the portal trip (`back to field: portal did not lead to the level ... wanted 6, in 9`, a stale return target, see section 5).
- **So the re-ask and the recovery routing were not the cause.** The zone model's pay-for-risk route (`typicalEnter` 8, `badFloor` 0.55 for balanced, card survtune) lets a zone in whose bad-pack margin is under the need, because the typical pack is comfortable and the zone pays. It lives on chickens and potions and nothing in it asks whether the hero can rebuy a belt. The legacy arm sat in zones with margin 1.1 to 1.7 and a gold income that kept it solvent (gold 900 to 1500 by minute 60 in the legacy run).

## 2. The fix (I: a rule in the data tables, tested on both builds)
`riskMinGold` per dial row in `internal/survival/params.json` (balanced 300, bold 150, cautious n/a): the risk route is closed to a hero whose pack gold is under it. The zone then falls to the normal rules (`Why` code `too_poor`, new, in the decision line as `not <zone>: too_poor`). A hero already in the zone is not affected (the stay rule). Code: `judge` in `internal/survival/controller.go`; the number is data, not a class check. Tests (`internal/survival/riskpurse_test.go`, `riskroute_test.go`): a poor hero is never risk-routed at any level 4 to 40 for the Paladin (bruiser) and the Sorceress (caster) tables; hand-made candidates for both archetypes take the risk zone at exactly the limit and refuse it one gold under; the rows carry the condition. `go test -p 1 ./internal/survival/ ./cmd/d2bot/` passes. Knob 0 (`bot.zoneModel` 0) is untouched, no default changed. Numbers 300 and 150 are U (one holdout, nothing fitted).

## 3. Results (V). Horizon 90 virtual minutes: the 45 minute builder lane cuts a run of 120 virtual minutes short, so every run here is read at minute 90 (the legacy runs of 149 and 163 reached 120). XP is "gained since the start".
Paladin (class 2, `AB_SPEED=5`, 4 runs at once, nice 15), arms differ only in `bot.zoneModel`:

| seed | arm | death | chick | xp@90 | gold end | binary |
|---|---|---|---|---|---|---|
| 149 | legacy | 0 | 0 | 37047 | 2886 | fixed (arm 0 does not use the rule) |
| 149 | margin before the fix | 1 | 3 | 10652 | -100 | before |
| 149 | margin after the fix | 0 | 3 | 27202 | 1067 | after |
| 163 (fresh) | legacy | 0 | 0 | 28574 | 2179 | |
| 163 | margin before / after | 0 / 0 | 0 / 1 | 30636 / 29252 | 4434 / 763 | |
| 167 (fresh) | legacy | 0 | 0 | 28357 | 2276 | |
| 167 | margin after | 0 | 0 | 29680 (minute 89) | 635 | after |

Pooled, fixed binary, 3 pairs: deaths 0 / 0, chickens 0 / 4, XP ratio 0.917 (95 percent interval of the per-seed difference [-18193, 12966]). **Verdict (abreport, slack 1/1, floor 0.95): FAIL (chickens 4 against 0 + 1)**. Before the fix on seeds 149 and 163 (2 pairs): deaths 0 / 1, chickens 0 / 3, XP ratio 0.629, FAIL.

Sorceress (class 4, no weapon, `AB_SPEED=5`, horizon 90 because the speed 3 batch does not fit the lane), seeds 101 103 107, fixed binary, 3 pairs: deaths 0 / 0, chickens 2 / 2, XP ratio 0.858, interval [-8621, 4698]. **Verdict: INCONCLUSIVE** (3 pairs, the interval spans the floor; seed 107 margin 35 percent under legacy, seed 101 equal, seed 103 margin 20 percent under). Not a pass.

## 4. What the fix did and did not do (I)
- Did: seed 149's death and the first-minutes potion burn are gone (the model now says `too_poor` at gold 100, takes Dark Wood margin 3.8 and Stony Field margin 3.3, `best_pay_eligible`); XP at minute 90 went from 10652 to 27202.
- Did not: the margin arm still ends poorer than legacy on all three Paladin seeds (gold 1067 / 763 / 635 against 2886 / 2179 / 2276) while it uses no more potions (44, 69, 51 against 60, 63, 75), so seed 149 still hits recovery with an empty belt at about minute 55 and chickens three times in 20 seconds in Cave Level 1 (life 38/154, `recovery: farm area Cave Level 1`, 0 potions). The recovery state farms with no potions by design; with the legacy arm solvent it never enters it. I read the cause as the model's zones paying less gold per minute for this weak (damage 3-6) hero, not as a bug in the recovery code (I, not shown: I did not isolate the gold gap per zone).
- XP: seed 149 is still 27 percent under legacy at minute 90; seeds 163 and 167 are at parity (+2 and +5 percent).

## 5. Side finding, not fixed (I)
`back to field: portal did not lead to the level ... wanted 6, in 9` (old run, seed 149 margin, minute 5): the return target of a trip (area 6) did not match the area the hero had portaled from (9, Cave Level 1), and the hero died right after arriving. It appeared in the pre-fix margin 149 run and once in the fixed margin 163 run (`portal led to another level`), so it is not gone: a trip returns to a stale target. It was not followed by a death there. Owner would be the recovery or trip code (`internal/bot/goal.go` advanceTrip, `internal/bot/broke.go`).

## 6. Recommendation (the user decides; nothing flipped)
1. **Keep `bot.zoneModel` at 0.** The Paladin holdout still fails the chicken rule (4 against 1 allowed) on fixed code; the failure is now only the recovery chickens after the hero went broke, not deaths. The Sorceress is not shown (3 pairs at 90 minutes, 0.86 XP, wide interval).
2. Keep the `riskMinGold` fix merged: it removes the seed-149 death and the burn (XP 10652 to 27202), costs nothing with the knob at 0 and is data.
3. Next step if the model is to go on: close the gold gap of the margin arm (why a model-chosen zone pays about a third of the legacy gold for this hero; per-zone gold per minute from the logs, then the `gold/min` term in `Eco` and the purse weight `lambdaMax`), and make recovery chicken less without potions (for example a potion floor before a recovery farm); then 8 fresh Paladin seeds and 8 Sorceress seeds at the same horizon, read at minute 120 (needs a lane longer than 45 minutes or two-wave runs).
4. A per-class default is not needed yet: the same rule is used for both builds.

## Provenance
Heroes `SaHero-<seed>` (Paladin and Sorceress), SETUP in `prep.abmerc` (declared, ledger in the instance dir): [test-characters](test-characters.md). d2d `d2d.survtune-base-fabae935-linuxbuild` (fork branch `didle/rebase-20261115`), d2bot built from this branch (before: main at 074c13225, after: tip of agent/survhold). Runs on the Tower builder (`labrun tower`, one job at a time, work dirs `d2d-lab/survhold` and `d2d-lab/survhold-fix` there); nothing ran on the Orion desktop, nothing on the herdr box. Processes of this card: none left (the lane stops its children).
