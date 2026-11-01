## Why

Decomp item PALSKILL (follows the offline Paladin probe, docs/classes/paladin-probe.md, open items PALAURA and "does an aura do anything on d2d"): the Paladin rules the bot depends on (Holy Bolt, Blessed Hammer, the aura set) were never read in game.exe 1.14d, and upstream's skills.md lists Blessed Hammer's path and Holy Bolt's heal as "not built".

## What Changes

- Trace in game.exe 1.14d (disasm + emulator) Holy Bolt's hit function 7 (FUN_005a9fb0), Blessed Hammer's do 73 (FUN_005d0040), Concentration's share (FUN_006461d0), the undead / demon damage function 5 (FUN_005ad2f0), the aura dos 65 / 66 and their pulse (FUN_005cf010, FUN_005cf3a0, FUN_0056cda0), with V/D/S/I/U tags.
- Differentials, game.exe natively vs the real d2d build: undead / demon flags (734 ids), element min / max of six Paladin skills (8,640 cases), aura calcs (21,760 cases), plus game-side runs of Concentration's influence and the hammer's flight and hits.
- In this repo: `docs/research/pal-skill-trace.md` (deviations D1..D7, bot findings B1..B5, drill recipe, matrix text), `tools/emu/pal_*` (own prefix), shared-file lines in `merge-lines.txt`. No Go change, no live game, no d2d slot, no PR (none proven).
