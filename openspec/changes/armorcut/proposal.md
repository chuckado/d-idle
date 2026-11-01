## Why

DRILLS26 D26-A: in a melee fight a Frozen Armor recast issued while the old armor was still up did not extend it (the armor was absent for up to 3.5 s in 300 s, both fork binaries). The bot counted the cast as done on the key press.

## What Changes

- `Fighter.buffPass` retries a buff cast during which the player was in hit recovery (bounded, mana-checked) and casts nothing mid-hit or mid-cast.
- Evidence, tests and the live recipe are in docs/armorcut.md.

## Impact

Combat buff upkeep only. No numeric spec gate changes. Awaits the live proof.

## Overlap (declared)

No open card owns internal/combat/fighter.go. fixcastgap, fixskilllvl, fixmanamodel and buildv3 own other combat files (castgap.go, skilllevel_test.go, profiles.json and tests); none overlaps and no glob of this card covers them.
