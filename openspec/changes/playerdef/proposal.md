## Why

The bot reads and derives the player's life, mana, resistances and attributes, and the d2d backend computes them itself. No trace compared game.exe 1.14d, d2d and the bot for these derived defence stats (decomp lane, docs/decomp-queue.md).

## What Changes

- New trace `docs/research/playerdef-trace.md`: Q1..Q14 with addresses and tags, deviations D1..D8 (d2d vs game.exe), bot findings B1..B4, drill recipe, differential recipes, proposed matrix text. Docs only; no Go code, no game, no ledger row (nothing was run: the runner allowed no emulator).
- Two read-only helper scripts `tools/emu/pd_dis.sh` (disassembler wrapper) and `tools/emu/pd_sh.sh` (grep in the upstream read checkout) for the follow-up differential.

## Impact

Docs and tools/emu only. No spec delta.
