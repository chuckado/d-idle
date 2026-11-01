## Why

Frost Nova is in the bot's early plan (BUILDV2: Frost Nova 1 at clvl 6, the crowd rule in `novaTick`). SKMISS left its ring untraced (D3: "tables at 0x6e1288, U"; reach about 12 subtiles, I) and SKLIGHT handed over the NextHit question for the nova rows. Decomp row 18 (SKNOVA): how game.exe 1.14d builds the Frost Nova ring, how d2d builds it, where they differ, and how the bot uses it; plus Static Field's radius unit.

## What Changes

- Trace in game.exe (disasm + emulator) the nova do (FUN_005c9b50 -> FUN_0056d400), the ring tables, the params FUN_0059fa30 reads, the flight of the 64 missiles and the hit rule for NextHit rows, with V/D/S/I/U tags; cross-reference the chill / freeze traces instead of redoing them.
- Emulator oracle `tools/emu/sk_nova_ring.py` (runs next to upstream's tools/emu): FUN_0056d400's params, the 64 real flights against the reading and against d2d's aim; a hit model (I) for the NextHit rule.
- In this repo only `docs/research/sk-nova-trace.md`, the oracle, and this card. Shared-file lines go to the merger in the trace's last section.
- No PR candidate unless the existing differential harness proves a d2d deviation.
