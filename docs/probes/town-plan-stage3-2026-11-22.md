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

## Evidence for making stage 3 the default (card plan3, 2026-11-22 night; builder long lane, lab d2d devp20, CHEATED setups, mechanics only)

### `act1.den.clear` with `town.bankFirst` 1, level 10, 8000 gold (jobs 20261122-221125-7b1a, 212018-1980, 213936-7ae3, 205642-49a0)
| class, seed | stage 0 | stage 3 | purse and stash numbers |
|---|---|---|---|
| Necromancer 11 | RED (record 1 0x0014, Den left after 36 s, farming gold) | GREEN | stage 0: bank chore deposited about 6000 of 8000 (`town budget` bank -6000, shop -900), purse 112 at the recovery restart |
| Necromancer 3 | GREEN | GREEN | stage 0: bank -7200, shop -650, purse left 150; stage 3: same deposit, then `withdraw 650`, purse 800 = keep line (`keep=800 spend=0 purse=8000 stash=0 verdict=ready`) |
| Paladin 11 | GREEN | GREEN | stage 0: bank -7200, shop -700, purse left 100; stage 3: withdraw 700, purse 800 |
| Sorceress 11 | GREEN | GREEN (job 7ae3) | |
| Sorceress 3 | GREEN | GREEN | stage 0: bank -7200, shop -650, purse 150; stage 3: withdraw 650, purse 800 |

Reading: the defect (purse emptied by the bank chore, hero short in kit, recovery farming) shows only for one Necromancer seed at stage 0; at stage 3 all six runs are GREEN and each purse is topped up to the keep line (800 at clvl 10) by one withdrawal. No run is worse at stage 3 than at stage 0.

### Stage 3 gate scenarios, Paladin 11 (job 20261122-223636-5342)
| scenario | stage 3 | stage 0 (`D2SCEN_KNOBS=town.plan=0`) |
|---|---|---|
| `town.plan.broke-after-bank` | RED: `wealth_between` wanted purse plus stash >= 7000, got 6453 (the assertion floor is not reached after the visit's purchases: U whether the 7000 floor or the shop spend is wrong) | RED: `recovery: start (gold 116 < kit 500)` appears, which the scenario forbids (log_line max 0): the I9 defect reproduces at stage 0 and not at stage 3 |
| `town.plan.malus-errand-gold` | BLOCKED: bot run cut before its virtual budget (not judged) | BLOCKED, same |

The malus scenario is NOT fixed: the cut is still there (cause U; the earlier note blamed the bot's own save point quitting the d2d, the log was not read again this time, the builder main lane was busy). The stage 3 `broke-after-bank` RED needs its floor checked before the scenario can be a gate.

### Is stage 3 safe to propose as the sandbox default?
Mostly yes for purse behaviour: the den scenario table above, with one seed per class pair and no regression, and `broke-after-bank` shows the stage 0 defect gone. Missing before a flip: (1) a verdict for `town.plan.malus-errand-gold` (BLOCKED at both stages) and a corrected floor for `broke-after-bank`; (2) a Paladin run at stage 3 on a longer stage (Blood Raven, 60 virtual minutes) to see the withdraw rule over many visits; (3) the sandbox itself runs stage 1 and `quests.boardSM` 1; the interaction of stage 3 with those two on the real friends board is U. Nothing was flipped.
