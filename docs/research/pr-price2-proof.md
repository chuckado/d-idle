# Proof: vendor prices follow-on, two claims (fix/vendor-prices-followon, stacked on #152)

> status: current (agent DECOMPPRICE2, Opus, 2026-10-31; decomp-queue items 10 and 11). Applies docs/process/upstream-pr-standard.md. Local upstream branch `fix/vendor-prices-followon` in the worker host's PR checkout (worktree `price3`), built on #152's head `8a45cfc9` (fix/vendor-prices, not modified); nothing pushed or posted. Base trace: [price-trace.md](price-trace.md); #152's proof: [pr-price-proof.md](pr-price-proof.md). Body: [pr-price2-draft.md](pr-price2-draft.md); one page: [pr-price2-evidence-pack.md](pr-price2-evidence-pack.md).
>
> status: #152 MERGED 2026-11-01: rebase onto upstream main is clean (merge-tree, docs/upstreamsync.md). The follow-on itself is not posted.

Tags: V = verified (disassembly read with `disasm.py` and game.exe's own function run in the emulator), D = read from disassembly only, S = upstream doc, I = inferred, U = unverified. Addresses are game.exe 1.14d.

## Commits (local, worker host; on top of #152's eight)

| Commit | What | Files |
|---|---|---|
| `c3bab73e` | tooling: `tools/emu/price_edge.py` (edge-case generator), `tools/emu/price_claims.py` (per-claim counts), README row | 3 |
| `e297025a` | D9 npc.txt mults divide by 1024 first above 0xffff | rules.hpp, test_rules.cpp, re/store.md |
| `312272a7` | D10 no repair without ItemTypes Repair or durability | rules.hpp, gamedata_load.cpp, test_rules.cpp, re/store.md |

Author and committer: the user's GitHub identity, the numeric-ID noreply form copied from `fix/waypoint-hostile-window` (`git log -1 --format='%an <%ae> | %cn <%ce>'`). Body `Agent: Claude Code`, no trailer. The whole stack against upstream/main (72132b99): 11 commits, 11 files.

## D9: the npc.txt multipliers divide by 1024 first above 0xffff (V)

`FUN_0062efb0` after the quarters and the D7 zero (`0x62fa90`): `FUN_00656900(hcIdx)` gives the npc.txt record (0x4c bytes: +4 "sell mult", +8 "buy mult", +0xc "rep mult", +0x10..0x18 quest flags A..C, +0x1c / +0x28 / +0x34 quest sell / buy / rep mults). Buy base B in `[ebp-0x10]`, sell S in `edi`, repair R in `[ebp-8]`.

| Step | Address | Code | Rule |
|---|---|---|---|
| buy | `0x62faac` | `cmp edx, 0xffff; jle` / `test ecx ([esi+4]), ecx; je` | B > 0xffff (signed) and mult != 0: `(B / 1024) * mult`, else `mult * B / 1024` |
| sell | `0x62fae5` | `cmp edi, 0xffff; jle` / `test ecx, ecx` (ecx still the BUY row's +4) / `imul [esi+8]` | S > 0xffff and "sell mult" != 0: `(S / 1024) * buy mult`, else `buy mult * S / 1024` |
| repair | `0x62fb1d` | `cmp ecx, 0xffff; jle` / `test [esi+0xc]` | same with rep mult |
| quests | `0x62fb70`..`0x62fc53` | 3 x: `FUN_0065c310(quests, flag, 0 or 1) == 1`, then the same three tests, each on its own mult (+0x1c / +0x28 / +0x34 + 4i) | same rule per quest mult, in order A, B, C |

Rounding (all six sites): `cdq; and edx, 0x3ff; add eax, edx; sar eax, 0xa` = signed division toward 0. The threshold is `> 0xffff`: 0xffff itself multiplies first, 0x10000 divides first. The multiply is a 32-bit `imul` (wraps above 2^31: see limits). A mult of 0 gives 0 on either path, so the zero test changes nothing except on the sale, which tests the "sell mult" column (+4) instead of its own: every npc.txt 1.14d row has a non-zero "sell mult" (V: the 17 vendors' rows read through the emulator, 960..2048), so it never shows.

d2d before (#152 head, `rules.hpp` item_price / repair_cost): `price * mult / 1024` in 64 bits for the npc and quest mults, with a `// ponytail:` naming this gap. After: `rules::npc_mult(price, mult)` = `price > 0xffff ? price / 1024 * mult : price * mult / 1024`, used for the npc mult and the three quest mults of buy, sell and repair. The quality extras already divided first (`base < 0x10000 ? ...`), untouched.

## D10: the repair gate's ItemTypes Repair and durability check (V)

`FUN_0062efb0` mode 3 calls `FUN_0062e660(item)` at `0x62efee`; 0 returns price 0.

| Step | Address | Rule |
|---|---|---|
| identified | `0x62e679` | item flag 0x10, else 0 |
| ethereal | `0x62e689` | item flag 0x400000 -> 0 |
| charges | `0x62e698` `FUN_0062a840(item, &full)` | the item has stat 204 entries and one has `(v & 0xff) < (v >> 8)` -> 1 (gate passes whatever the type) |
| ItemTypes Repair | `0x62e6b2`..`0x62e6c6` | `FUN_0062b400` (own type) -> `FUN_00555c90` row, byte +8 (Repair) == 0 -> 0 |
| restock | `0x62e6c8` `FUN_0062e5d0` (esi = item) | ItemTypes Repair, Throwable (`FUN_0062ba80`, +0x10), stackable (`FUN_006289f0`, items +0x132), not ethereal -> 1 |
| durability | `0x62e6d2` `FUN_00629930` | items +0x113 (nodurability) == 0, +0x112 (durability column, unsigned byte) > 0, `FUN_00625e00` (max durability, total) != 0, stat 152 total <= 0 -> 1, else 0 |

Then the price: mode 3 skips the durability term for a quiver (`FUN_0062e740`), a throwable or an item `FUN_00629930` refuses (R stays the full base, `0x62f9c2`..`0x62f9ef`); with durability, `0x62fa09`: max 0 or dur >= max -> R = 0; replenish (stat 252) -> `(max - 1) * R / max` when dur < max - 1, else 0; otherwise `(max - dur) * R / max`.

d2d before: `repair_cost` refused ethereal and unidentified items (D4 in #152), then priced any item whose SAVED max durability is above 0 and whose durability is below max. It never read ItemTypes Repair (no field) nor nodurability / the durability column. Real items this changes (1.14d data, V: weapons.txt / itemtypes.txt read through the emulator): bows (`bow`: Repair 0, nodurability 1: 30 codes), crossbows (`xbow`: Repair 1, nodurability 1: 12 codes), throwing potions (`tpot`: Repair 0, nodurability 1, a restocking stack in d2d's `restocks`). d2d's `generate_item` gives these their durability column as max and a random current below it, so d2d priced a repair for them where game.exe asks nothing (price 0, and the store refuses). After: a new `ItemType::repair` (ItemTypes "Repair", loaded in `gamedata_load.cpp`), and `repair_cost` returns the charges-only path unless the type has Repair and the item restocks or is durable (`!no_durability && durability column > 0 && max != 0 && !indestructible`).

## The differential (V)

game.exe side: `price_oracle.py game` (#152's oracle: `FUN_0062efb0` and its callees run natively on the loaded tables; hooks: the stat getters answer from the case, quest flags 0). d2d side: `test_game --price` of each REAL build (#152's harness). The formula `price_model.py` equals game.exe on every case of both sets (`model 120000/120000`; #152's set 100,000/100,000), so its switches `bigmul` (D9) and `nodur` (D10) name the cases each claim is about: a column counts for a claim when switching that term to d2d's old behaviour moves game.exe's price.

Builds (worker host, gcc, Release, `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`): upstream = `1f8f739a` (#152's tooling commit: upstream/main's rules with the harness), #152 = `8a45cfc9`, D9 = `e297025a`, HEAD = `312272a7`.

**Edge set** (`price_edge.py 120000 11`: 40 % armors with a defence that puts the base around 0xffff / 0x10000 / x4 / x16 or anywhere up to 2^20, 30 % items without ItemTypes Repair / with nodurability / without a durability column, durability 0, 1, 2, max - 1, max, 255, 30 % #152's mix with edge durability and stack sizes; 17 vendors and 3 difficulties cycled). Mismatches buy / sell / repair:

| Build | D9 cases (46,730 / 545 / 14,356) | D10 cases (0 / 2,593 / 21,861) | clean (114,900 / 114,900 / 101,327) | all 120,000 |
|---|---|---|---|---|
| upstream | 46,730 / 545 / 14,356 | 0 / 944 / 15,519 | 64,322 / 50,062 / 50,444 | 67,969 / 53,901 / 66,689 |
| #152 head | 46,730 / 545 / 14,356 | 0 / 825 / 15,519 | 46,730 / 545 / 28,473 | 50,287 / 4,221 / 44,314 |
| + D9 | **0 / 0 / 0** | 0 / 825 / 15,519 | 0 / 0 / 14,117 | 3,557 / 3,676 / 29,958 |
| + D10 (HEAD) | 0 / 0 / 0 | 0 / 825 / **0** | **0 / 0 / 0** | 3,557 / 3,676 / 14,439 |

D9 fixed 46,730 / 545 / 14,356 columns, newly wrong 0. D10 changed only repair columns: 15,519 changed, all fixed, newly wrong 0. Clean = not a stack or quiver and the whole-item repair term (D5) does not move it. The 825 sell columns left in the D10 set are stacks (their sale uses the repair gate inside the stack formula, D6) and sit outside clean. What is left in "all" is D6 (stacks) and D5 (whole item: 1 in game.exe, d2d's 0 sentinel).

**#152's own set** (`price_oracle.py gen 100000 7`; the #152 build's output is byte-identical to #152's recorded `dH.tsv`):

| Build | D9 cases (9,232 / 154 / 1,608) | D10 cases (0 / 809 / 6,481) | clean (93,262 / 93,262 / 82,292) |
|---|---|---|---|
| upstream | 9,232 / 154 / 1,608 | 0 / 762 / 5,163 | 33,316 / 32,629 / 27,195 |
| #152 head | 9,232 / 154 / 1,608 | 0 / 759 / 5,163 | 9,232 / 154 / 5,947 |
| HEAD | **0 / 0 / 0** | 0 / 759 / **0** | **0 / 0 / 0** |

Per claim on real builds: D9 buy 46,730 + 9,232, sell 545 + 154, repair 14,356 + 1,608 mismatches before, 0 after; D10 repair 15,519 + 5,163 before, 0 after (both sets together 220,000 cases).

Confirmation of the queue's numbers: D9 buy 9,232 and repair 1,608 are exact; the sell figure 138 was the single-term bucket of `price_terms.py`: 154 sell columns move with the term, all wrong before (16 also carry another term). D10's 4,238 was also the single-term bucket: 5,163 repair columns move with the gate and all differ on #152's head. Both claims hold as stated; only those two counts are corrected.

## Their checks (worker host, the HEAD build)

- `test_rules` ok (new: `big` 100000 buys for 99,328 and 49,664 after quest A, repairs (worn to 0 of 40) for 49,664 = 100000 / 1024 x 512 (rep 512 in the test table), not 50,000; a worn bow and a worn crossbow repair for 0), `test_combat` ok, `test_protocol` ok, `test_game` rc 0 (with the game data dir).
- clang-tidy with the repo's `.clang-tidy` (identifier length 3, misc-include-cleaner) over the touched TUs (`gamedata_load.cpp`, `test_rules.cpp`, `test_game.cpp`; `rules.hpp` through them): 0 findings in our files or lines; the only findings are 4 in `components/drlg/outdoor.hpp:1052` and `:1090` (structured bindings `x`, `y`), upstream's, untouched by this branch (the local clang-tidy may be newer than CI's).
- Shadow sweep (gcc `-Wshadow -Wunused -Wsign-compare -fsyntax-only`) on `test_rules.cpp` (includes `rules.hpp`) and `gamedata_load.cpp`: 0 warnings. No new local reuses an upstream name in an enclosing scope (`dear`, `worn_bow`, `worn_crossbow`, `info`, `kind`, `mends`, `durable` checked by hand; no range-for added).
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick`: 0 FAIL, 1 WARN (build skipped by --quick; built and tested above).

## Callers (what each now does differently)

`npc_mult` (D9): `item_price` (store buy, sell, the hover prices, the town buy in `apps/d2d/town/net.cpp`) and `repair_cost` (store repair, Repair All's total, the repair hover): a price above 0xffff before a mult now loses its low 10 bits first, as in game.exe (a few gold on dear items, more after several quest mults). `repair_cost` (D10): bows, crossbows and throwing potions (and any item whose items row has nodurability or no durability column) are no longer repaired, nor counted in Repair All, unless charges are missing. Nothing else calls them; `restocks` and the repair itself are untouched.

## Limits

- No stat list in the oracle: charges, staff mods, magic stat costs and sockets are not in either differential. With charges missing, game.exe passes the gate for any type and then adds the FULL base x rep mult when the item has no durability (indestructible, bows): d2d charges only the charges (a `// ponytail:` in the commit); not claimed.
- Quest mults: the rule is read at four sites (D) and d2d's quest path now uses the same helper, but the oracle's quest flags are 0, so the quest part is covered by `test_rules` (49,664) and the disassembly, not by the differential.
- 32-bit wrap: game.exe multiplies in 32 bits (`imul`), d2d in 64. The generator keeps every base below 2^20 and cost x defence below 2^31 so neither side wraps; above that the two differ (unreachable with 1.14d items: no real price comes near).
- Stat 152 (indestructible): game.exe needs the total > 0, d2d any non-zero prop (from #152's verifier); stat 252 replenish repair `(max - 1) / max`: not modelled by d2d, not in the oracle.
- Whole item repair (D5) and stacks (D6) remain as in #152.
- Nothing verified live on the original game.

## Verifier

Not run yet: needs an independent Opus agent (read-only). Brief: break D9 and D10 on `fix/vendor-prices-followon`. Commands (from `tools/emu` of a checkout of the branch, `D2_MPQ_DIR` set, the build in `build/`):

```sh
cmake --build build --target test_game test_rules test_combat test_protocol && build/tests/test_rules
uv run python price_edge.py 120000 11 > cases.tsv && uv run python price_oracle.py game cases.tsv > g.txt
uv run python price_model.py cases.tsv g.txt | tail -1                                   # model 120000/120000
../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d.txt
uv run python price_claims.py cases.tsv g.txt d.txt                                      # bigmul 0/0/0, nodur repair 0, clean 0/0/0
# before: build 8a45cfc9 (#152) and 1f8f739a (upstream rules) in their own worktrees, same --price run -> bigmul 46730 / 545 / 14356, nodur repair 15519
```

(The game side takes about 16 ms a case: split `cases.tsv` and run chunks in parallel.)
