# fixcrowd: the melee cast guard uses the real melee reach (card fixcrowd, DRILLS11 F18)

> status: current. Offline fix; live proof pending. Tags: V = verified (how stated), S = secondary, U = unverified.

## Problem (V, DRILLS11 logs `2026-11-01-drills11-portalmelee1..3.txt` on branch agent/drills11)
`meleeCastGuard` (docs/fixportalmelee.md) fires with 3 or more hostiles "adjacent", and adjacent was `crowdAdjacent = 3` subtiles (marked U). Live, 6 zombies lured to the character, 142 reads: 3 monsters within 8 subtiles in all 3 runs, never 3 within 3 (the third nearest at best 5.7 subtiles; 7.4 in one run's log). So Safety stayed `town-portal` at 34 percent life, no potion, one scroll: the guard never fires live. The one-adjacent control drill passed (portal as before).

## Units and reach (derived)
- Units: `Target.Dist` is `dist(player.X, player.Y, unit.X, unit.Y)` (internal/combat/target.go), world subtile coordinates, 5 subtiles per tile, centre to centre (V, code read). The drill logs use the same unit.
- Reach: in the DRILLS11 reads, monsters in an attack mode (A1, A2) stood at 1 to 7 and 9 subtiles (57 of 58 samples at 1 to 7 in run 1: 1:14 2:15 3:6 4:5 5:7 6:4 7:5, 9:1). Centre to centre distance includes the body radius of both units, so a zombie hits from several subtiles. The bot's own melee skill reach is the same order: `meleeRange` 10 minus `RangeMargin` 3 = 7, and `defaultMeleeRange` 7 (U) (internal/combat). Default 8 covers every attacker sample but one (9) and equals `crowdRadius` (8, the crowd count). V for zombies (drill data), U for other monsters: the monsters table in the repo carries no melee range (internal/data has missiles only), so no per-monster value exists.

## Fix
- Knob `combat.crowdAdjacent` (subtiles, default 8, range 3 to 12, docs/knobs.md regenerated with `go run ./tools/knobdump`); the constant `crowdAdjacent = 3` is removed. `Crowd(p, hs, reach)` takes the reach; `Fighter.Safety` and the resource input (`resInput`) pass the knob.
- Decision unchanged: 3 or more adjacent in a fight, a TownPortal becomes a ready potion, else Chicken; AbortFight leaves the game. The portal cast itself is untouched.
- Side effect (intended): the crowd rule's `surrounded` (adjacent >= 3) uses the same count, so a crowd with 3 monsters at 4 to 8 subtiles now counts as surrounded: it drinks (or retreats) instead of casting the portal. `TestSafetyCrowdFromSnapshot` pins the reach to 3 to keep testing the crowd portal path.
- Override: set `combat.crowdAdjacent` to 3 for the old behaviour.

## What the cut-cast data implies (V, DRILLS11 D-castcut, 14 trials)
- The portal appears 320 ms (4 steps) after the click, the player is in cast mode 6 steps. Five zombie hits (damage 2 to 3) at steps 2 and 3 did NOT cut the cast (mode stayed 10, scroll spent once). A hit outside a cast staggers (mode 4). The `debug hit` cheat does cut the cast, so the cheat is not evidence.
- So the rule's premise (melee hits cut a cast) is unsupported for weak hits and untested for hard hits. The rule stays as a conservative safety guard: chickening and drinking are cheap compared with losing a character. A drill with HARD hits needs the fork debug verbs of card forkdbg (not on main yet); that is the open proof of the premise.

## Verified offline
`go test -p 1 -count=1 ./internal/combat ./internal/env ./internal/resource ./tools/knobdump`: `TestCrowdAdjacentReach` (1/5/8, 1/2.4/5.7, far monsters, boundary, reach 3 control), `TestSafetyMeleeCutCastDrillGeometry` (fake game: 1/5/8 Chicken, single adjacent and far monsters TownPortal, knob override 3 TownPortal), plus the unchanged `portalmelee_test.go` and `crowd_test.go`.

## Live proof (pending, not run)
Rerun the DRILLS11 group P drill on the worker host with the lab build, same as the README DRILLS11 section:
`D2D_AGENT=fixcrowd D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh fixcrowd --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- <built>/d2drill -group P -only D-portalmelee -label fixcrowd-portalmelee1` (repeat 3 times) and the control with `-only D-portalmelee.ctl`. The drill's wait for "3 within 3 subtiles" must become "3 within combat.crowdAdjacent" (internal/drill, not owned here). Expect Safety `chicken` with 3 zombies in reach and the scroll unspent; the control still `town-portal` with the scroll spent.

Follow-up (a later drill agent, after drills11 merges): change the drill's wait from '3 within 3 subtiles' to '3 within crowdAdjacent' in internal/drill (owned by the drills cards).
