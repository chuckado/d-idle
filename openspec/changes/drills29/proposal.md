## Why

Card armorcut (the retry of a Frozen Armor recast cut by a hit) was never exercised live: DRILLS28 saw no cut line in four runs, and its D26-A spacing rule failed by 4.7 s because of a harness dead zone. DRILLS29 holds a pack in melee at the due time and fixes the dead zone.

## What Changes

- Drill D29-C (group W): pack held adjacent at the buff due time, the bot's own Fighter.Buff, per-recast record (cmd/d2drill, internal/drill; harness only).
- D26-A: the bot's Buff pass also runs inside the spawn and lure loop, so the pack handling no longer starves the 114 s recast.
- Results in docs/probes/drills (drills29 files, README section DRILLS29), a primitives row, a `live:` line on card armorcut. No bot code changes.
