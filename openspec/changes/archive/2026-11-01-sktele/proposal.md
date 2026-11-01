## Why

Decomp row 16 (SKTELE, sorc-skill-matrix rank 3): Teleport is the Act 2 movement skill (clvl 18, F2 nav hops), but its rule in game.exe was UNTRACED: the destination (range, snap to a free spot, line rule, refusals), the mana on a refused cast, and what happens on arrival. d2d's `spot()` puts the player at the exact clicked point at any distance; if game.exe differs, d2d fakes nav speed or stalls hops the real game would make.

## What Changes

- Trace in game.exe 1.14d (disasm + emulator) the Teleport start refusals (FUN_0056f640), srvdofunc 27 (FUN_005ca360), the warp FUN_00554ea0 (target room, nearest free spot FUN_0064dea0, position set, the player tail: 20-step ring, warp spot, pets' warp), with V/D/S/I/U tags.
- An emulator oracle of the destination rule (`tools/emu/sk_tele_spot.py`) on synthetic rooms, compared with a line copy of d2d's rule (no PR claim: a line copy is the "before" only).
- `docs/research/sk-teleport-trace.md`: deviations D1..., bot findings B1..., drill recipe with the FORKDBG verbs, matrix cell updates, and the shared-file lines for the merger (decomp-queue row 16, BACKLOG, INDEX, CHANGELOG), which this card does not own.
- No upstream candidate unless a deviation is proven on the real d2d build with an existing harness (nothing pushed or posted).
