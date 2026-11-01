## Why

fixcrowd (finding F18) moved the melee cast guard's reach from 3 to the knob combat.crowdAdjacent (default 8). Its live proof was open: the DRILLS11 drill itself waited for 3 monsters within 3 subtiles. DRILLS12 changes the drill and runs it live.

## What Changes

- The D-portalmelee wait is "3 monsters within combat.crowdAdjacent", read from the tuning the harness hands the bot (never a copy of the number); the 1-adjacent control is unchanged.
- The drill counts reach with the bot's own measure (snapshot act subtiles), prints the bot's view, and waits again after the low-life cheat (bounded).
- Results in docs/probes/drills (drills12 files, README section DRILLS12), primitives rows, `live:` lines on fixcrowd and fixportalmelee.
- No product code changes.

## Impact

Overlap (declared as DRILLS11 did): cmd/d2drill, internal/drill, docs/probes/drills and docs/primitives.md are covered by the standing drills globs of m1-liveproof.
