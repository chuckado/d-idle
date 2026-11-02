## Why

docs/coverage/rule-coverage.md lists monster.ai at 0 of 146 rows: nothing compared against game.exe 1.14d. Wake and aggro were traced (mon-wake-trace.md), but not what a woken Act 1 and 2 monster does. Upstream's own oracle (tools/emu/mon_ais.py) leaves out the Act 1 core AIs (Skeleton, Zombie, Fallen, Fallen Shaman, Goatman, Quill Rat, Brute, Blood Hawk, Wraith, Corrupt Rogue, Andariel, ...), whose d2d code is a hand port.

## What Changes

- docs/research/monster-ai-trace.md: per-AI table (146 rows), 25 Act 1 and 2 core thinks run in game.exe's own code under the emulator and replayed through d2d's rules::mon_think, deviations D1.., bot findings B1.., drill recipe, matrix-cell and ledger proposals.
- tools/emu/mai_*: the harness (extends upstream's mon_ais.py without editing it; the d2d side is a scratch worktree of upstream f8aab0e8 on the worker host).
- docs/coverage/ledger.tsv: rows proved (case counts in the doc).

## Impact

Docs and tools only. No Go, no spec, ADR or glossary change. No game, no slot. Nothing is pushed or posted to the d2d fork or GitHub.
