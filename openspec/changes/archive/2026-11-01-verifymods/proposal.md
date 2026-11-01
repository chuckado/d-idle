## Why

docs/process/upstream-pr-standard.md section 5: every upstream candidate needs an independent adversarial verifier before the coordinator may ask the user for a go. Candidate: fix/boss-stats (boss and minion spawn stats, claims D1 / D2 / D3 / B, card buildmods).

## What Changes

- Own re-derivation from the disassembly of which MonLvl level game.exe reads for each boss stat, spawn vs per hit, including FUN_005a2620 / FUN_005a6ff0 and the other FUN_006538a0 callers.
- Own oracle and d2d mode (tools/emu/verify_mods.py, verify_mods_mode.inc, verify_mods_compare.py), run against real d2d builds of upstream main, each claim commit and the head.
- Verifier section in docs/research/pr-mods-proof.md, verdict row in docs/research/pr-bundle-queue.md, a BACKLOG line.

## Impact

Docs and verifier tooling only. Offline: no game, no d2d slot (the det state hashes stay open). Nothing pushed or posted; the author's branch and worktrees untouched.
