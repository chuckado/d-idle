# Necromancer earned Act 1 drill (card necrodrill, 2026-11-22)

> status: current. Verdict: BLOCKED, NOT ready to play. Tags: V verified in a named run, S read in a log, U unverified. Builder LONG lane, lab d2d devp20 (`--read-combat`), `combat.minionRead` 1, build `necromancer_summoner_sme`.

## 1. Runs

| run | what | verdict |
|---|---|---|
| 20261122-172117-b149 (V) | `necro.act1-earned`: a fresh clvl 1 Necromancer, town start, goal `campaign:normal_act1`, NO setup write, `writes after start: 0` | BLOCKED: quest record 6 (Andariel) not observed in 400 virtual minutes. The job was cut at 3604 s wall by the 60-minute hard limit of a detached job (the scenario's own 140-minute wall budget never applied); the result file was written at 3586 s |
| 20261122-182247-6965 (V) | stage 1 as the existing act1 scenarios with `-class Necromancer -seed 11` (CHEATED level fixtures, default table `necromancer_summoner_normal30`): `act1.den.clear` and `act1.burial.blood_raven` | `act1.den.clear` RED (record 1 read 0x0014, wanted bit 0 set; Start was tapped only 4 wall minutes before the end); `act1.burial.blood_raven` BLOCKED (bot run cut before its virtual budget; the hero ended at clvl 16 in the Rogue Encampment) |

The ledger refusal ("missing bot_commit") is a launch detail: `run.sh` reads `git rev-parse` in the synced mirror, which has no `.git`; passing `env D2LOD_BOT_COMMIT=<hash>` on the labrun command line fixes it (V: no refusal in run 6965).

## 2. How far the fresh hero got (V from the bot log and result file of run b149)

- clvl 14, 5381 kills, 77 boss kills, 0 deaths, 168 potions, 13 level-ups in 24008 active seconds (11448 xp/h). Level by wall-time hour: 1, 6, 9, 12, 13, 13, 14.
- Quests: Den of Evil done (`Quest done: den_of_evil` after about five virtual minutes; record 1 ended 0x2001). Nothing else: no Blood Raven, no Cain, no Countess, no Andariel.
- Where the time went: 88 "Quest: fighting (Level up to N)" states (N = 12 to 15), 45 "Recovery: farming gold", 102 town trips (heal below the resume line 39, healing potions 25), 327 area changes (Rogue Encampment to Cold Plains and back about 60 times), stash 29341 gold with 423 in the purse, skill points all spent (15 hard points, `skill_levels` Raise Skeleton 5, Skeleton Mastery 5, Clay Golem 1).
- What stopped it: the board said "Training toward Sisters to the Slaughter: N levels to go" (10 lines) and the bot kept levelling; 38 "quest board step" actions failed, 37 with `exit to Cold Plains: nav: click {149 330}: no path: the click would cross the trigger area of an unwanted exit`, three failsafe-ladder escalations, and one `back to field ... portal led to another level`. The campaign goal without the board-taps driver starts no chunk (autoStart is never), so only the Den (a goal of its own) completed.

## 3. Blockers, ranked

1. Nav: the Rogue Encampment exit to Cold Plains click is refused as crossing an unwanted exit trigger (37 failures in one run). It sits on every Act 1 quest. Not Necromancer-specific (U: the Paladin runs used the same exit; the dev logs of act1more show the same family for the Outer Cloister).
2. Quest start: the drill cannot start the chunks after the Den. The `board-taps` driver taps one quest per scenario; a whole-act run needs a driver that taps every offered chunk in order (the Cain scenario taps only its own). Until then an earned multi-quest drill cannot be built from the scenario format.
3. Pace: clvl 14 after 400 virtual minutes (11.4k xp/h) with 102 town trips; the "Level up to N" training loop is the time sink, and the hero at that level is below the board's readiness line for Blood Raven (U: the board's reason is not logged, same finding as act1more finding 3). The army kills (5381 kills, 0 deaths), so this is throughput, not survival.
4. Tooling: a detached labrun job has a 60-minute hard limit regardless of `budget.wall_min`; a 400-virtual-minute drill needs stages, and `act1.burial.blood_raven` for a Necromancer hit a bot-run cut before its budget (cause U, the log tail shows a normal stop).
5. Den for the Necromancer under the act1 scenario (CHEATED level 10, default table) did not complete in 45 virtual minutes although the fresh hero finished it in five (U why: different build table and fixture level; read the Den log of run 6965 before drawing a conclusion).

## 4. Next cards

- `nav: Rogue Encampment to Cold Plains exit click refused as an unwanted exit trigger` (owner nav; blocker 1).
- `scenario driver: tap every offered chunk of a campaign in order (board-taps all)` (owner scenario/provingground; blocker 2), then stage the drill as three jobs each under 60 minutes: Rogue Encampment to Cain (record 4), to Countess, to Andariel, restoring the saved hero between stages.
- `quest board: log why a chunk is not startable at the hero's level` (blocker 3, also act1more finding 3).
- `scenario: necro stage runs with the SME build and the Necromancer matrix on act1.*` (blocker 5).
- Test character: none created by this drill (the builder runner creates and deletes its own heroes; the provenance is the scenario's `provenance` text). A hidden-instance Necromancer in docs/probes/test-characters.md waits for the 23:00-07:00 EST window.

## 5. Follow-up runs: the exit refusal, the knobs, and why the Necromancer runs are cut (2026-11-22, evening)

### 5.1 Is the Cold Plains exit refusal fixed by `nav.warpRead` and `nav.fullGrid`? (V for the counts, U for the conclusion)
| run | scenario, seed | knobs | verdict | `unwanted exit` lines, `mouth gate` lines |
|---|---|---|---|---|
| 20261122-190413-c261 | `act1.burial.blood_raven`, Paladin 11 and 3 | warpRead 1, fullGrid 1 | GREEN | 0, 0 |
| 20261122-190413-c261 | same, Necromancer 11 and 3 | warpRead 1, fullGrid 1 | BLOCKED (cut, section 5.2) | 0, 0 |
| 20261122-195405-6208 | same, Necromancer 11 | OFF (defaults) | BLOCKED (cut, section 5.2) | 0, 0 |
| 20261122-195405-6208 | `act1.den.clear`, Necromancer 11 | warpRead 1, fullGrid 1 | RED, record 1 0x0014 as before | 0, 0 |

The refusal does not appear in these scenarios with the knobs OFF either, so they cannot show that the knobs remove it. The 37 refusals of run b149 came from a fresh hero whose board step walked from the Blood Moor to Cold Plains after hours of play, knobs OFF; no staged scenario reproduces that walk. Evidence for a default flip of `nav.warpRead` and `nav.fullGrid`: none from this card (mapverify #1108, 55 to 0 mouth-gate refusals on the Countess, remains the only measurement). A direct test needs a stage that starts the hero in the Blood Moor and sends it to Cold Plains with the knobs OFF first (to reproduce) and ON second.

### 5.2 Why the Necromancer `act1.burial.blood_raven` runs are cut (V, bot log of 20261122-190434 seed 11)
The hero is not slow, it is stuck from the first minute. clvl 16 fixture, 8000 gold, 6 mana potions in the pack and none in the belt:
1. the town bank chore deposits 6000 of 8000 gold (`byChore stash -6000`, `town.bankFirst` default; the act1more scenarios set it to 0, the Necromancer run used the same file and was cut anyway);
2. the kit gate asks for mana potions 0/3 and the shop plan says `wantMP 0` because the pack holds 6 (`packMP 6, beltMP 0`);
3. the quest step then refuses with `not ready: 0 mana potions < 2 (caster)` (internal/task/gate.go:59, World.ManaPotions = `in.Belt.MP`, internal/bot/task.go:357, belt only), forever: `needs_attention` from 19:04:47 to the stop at 19:33:40, 16 kills in 29 wall minutes.
The guard counts belt plus pack for mana only at `town.plan` stage 2 or `quests.boardSM` 1 (internal/bot/routeguard.go:368, 395), both default 0 and absent from the scenario knobs, while healing potions already count the pack (task.go:356). A Paladin never meets it (no mana need). So the cut is a bot defect, not the 60-minute job limit and not a Necromancer pace problem: the same family as act1more finding 2 (belt against pack).

### 5.3 Cards
- `quest gate: count pack mana potions like healing ones` (owner task/bot): fix `World.ManaPotions` at stage 0 and 1, or make `quests.boardSM` the default; must-fail test on Necromancer and Sorceress with mana in the pack and none in the belt. Then rerun `act1.burial.blood_raven` for the Necromancer.
- `scenario stage: Blood Moor to Cold Plains walk, knobs OFF then ON` to settle the default flip.
- `act1.den.clear` for the Necromancer stays RED with and without the knobs (record 1 0x0014 against 0x0001); the Den log of run 6208 (V, read): 409 kills, 4 boss kills, 0 deaths, 33 potions in 45 virtual minutes, but the hero went back to town again and again for `low on mana potions` and `mp belt columns below 50%` (Cave, area 9, then area 13 and Cold Plains), so mana potion supply, not survival, set the pace and the Den bit was not set in time.

## 6. Stage results after the fixes (2026-11-22, late evening) and the revised blocker list

| stage run | verdict | note |
|---|---|---|
| 20261122-205642-49a0 `act1.burial.blood_raven`, Necromancer 11 | GREEN (60 virtual minutes) | after the pack-mana fix (manapack, #1135); before it: stuck 29 minutes on "0 mana potions < 2" |
| 20261122-205642-49a0 `act1.den.clear`, Necromancer 11 | RED, then GREEN | RED cause: the bank chore moved 6000 of 8000 gold into the stash, the hero left the Den after 36 s to farm gold (denfix, #1140); with `town.bankFirst 0` GREEN for the Necromancer and the Paladin (job 20261122-212018-1980). Record 1 0x0014 = 0x4 Akara talked, 0x10 U, not the completed bit 0 |
| 20261122-213319-0b57 `nav.bloodmoor-coldplains`, Paladin and Necromancer, seeds 11 and 3, knobs OFF and ON | GREEN 8 of 8 | 0 `unwanted exit`, 0 `mouth gate` lines in all eight logs, Cold Plains entered in each: the Blood Moor to Cold Plains walk does not reproduce the 37 refusals of the fresh drill with or without `nav.warpRead` and `nav.fullGrid`. So there is still no default-flip evidence for the two knobs from this card |

Revised blockers for the earned Act 1 drill (ranked):
1. Whole-act run: the tap driver now exists (`trigger.tap_quest` list or `*`, tapall #1130) but no multi-quest stage scenario uses it yet; stages must restart from declared level fixtures (the runner cannot restore a save) and each job must stay under the 60-minute limit of a detached job.
2. Bank chore against the quest kit: with the default `town.bankFirst` the purse is emptied before the kit is bought (card in progress; a Necromancer needs many mana potions, so it shows first).
3. Fresh-hero pace: clvl 14 in 400 virtual minutes (section 2); not re-measured after the fixes.
4. The fresh-drill exit refusal (37 lines, knobs OFF) is not reproduced by any staged scenario: U which walk and which position produced it.
Is the drill feasible in stages now? Yes for single quests from fixtures: Den and Blood Raven are GREEN for the Necromancer on the lab d2d. Not yet for the chain Rogue Encampment to Andariel; the next step is stage scenarios (`tap_quest: "*"`, a level fixture and a 55-minute budget each) for Cain, the Countess and Andariel. Still NOT ready to play.
