# Survival A/B with the cost-event re-ask, Paladin 8 seeds, holdout, Sorceress screen, 2026-11-17

> status: current (TASK-595, card survreask). Follows [the 8-seed rerun](survival-ab-8seed-rerun-2026-11-16.md). Tags: V verified (run logs, `abreport -death-slack 1 -chicken-slack 1`), I inferred, U unverified. Lockstep d2d only, heroes CHEATED for provenance (SETUP clvl 12, 500 gold, free Rogue, weapon `hax`): a relative A/B, not evidence about game behaviour. Every one of the 30 logs ends `writes after start: 0` (V).

## What changed (V: merged, `go test -p 1`, race gate on the worker host)
A cost event now makes the next leveling pick ask the zone model again instead of waiting for the 3 minute dwell (`internal/bot/zonemodel.go`, `bot.zoneModel` 1 only; knob 0 records nothing, tested):
- events: a chicken, a death, a merc death, two emergency (rejuvenation) drinks within a minute (`potion_crisis`);
- limits: at most 6 re-asks per 10 minutes, none closer than 5 seconds; the model's own hysteresis (switch margin, step-down rows) still decides whether to leave, so a re-ask that finds the zone fine says `stay`;
- the log line is `zone reask: reason=chicken k 1.64 -> 3.58 choice 5 (leave)` (k is k_zone x k_hero of the zone the hero is in, from `Decision.CurK`; `choice` is the area id).
Tests (two classes, Sorceress and Paladin, fake clock): re-ask after a chicken and the model's zone is used, no re-ask inside the gap, cap per window and reopening after it, the seed-149 series (a chicken every 7 s; the model closes the zone on the second) leaves after at most two, potion crisis needs two drinks, knob off leaves nothing pending.
Not covered: hard hit events from the ledger as a trigger (the ledger folds those into k already; no separate event exists in the bot).

## Results (V; legacy and margin arm both rerun on this binary, `AB_SPEED=5`, four runs at a time, nice 19)
| set | pairs | XP ratio | 95 percent interval of the difference | deaths legacy / margin | chickens legacy / margin | verdict |
|---|---|---|---|---|---|---|
| Paladin 8 seeds 101 103 107 109 113 127 131 137 | 8 | **1.005** | [-3884, 4326] | 1 / 0 | 15 / 2 | **PASS** |
| holdout 139 149 151 157 | 4 | 0.951 | [-14437, 10244] | 0 / 1 | 1 / 3 | **FAIL** (chickens 3 > 1 + 1) |
| Sorceress 101 103 (107 short in both arms) | 2 | 0.700 | [-13692, 2127] | 0 / 0 | 1 / 0 | INCONCLUSIVE (2 pairs, under the 3 minimum) |

Pooled over the 12 Paladin seeds: deaths 1 / 1, chickens 16 / 5, mean XP difference about -550 per run (ratio about 0.99). Potions, merc deaths (5 a run in both arms, as before: the controller does nothing for the mercenary) and clvl are in the `abreport` output of the logs (worker host `d2d-lab/survreask`, `d2d-lab/survreask-sorc`).
Re-asks in the 15 margin runs: 5 lines (4 chicken, 1 potion crisis), 4 of them left the zone. k behaviour: it rises after the event (1.64 to 3.58, 1.17 to 1.52) and the model then moves to a lower zone; on seed 149 k read 3.31 at the event and 2.43 after the area change (the zone left had the large correction).

## What this says about the legacy arm
With `AB_SPEED` 5 the legacy arm was NOT safe on this batch: seed 101 four chickens and a death, seed 127 ten chickens (the 8-seed rerun at speed 3 had 4 chickens over 8 seeds). The controller arm had 1 and 0 on the same seeds. So the pair now shows a safety gain on the tuning seeds (15 chickens against 2), but the legacy spread over runs is the larger effect (0 to 10 on one seed), the speed was not the same as in the earlier batch, and 8 seeds cannot separate the two (U). Do not read it as proven.

## Seed 149 again (V: `zone choice`, `safety decision`, `character died` lines of `margin/149.log`)
The holdout failure is seed 149 once more, but the cause this time is not the one the re-ask fixes. Two chickens (areas 4 and 9, 16 s apart) and one death (in town, area 1) happened inside the bot's recovery state: `recovery: start (gold 105 < kit 500)`, `Recovery: farming gold`, `recovery: farm area Stony Field`. The recovery trip returns through `back to field: portal in town` to the zone the hero had left (Dark Wood, corrected bad-pack margin 0.7, k 3.3) and then picks its own gold-farm area; none of that goes through the leveling pick, so the cost event stayed pending for about 8 wall minutes (the next normal pick), where the re-ask fired (`reason=chicken k 3.31 -> 2.43 choice 5 (leave)`) and the model left Dark Wood. The re-ask itself works as built; the missing part is the recovery path (I, not yet changed): the recovery return portal and gold-farm area should ask the zone model (or at least avoid a zone with a pending cost event). Owner would be `internal/bot` recovery code; not part of this card.

## Recommendation (plain)
1. **Do not flip `bot.zoneModel` yet.** The Paladin gate on the eight tuning seeds passes again (XP 1.005, no more deaths, far fewer chickens in this batch), and the Sorceress is not shown: two pairs only, XP 0.70, seed 107 ran into the wall cap in both arms (a rerun needs a longer cap, `AB_SPEED=3`). The four fresh holdout seeds fail the chicken part by one because of seed 149, whose chickens come from the recovery state, which the re-ask does not reach.
2. The re-ask code is safe to keep merged (knob 0 changes nothing, tests on two classes, race gate green). It is what the doc of 2026-11-16 asked for and it does fire and leave after a chicken.
3. Next step if the user wants the model on: route the recovery trip through the model (portal back and gold-farm area), then 8 fresh Paladin seeds plus the Sorceress 8 seeds; per-class default needs the knob to become data first.

## Provenance
Heroes `SaHero-<seed>` (Paladin and Sorceress): [test-characters](test-characters.md). d2d `d2d.survtune-base-fabae935-linuxbuild` (fork branch `didle/rebase-20261115`), d2bot built from agent/survreask at the merge commit's content. Work dirs on the worker host: `d2d-lab/survreask`, `d2d-lab/survreask-sorc`. Processes of this card: stopped, registry empty.
