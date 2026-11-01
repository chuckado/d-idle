## Why

The user asked (2026-11-01) whether Frozen Armor behaves on d2d as in game.exe. SKARMOR (docs/research/sk-armor-trace.md D1) proved the game side (FUN_005cac50 follows the freeze rule) and read d2d's `buff_events` as stunning and chilling every damaging striker, but no d2d-side run existed, so no PR candidate could be proposed.

## What Changes

- On the worker host (not in this repo): local upstream candidate `fix/armor-hit` (base upstream main 72132b99): a d2d test mode `test_armor_hit --armor-hit` that drives the real `Fight::buff_events`, the game-side oracle `tools/emu/armor_hit.py` (FUN_005cac50 natively), and a claim commit that makes `buff_events` follow FUN_0057b230 (resisted length, specials chilled, plain monsters frozen at coldeffect < 0 / MonsterFreezeDivisor).
- In this repo: docs/research/pr-armorhit-proof.md (evidence pack, the answer to the user's question, the bot implication, a fork proposal), pr-armorhit-draft.md (PR body, limits first), a queue line in pr-bundle-queue.md, copies of the tooling (tools/emu/armor_hit.py, tools/emu/armor_hit_mode.inc), this card.
- Nothing pushed to the public repo or posted upstream; the private fork is not touched (proposal only).

## Impact

Docs only here. d2d soaks with Frozen Armor are optimistic against bosses and champions (Andariel first) until the fix runs.
