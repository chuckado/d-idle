# PR candidate proof: walk velocity (monster RN, chill, floor, velocitypercent)

> status: local candidate, proof complete for the oracle, independent verifier NOT run (card decompwalk, 2026-10-31). Nothing posted. Trace: docs/research/walk-speed-trace.md.

## Claims (one per commit, branch `fix/walk-velocity` on upstream d0ea14b7, worker host `trace/walk` worktree)

| commit | claim | game.exe evidence (V) | d2d before (V-d2d) |
|---|---|---|---|
| 96953bfe | a monster in RN moves at MonStats Velocity x walk %, not Run | `00623f50` stores `00621360` x % / 100 at path +0x7c for every moving mode (`006214a0`: WL, RN); `00621360` reads MonStats +0x32 for a monster whatever the mode; the loader names +0x32 Velocity (descriptor at 0x65161a) and +0x34 Run (0x651640); +0x34 is read only by pace AIs and charges | ai.cpp:2727 `cells_per_sec(type_info.run)` for RN |
| 7e8feb87 | chill adds MonStats coldeffect to velocitypercent; the walk % is at least 25 | `0057af80` puts coldeffect (MonStats byte +0x168 + difficulty; -50 for a player) on stat 0x43 as a stat list; `00623f50` 0x62414e floors the sum at 0x19 | ai.cpp:2724..2726 multiplies by (100 + coldeffect) / 100, floor 10 |
| 97d55732 | a player's velocitypercent adds as it is; only FRW is diminished, for any v != 0 | `00623f50` adds stat 0x43 raw to `00621740` row 4 (stat 96 only); `00621740` skips only v == 0 | strike.cpp:185..186 folds velocitypercent into frw; combat.hpp:170 leaves v < 0 undiminished |

## Differential

`tools/emu/walk_speed.py gen|game|cmp` + `tools/emu/walk_speed_harness.cpp`: game.exe's code from 0x624113 to 0x6241ce under unicorn (upstream tools/emu/emu.py, data tables built from the patched MonStats / MonStats2), d2d transcribed line for line before and after. Exhaustive grid (see the script header): 712,362 cases, 425,665 differ before (player 208,586, monster 217,079), **0 after**. Per term: M1 34,190; M2 124,301; M3 58,588; P1 207,466; P2 1,120. Tolerance: 1 unit of 1/4096 subtile a tick (float vs integer truncation).

## Limits

- The d2d side is a transcription (as the hit_len template), not d2d's binary; the patched code compiles and upstream's tests pass (evidence pack).
- What feeds stat 0x43 (boss speed +20, pace, auras) is d2d's model, taken as given; the claims are only about how the sum becomes a velocity.
- Pets (pets.cpp:217 `max(run, velocity)`) and mercs (merc.cpp:287) move by their own code; not checked, not changed.
- frw == -150 would divide by zero in game.exe; the patch guards it (unreachable).
