# Search for Cain, drill L3: the real bot chain on a d2d lab slot (2026-11-19)

> status: result of drill CAIN-L3 (task `cainl3`), follow-up to `search-for-cain-l2-2026-11-18.md`. NOT Game.exe, no Wine instance, nothing on the sandbox: one d2d lab slot on the builder, lab binary with `--lab-verbs` (prep only), throwaway Sorceress `Cainy-11` (seed 11). The HERO IS CHEATED/IMPOSSIBLE (SETUP writes: clvl 24 to 32 with its experience, 8000 to 20000 gold, `tools/ab/prep`, ledger `prep-ledger.jsonl`): mechanics only, never evidence of balance. Tags: V verified live (fixture or log line below), S secondary (source), U unverified. Fixtures: `docs/probes/cainl3/testdata/`. Harness: `docs/probes/cainl3/run.sh` (prep, then the real `d2bot --backend d2d --managed --ladder --watch --goals campaign:normal_act1`, like the sandbox) and `poll.py` (plays the player: `GET /api/quests`, `POST .../start` and `.../handin` on the watch routes, never anything else).

## Method

`run.sh SEED CLVL WALLMIN SLOT`: `prep` makes the hero (SETUP, ledgered) and saves, the bot loads it with a d2d WITHOUT `--lab-verbs`, so no gameplay write can happen after the start (`writes after start: 0` by construction; refused-verb count is 0 in every log). `poll.py` taps Start for the offered chunk of `search_for_cain`, taps Hand in when the board says awaiting, and logs every change of the board view. The bot itself does everything else (route, fights, potions, town trips, tree, Akara, stones, portal). Wall times are the bot log's clock; the lockstep d2d runs about 9 times faster than wall time.

Run knob (declared): from run 6 on `combat.crowdAbandonN` is 0 (override file in the instance state dir, `TUNE` of `run.sh`). Reason: at the default (15) the bot chickens (Save and Exit) at every arrival in the Underground Passage and the Dark Wood of this seed (21 to 30 hostiles, "stale pack"), over twenty times in a row, and never reaches the tree (`testdata/run2-up1-chicken-loop.txt`, `run3-up1-crowd.txt`, a clvl 18 and a clvl 24 hero). That is a survival-design finding, not a quest finding; see Open items.

## Timeline (what each step did, unattended)

| step | result | evidence |
|---|---|---|
| accept `search_for_cain` on the board (POST start, chunk `scroll`) | PASS in every run: `ok:true state:accepted`, the board shows `running take_scroll` within 5 s, the bot's log line `quest board: accepted search_for_cain/scroll` | V `poll.log` |
| walk to the Dark Wood (Cold Plains, Stony Field, Underground Passage 1, Dark Wood) and activate the waypoint | PASS (about 100 s wall at clvl 32; at clvl 18 and 24 only with the crowd rule off) | V bot log |
| Dark Wood: find the tree (CAIN-7 room walk) | FAIL, then PASS after fix 1. The hero walked straight to the tree room (222,218), 4 subtiles from the preset (222,222), and stood there for good: `still exploring for object class 30 (7 s of 720 s)` repeated 400 times in 90 s with the game clock frozen (`now` constant, lockstep). Root cause below | V `testdata/run1-livelock-tree.txt`, `tree-class-minus1.txt` |
| click the tree, pick up the Scroll of Inifuss (bks, CAIN-3 polling) | PASS after fix 1 (about 20 s after the walk): `Found Scroll of Inifuss`, quest word `search_for_cain=0x0008` | V log |
| `read` chunk, hand-in to Akara (deciphered Key to the Cairn Stones, bkd) | PASS when the tap lands while the hero is in town (`npc: talk Akara pick bark`, outcome `confirming`, T3); FAIL in run 7: tap taken in the Dark Wood, answered "not in town" three times in one second, chunk parked. Fix 3 | V log |
| Stony Field: stones by class cycling with ON verification (CAIN-4) | PASS in run 6: the stones chunk ran in about 7 min wall (the first 230 s virtual of exploring for class 17, then the five stones and the Tristram portal) | V log |
| Tristram portal, arrival in level 38 | PASS: `Entered Tristram`, quest word `0x0018` (bit 4 added by the arrival, bit 3 from the scroll) | V log |
| the three Champion Skeleton Archers, the cage, quest record 4 bit 1 (CAIN-5b) | NOT REACHED. Right after the arrival the board lost the progress of the first two chunks (Bug 2), and the hero went back to town and then to the Dark Wood again | see below |
| town portal back, report to Akara (record 4 = 0x2011) | NOT REACHED | |

## Bugs found, with the minimal repro and the fix

All three fixes are on branch `agent/cainfix2` with a must-fail test each (per-package `go test -race -count=20 -p 1`).

### Bug 1 (fixed, V): the Tree of Inifuss is class -1 in the snapshot, so the search for class 30 never matches

- Evidence: `testdata/tree-class-minus1.txt` (the diagnostic line of this drill): player (222,218), preset (222,222), `ofClass 0`, and in the objects around the hero `c-1/id268435499/(222,222)`: unit id `NPCIDBase + 43`, class -1, at the preset spot. The tree is also an `npcs` row (named, class-less). `d2dsrc` files such a row as an object of class -1 (`npc()`), and `fieldObjects` then skips every preset whose index `npcs` lists (`seen`), so the preset class (30) is never attached.
- Why the livelock: the walk to the preset ends with `nav: no path: already there: empty path`, no game tick passes (lockstep: only a Sleep or a walk steps the game) and the task step retries at once, 400 times in 90 s, for ever. In real time the same loop would run for 12 minutes and then park.
- Minimal repro: `internal/d2dsrc` `TestNpcsListedObjectGetsItsPresetClass` (an `npcs` row plus a preset row with the same index and spot).
- Fix: `classFromPresets` (internal/d2dsrc/items.go) gives a class -1 `npcs` object the class of the preset with the same index at the same spot (within 2 subtiles). Class-free (read from the preset table, no build data). Live: with the fix the tree was found and the scroll taken 20 s after the walk, in every run that got to the tree.
- Relation to `treefix3` (merged before this run, U theory "hidden flag in the sandbox build"): the lab build prints hidden 0 and the tree row was never skipped for that reason; the cause here is the `npcs` listing. The sandbox symptom ("object class 30 not found", hero standing on the tree) fits this cause just as well; U for the sandbox build.

### Bug 2 (fixed, V): the board forgets finished chunks once the evidence vanishes

- Evidence: run 6, 10:31 (`history` of `GET /api/quests` went `["read","scroll"]` to `["read"]`, the stones chunk never appeared in it, and the board offered `scroll`/`take_scroll` again): the scroll chunk is done only while the Key to the Cairn Stones (or the bark) is in the pack, the stones chunk only while the hero stands in Tristram or holds a portal item (`opened`, internal/task/act1.go). The key is used up by the portal and the first potion trip from Tristram puts the hero in town: both chunks read open again, the board runs `take_scroll` and walks back to the Dark Wood (a second scroll cannot exist).
- Fix: `internal/bot/questboard.go` `chunkFact`: an item or auto chunk seen done stays done for this hero and difficulty for the rest of the process (`reached`); hand-in chunks already had `handed`. Test: `TestCainAutoChunksStayDoneAfterTheEvidenceVanishes` (the same two worlds, Tristram then town; fails without the fix). Class-free. Limit (U): a restart loses it, then the evidence rules apply again (the quest word 0x0018 would be the durable evidence, but `task.QuestState` carries only the stage, not the bit; a follow-up).

### Bug 3 (fixed, V): a hand-in tap is answered "not in town" three times and parks the chunk

- Evidence: run 7, 10:43:04: `quest board: hand-in failed search_for_cain/read attempt 1 err not in town`, then parked after 3 in the same second. `docs/questboard.md` says the tap waits for the next town visit; `boardServeHandIn` took the request off the queue wherever the hero stood.
- Fix: the request stays in the queue until the snapshot says town (`internal/bot/questboard.go`). Test `TestQuestBoardHandInWaitsForTown` (fails without the fix).

## Open items (not fixed here)

1. Survival at this seed: with the default crowd rule the hero chickens at every arrival in the Underground Passage (26 to 30 hostiles) and the Dark Wood (21), clvl 18, 24 and 32 alike, because mana stays at 0 (the leveler puts everything into vitality; belt mana potions run out; `combat out of mana: backing off`). The lab d2d spawns denser packs than Game.exe is known to (U). The sandbox runs at the default rule: a user tapping Start there may see the same chicken loop on the way to the Dark Wood. Not a quest-chain bug; owner: combat/safety.
2. The cage and everything after it (archers at the arrival, record 4 bit 1, the portal back, the report to Akara) were NOT driven by the bot in this drill (see the timeline). L2 has the hand-sent version (V).
3. `quest board` evidence of the first two chunks is not durable across a bot restart (Bug 2 limit).

## What the user should expect when tapping Start on the sandbox

Tree and scroll should now work (U for the sandbox build, V on the lab slot); the report to Akara after the scroll should wait for the town visit; after the stones and the portal the board should no longer fall back to the scroll. The cage fight and the final report have not been driven by the bot yet: expect the first real test there. If the hero keeps leaving the game in the Underground Passage, that is Open item 1.
