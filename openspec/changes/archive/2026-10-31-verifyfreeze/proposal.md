## Why

The freeze-rules candidate `fix/freeze-rules` (card decompfreeze: C5 resisted / divided freeze, C3 special monsters chilled, C4 no half freeze for Blood Raven and the Countess) needs the independent verifier of docs/process/upstream-pr-standard.md section 5 before the coordinator may ask the user for a go. An earlier verifier note (verifychill) said champions were frozen; the candidate says they are chilled. This card settles that with a native spawn run.

## What Changes

- Re-derive FUN_0057b230, FUN_0057af80, the resist row for record +0x34, FUN_00639db0 and the spawn-path flag writes from the disassembly before reading the author's docs.
- `tools/emu/verify_freeze_spawn.py`: game.exe's own population on real levels, every spawned monster's flag word and the outcome of the real FUN_0057b230 on it.
- `tools/emu/verify_freeze.py` + `tools/emu/verify_freeze_mode.inc`: own generator (seed masked to 32 bits in both the Python and the C++ copy), game.exe's FUN_0057c6c0 natively against the REAL d2d `Fight::strike` built from #155's head, each claim commit and the head.
- A verifier section appended to `docs/research/pr-freeze-proof.md` (card decompfreeze is archived), the freeze row of `docs/research/pr-bundle-queue.md`, one `docs/BACKLOG.md` line. Nothing pushed, posted or changed on the author's branch.

owns: tools/emu/verify_freeze.py, tools/emu/verify_freeze_spawn.py, tools/emu/verify_freeze_mode.inc, docs/research/pr-freeze-proof.md, docs/research/pr-bundle-queue.md, docs/BACKLOG.md, openspec/changes/verifyfreeze/**
exclusive: none (offline; own throwaway worktrees of the candidate on the worker host, no game instance, no d2d slot)
gates: none (verification; no spec numbers change)
