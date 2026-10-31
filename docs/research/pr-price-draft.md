# PR draft: vendor prices as game.exe prices them (five claims)

> status: DRAFT (agent PRICEPR, 2026-10-31). Nothing is pushed, posted or filed. Branch `fix/vendor-prices` in the worker host's PR checkout (worktree `price2`), six commits on upstream/main d0ea14b7, 8 files. Proof: docs/research/pr-price-proof.md. The body below follows the earlier PRs' template; one section per claim (bundle rule, docs/research/pr-bundle-queue.md).

````
## What changes
Five claims, one commit each (plus one tooling commit), all in `components/rules/rules.hpp` (`item_price`, `repair_cost`), each with a `tests/test_rules.cpp` case and its line in `docs/research/re/store.md`:

1. **An armor is priced by its defence.** game.exe starts an armor (ItemTypes armo) at cost x its own defence / maxac, for buy, sell and repair; d2d used the bare cost. A Quilted Armor (maxac 11) with 8 defence now buys for 8/11 of its cost.
2. **A class item sells for a quarter.** An item whose ItemTypes Class is set (wands, orbs, pelts, Barbarian helms, Paladin shields, ...) sells for a quarter of the usual price, after the ethereal quarter; buying is unchanged.
3. **Quality extras only on identified items.** An unidentified magic, rare, set or unique item buys and sells as a normal one; an unidentified low quality item is not halved.
4. **An unidentified item cannot be repaired.** Its repair price is 0, so neither Repair nor Repair All touches it (as for ethereal items already).
5. **A broken ethereal item sells for the minimum.** An ethereal item with durability at 0 sells for 1.

Not changed (each a `// ponytail:` in the code): a whole item's repair price stays 0 (game.exe gives 1 and its Repair All skips it), stack quantities, the stats' own costs and staff mods, the divide-first NPC mult above 0xffff, the repair gate's ItemTypes / durability-column checks, a sale over the purse cap.

## Evidence from game.exe 1.14d

| Address | What it shows |
|---|---|
| `FUN_0062efb0` | The price function: buy 0, sell 1, gamble 2, repair 3; builds the buy, sell and repair bases together. |
| `0062f0f2` | Claim 1: `FUN_00629bb0(item, 0x32)` (armo); base = cost x `FUN_006253b0(item, 31, 0)` / maxac (items +0xd0) when maxac - minac + 1 and maxac are non-zero; all three bases. |
| `0062f99d`, `FUN_0062c060` | Claim 2: the item's own ItemTypes row, Class byte < 7 -> sale base / 4, after the ethereal / 4 (`0062f888`). |
| `0062f13c`, `0062f182` | Claim 3: identified flag 0x10 read into edi; `test edi, edi; je 0062f85e` skips every quality extra and goes to the sockets. |
| `0062efee`, `FUN_0062e660` | Claim 4: repair (mode 3) returns 0 when the gate refuses; the gate needs flag 0x10 and not 0x400000. |
| `0062fa60`, `FUN_00629930`, `0062fd75` | Claim 5: sell with ethereal, durability (nodurability 0, durability column > 0, max > 0, not indestructible) and stat 72 <= 0 -> sale base 0 before the NPC mults; the result is floored at 1. |

## How to re-check

```sh
python3 tools/pr_guard.py
cmake --build build && ctest --test-dir build --output-on-failure      # test_game needs D2_MPQ_DIR and D2_PATCH_INSTALLER
# the oracle: game.exe's own FUN_0062efb0 (buy, sell, repair) in the emulator on the loaded tables
# (stat getters answer from the case; quest flags 0), against rules::item_price / repair_cost:
cd tools/emu && uv run python price_oracle.py gen 100000 7 > cases.tsv && uv run python price_oracle.py game cases.tsv > g.txt
../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d.txt
uv run python price_model.py cases.tsv g.txt                 # model 100000/100000: the formula as read equals game.exe
uv run python price_terms.py cases.tsv g.txt d.txt           # differences by term; last line: outside the unmodelled terms
```

The last line of `price_terms.py`, on the 100,000 cases: before (upstream) buy 24,084 / sell 32,472 / repair 21,248 differ; after each commit 3,933 / 20,070 / 7,464 (armor), 3,933 / 4,189 / 7,464 (class), 0 / 233 / 7,464 (identified), 0 / 233 / 0 (repair), 0 / 0 / 0 (broken ethereal), of 84,030 / 93,095 / 75,112. No line that matched before differs after, other than old coincidental matches (an unidentified low quality armor, a class javelin stack of 4) that the next claim or the stack term owns. All cases: 39,626 / 38,939 / 42,705 -> 15,527 / 6,406 / 20,695; what is left is the terms listed under "Not changed".

Limits: the oracle's item has no stat list (no magic stat costs, staff mods, charges or sockets; d2d's socket and charge terms are untouched); no quests or reduced prices in the cases; body parts and tomes are outside the generator, and it gives random durability to items without a durability column (that only feeds the repair-gate term, not a claim). Nothing checked live on the original game for these terms. `tools/emu/price_*.py` are new tools (CODEOWNERS review).

Presubmit `--quick`: 0 FAIL; the build and ctest were run by hand (above).

## Screenshots
Not applicable.
````
