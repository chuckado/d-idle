## Why

The Sorceress is the bot's primary class and `docs/coverage/rule-coverage.md` showed skills.sor with five rows that had no ledger line (Inferno, Blaze, Fire Ball, Enchant, Glacial Spike). This card decomposes them (game.exe vs d2d vs the bot) with the existing SKMISS, SKGROUND2 and necskill harnesses and closes the ledger.

## What Changes

- docs/research/sor-rest-trace.md: questions Q1.., deviations D1.., bot findings B1.., differentials, matrix cell updates, drill recipe.
- tools/emu/sk_rest_* and sorcrest_*: oracles (hit pipeline, burst radius, shared burst roll, calc differential for the 30 Sorceress rows), the d2d-side test sk_rest_test.cpp and the build and run scripts.
- docs/coverage/ledger.tsv (two lines) and the regenerated docs/coverage/rule-coverage.md.

## Impact

Docs and offline tooling only; no Go code, no live game, no d2d slot. The d2d build used lives in a scratch copy on the worker host; no fork branch, nothing pushed or posted. No PR candidate (D1 proven on a real build, fix not built).
