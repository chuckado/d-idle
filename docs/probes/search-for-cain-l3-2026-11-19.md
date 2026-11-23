# Search for Cain, drill L3: the real bot chain on a d2d lab slot (2026-11-19)

> status: result of drill CAIN-L3 (task `cainl3`), follow-up to `search-for-cain-l2-2026-11-18.md`. NOT Game.exe, no Wine instance, nothing on the sandbox: one d2d lab slot on the builder, lab binary with `--lab-verbs` (prep only), throwaway Sorceress `Cainy-11` (seed 11). The HERO IS CHEATED/IMPOSSIBLE (SETUP writes by `tools/ab/prep`: clvl 18, 24, then 32 with its experience, 3000 to 20000 gold, ledger `testdata/prep-ledger.txt`): mechanics only, never evidence of balance or of how a normal hero fares. Tags: V verified live (fixture or log line named), S secondary (source), U unverified. Fixtures: `docs/probes/cainl3/testdata/` (log excerpts as `.txt`, the repo ignores `*.log`). Harness: `docs/probes/cainl3/run.sh` (prep, then the real `d2bot --backend d2d --managed --ladder --watch --goals campaign:normal_act1`, the sandbox's own flags) and `poll.py` (plays the player: `GET /api/quests`, `POST .../start` and `.../handin` on the watch routes, nothing else).

## Method

`run.sh SEED CLVL WALLMIN SLOT` (env `GOLD`, `TUNE`): `prep` makes the hero (SETUP, ledgered) and saves; the bot loads it with a d2d WITHOUT `--lab-verbs`, so no gameplay-affecting write can follow the start (`writes after start: 0` by construction). `poll.py` taps Start for the offered chunk of `search_for_cain` and Hand in when the board says awaiting, and logs every change of the board view. Everything else (route, fights, potions, town trips, tree, Akara, stones, portal, cage) is the bot. Wall times below are the bot log's clock; the lockstep d2d runs about 4 to 9 times faster than wall time.

Declared run knob: from run 6 on `combat.crowdAbandonN` is 0 (override file in the instance state dir, `TUNE`). At the default (15) the bot chickens (Save and Exit) at every arrival in the Underground Passage and the Dark Wood of this seed (21 to 30 hostiles, "stale pack"), over twenty times in a row at clvl 18, 24 and 32, and never reaches the tree (`run2-up1-chicken-loop.txt`, `run3-up1-crowd.txt`). A user on the sandbox runs the default: see Open item 1.

## Result in one line

With the three fixes below plus the two earlier ones of this branch family the real bot chain ran UNATTENDED from the board tap to a finished quest on the lab slot in about 9 minutes wall (run 10, 11:31 to 11:41): scroll, Akara read, stones, Tristram portal, archers, cage, return to camp, quest word `0x2019` (done), log `Quest done: search_for_cain`. Run 10 used the declared knob and a clvl 32 hero.

## Timeline of the final run (run 10; times are bot log wall clock)

| step | result | evidence |
|---|---|---|
| accept on the board (POST start, chunk `scroll`) | PASS 11:31:15, `ok:true state:accepted`, board shows `running take_scroll` in 5 s | V `run10-poll.txt` |
| walk to the Dark Wood, waypoint | PASS (Stony Field, Underground Passage 1, Dark Wood, about 90 s wall) | V `run10-chain-excerpt.txt` |
| find the tree (CAIN-7 room walk) | PASS after fix 1; without it FAIL (livelock, Bug 1) | V `run1-livelock-tree.txt`, `tree-class-minus1.txt` |
| click the tree, pick up the Scroll of Inifuss (CAIN-3) | PASS, quest word `0x0008` | V excerpt |
| `read`: Akara deciphers (Key to the Cairn Stones bkd) | PASS, hand-in `confirming` (T3) at 11:36:46 once the tap landed with the hero in town; with the tap in the field FAIL before fix 3 (Bug 3) | V `run7-handin-not-in-town.txt` |
| Stony Field stones (CAIN-4, class cycling, ON check) | PASS in runs 6, 8, 9 and 10: 2 to 7 minutes wall incl. potion trips | V excerpt, `run8-stones-stuck-after-key-used.txt` |
| Tristram portal, arrival level 38 | PASS 11:39:38, quest word `0x0018` | V excerpt |
| three Champion Skeleton Archers, cage (CAIN-5b) | PASS: two potion trips back through the town portal into Tristram, then at 11:40:53 `3 hostile units are still near the cage; not opening it` (danger retry), cage opened at 11:40:55: word `0x201a` (bit 1 = free) | V excerpt |
| return to camp | PASS, same second (town portal) | V excerpt |
| report to Akara (record 4 `0x2011` in L2) | The quest word went `0x201a` to `0x2019` (bit 0 done, bit 1 cleared, bit 3 kept) in the same second the hero arrived in camp and opened Akara's potion shop (`npc: interact class 148`), and the bot logged `Quest done: search_for_cain`. The board's `report` chunk was NOT offered or run: the greeting of the routine shop visit completes the quest (the L2 hand-sent talk did the same). The `read` hand-in stays `confirming` | V excerpt, `run10-poll.txt` |

Before the fixes the same chain failed at: tree (runs 1 and 3 to 5, Bug 1), report-in-field (run 7, Bug 3), stones chunk never done after the key was used up (run 8, Bug 2), and the board dropping the cage chunk after the first chicken from Tristram (run 9, Bug 4).

## Bugs found, minimal repro and fix (branch `agent/cainfix2`, must-fail test each, `-race -count=20`)

1. Tree of Inifuss is class -1 in the snapshot (V). Evidence: `tree-class-minus1.txt`: player (222,218), preset (222,222), `ofClass 0`, and an object `c-1/id268435499/(222,222)` (unit id NPCIDBase+43) at the preset spot: the tree is also an `npcs` row (named, class-less); `d2dsrc` filed it as class -1 and `fieldObjects` skips every preset index that `npcs` lists, so the class (30) was never attached. The walk to the preset ends `nav: no path: already there: empty path`, no game tick passes (lockstep) and the step retried 400 times in 90 s for ever (clock frozen, `now` constant); in real time it would idle 12 minutes and park. Fix: `classFromPresets` (internal/d2dsrc/items.go); test `TestNpcsListedObjectGetsItsPresetClass`. Related `treefix3` (hidden-flag theory, U for the sandbox build): the lab build prints hidden 0; the `npcs` listing fits the sandbox symptom just as well (U).
2. Finished chunks were forgotten (V). The scroll chunk is done only while the key is in the pack, the stones chunk only in Tristram or with a portal item, and a chunk "after" another needs it done: after the key was used up the whole chain read open again (history went `["read","scroll"]` to `["read"]`, run 6; stones stuck in progress with `all steps done but the quest does not read as finished`, run 8). Fix: `reached`/`keptDone` in `chunkFact` (internal/bot/questboard.go): an item or auto chunk seen done stays done for this hero and difficulty, and counts for the chunks after it. Tests `TestCainAutoChunksStayDoneAfterTheEvidenceVanishes`, `TestCainStonesChunkDoneAfterTheKeyIsUsedUp`. Limit (U): a restart loses it.
3. A hand-in tap answered "not in town" three times in a second and parked the chunk (V, run 7). `boardServeHandIn` took the request off the queue in the field. Fix: it stays queued until the snapshot is in town; test `TestQuestBoardHandInWaitsForTown`.
4. The free-Cain chunk read done outside Tristram (V, run 9): the cage step is "not needed" while the hero is not in Tristram, which the board read as done evidence, and so was `home` in town; after the first chicken from Tristram the board dropped the chunk and trained. Fix: `Step.Evidence` (internal/task/step.go), cage and home name the quest word (`free`, plus a hero in camp) as their board evidence; test `TestCainFreeChunkIsDoneOnTheQuestWordOnly`. Only the Cain chunk uses it; other chunks with an area precondition may read vacuously done the same way (U, not audited).

## Open items (not fixed here)

1. Survival (V, runs 2 to 5): at the default crowd rule the hero chickens at every arrival in the Underground Passage (26 to 30 hostiles) and the Dark Wood (21), clvl 18, 24 and 32 alike; mana stays 0 (the leveler puts everything into vitality, belt mana runs out, `combat out of mana: backing off`). The lab d2d may spawn denser packs than Game.exe (U). The sandbox runs the default, so a Start tap there can show the same loop on the way to the tree; not a quest-chain bug, owner combat/safety. A normal clvl 32 hero is not a normal hero: do not tune on this.
2. The board's `report` chunk was not exercised: the quest finished through the shop visit first. Whether `report` ever runs when the hero walks straight to Akara is U; the `read` hand-in staying `confirming` after the quest word is done is cosmetic (T3, never promoted).
3. Chunk memory is per process (fix 2): after a bot restart in the middle of the chain the evidence rules apply again. The durable evidence would be the record bits (0x0008 bark, 0x0010 Tristram), but `task.QuestState` carries only the stage.
4. Cage fight at the default rule: not run (knob 0 in run 10).

## What the user should expect when tapping Start on the sandbox

If the hero does not chicken on the way (Open item 1): tree, scroll, Akara, stones and the Tristram portal work; the cage fight happens with the potion trips through the town portal; the quest can finish at the first visit to Akara, before the board's report chunk. These need the merged `cainfix2` build; the tree fix is the only one proven on a slot against the sandbox's symptom (U that the sandbox build lists the tree through `npcs` as the lab build does).
