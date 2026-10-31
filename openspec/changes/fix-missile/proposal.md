## Why

The decomp of monster attacks (docs/research/monster-attack-trace.md, findings B1 and B3) found two bot-side primitive bugs behind d2d's lower kill rate. B1: `internal/d2dsrc` never fills `Snapshot.Missiles`, so the missile dodge and sidestep never fire on d2d (0 vs 0.24 dodges per kill on Game.exe). B3: `CastGapMs` 450 is shorter than the 560 ms early Sorceress cast, so the fighter re-issues casts that d2d drops (docs/d2d-casts-real.md: 1.6 to 1.75 commands per real cast at 450, 1.02 to 1.04 at 650).

## What Changes

- B1 (d2d-only): the d2d dev interface has no missile list (`state missiles=` is a count). The bot side is built against a requested verb `debug missiles` (spec in docs/fixmissile.md): d2dctl parses it, `d2dsrc.Source.FillMissiles` (new file, called from `d2denv.Reader.Snapshot`; items.go and source.go belong to the open card fix-obj) fills `Snapshot.Missiles` like Game.exe's reader (Class = missiles.txt id, OwnerType/OwnerID, subtiles) and latches the verb off on the first `err`. Game.exe's path is untouched.
- B3 (shared): `CastGapMs` default 450 -> 650 (above the 560 ms cast), knob range unchanged.
- Held, policy: B2 (run policy into packs), B4 (TownPotionMin 2, chicken exits). Also held: `MissileDodge` (sidestep) stays off until a live pass.
- Delta specs: backend-contract ADDED, combat-loop ADDED.
