# Cain tree stall: record 0x0018 means the scroll is long gone (cainscroll, 2026-11-21)

> status: current. Tags: V verified (source named), S secondary (fork source read, not run), U unverified. Setup of every scenario run here is CHEATED (quest word, level, waypoints written as SETUP): mechanics only, never balance or play evidence.

## Question

Sandbox hero Doofer (Paladin clvl 18, release 0.0.261121135709, deploy 23): the board walks to the Tree of Inifuss, clicks it, no scroll drops within 6 polls, three tries, chunk `search_for_cain/scroll` parked. Quest words in the log `memory quest words`: `search_for_cain=0x0018`. Earlier the stones chunk was parked after "Have Akara read the scroll" before any tree click.

## Findings

1. **Where the scroll is: nowhere, it was used.** Record 4 (Search for Cain) = `0x0018` is bit 3 (0x0008, "under way", the hero left town) plus bit 4 (0x0010, "the stones done": the Tristram portal was opened). V (sandbox log `memory quest words`, 13:51 and 14:12 UTC, unchanged all run; fork comment on `CainQuest`, S: "bits: 0 done, 1 reward due, 2 Akara told, 3 under way, 4 the stones done"; internal/quest/state.go Interpret: bits 2,3,4 = Started). To light the stones the hero must have taken the Scroll of Inifuss (`bks`), had Akara decipher it into the key (`bkd`) and used the key at the fifth stone. The pack in the sandbox is `yps x2, tbk x1` (V, `inventory contents` line 14:23 UTC), the stash holds jewels, charms, gems, rings (no `bks`, no `bkd`). Cain himself is not freed (no bit 0 or 1). So the hero reached Tristram on an earlier deploy and left; the key is used up (S: the fork's fifth stone takes the only `bkd`).
2. **Why the tree gives nothing (H3, S).** On join the fork's `CainQuest` rebuilds its state from the record: bit 4 set gives `state = 5`, `rejoined = deciphered = stones_open = tree_used = true` (fork `components/rules/quests.hpp`, `CainQuest` join block, read on the worker host). `tree()` returns false when `tree_used` (also with bit 0 or 1, or a scroll in the pack). game.exe is the same by decomp (docs/research/decomp40-cairn-portal.md Q3: one drop per game, none for a player with bits 0 or 1 or holding `bks`/`bkd`; the join rewind with bit 4 is S). So clicking the tree can never succeed for this hero: waiting longer (H2) does not help; the six polls were not too few. The click animation (H2) is a non-issue here (U: no probe timed a real drop; the L3 lab run's drop came from a fresh hero).
3. **Why the board disagreed (the real bug).** The `scroll` chunk's steps read done only from the pack (`bks` or `bkd`) or from Cain being free (`task/act1.go cainNodes`); `World.Quests` carried only the folded stage, never the word, so bit 4 was invisible to the steps. After the stones were done the board therefore re-ran the whole chain from the waypoint, the same blindness that made the stones chunk park earlier (needs `bkd`, key gone) and that the L3 doc noted ("the durable evidence would be the record bits (0x0008 bark, 0x0010 Tristram), but `task.QuestState` carries only the stage": doc search-for-cain-l3).
4. **H1 (the loot rules sold or vaulted it): not supported, but a real gap.** No log line of the sandbox run sells, drops or vaults `bks`/`bkd`/`hdm` (container log from 13:51 UTC only, the earlier 21:39 EST run is not in it, so U for the time before). The loot engine protects every base with the items.txt `quest` column (`lootrules/eval.go` invariant:quest), but the town's own never-sell list (`town/stashfull.go neverSell`) had the Act 2 and Act 3 to 5 quest items and not `bks`, `bkd`, `hdm` or `leg`, and it ignored the column. If the data build's `quest` column is 0 for them (U: the sandbox Misc.txt was not read), the stash-full sell or the 196-item sale could have taken them. Both are closed below.
5. **Which bits mean what for the board (S fork, V probes):** bit 2 Akara told, bit 3 under way (left town), bit 4 stones done = scroll taken, read, key used, Tristram portal exists (V probes l1 to l3: bit 4 set on arrival in Tristram), bit 1 cage opened / reward due, bit 0 done. Only bit 4 proves the scroll steps over; bit 3 does not (the hero merely left town: the L3 first scroll still has to be taken).

## Fix (this branch)

- `task.QuestState.Raw` carries the record word (`cmd/d2bot/task.go` Read sets it for every started or done quest), `World.QuestWord(q)` reads it. `cainNodes`: `passed = free or word bit 4`; the waypoint, travel, tree, read, to-stony, stones and to-tristram steps read done from it, so the board offers `free_cain` (the cage step already re-enters through the Stony Field portal, V proving ground cain.chain). A fresh hero (word 0 or only bits 2/3) behaves as before.
- Quest items are data: `data/questitems/codes.json` (Acts 1 and 2: `bks bkd hdm leg`, `box msf vip hst tr1 ass`) plus `loot.QuestItemCodes`; `loot.QuestCode` backs the items.txt column. The town's sell, stash-full drop and shadow facts (`NeverSellReason`, `sellProtected`, `neverSellUnit`) and the loot rules item (`Quest: b.Quest || ... || QuestCode`) use it.
- Lost scroll recovery (U, not coded): a hero with bit 4 clear that lost `bks` has the tree marked used for this game only (`tree_used` is per game, not in the record), so leaving and re-entering the game gives a fresh tree (S: fork rejoin rule needs bit 4 for `tree_used`). A hero with bit 4 set has no recovery: Cain is reached through the Stony Field portal, which needs no key (S: `stones_open`). The board therefore offers `free_cain` directly. A bot restart as a recovery step is a decision for the owner of the board (not built).

## Scenario verdicts (builder lab slot on the second worker, Paladin 11)

V, second worker (Orion) lab slots, lab d2d devp16, bot of this branch vs origin/main plus the new scenarios, `tools/provingground/run.sh -parallel 1`:

| scenario | unfixed | fixed |
|---|---|---|
| cain.scroll-already-taken Paladin 11 | RED (chunk started, walked to the Dark Wood) | GREEN |
| cain.scroll-already-taken Sorceress 3 | not run | GREEN |
| cain.scroll-already-taken.key Paladin 11 | not run | GREEN |
| cain.scroll Paladin 11 | not run | GREEN |
| cain.scroll Sorceress 3 | RED (died once, still exploring for the tree) | RED, same (pre-existing on this host, not a regression) |
| cain.chain Paladin 11 | GREEN 80 virtual min | GREEN 80 virtual min (a first fixed run under host load 25 was cut by the wall limit: BLOCKED, not judged) |

The first version of the scenario asserted only "no tree click" and was GREEN on the unfixed tree because the hero never found the tree in 20 virtual minutes; it now asserts the chunk never starts. Every setup is CHEATED (quest word written), mechanics only.

## Not covered

- Whether game.exe's own join logic agrees with the fork for bit 4 (S only).
- The sandbox Misc.txt `quest` column for `bks`/`bkd` (U).
