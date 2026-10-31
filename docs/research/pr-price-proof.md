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
- Nothing verified live on the original game; the D3.r drill price (19 for a Short Staff at Charsi) agrees with both sides and is unchanged.

## Independent verifier

VERIFIER_REPORT
