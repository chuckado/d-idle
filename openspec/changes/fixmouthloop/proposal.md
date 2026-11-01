## Why

DRILLS19 (live proof of fixmouth4) passed the gate itself, but after the gate's 3-rewrite cap the walk fails with `ErrMouthGate` and the Fighter restarted a step-away at once on every engage tick: 117 times at (36,140) in one Den cycle until the phase limit, because monsters blocked the away route. Nothing logged it. Details: docs/fixmouthloop.md.

## What Changes

- `internal/combat`: `Fighter.move` remembers a walk that failed with `nav.ErrMouthGate` (position, time, count), logs one debug line per episode, and does not start another walk from that spot for 3 s (first 3 holds) and then 30 s; a skipped walk costs a poll and the fight goes on in place. Position change of more than 2 subtiles or a successful walk clears it.
- Tests (sim with a held mover: 120 walks without the back-off, 3 with it).

## Impact

Movement guard only; no change to which areas are visited or when the bot leaves (policy cards stay paused). fixmouth2's evade cap and fixmouth4's gate are unchanged. Needs a live DW.m3 rerun.

Overlap (declared): internal/combat/fight.go (move: one check before GoTo, one note after) and internal/combat/fighter.go (one field) are owned by open cards (fixmouth2 and earlier) and are not listed in owns.

## Shared-file lines for the merger
- docs/BACKLOG.md: FIXMOUTH5 (DRILLS19 hold loop): done offline in fixmouthloop (back-off in Fighter.move); live DW.m3 proof open. The drill DW.m4 setup defect is not covered.
- docs/INDEX.md: docs/fixmouthloop.md.
