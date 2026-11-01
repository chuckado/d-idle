## Why

SKRESIST (docs/research/sk-resist-trace.md, branch agent/skresist) proved deviation D1 on the real d2d build: a curse or enemy aura that lowers a monster's resistance counts its full value on an immune monster, while game.exe divides a negative change by 5 when the monster's base resistance is 100 or more (FUN_005c3540). Hell and Nightmare immunities break on d2d that game.exe keeps. The trace stopped at "harness ready, no fix"; this card builds the local upstream candidate.

## What Changes

- On the worker host (not in this repo): local branch `fix/curse-immune` in a new worktree `skresistfix` of the PR checkout `skresist` (base upstream main 72132b99): one claim commit (`Fight::apply_target_stats` takes the monster's base resistances and counts a negative change as a fifth when the base is 100 or more) and one tooling commit (`tests/game/test_sk_resist.cpp` driving the real `Fight::target_of`, `tools/emu/sk_resist.py`).
- In this repo: `docs/research/pr-curseimmune-draft.md` (PR body draft), `docs/research/pr-curseimmune-proof.md` (commands, counts, checks), this card, OWNERS lines.
- Nothing is pushed or posted; the private fork is not touched (a later card carries it after an independent verifier).

## Impact

Docs only here. D2 (no -100 floor on magic and physical in `missile_blow`) is a different code path, unreachable with 1.14d data, and is NOT in the claim.
