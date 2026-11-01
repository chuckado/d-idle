## Why

Decomp queue row 20 (SKMELEE, widened by the Paladin recon): the first levels of every melee class and the Sorceress staff rung run on the player's plain Attack, and nobody had traced how game.exe 1.14d rolls its to-hit and physical damage (weapon share, Strength bonus, shield carrier) or compared it with d2d's `player_blow` / `make_fighter`.

## What Changes

- Trace in game.exe (disasm plus the unicorn emulator) the plain Attack: do 1 (FUN_0056f070), the to-hit (FUN_0057ec10 -> FUN_0057d9b0, FUN_00622560, FUN_0057d8a0), the damage build (FUN_0057b7d0 -> FUN_0057b420, FUN_0057b660, FUN_0057b6a0), with V/D/S/I/U tags.
- An emulator oracle `tools/emu/sk_melee.py` (game.exe's own functions vs the reading and vs a port of d2d's rule), 100,000 cases on two seeds.
- In this repo: `docs/research/sk-melee-trace.md` (deviations, bot findings, drill recipe, matrix cells, shared-file lines for the merger).
- No d2d PR candidate (none proven with an existing differential harness on a real d2d build); no Go code, no game, no slot.
