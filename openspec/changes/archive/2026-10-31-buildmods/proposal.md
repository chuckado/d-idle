## Why

Decomp item 13 (docs/research/monster-mods-trace.md) found that d2d's spawn-time champion / unique / minion stats deviate from game.exe 1.14d (D1 boss rows at the boosted level, D2 minions without the leveladd, D3 ghostly slowed). The coordinator asked for the clean, provable parts to be built as separate claim commits on upstream main, with a differential on real d2d builds, and for D1 to be re-derived skeptically because it changes how strong every champion and unique is.

## What Changes

- A local PR candidate in the worker host's PR checkout (branch `fix/boss-stats`, nothing pushed): claim commits D1 (narrowed: life, defence and experience only), D2 (minion leveladd and experience), D3 (ghostly speed), B (berserker leveladd), and one tooling commit.
- `docs/research/pr-mods-proof.md` and `docs/research/pr-mods-draft.md` (proof with counts per claim, verifier commands, limits, PR body draft).
- `tools/emu/boss_level.py` and `tools/emu/boss_level_test.cpp` (copies of the PR's oracle and test mode, new names).
- One line each in `docs/decomp-queue.md`, `docs/research/pr-bundle-queue.md`, `docs/BACKLOG.md`; a correction banner line in `docs/research/monster-mods-trace.md`.

owns: docs/research/pr-mods-*.md, tools/emu/boss_level*, docs/research/monster-mods-trace.md (banner), docs/decomp-queue.md, docs/research/pr-bundle-queue.md, docs/BACKLOG.md (one line each), openspec/changes/buildmods/**
exclusive: fork-branch:fix/boss-stats (the worker host's PR checkout; offline, no game instance, no d2d slot)
gates: none (research and a local candidate)
