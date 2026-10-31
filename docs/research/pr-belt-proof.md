# Proof: a picked-up scroll goes into a belt column of the same scroll (trace/belt)

> status: current (agent DECOMPBELT, 2026-10-31). Applies docs/process/upstream-pr-standard.md. Branch `trace/belt` in `~/d2d-lab/pr/belt` on the worker host, two commits on upstream/main f8aab0e8: `6c2c7347` (tooling: `tools/emu/belt_pickup.py`, `tests/test_game.cpp --belt-pickup`, a `tools/emu/README.md` row) and `41bd5a88` (the claim: `components/rules/rules.hpp` pick_up, a `tests/test_rules.cpp` case, the `docs/research/re/combat.md` step 5 note). Local only; nothing pushed or posted. Trace: docs/research/belt-pickup-trace.md.

## Claim (one, one commit)

On a pickup (C->S 0x16 action 0, `FUN_00563560`), game.exe's belt step lets a Scroll of Identify (`isc`) or Town Portal (`tsc`) into the belt when a column's bottom item is the same scroll, at that column's lowest free box within the belt's boxes. `FUN_00628ba0` returns 1 for autobelt items (Items +0x131) and otherwise calls `FUN_0063c560` and returns 0 only when it gave -1 (0x628bfd `cmp eax, -1; jne`); `FUN_0063c560` returns only 0 or 1 (0x63c56b / 0x63c5e1 / 0x63c5f5 `xor eax, eax`, 0x63c5ea `mov eax, 1`), so its "not isc / tsc" test never refuses. `FUN_0063c600` then picks the box with no isc/tsc test. d2d's `rules::pick_up` (rules.hpp:1936) skipped the belt for isc/tsc unless autobelt (neither is), so they went to the inventory. Fix: drop that condition.

## The chain, tagged

| Link | What it shows | Tag |
|---|---|---|
| `FUN_00480930` case 4 -> 0x16 | the client's pickup click, action = inventory panel open | V (read) |
| `FUN_00548b00` 0x548c8e | action 0 -> `FUN_00563560` | V (read) |
| `FUN_00563560` 0x563703..0x56372f | `FUN_0062bad0`, then `FUN_00628ba0`, then `FUN_0063c790`; any 0 -> `FUN_005600a0` (inventory) | V (read) |
| `FUN_00560020` before it | a scroll goes into a tome with room first (unchanged; d2d does the same) | V (call), S (combat.md) |
| `FUN_0062bad0` | own ItemTypes row, Beltable +0x19 (scro: 1) | V (read + oracle), D |
| `FUN_00628ba0` / `FUN_0063c560` | as in the claim | V (read + oracle) |
| `FUN_0063c600` / `FUN_00628a40` | 1x1, belt boxes by belts.txt, bottom match by class or group, lowest free box; autobelt fallback (isc/tsc: 0) | V (read + oracle) |
| `FUN_0063c790` -> `FUN_0063afd0` | the box is used as is | V (read) |

No U link. The oracle hooks NOTHING: fake units {type 4, class} (every function on the path reads only +0 and +4, V by the listings) and a fake inventory (magic 0x1020304, grids at +0x14, count +0x18; grid 0 body 13x1 with the belt at slot 8, grid 1 belt 16x1: the descriptors at 0x74479c / 0x7447b4, D).

## The differential (V)

`belt_pickup.py gen 100000 1`: 290,000 grid cases (8 belts {none, lbl, vbl, mbl, tbl, hbl, zlb, uhc} x bottom row over {-, hp1, mp1, rvs, isc}^4 x {bottom only, columns full to the belt's rows} x 29 picks: every Beltable 1x1 code incl. isc, tsc, 0sc, and key, rin, jew, amu, aqv, gcv) + 100,000 random stacked belts (any kinds, 10 % of columns above the belt's rows). `game` runs 62bad0, 628ba0, 63c600; `test_game --belt-pickup` runs `rules::pick_up` with the same belt and an empty 10x4 inventory.

| Build | Lines differ (of 390,000) | by pick |
|---|---|---|
| 6c2c7347 (before) | 3,111 | isc 2,859, tsc 252; every other code 0 |
| 41bd5a88 (after) | 0 (md5 984ef9a4... equal) | |

Their checks on 41bd5a88: `test_rules` ok (with the new case: an isc onto a belt whose column 3 bottom is isc -> box 7), `test_combat` rc 0, `test_game` rc 0 (27 OK, MPQs + LODPatch_114d.exe), `pr_guard.py upstream/main` 0 FAIL 0 WARN, `presubmit.py --quick` 0 FAIL 2 WARN (ruff absent; build by hand).

## Callers (nothing else moved)

`rules::pick_up` has one game caller, `Loot::take` (loot.hpp:230: the player's ground pickup); the other `pick_up` overloads (rules.hpp:963 cursor, c2s.hpp packet builder) are different functions. Changed behaviour: only a picked isc/tsc with no tome room and a same-scroll belt column. Potions, `0sc`, non-Beltable items: 0 of 386,889 lines changed.

## Limits

- The step order around the belt (tome first, body step absent in d2d: upstream ponytail) is not in the claim; the cases have no tome and no inventory items.
- `FUN_0055d710` (auto-equip) is not run; no scroll fits a body slot.
- Client-side gaps found on the way (not in this PR): pickup with the inventory open goes to the cursor in game.exe (action 1), d2d always auto-places; shift-click to the belt (0x63 -> `FUN_0054d520` -> 63c600) and ctrl-click sell (0x33) have no d2d path. Not oracle-run.

## Verifier

Independent second agent (Opus, read-only, d2-reviewer), report verbatim:

> VERDICT: the claim holds, with limits. I could not break it. Checked against trace/belt at 41bd5a88; its diff from b80afdce touches only docs/research/re/combat.md. I wrote no files in the repo, made no worktree and left nothing running. My /tmp scratch is deleted.
>
> Attempts (V = read in the disassembly or run):
> 1. **628ba0 / 63c560 (V, disasm.py).** 63c560 has four exits. Three return 0 (`xor eax,eax` at 63c56b, 63c5e1, 63c5f4) and one returns 1 (63c5ea). It never returns -1. The isc/tsc test at 63c59c/63c5a3 only sends it to "return 0". 628ba0 returns 0 only on `cmp eax,-1`, so it returns 1 for every item with a record. In 563560 (5636ff-56371d) the caller tests only for 0. I did not scan the whole binary for other callers of 628ba0.
> 2. **Oracle soundness (V).** Every function on the path reads only +0 and +4 of an item unit: 62bad0, 62b400, 628ba0, 63c560, 63c600, 628a40, 6286c0, 621ed0. 63bde0 and 63add0 read the inventory at +0, +0x14 and +0x18. One catch: 63add0 returns 0 unless the grid's width and height match the templates at 0x74479c / 0x7447b4. The many belt placements the oracle prints show the fake grids do match. 63c600 hard-codes resolution mode 0 in 660cb0, while d2d reads the 800x600 half of belts.txt; 0 mismatches across all 7 belt rows shows numboxes is the same.
> 3. **Fresh seed (V).** `gen 30000 9177` gave 320,000 cases: 0 mismatches, md5 equal.
> 4. **My own adversarial set (V).** 60,000 cases: 15 belt codes including exceptional and elite, plus no belt. Every box filled at random with gaps and above the belt's rows, 55% of filled boxes being isc/tsc/0sc. Picks: 60% scrolls, the rest all 27 Beltable 1x1 codes the game reports, and non-belt items. Result: 0 mismatches, md5 equal.
> 5. **Scroll stacking (V, read).** The stacking step 560020 runs before the belt (56367f, then done) and d2d's stacking branch returns before the belt too. The fix only matters when no tome has room.
> 6. **Callers (V, read).** loot.hpp:230 is the only production caller. Scroll use (world.cpp:2328, 2598) does not care where the scroll sits.
>
> Limits to disclose:
> - The oracle calls 63c600 directly. It does not run 63c790's placement (63afd0), the stacking step or auto-equip.
> - I did not build the BEFORE binary, so the author's 3,111 "before" mismatches are not reproduced.
> - Existing simplifications in d2d, unchanged by this PR: auto-equip is not modelled, and stacks match on code only.

**PROOF: complete** (differential 3,111 -> 0 over 390,000 author cases and 0 over 380,000 verifier cases; chain V/D; checks green; verifier holds). Limits added: 63afd0 placement not run in the oracle; the before count is the author's only.
