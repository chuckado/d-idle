## Why

SKGROUND (card skground, merged) left Hydra and Fire Wall unread in game.exe, the Blizzard shard hit area (D4) as a read plus an inference, and the shard offsets (D1) proven only on the game side. This card runs the real game.exe routines under the emulator against a REAL d2d build for all three and builds a local PR candidate for D1 if the real-build differential proves it.

## What Changes

- docs/research/sk-ground2-trace.md: Fire Wall (do 24, maker do 6, hits), Hydra (start check, do 144), D4 and D1 on a real d2d build, deviations D9.., bot findings B6.., drill recipe, matrix cells, shared-file lines for the merger.
- docs/research/pr-blizzard-draft.md: draft text of the local candidate (nothing pushed or posted).
- tools/emu/sk_gnd2_*: oracle scripts (hit, hit_cmp, fw, fw2, fw_cmp, row, blizz_d2d, blizz_sum) and the d2d-side test sk_gnd2_test.cpp; sk_ground_blizz.py and sk_ground_run.sh (skground) are used unchanged.

## Impact

Docs and offline tooling only; no Go code, no live game. The PR candidate lives on the worker host in a scratch checkout of upstream (local branch fix/blizzard-shard-offsets, not pushed).
