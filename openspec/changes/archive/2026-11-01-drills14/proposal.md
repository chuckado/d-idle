## Why

DENBOUNCE (docs/fixdenbounce.md on agent/fixdenbounce): the bot enters a cave level and is back outdoors milliseconds later with nothing at INFO level between. The call that leaves the cave was unproven (sweepPass, nav mouthStep, the d2d mover).

## What Changes

- Drills DW.m3 (Blood Moor to Den of Evil) and DW.m4 (Cold Plains to Cave Level 1) in group F: the bot's exit walk into the cave, then idle, Fighter.Clear and Navigator.Explore one at a time; a phase that ends outdoors is the bounce and its debug log (nav walk callers) is printed.
- A debug-only diagnostic in nav mouthStep (`nav mouth check`, `nav mouth exit`): no behaviour change, with a test.
- Live logs `docs/probes/drills/2026-11-01-drills14-*.txt` and a DRILLS14 section in docs/probes/drills/README.md.
- No fix: the smallest fix is proposed in the README section.
