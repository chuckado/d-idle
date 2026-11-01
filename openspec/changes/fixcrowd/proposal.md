## Why

DRILLS11 finding F18: the melee cast guard (card fixportalmelee) counts attackers within 3 subtiles, but live melee reach is larger. Over 142 reads of 6 lured zombies, 3 monsters were within 8 subtiles in every run and never 3 within 3 (the third nearest at best 5.7), so the guard never fires at live distances and the bot still casts the portal at 34 percent life with no potion.

## What Changes

- New knob combat.crowdAdjacent (subtiles, default 8): the radius of "adjacent" for the crowd rule and the melee cast guard. The constant crowdAdjacent = 3 is gone; Crowd takes the reach as a parameter.
- No new policy and no change to the portal cast. The guard's decision (potion, else chicken, with 3 or more adjacent) is unchanged.
- The crowd rule's "surrounded" uses the same count, so it also sees monsters 4 to 8 subtiles away as in melee.
- Docs state the weak-hit premise is unverified (DRILLS11 cut-cast data).

## Impact

Offline fix; live proof pending (docs/fixcrowd.md). No spec names crowdAdjacent or the guard, so no spec delta (grep openspec/specs).
