## Why

DECOMPREPAIR proved (docs/research/price-trace.md, docs/research/pr-price-proof.md; model equal to game.exe on 100,000 emulator cases) that d2d's vendor prices deviate from game.exe 1.14d's `FUN_0062efb0` in several separable terms. The coordinator asked for the local upstream PR candidate (agent PRICEPR, 2026-10-31). No observable behaviour change here: research docs and oracle scripts in this repo; the d2d patch lives in the local upstream clone on the worker host, nothing pushed or posted.

## What Changes

- Local upstream branch `fix/vendor-prices` (worker host) on upstream/main: a tooling commit (oracle, model, per-term attribution, `test_game --price`) and one commit per proven claim (D1 armor defence, D2 class item sell quarter, D3 identified gate on quality extras, D4 no repair of unidentified items, D7 broken ethereal sells for the minimum), with upstream unit tests and `// ponytail:` notes for what stays out (D5, D6, D8, D9, D10, the purse-cap sale).
- `docs/research/pr-price-{proof,draft,evidence-pack}.md` finished: per-claim before/after, after-run, upstream checks, verifier report, limits.
- `tools/emu/price_*.py` here kept in step with the branch copy (two new d2d-behaviour switches `bigmul`, `nodur`, and the clean-subset line in `price_terms.py`).
- `docs/decomp-queue.md` item 5, `docs/research/pr-bundle-queue.md` price row, `docs/research/price-trace.md` PR state.
