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
| same, setup with no healing potion anywhere | | visit settled the purse (`purse=8000` to stash 6390 after the first purchases), chunk started within 1 virtual minute; the run was then cut by a d2d channel loss (BLOCKED, infrastructure); the rerun (job 20261122-135010-5fd8) read the same: BLOCKED |

Reading of the bot logs of both setup-2 runs (V, the `town plan: purse` lines): first visit `keep=1610 spend=210 purse=8000 stash=0`, next visit `purse=720 stash=6390` (8000 settled to the stash, the shop paid from the purse), later visits `purse` back at the keep line minus the spend (withdrawals of the gap), stash falling only by those top-ups. Two harness limits stop a verdict: (1) both runs end with the bot's own save point (`save point: saving and re-entering the game`, xp at risk from the CHEATED level) which quits the d2d, so the runner reports BLOCKED, not RED or GREEN; (2) the hero has 55 life at clvl 22 (CHEATED level without stats), burns its potions and goes town, field, town about every 6 real seconds (guard `ran but is still needed`): the same ping-pong the stage 2 guard scenarios show, not caused by the purse (U: not compared with a stage 2 run of the same setup). The scenario asserts `stash_between` and `gold_between` are therefore NOT judged.

U: the Sorceress rows were not run; the run budget went to the Paladin reruns.
