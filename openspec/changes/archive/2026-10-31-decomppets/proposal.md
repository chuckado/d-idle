## Why

The merc policy (openspec/specs/merc-policy) and the early Sorceress plan assume how mercs, pets and monsters pick targets, follow and survive, and the soak deaths near pets raised "is d2d / our assumption right?" (decomp-queue item 8, beyond PR #134's get-hit rule). This card traces the pet and merc AIs and the monsters' target search in game.exe 1.14d, compares d2d's code, and lists what the bot assumes. No observable behaviour change: research, docs and emulator scripts only; no Go code, no game instance.

## What Changes

- `docs/research/pets-merc-trace.md`: Q1-Q4 with addresses and tags, d2d's version, the bot's model, findings and deviations tables, the answer the drills should expect, the DRILL RECIPE.
- `tools/emu/pets_merc*.py` (+ a C++ harness against d2d's real rules code): the differential for the pure parts (the monster's target search between the player, its merc and its pets; the merc's search).
- `docs/decomp-queue.md` item 8 row and result line; the `docs/research/README.md` and `docs/INDEX.md` lines; a `docs/research/pr-bundle-queue.md` row only if a candidate exists.

owns: docs/research/pets-merc-trace.md, docs/research/pr-pets-*.md, tools/emu/pets_merc*, openspec/changes/decomppets/**
exclusive: none (offline; no game instance, no d2d slot, no Go changes)
gates: none (research)
