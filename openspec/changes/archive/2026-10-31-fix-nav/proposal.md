## Why

Post-fix soak 1 (docs/probes/postfix-soak1.md, Sorceress, d2d): 16 of 33 waypoint-to-town attempts ended `the level changed during the walk 6 -> 20` and the picker hid Dark Wood and Black Marsh behind `no activated waypoint` (60 lines). Log reading (docs/fixnav.md): the character arrived in Black Marsh inside the Forgotten Tower mouth and the same step-away (identical from/to) was repeated and bounced each time; and the town's waypoint-to-town trip touched waypoints the runner never learned about.

## What Changes

- internal/nav walk.go: a step-away from a cave mouth that ends in another level counts as a bounce for that exit; the next step-away from it turns the direction (0, 90, -90, 45, -45, 135, -135, 180 degrees), memory 10 minutes. Shared (nav), seed-dependent trigger geometry is d2d behaviour.
- internal/nav waypoint.go: touching or travelling to a waypoint is recorded (`DrainActivated`); internal/bot drains it into its waypoint memory before the picker and the travel planner read it. Shared.
- internal/nav act2hints.go: level-relative waypoint positions of towns 75, 103, 109 (measured, DRILLS4), used only on d2d (level map with an Origin). Game.exe untouched.
- Delta: backend-contract ADDED requirements.
