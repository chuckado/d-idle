# Town planner, stage 3 (purse), 2026-11-22

> status: current. Design: docs/design/fable-review-2026-11-21/town-planner-b-model.md (card tp-purse, TASK-757). Setups are CHEATED (level, gold, items by SETUP verbs): mechanics only, never evidence about play or balance.

## What stage 3 does (`town.plan` 3; 0 stays the default)

- One purse target per visit: carry line (`town.stashGoldBase` + `town.stashGoldPerLvl` x clvl, never below the minimal kit) plus the gold the supply plan spends this visit (`Plan.Spent`, includes a `gold_carry` line). The stash stop deposits the surplus; the shop withdraws the gap (never more than the stash holds). `town.bankFirst` and `town.withdrawShop` do not matter at stage 3. One log line per visit: `town plan: purse keep=.. spend=.. purse=.. stash=.. verdict=..`.
- Recovery and poverty judge WEALTH (purse plus stash; a stash that was never read counts 0).
- Contract `docs/contracts/town-basics.md` "Bank before the shop" carries the new sentence.
- Code: `internal/town/{chores,stash,basicsfix,town}.go`, `internal/bot/{broke,leveling}.go`; tests `internal/town/purse_test.go` (no tables, runs everywhere), `purserig_test.go` (two classes, needs the game tables, skips on a box without them).

## Scenario record (builder long lane, lab d2d, bot of this branch)

| scenario | today's code (stage 2 code, knob 3 unknown to it) | stage 3 |
|---|---|---|
| `town.plan.broke-after-bank` Paladin 11 | GREEN (the I9 shape did not reproduce at clvl 3 with 8000 gold; U why) | GREEN |
| `town.plan.malus-errand-gold` Paladin 11, first setup (4 healing in the belt) | RED: no town visit before the chunk, 8000 carried into the field, hero died at life 55 (setup fault: no visit means no settle) | same fault |
| same, setup with no healing potion anywhere | | visit settled the purse (`purse=8000` to stash 6390 after the first purchases), chunk started within 1 virtual minute; the run was then cut by a d2d channel loss (BLOCKED, infrastructure); see the rerun line below |

U: the Sorceress rows were not run; the run budget went to the Paladin reruns.
