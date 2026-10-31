## Why

The bot's gold decisions (repair gate, sell ranking, potion budget, stash cap) rest on UNVERIFIED price guesses, and the drills showed "repair not confirmed" and a purse cap. Decomp lane item 5 traces game.exe 1.14d's vendor price function and compares d2d and the bot.

## What Changes

- docs/research/price-trace.md: repair, sell and buy formulas and the gold caps with addresses (V/I/U), the d2d deviations D1-D7, the bot findings B1-B6 with file:line, the table the bot should use.
- docs/research/pr-price-proof.md and tools/emu/price_oracle.py, price_model.py, price_terms.py: the emulator oracle, the formula as code (equal to game.exe on 100,000 cases) and the per-term differential against d2d.
- Index lines: docs/decomp-queue.md item 5, docs/research/README.md, docs/INDEX.md, docs/research/pr-bundle-queue.md row.
- No code of the bot changes; findings go to the fix queue.
