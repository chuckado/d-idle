## Why

Decomp row 17 (SKLIGHT, docs/research/sorc-skill-matrix.md section 5 rank 4): Lightning and Chain Lightning are the lightning plan's main attack from clvl 26 to 90. SKMISS (docs/research/sk-missile-trace.md) left the per-frame hit rule of the piercing `lightningbolt` (its D4), Chain Lightning's hop rule and the NextHit handling as S / U, and the bot's use of the two skills (ranges, cast gap, immune fallback) was never checked against game.exe.

## What Changes

- Trace in game.exe 1.14d (disassembly) Chain Lightning's do 26 and hit function 12, the hop search FUN_0056bd10 / FUN_0056b7e0 / FUN_0056b3e0, the missile-unit collision FUN_005ae1f0 / FUN_005adf10 (one unit a frame, the last-unit skip, NextHit as the unit state `justhit`), the to-hit gate, and how Lightning Mastery and the synergies reach the hit; with V / D / S / I / U tags.
- Compare with d2d (upstream main 72132b99) and with the bot (internal/combat): deviations D1..., bot findings B1... (described, not changed).
- In this repo: `docs/research/sk-lightning-trace.md` (drill recipe for the fork verbs, matrix-cell updates, shared-file lines for the merger) and `tools/emu/sk_light_rehit.py` (a model of the two hit rules, not a proof).
- No shared file is edited (decomp-queue, BACKLOG, INDEX, CHANGELOG, research README): the trace ends with the lines for the merger.
