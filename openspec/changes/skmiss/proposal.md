## Why

Before act progression goes further the user asked for a focused check of the Sorceress skills. SKILLINV's matrix (docs/research/sorc-skill-matrix.md) leaves the player missile pipeline UNTRACED apart from Fire Bolt's one live mana / damage proof. Decomp row 14: how game.exe 1.14d spawns, flies and lands the plan's missile skills (Fire Bolt, Charged Bolt, Ice Bolt, Frost Nova, Lightning, Chain Lightning), what d2d does, what the bot assumes.

## What Changes

- Trace in game.exe (disasm + emulator) the spawn (skill start / do, mana payment, missile rows, counts, aim), the flight (reusing upstream's proven flight oracle) and the hit (the roll, resistance, life), with V/D/S/I/U tags.
- Differential oracle on real d2d builds for the hit (skill, level 1..20, synergies, mastery, resistance, pierce, life, seed); a structural comparison and a drill recipe for the stateful parts.
- In this repo: `docs/research/sk-missile-trace.md`, `tools/emu/sk_missile_hit.py` and `tools/emu/sk_missile_hit_test.cpp` (copies of the upstream-candidate oracle and test mode), decomp-queue row 14, BACKLOG and INDEX lines.
- If a deviation is proven on the real patched d2d build: a LOCAL upstream candidate branch on the worker host (nothing pushed or posted).
