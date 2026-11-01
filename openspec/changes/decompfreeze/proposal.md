## Why

The chill verifier (card verifychill) and decomp item 6 follow-up found unpatched freeze deviations by reading only: C3 (game.exe turns a freeze into a chill on special monsters, d2d freezes and stuns any monster), C4 (d2d's `half_freeze` on Blood Raven and the Countess) and C5 (d2d's freeze ignores cold resistance and MonsterFreezeDivisor, adds a chill, and freezes at coldeffect >= 0). The next upstream bundle needs a verified candidate ready.

## What Changes

- Re-derive the freeze rules independently from the disassembly (FUN_0057b230, FUN_005aa8b0 / FUN_005a8f20, the table at 0x732980 row 6, the flag-8 makers, the state-118 call) and decide C3 / C4 / C5.
- Local upstream candidate `fix/freeze-rules` on #155's head ddc8281d on the worker host: one claim commit per proven deviation and one tooling commit (`tools/emu/freeze_rules.py`, `tests/game/test_freeze_rules.cpp`); nothing pushed or posted.
- In this repo: `docs/research/freeze-rules-trace.md`, `docs/research/pr-freeze-proof.md` and `-draft.md`, copies of the oracle and the test mode under `tools/emu/`, one row each in `docs/decomp-queue.md`, `docs/research/pr-bundle-queue.md` and `docs/BACKLOG.md`.
