## Why

DRILLS24 measured 18 ticks per hop (not 64), so the 36 subtile minimum leg of teletune saved mana, not time; the straight-line hop cell lengthened Halls of the Dead 1 (755 against 578 ticks); and the leg rule kept Teleport from leaving a map-walkable pocket. docs/teletune2.md has the derivation.

## What Changes

- internal/nav: the hop landing cell is the farthest planned-path cell within cast range; a hop is taken only if it skips more than `nav.teleMinLeg` of path (default 13 = 0.5625 x 18 + 3). After a walk stall at the character's spot the leg rule is off and the mana floor is `nav.teleEscapeReservePct` (new, default 15). Greedy cell kept for islands.
- internal/env: knob default and the new knob (docs/knobs.md regenerated).
- Unchanged: level rules, towns, border exits, boss teleport stops, no-Teleport characters, `teleportBadHops` (now proven by a unit test).

## Impact

Offline only; awaits the live legs and the pocket in docs/teletune2.md. No spec names hop rules, so no delta.

Overlap (declared): card teletune (still open) owns internal/nav/teletune.go and docs/teletune.md; this card edits internal/nav/teletune.go and teleport.go (plus teleport_test.go, internal/env/tuning_nav.go, cmd/d2drill/main.go help text, docs/knobs.md) without listing them in owns. The d2drill drills of drills24 stay as they are.
