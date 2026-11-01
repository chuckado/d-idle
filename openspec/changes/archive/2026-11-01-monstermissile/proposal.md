## Why

The rule-coverage table (docs/coverage) ranks "missiles 658 of 684 untested" as the top gap: monster missiles (arrows, bolts, spit, pots) decide most damage the bot takes in Acts 1 and 2, and only the player skill path and the hit on the player were traced.

## What Changes

- docs/research/monster-missile-trace.md: how game.exe 1.14d launches a monster missile (level, damage record, difficulty scaling, brackets, elements), the per-row data and record for the Act 1 / 2 shooters, deviations D1-D6 against d2d, bot findings B1-B6, no PR candidate (why and how to make one), a drill recipe, and ledger / matrix updates.
- tools/emu/mm_*: table dump, the game.exe launch oracle (213 launches under the emulator), the level-bracket probe and the summary generator.
- docs/coverage/ledger.tsv and rule-coverage.md: one READ-ONLY line for 48 missile rows (game side V run, d2d side read).

## Impact

Docs and emulator scripts only. No Go, no live game, no d2d slot, no spec, policy, ADR or glossary change. Missing glossary term for the coordinator: "missile level" (the number a missile's level brackets use: monster skill bonus + 1 for a MissA shot) and "launch record" (the 0x7c-byte damage record FUN_0064b860 builds).
