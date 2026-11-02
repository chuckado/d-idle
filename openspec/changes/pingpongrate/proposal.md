## Why

Gate 1 rerun 2 (docs/soaks/gate1r2-20261101.md, Review 2 sections 3 and 4): the ping-pong guard fired twice correctly (windows A and B, 847 and 830 virtual seconds) but missed cluster C: 52 virtual minutes, 8 switches between Blood Moor and Cold Plains, 26 kills worth 18 XP. `noteArea` (internal/bot/goal.go:1196, `pp.kills != kills`) restarts its record on any kill, so a trickle of kills in emptied fields keeps it from firing.

## What Changes

- internal/bot/goal.go `noteArea`: a kill resets the record only when it is progress: the kills of the span must add up to at least a floor of XP or gold (a knob, default from the numbers above), otherwise the alternation count and span continue.
- Test: a replay of cluster C (8 switches over 52 virtual minutes, 26 kills of 0 to 2 XP) must fire; windows A and B must still fire; a field with real kills must not.

## Impact

No spec delta. Overlap (declared): soakgiveup (merged, the guard's code) is extended.
