## Why

The trace docs/research/sk-resist-trace.md section 6 lists bot findings B1 to B4 in the monster resistance model: no -100 floor, pierce ignored in the damage factor, a comment that claims a monster difficulty penalty.

## What Changes

- `expFactor` floors at -100; `Target.Pierce` / `Target.EffRes` apply the player's resistance pierce only while the total is under 100; the comment is corrected; B3 left (trace tags it S). See docs/skresistbot.md.
- Overlap (declared): `internal/combat/fighter.go` is owned by the open card armorcut; one comment line there (the `refineResists` doc) is left for the merger, so it is not in owns.
- No live proof owed; archive-ready.
