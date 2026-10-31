# Proof: vendor prices, five claims (fix/vendor-prices)

> status: current (agent PRICEPR, Opus, 2026-10-31; builds on DECOMPREPAIR's trace). Applies docs/process/upstream-pr-standard.md. Local upstream branch `fix/vendor-prices` in the worker host's PR checkout (a new worktree `price2` off upstream/main d0ea14b7), six commits, 8 files; nothing pushed or posted. Trace and formula: [price-trace.md](price-trace.md). Body: [pr-price-draft.md](pr-price-draft.md); one page: [pr-price-evidence-pack.md](pr-price-evidence-pack.md).

## Commits (local, worker host)

| Commit | What | Files |
|---|---|---|
| `d8dae7b6` | tooling: `tools/emu/price_oracle.py` (game.exe FUN_0062efb0 in the emulator), `price_model.py` (the formula, = game.exe 100,000/100,000), `price_terms.py` (per-term attribution, clean-subset line), `test_game --price`, README row | 5 |
| `fcfcd82f` | D1 an armor is priced by its defence | rules.hpp, test_rules.cpp, re/store.md |
| `3566b42d` | D2 a class item sells for a quarter | same three |
| `95557754` | D3 quality extras only on identified items | same three |
| `02e77e74` | D4 an unidentified item cannot be repaired | same three |
| `1bfff15c` | D7 a broken ethereal item sells for the minimum | same three |

Authored `shatteredeagle`, body `Agent: Claude Code`, no trailer. Each claim commit also carries the `// ponytail:` notes for what its area leaves out (D8 under D3, D9 under D2, D5 and D10 under D4, the purse-cap sale under D7).

## The claims and their chain (game.exe 1.14d)

| # | Claim | Address, what it shows | Tag |
|---|---|---|---|
| D1 | An armor starts at cost x defence / maxac for buy, sell and repair | `0x62f0f2`: `FUN_00629bb0(item, 0x32)` (ItemTypes armo, Equiv walk); `FUN_006253b0(item, 31, 0)` (the item's own defence, base layer) x items +0xe0 / items +0xd0, only if maxac - minac + 1 != 0 and maxac != 0; written to all three bases | V (read + oracle) |
| D2 | A class item's sale price is quartered again, after the ethereal quarter | `0x62f99d`: `FUN_0062c060` (item's own ItemTypes row, byte +0x21 < 7) -> S / 4 | V (read + oracle) |
| D3 | Quality extras only when identified | `0x62f13c`: `FUN_006280a0(item, 0x10)` into edi; `0x62f182 test edi, edi; je 0x62f85e` skips the whole extras block (affixes, set, unique, low quality) straight to the sockets | V (read + oracle) |
| D4 | No repair of an unidentified item | `0x62efee`: mode 3 calls `FUN_0062e660`, 0 -> returns 0; `FUN_0062e660` returns 0 unless item flag 0x10 is set and 0x400000 is not | V (read + oracle) |
| D7 | An ethereal item with durability at 0 sells for 1 | `0x62fa60` (mode 1): `FUN_0062a8d0` (0x400000), `FUN_00629930` (nodurability 0, durability column > 0, max durability != 0, stat 152 <= 0), max durability != 0, stat 72 <= 0 -> `xor edi, edi` (S = 0) before the NPC mults; `0x62fd75` floors the sale at 1 | V (read + oracle) |

No U link. d2d before the patch (`components/rules/rules.hpp` on d0ea14b7): `item_price` started every item at its cost, applied the quality switch whatever the identified flag, quartered only ethereal sales; `repair_cost` refused only ethereal items.

## The differential (V)

Cases: `price_oracle.py gen 100000 7` (every item code with a cost except body parts and tomes, qualities 1-8, identified or not, ethereal, every durability state incl. maxdur 0, stack sizes, 17 vendors, 3 difficulties). game.exe side: `price_oracle.py game` (hooks: the stat getters answer from the case; quest flags 0; no stat list, no sockets). d2d side: `test_game --price` (the real `rules::item_price` / `rules::repair_cost`). The d2d md5 of the upstream run equals DECOMPREPAIR's (954eeb4c...); game 6e5fa976...; the oracle and the generator re-ran equal on a 3,000-case slice after the tooling edits.

Mismatches (buy / sell / repair, of 100,000 each), after each commit:

| Build | All cases | Clean subset (no unclaimed term moves the price) | Fixed by this commit | Newly wrong |
|---|---|---|---|---|
| upstream (tooling) | 39,626 / 38,939 / 42,705 | 24,084 / 32,472 / 21,248 | | |
| + D1 | 19,475 / 26,537 / 28,920 | 3,933 / 20,070 / 7,464 | 20,152 / 12,405 / 13,785 | 1 / 3 / 0 |
| + D2 | 19,475 / 10,648 / 28,920 | 3,933 / 4,189 / 7,464 | 0 / 15,895 / 0 | 0 / 6 / 0 |
| + D3 | 15,527 / 6,675 / 28,920 | 0 / 233 / 7,464 | 3,948 / 3,973 / 0 | 0 |
| + D4 | 15,527 / 6,675 / 20,695 | 0 / 233 / 0 | 0 / 0 / 8,225 | 0 |
| + D7 (HEAD) | 15,527 / 6,406 / 20,695 | **0 / 0 / 0** (of 84,030 / 93,095 / 75,112) | 0 / 269 / 0 | 0 |

Each commit moves only its own column(s): D2 and D7 change no buy or repair line, D4 no buy or sell line. The "newly wrong" cases were coincidental matches before: the four after D1 are unidentified low quality armors (the missing D3 term had cancelled the missing D1 term); the six after D2 are four of the same and two Matriarchal Javelin stacks of 4 (the missing class quarter had cancelled the missing x4 quantity, D6). All are fixed by D3 or sit in the stack set. Single-term buckets of `price_terms.py` before -> after their commit: D1 armor 19,122 / 12,395 / 13,525 -> 0 / 16 / 0 (the 16: druid pelts whose defence / maxac is about 1/4, standing in for the class quarter; gone from the bucket with D2); D2 class 9,762 -> 0; D3 unid 2,939 / 2,520 -> 0; D4 unidrep 4,460 -> 0; D7 ethzero 233 -> 0.

The clean subset: a case's column counts unless the item is stackable or a quiver (D6) or switching `whole` (D5), `bigmul` (D9) or `nodur` (D10) in the model moves its price. `price_model.py` with no switch equals game.exe on all 100,000.

## Residual at HEAD (not claimed; each a `// ponytail:`)

| # | Term | game.exe | d2d | Single-term cases at HEAD |
|---|---|---|---|---|
| D5 | whole repairable item | 1 (Repair All `FUN_0062fe60` skips undamaged items) | 0, the "nothing to repair" sentinel `store_repair` and `repair_all_cost` rely on | repair 10,841 |
| D6 | stack quantity, stack extras / maxstack, restock rounding | B x qty, throw stacks S = max x S - missing x R | priced once | the "several" buckets (buy 6,295, sell 6,216, repair 3,795) and stack 35 / 213 |
| D9 | NPC mult above 0xffff | (base / 1024) x mult | base x mult / 1024 | buy 9,232, sell 138, repair 1,608 |
| D10 | repair gate | ItemTypes record byte +8 and `FUN_00629930` (durability column, nodurability) | the save's max durability alone | repair 4,238 (many are generator-made durability on items with no durability column) |
| D8 | stat costs `FUN_00628e70`, staff mods `FUN_0062edd0` | added (I) | none | not in the oracle (no stat list) |
| - | sale over the purse cap | not traced (U) | clamped at clvl x 10,000 | not in the oracle |

D9 and D10 are new gaps found on the way (separable; candidates, not part of this PR).

## Their checks (worker host, gcc 16, Release, `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`)

- `test_rules` ok (new cases for each claim: armor 750 / 187 and repair 281, class 1000 / 62 and ethereal class 15, unidentified 1000 and low quality 250, unidentified repair 0, broken ethereal 1 vs 62), `test_combat` ok, `test_game` rc 0, `test_protocol` ok.
- Full build rc 0; ctest 53 of 58 pass, 2 skipped (test_exe_tables, test_iso9660); fails test_mpq, smoke_d2d, pad_use_d2d, quests_act1, quests_act4, and upstream d0ea14b7 built the same way fails exactly these five (environment: data dir, audio, `debug equip tsc`; no price code on their paths): unchanged.
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick`: 0 FAIL, 1 WARN (build skipped by --quick; built and tested by hand above). Ruff findings in the oracle fixed (unused import; the signed return now plain arithmetic).

## Callers (what each now does differently)

`item_price`: `store_buy`, the store hover (`apps/d2d/store.cpp`) and the town buy (`apps/d2d/town/net.cpp`): armor at its defence; unidentified stock would lose its extras (store stock is made identified by `make_store_item`). `store_sell` and the sell hover (`apps/d2d/ingame.cpp`): armor scaling, class quarter, identified gate, broken ethereal 1. `repair_cost`: `store_repair`, `repair_all_cost`, `store_repair_all` and the repair hover: armor scaling; unidentified items are no longer repaired nor counted in Repair All. Nothing else calls them.

## Limits

- The oracle's item has no stat list: magic stat costs, staff mods, charges and sockets are not in the differential (d2d's socket and charge terms are untouched by the patch). No quests, no reduced prices in the cases.
- Body parts and tomes are outside the generator; the generator gives random durability to items without a durability column (that feeds D10, not the claims).
- From the verifier: starter items (item flag 0x20000) price 1 in every mode in game.exe (`0x62f00b`), d2d prices them normally (now armor-scaled); repair with replenish (stat 252) is (max - 1) / max of the base; the cube's kUseItem keeps the old defence when the code changes, which now feeds the armor price. Not claimed, not in the oracle (no stat list, no flags beyond identified / ethereal).
- Nothing verified live on the original game; the D3.r drill price (19 for a Short Staff at Charsi) agrees with both sides and is unchanged.

## Independent verifier

Independent Opus agent (read-only, fresh brief, 2026-10-31). Report verbatim:

```
INDEPENDENT VERIFIER REPORT: fix/vendor-prices (price2, HEAD 1bfff15c on d0ea14b7)

VERDICT
- D1 armor price = cost x defence / maxac: HOLDS. Seen in game.exe at 0x62f0f2. FUN_00629bb0(item, 0x32) checks the type: ItemTypes row 0x32 is "armo" (read from the loaded table), and the test uses a precomputed Equiv bitmask over type and type2. The guard (maxac-minac+1 != 0, maxac != 0) and the 32-bit imul/idiv match d2d. One result covers buy, sell and repair (ebp-0x10, -4, -8). FUN_006253b0 reads the stat list's base layer (FUN_00624ed0 reads [statlist+0x24]; the total getter FUN_00624f60 reads +0x48), so enhanced defence in props is correctly left out.
- D2 class item sell /4 after the ethereal /4: HOLDS. The ethereal quarter is at 0x62f888 and FUN_0062c060 at 0x62f99d (sell only). That function checks ItemTypes+0x21 (Class) < 7 on the item's own type (FUN_0062b400), with no Equiv walk. Both quarters truncate toward zero.
- D3 quality extras only when identified: HOLDS. 0x62f146 tests item flag 0x10, then 0x62f182 jumps to 0x62f85e, skipping the quality jump table and the 0x62f822 adds. Quality 0 and 9 also match.
- D4 unidentified item cannot be repaired: HOLDS. For mode 3, 0x62efee calls FUN_0062e660, which returns 0 unless flag 0x10 is set and flag 0x400000 is clear. This happens before charges and durability are looked at, so an unidentified wand with missing charges also costs 0 in game.exe. d2d's new early return matches.
- D7 broken ethereal sells for 0, then the minimum 1: HOLDS. 0x62fa60 checks FUN_0062a8d0 (flag 0x400000), then FUN_00629930 (ItemsTxt +0x113 == 0, +0x112 > 0, max durability != 0, stat 152 <= 0), then max durability != 0 and stat 72 <= 0, then `xor edi,edi`. This comes after both quarters and before the NPC mults; the final minimum 1 is at 0x62fd75. Small difference: d2d's indestructible() counts any non-zero stat-152 prop, while game.exe needs the total to be > 0. Negative or cancelling values are not realistic.

WHAT I RAN (all on the worker, my files in the verifier's own scratch dir)
1. Disassembly of FUN_0062efb0 (whole), 62c060, 62e660, 629930, 62a8d0, 629bb0, 6253b0/625480, 624ed0/624f60, 6292f0, 62a840.
2. vtypes.py: for all 659 item codes, game.exe's own FUN_00629bb0(0x32) and FUN_0062c060 compared with d2d's logic (type_is Equiv walk, ItemTypes Class non-empty). 0 mismatches. 111 class items in types abow ajav ashd aspe h2h h2h2 head orb pelt phlm, including 60 armor-file class items (pelts, barbarian helms, paladin shields, necromancer heads), which are both armor-scaled and class-quartered. Only gems have a type2, and none of them is armor.
3. edge.py: 13,920 hand-made cases. 22 codes (cap lea buc pa1 dr1 ba1 ne1 ob1 am1 ktr hax wnd tkf jav aqv key 7cr ci0 uap xpk 9ja am5) x quality 0..9 x identified x ethereal x durability {0/M, M/2, M/M, 0/0, 5/0, 1/M} x defence {0, maxac, 2*maxac+7, -1} x quantity {1, 3} for stacks. Vendor cycles through all 17 VENDORS, difficulty 0..2. Both sides were run. Raw mismatches buy/sell/repair: 1608/1219/926. price_terms.py puts all of them in stack, bigmul, whole or nodur. Outside those: buy 0/10812, sell 0/11000, repair 0/10022.
4. Rebuilt test_game and test_rules from a fresh HEAD worktree with the same cmake flags. test_game --price on c100k.tsv reproduces dH.tsv byte for byte, so the author's binary is not stale.
5. Unit tests, HEAD tests built against rules.hpp of each commit: d0ea14b7 and the 4 intermediate commits abort; HEAD prints "test_rules: ok".
6. Mutation test (HEAD with one claim reverted at a time). D1, D2, D3, D4 and D7 each make their own assertion fail, so every claim has a test that catches its removal.

COUNTEREXAMPLES: none found against D1-D7. Mismatches remain only in the declared unclaimed terms, e.g. "8290 tkf 7 0 0 0 0 0 0 4 1 0 1 4 1 0 halbu 1": game 37 2244 636, d2d 5006 2503 621 (stack, D6).

LIMITS AND WRONG ASSUMPTIONS TO KNOW
- Stacks and quivers (tkf, jav, am5 class javelins, aqv) are set aside wholesale by price_terms. The claims run before the stack branch in game.exe, but stacks did not test them in isolation.
- Unclaimed terms that price_terms does not list:
  - Item flag 0x20000 (starter item) returns 1 for every mode (0x62f00b). d2d starter bucklers are now armor-scaled.
  - Repair with stat 252 (replenish durability) prices (max-1)/max of the base whatever the current durability (0x62fa22-0x62fa3c).
  - The oracle answers 0 for both. So "0 left outside unclaimed terms" holds only for items with no stat list.
- Sockets: checked by reading the disassembly only (the oracle's inventory is 0). FUN_006292f0 adds socketed cost/2 after the extras (on the unidentified path too), before the quarters and the D7 zero. d2d's item_price order matches. Repair: game.exe adds sockets to the repair base, d2d does not (an existing, documented ponytail).
- d2d Item.defense defaults to -1. Every armor path I found sets it: generate_item (both seed paths), the d2s/net parser, and starter items through generate_item. Gamble stock has -1 but is priced by gamble_price. Not traced: cube kUseItem keeps the old defence when the code changes (cube.hpp:341), which now feeds into the price.
- Callers whose behaviour changes:
  - store_buy, store_sell, ingame sell hover, store hover, net.cpp buy/sell cost: new prices.
  - store_repair: returns false for unidentified items.
  - repair_all_cost and store_repair_all (commands.cpp): add 0 for unidentified items.
  - No other users of item_price or repair_cost in components/ or apps/.
- The author's ctest run has 5 failures: test_mpq (damaged MPQ assert), smoke_d2d, pad_use_d2d, quests_act1, quests_act4 (no voice MPQ, quake timeouts). They look like environment problems, but I did not check whether upstream fails the same way.
- No live game, so whether base stat 31 equals the saved defence is inferred from the disassembly and the save format, not observed.

CLEANUP: the price2 checkout is untouched (status clean, HEAD 1bfff15c). My worktree was removed with git worktree remove and pruned, and my build dir was deleted. Nothing of mine is running (ps check). Case and output files stay in the verifier's scratch dir.
```

Author's answers: the ctest baseline was checked after the report (upstream d0ea14b7 fails the same five, see "Their checks"). The starter-item flag (0x20000: price 1 in every mode), the replenish stat 252 repair term and the cube's kept defence are new limits, listed under Limits; none is in a claim (starter items were already mispriced by d2d before the patch: game.exe gives 1 whatever the item).
