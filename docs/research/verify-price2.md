# Independent verifier: vendor prices follow-on (D9, D10)

> status: current (card verifyprs2, 2026-10-31). Verifier section for `fix/vendor-prices-followon` (worker host worktree `price3`, head 312272a7 on #152's head 8a45cfc9); the author's proof is pr-price2-proof.md (card decompprice2). Offline, throwaway worktrees; the author's worktree and branch untouched; nothing pushed or posted.

Verdict: **D9 HOLDS WITH LIMITS, D10 HOLDS WITH LIMITS** (no counter-example on 1.14d data). **NOT READY** to open: one presubmit FAIL to fix, and #152 must merge first (steps at the end).

## Re-derived from the disassembly (before reading the author's oracle)

- Column offsets, from the npc.txt loader's descriptors (0x656650..): record +4 "sell mult", +8 "buy mult", +0xc "rep mult", +0x10/14/18 questflag A..C, +0x1c.. questsellmult, +0x28.. questbuymult, +0x34.. questrepmult, +0x40.. max buy. d2d's loader swaps them into our buy / sell (gamedata_load.cpp:1434), as game.exe uses them.
- D9, FUN_0062efb0 0x62faac..0x62fc53: three prices in parallel, the player's buy at [ebp-0x10] x "sell mult", the player's sale in edi then [ebp-4] x "buy mult", the repair at [ebp-8] then edi x "rep mult". Each step: price > 0xffff (signed `jle`) and the mult tested non-zero -> (price / 1024 toward zero: `cdq; and edx, 0x3ff; add; sar 0xa`) x mult, else mult x price / 1024 toward zero, 32-bit `imul`. The sale's test is on ecx = the BUY's mult (+4, loaded at 0x62fab7 / 0x62fad1), not its own; the repair and every quest step test their own. Quests: 3 iterations, flag non-zero and FUN_0065c310(rec, flag, 0) == 1 or (flag, 1) == 1, then the same rule with the quest mults. Same rule as the author's, including the sale's column quirk (the ponytail at `npc_mult` states it correctly: no 1.14d npc.txt row has a 0 "sell mult", checked: every row is non-zero in all three columns).
- Every other `cmp ..., 0xffff` followed by a signed `/ 1024` in game.exe (18 sites in the whole binary): 0x628df4 (FUN_00628d30, charges: d2d `charges_repair_cost` divides first too), 0x628f72 / 0x62907d / 0x629165 / 0x6291e8 (FUN_00628e70, the stats' own costs: not modelled in d2d, #152 ponytail), 0x62f1ab..0x62f8bc (the quality extras: d2d's `extra` divides first) and the six D9 sites. No other price scaler.
- D10, FUN_0062e660: item unit (type 4) with data, identified (flags 0x10), not ethereal (0x400000); FUN_0062a840 (stat 204 entries, one with value & 0xff < value >> 8, signed) -> 1; else ItemTypes +8 (Repair) of FUN_0062b400's type must be set; then FUN_0062e5d0 (ItemTypes Repair of FUN_00628700's type, Throwable +0x10 via FUN_0062ba80 on the item's own type row, Items +0x132 stackable, not ethereal) or FUN_00629930 (Items +0x113 nodurability 0, +0x112 durability unsigned > 0, FUN_00625e00 = stat 73 != 0, stat 152 total <= 0). Same rule as the author's.
- Other callers of FUN_0062e660: 0x62efee (the repair price: 0 when refused; the claim's site), 0x62fc86 (FUN_0062efb0's price of a partly used stack, after the mults: d2d does not model that path, #152 ponytail "restocking stacks take another path"), 0x62fea7 (FUN_0062fe60, the Repair All sum: it adds FUN_0062efb0 mode 3 per item, so the gate reaches it through the price; d2d's `repair_all_cost` sums `repair_cost`, same effect), 0x5761cb / 0x5781fa (the server's repair handlers; d2d's `store_repair` refuses at cost 0), 0x4b1f96 (client). d2d has one gate (`repair_cost`), used by `store_repair`, `repair_all_cost`, `store_repair_all`.
- d2d sweep: `item_price` and `repair_cost` are the only npc / quest multiplier users (`item_text.hpp:226..236` is a stale comment only); `gamble_price` uses no npc mult.

## Own differential

`tools/emu/verify_price_followon.py` (seed 8181, `gen 100000`) + `verify_price_followon_mode.inc` (`test_rules --verify-price`, pasted into a throwaway worktree's test_rules.cpp). Game side: game.exe's own instructions under unicorn: D9 runs 0x62faac..0x62fc59 with the three locals and a fake npc.txt record (FUN_0065c310 hooked from the case's quest flags); D10 calls FUN_0062e660 with all its helpers native against fake Items and ItemTypes rows (hooked: the stat 204 list, stat 152, stat 73). d2d side: the REAL `rules::item_price` / `rules::repair_cost` of each build on a synthetic table (an item of cost P, normal quality, no armor; for repair max durability 1 and durability 0, so the repair base is P); the d2d functions' own ends (at least 1, the sale under max buy = INT_MAX here) are applied to the game value too.

104,422 cases: D9 56,656 (buy / sell / repair, prices around 0, 1024, 0xffff, 0x10000, 0x103ff / 0x10400, up to 2^31 - 1 and negative; mults 0, 1, 1023..1025, the npc.txt range, 65535, negative; 0 to 3 quest flags on with their own mults, done on save bit 0 or 1), D10 47,766 (all 256 combinations of identified, ethereal, Repair, Throwable, stackable, nodurability, charges missing, stack full with 10 durability edges, plus random items: max durability -100..3000, durability column 0..255, stat 152 -3..3, stacks 1..250, two charged skills).

Builds: before = #152 head 8a45cfc9, D9 = e297025a, head = 312272a7. Mismatches:

| bucket (1.14d-reachable) | cases | before | after D9 | head |
|---|---|---|---|---|
| D9 buy (npc mult only) | 6,548 | 2,898 | 0 | 0 |
| D9 buy with quests | 9,689 | 4,454 | 0 | 0 |
| D9 sell (npc mult only) | 5,983 | 2,563 | 0 | 0 |
| D9 sell with quests | 9,181 | 4,201 | 0 | 0 |
| D9 repair (npc mult only) | 6,295 | 2,770 | 0 | 0 |
| D9 repair with quests | 9,460 | 4,373 | 0 | 0 |
| of these: only a quest step crosses 0xffff (P <= 0xffff) | 13,415 | 1,373 | 0 | 0 |
| D10 repair gate (stat 152 >= 0) | 41,663 | 2,235 | 2,235 | **0** |

D9 total 21,259 -> 0 of 47,156; D10 2,235 -> 0 of 41,663. The quest multipliers (covered in the author's run by a unit test only: its oracle had quest flags 0) are in this differential through the real functions: 28,330 cases with quests on, 1,373 where only a quest step decides, all 0 after.

Off-data buckets (not 1.14d-reachable, reported, not counter-examples):

| bucket | cases | head differs | why |
|---|---|---|---|
| D9 int32 overflow (an exact intermediate leaves int32) | 8,336 | 5,705 | game.exe's 32-bit `imul` wraps, d2d computes in 64 bits; needs a price x mult past 2^31 (a price above 0xffff over about 1.05e9 with mult 2048, or a mult of 32,768 and up) |
| D9 sale with the "sell mult" column 0 (the quirk) | 1,164 | 406 | game.exe tests the buy's column; no 1.14d npc.txt row has a 0 there (ponytail in the patch) |
| D10 stat 152 negative | 3,846 | 67 | game.exe tests stat 152 <= 0, d2d `indestructible()` any non-zero value; no 1.14d source gives a negative item_indesctructible (UNVERIFIED beyond properties.txt's "indestruct", which sets 1) |
| D10 d2d kind 2 != ItemTypes Throwable | 2,257 | 231 | d2d's `restocks()` is stackable && weapons.txt, game.exe's FUN_0062e5d0 is Repair && Throwable && stackable; in 1.14d data the only stackable weapons that are not Repair && Throwable are the throwing potions (tpot: Repair 0), which D10 refuses through Repair on both sides (checked: weapons / armor / misc.txt x itemtypes.txt) |

## Rebase recipe and CI class

- `git rebase --onto upstream/main 8a45cfc9` does NOT apply today: #152 is not merged and upstream main is still 72132b99; the first commit conflicts in tools/emu/README.md, e297025a / 312272a7 in rules.hpp, test_rules.cpp, store.md (they stand on #152's lines). Simulated on today's upstream main with #152 landed: as a merge commit and as a squash, the recipe replays cleanly, 3 commits, 7 files (the author's numbers).
- pr_guard: 0 FAIL 0 WARN against 8a45cfc9, against upstream/main, and for each of the three commits against its parent.
- presubmit --quick with ruff on PATH: **1 FAIL**: `tools/emu/price_claims.py:41: B905 zip() without an explicit strict= parameter` (upstream's RUFF_RULES include B; tools/emu/pyproject.toml requires Python >= 3.12, so the pinned ruff 0.8.6 flags it too). Without ruff on PATH it reports 0 FAIL, 2 WARN (ruff missing, build skipped), which is how the author's run passed. Fix: `zip(m.price(case), normal, strict=True)` (both lists come from the same case).
- test_game (data dir), test_rules, test_combat, test_protocol: rc 0 on 312272a7.
- clang-tidy (the repo's .clang-tidy) and gcc `-Wshadow -Wunused -Wsign-compare -Wextra` over the TUs tools/tidy-pr-select.py picks for the touched files: see the line below.
- One claim per commit (D9 e297025a, D10 312272a7, tooling c3bab73e). Ponytails: `npc_mult`'s (the sale's column) is accurate; repair_cost's new ponytail (a refused item with missing charges still adds its base in game.exe) matches FUN_0062efb0's mode 3 reading in the author's proof, not re-traced here (UNVERIFIED by this verifier).

## Limits

1. The D9 game side runs only the multiplier block (0x62faac..0x62fc59) with the three bases given; how game.exe computes the bases before it is #152's claims.
2. The D10 comparison is on the gate, not the price: the expected d2d result is "gate passes and #152's model has something to repair" (charges missing; a restocking stack short of its max stack; otherwise durability below max). A whole item prices 0 in d2d and 1 in game.exe (#152's ponytail).
3. FUN_0065c310's quest flag reading is taken as d2d's `quest_flag(diff, flag, bit)`; quest flags 1..3 stand for npc.txt's.
4. game.exe's 32-bit wrap above 2^31 and d2d's `int(...)` of a 64-bit price differ (off-data bucket above).
5. Det state hashes not run (needs a d2d slot; not allowed in this card).

## Remaining steps before the coordinator asks for a go

1. Fix B905 in `tools/emu/price_claims.py:41` (`strict=True`) on the author's branch, re-run presubmit with ruff on PATH.
2. After #152 merges: `git rebase --onto upstream/main 8a45cfc9` (verified clean for a merge or a squash of #152 on 72132b99), re-run pr_guard and the upstream tests.
3. Det state hashes on a d2d slot if the standard requires them for a rules-only change.
