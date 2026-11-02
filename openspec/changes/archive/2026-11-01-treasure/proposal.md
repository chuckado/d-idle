## Why

Coverage gap `treasure` (0 of 852 classes compared, docs/coverage/gaps.tsv rank 5): treasure classes decide the potions, scrolls, gold and gear the Sorceress and Paladin plans live on, and nothing said how game.exe 1.14d picks the class for an Act 1 / Act 2 monster or chest, how many items a kill drops, or how our bot's gold and supply assumptions compare.

## What Changes

- docs/research/treasure-trace.md: questions Q1.., d2d deviations D1.., bot findings B1.., drill recipe, matrix cell text, shared-file lines.
- tools/emu/tre_*: runners for the upstream drop differential against a given d2d build, a selection oracle (game.exe's FUN_005a6600 path against d2d's rule), and an exact class-odds tool checked by native rolls.
- docs/coverage/ledger.tsv, gaps.tsv and the generated rule-coverage.md: one row for the proven classes.

## Impact

Docs and offline tooling only. No Go code, no live game, no d2d slot. Nothing pushed or posted upstream.
