## Why

The local d2d PR candidate `fix/blizzard-shard-offsets` (card skground2, commits 9a1a90eb and 681d52a9) needs the independent verifier of docs/process/upstream-pr-standard.md before it may be offered upstream. no observable behaviour change: docs and an offline harness only.

## What Changes

- Re-derive game.exe's shard rule (FUN_005a9820) from the disassembly with an own generator and emulator harness (`tools/emu/verify_blizzard.py`), and an own d2d probe (`tools/emu/verify_blizzard_mode.inc`) run on real d2d builds of the base (7d5b3a5c) and the candidate (681d52a9).
- Diff review of both commits, other callers of the offset code, merge-tree against the current upstream main and the open PRs #159 / #160.
- Verdict in `docs/research/pr-blizzard-verifier.md`. Nothing pushed or posted upstream.
