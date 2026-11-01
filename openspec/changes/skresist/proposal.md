## Why

Decomp-queue row 21 (SKRESIST, sorc-skill-matrix rank 8): the Sorceress's damage plan (lightning vs immune packs, cold vs resistant bosses) and the bot's skill choice depend on how game.exe 1.14d builds a monster's resistances by difficulty, how curses, auras and masteries lower them, the 100 immunity line and the -100 floor, and whether d2d gives the same numbers.

## What Changes

- Trace in game.exe (disasm + emulator) the base resistances at spawn, the boss-mod adds (cited), the curse / aura reduction against immunes, the pierce and the effective resistance, the damage step, with V/D/S/I/U tags.
- An oracle that runs game.exe's own code natively and a test mode on the REAL d2d build (upstream main) on the same cases; deviations D1.., bot findings B1.. (described, not changed), drill recipe, matrix cell lines.
- In this repo: `docs/research/sk-resist-trace.md`, `tools/emu/sk_resist.py`, `tools/emu/sk_resist_test.cpp` (copy of the d2d-side test mode). Shared files (decomp-queue, BACKLOG, INDEX, matrix) are NOT edited: their lines are at the end of the trace for the merger.
- No game, no d2d slot, nothing pushed or posted; no Go change.
