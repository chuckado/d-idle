## Why

The lab fork now has the real `debug missiles` verb (docs/forkcheats.md). FIXMISSILE wrote the bot side against the requested shape. The real verb also lists a monster's hit effect (for example shamanexp: speed 0, owner -1), which the bot must not dodge.

## What Changes

- `d2dctl.ParseMissile` keeps the 10 leading fields and now reads the two optional extras (friendly, name); the raw parse stays faithful.
- `d2dsrc.Source.FillMissiles` drops rows with speed 0 (hit effects). Hostility stays in combat: owner type 1 and not a friend is hostile, the player's own (owner type 0, own id) is not.
- Tests with the real sample rows; docs/fixmissile.md says the real verb exists.
- Delta spec: backend-contract MODIFIED.
