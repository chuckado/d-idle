# PR candidate: vendor prices (proof)

> status: proof of deviations complete, PR candidate NOT built (agent DECOMPREPAIR, 2026-10-31; time box). Trace and formula: [price-trace.md](price-trace.md). No draft or evidence pack yet: they follow the fix commits.

## What is proven

game.exe's `FUN_0062efb0` (buy 0, sell 1, repair 3) run in the emulator against d2d's `rules::item_price` / `rules::repair_cost` (upstream 0dc4c62a) on the same 100,000 items: buy 39,626, sell 38,939, repair 42,705 differ. The formula written in price-trace.md, coded as `price_model.py`, equals game.exe on all 100,000, and switching one of its terms to d2d's behaviour explains each single-term difference (table D1-D7 in price-trace.md). V.

## How to re-run (worker host, the PR checkout's `price-trace` worktree, branch `trace/price`, tooling commit f38cb9af)

```sh
cd tools/emu    # D2_MPQ_DIR = a folder with the MPQs and bin/game.exe
uv run python price_oracle.py gen 100000 7 > c.tsv
uv run python price_oracle.py game c.tsv > g.tsv          # about 4 min in 8 parallel parts
../../build/tests/test_game --price c.tsv | grep -E '^[0-9]+ ' > d.tsv
uv run python price_model.py c.tsv g.tsv                   # model 100000/100000
uv run python price_terms.py c.tsv g.tsv d.tsv             # differences by term
```

md5 of the 100k run: game 6e5fa976dea4f2bdaf4a1f7b140d5312, d2d 954eeb4cd3c1856329790cf1f431105e (seed 7).

## Limits

- The oracle's item has no stat list: magic stat costs (`FUN_00628e70`), staff mods (`FUN_0062edd0`), charges and sockets are read from disassembly only (I); d2d models charges and sockets on buy/sell, none of the stat terms.
- No quests and no reduced prices in the cases (both sides model the same mults; not compared).
- Body parts (ItemTypes 0x28) and tomes (PR #127) excluded from the generator.
- Affix ids: d2d's MagicPrefix index p is game.exe's record `prefixes + (p - 1) * 0x90` (lined up by the cost columns); every MagicSuffix cost is 0 in 1.14d.
- Nothing verified live on the original game; the D3.r drill (d2d) price 19 agrees with both sides for a normal weapon.

## Next (for the PR)

One claim per commit on `trace/price`: D1 armor defence scaling, D2 + D7 sell (class / 4, ethereal at 0 durability), D3 + D4 identified gate, D5/D6 if the bundle has room; `test_game --price` and the oracle as the tooling commit; upstream tests with `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`, pr_guard 0/0, independent Opus verifier. Local only.
