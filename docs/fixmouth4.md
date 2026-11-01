# fixmouth4: one click gate for unwanted exits, on every click sent during a stay

> status: IMPLEMENTED OFFLINE, live proof pending. Evidence: DRILLS18 section of docs/probes/drills/README.md (branch agent/drills18, log 2026-11-01-drills18-m31). Follows docs/fixmouth.md, docs/fixmouth2.md and docs/fixmouth3.md. V = read from those logs, S = simulated here, U = unverified.

## Why three iterations did not converge

Each of fixmouth, fixmouth2 and fixmouth3 closed the bounce site the last drill showed, and each did it the same way: a filter on a PLANNER, evaluated at the START of what that planner decides.

| Card | Guard | Applied at |
|---|---|---|
| fixmouth | `mouthStep` step-away with `mouthArrivalR` | the first observation of each walk |
| fixmouth2 | `NearMouth` for the Fighter, evade cap | the Fighter's move target, the evade start |
| fixmouth3 | stay guard, `leavesExit`, one-click step-away, roam filter 17 | the step-away target at its original start; roam targets; Fighter targets |

The character is not only moved by planners. It is moved by every click the bot sends, and many of those clicks are not the click a planner checked: the walk's replan after a monster stall, the stuck ladder (re-click, side steps, random hop, retreat from a pack), a fight that runs INSIDE a walk through the `Yield` hook, and basic-attack swings, which in D2 walk the character to the monster (or to the ground point when the monster moved). Knockback and blocked steps change the start point after the check. So a guard checked once at plan start is stale as soon as anything else moves the character, and every new drill found the next mover. DRILLS18 shows all of these in one chain (V, m31 log):

1. `Fighter.move` to (37,150), 15.1 from the exit (35,135): just outside `mouthApproachR` 15, allowed.
2. The walk's `mouthStep` at (37,148), 13.2 from the exit, steps away to (40,168) (`stay=true r=16`): checked by `leavesExit` at 13.2. Correct.
3. The step-away is blocked by a monster (`nav stuck`, `stalled by monsters, yielding`): the fight now runs inside the step-away walk.
4. The fight's approach moves are refused (`combat move skipped: the point is in a cave mouth`, 7 times), but its basic-attack swings (no bound attack skill; the left click on the Fallen at the mouth) walk the character to (36,138), 3.2 from the exit. No guard looks at a swing.
5. The step-away walk replans from (36,138); the new path follows the corridor toward the exit and its first waypoint click (37,140), 5.4 from the exit, fires the trigger 16 ms later. `leavesExit` was evaluated at step 2 only.

The structural answer: the guard must be an INVARIANT ON EVERY CLICK the bot sends while a stay guard holds, not a filter per planner. Then it does not matter which planner, recovery or fight produced the click, nor where the character was when it was planned.

## Design

### The click gate (internal/nav/mouthgate.go)

`nav.New` wraps the controller: `n.in` is a `mouthGate` that embeds the real controller and overrides `ClickWorld`. Every world click nav sends (walk follow step, stuck ladder, side steps, hop, retreat, spiral, pocket leave, gate push, waypoint approach) goes through it; no call site has to remember. A test pins this (`TestEveryNavClickGoesThroughTheGate`).

Active when (`n.gateActive`): the current level is not a town and either the stay guard holds for it (`stayLevel`, fixmouth3) or the arrival flag is set (fixmouth). Not tied to `mouthGuard` (config, off by default): overworld levels list exits on their edges, and a gate there would clip walks along the edge for no observed failure. Exits with `Dest == wantExit` are never gated (ExitTo walks onto them).

The check (`mouthClickOK(from, to, exit)`), with `from` the character's current position read at the moment of the click (`Pos()`, falling back to the last observation):

- `from` outside `mouthTriggerR`: the straight segment from `from` to `to` must stay at least `mouthTriggerR` from the exit.
- `from` inside `mouthTriggerR` (arrival offset, after a swing, after knockback): the segment must never get nearer to the exit than `from` is (`leavesExit`, half a subtile of slack), and must end no nearer than `from`. Moving outward is the only allowed motion.

A refused click is rewritten, never dropped silently:

- outside: CLIPPED to the last point of the segment that stays outside the bound (half a subtile of margin), when that is at least `gateMinClip` (2) subtiles long; the walk makes progress up to the boundary;
- inside, or no clip room: the AWAY vector (exit to character), rotated 0, +-30, +-60, +-90 degrees, length `mouthAway` (20) then 10, snapped to a walkable cell within 3, and the candidate must itself pass the check;
- nothing passes: HOLD. No click, the click returns `ErrMouthGate` (wraps `ErrNoPath`), logged with the reason. The walk's main click turns that into a failed walk; roam and explore already treat `ErrNoPath` as "drop this target", the Fighter's `move` logs "move failed" and spends a tick.

Bounded: at most `gateMaxRewrites` (3) rewrites per walk, then HOLD. A walk whose target lies behind the mouth therefore ends in at most three rewritten clicks plus one hold, instead of shuttling between the boundary and the away point. Never a loop, never a wait.

### The trigger bound

`mouthTriggerR` = 15 (the same value as `mouthApproachR`). Evidence (V): clicks that fired reached 13.0 (Den, DRILLS17), 11.7 and 11.2 (Cave Level 1, DRILLS17) and 5.4 (Den, DRILLS18) from the listed exit position; clicks 16.2 to 18 away never fired (DRILLS16). 15 covers 13.0 with a margin of 2 and stays under 16.2. The real bound, its shape (the trigger area is seed-dependent, docs/d2d-loop.md) and whether the game's own short path from a click can bend into it are U. The exact radius is no longer critical: the 15.1 target of step 1 is allowed, but every click of the walk to it is gated, so the walk cannot cross into the disc on the way.

### Swings (internal/combat)

The left-hand click of `cast` (basic attack, melee profiles, left skills) is a movement order: the game walks the character into melee range, or to the ground point when the monster has moved (V, step 4 above: the character went from (38,151) to (36,138) with only swings and refused moves in between). Before such a click `engage` asks the mover (`MouthClickOK(from, to)`, an optional interface like `NearMouth`); when it refuses, the swing is skipped with `combat swing held: the click would walk into a cave mouth` and the tick costs a poll (the lockstep clock must move). A swing at a target within `swingNoWalk` (2.5 subtiles) is allowed: the character is already in melee range and does not walk (U: the melee range of the basic attack in subtiles). Right-hand casts at a point (`CastAt`) do not move the character and are not gated (U for movement skills such as Charge or Leap Attack, which no profile binds today).

### The step-away replan

The step-away walk (`oneClick`) carries the exit it flees. On every replan it re-checks the first waypoint it would click with `leavesExit` from the CURRENT position; when the replanned path heads toward the exit, it computes a new away target from the current position (`mouthStep` again, which also turns the direction after a bounce) and walks there. The gate would rewrite the click anyway; the re-check keeps the walk's own goal honest so it does not fight the gate.

### Kept

Wanted exits are never gated or stepped away from (tested both ways). The fixmouth2 evade cap stays. `NearMouth`, the roam filter, `stayLevel` and `mouthStep` stay as fixmouth3 left them. Policy cards stay PAUSED: no change to which areas are visited or when the bot leaves.

## Tests (`go test -p 1 -count=1 ./internal/nav ./internal/combat`)

nav (`internal/nav/fixmouth4_test.go`):
- `TestMouthClickOKGeometry`: table of the logged geometries: the DRILLS18 chain (target (37,150) at 15.1 from (37,148), allowed only while the segment stays out; the replan click (37,140) from (36,138) refused; the step-away (40,168) from (37,148) allowed), DRILLS17 fired clicks at 13.0 and 11.7 refused, 16.2 to 18 allowed.
- `TestGateRewritesDrills18ReplanClick`: the gate rewrites the 5.4 click from 3.2 to a point that leaves the exit.
- `TestGateClipsAGrazingClick`, `TestGateHoldsAfterRewriteCap`.
- `TestGateProperty`: 2000 random positions and click targets in a synthetic level with one unwanted exit: no emitted click's segment violates the invariant while the stay holds.
- `TestGateIgnoresWantedExit`, `TestGateOffWithoutStay`, `TestGateOffInTown`.
- `TestStepAwayReplanRetargets`.
- `TestEveryNavClickGoesThroughTheGate`.
- `TestSweepWithPackAndBlockedStepAwayNeverCrosses` (S): Clear-like yield plus Explore sweep, a pack near the mouth whose yield drags the character to 3.2 from the exit (the DRILLS18 swing) and a step-away blocked by it: zero clicks violate the invariant.

combat (`internal/combat/fixmouth4_test.go`):
- `TestSwingHeldTowardMouth`: basic attack at a Fallen at the mouth from 8 subtiles: no left click while the mover refuses; time passes.
- `TestSwingAllowedInMeleeRange`, `TestSwingAllowedWithoutGate`.

The sims have no tile warps, so "zero level changes" is proxied by "zero clicks that violate the invariant".

## Live proof recipe (not run: no game in this task)

From the fixmouth4 worktree, `d2drill` built there, on the worker host (same commands as DRILLS18):

```
D2D_AGENT=fixmouth4 D2D_BIN=<fork>/bin/d2d.lab-cheats3 tools/agentrun.sh fixmouth4 --ttl 300 -- tools/d2dslot.sh run -n 0 -t 200 -- <built>/d2drill -group F -only DW.m3 -label fixmouth4-m3-N -v
```

N = 1, 2, 3, then `-only DW.m4 -label fixmouth4-m4`. PASS: in all four runs, zero `level changed during the walk` errors caused by a mouth bounce and zero `Entered` cycles within 1 s. The pocket-stall town portal escape is counted separately (see below). Baseline DRILLS18: one mouth bounce per DW.m3 run, DW.m4 zero. On FAIL attach the filtered log and check: `nav mouth gate` lines (verdict, from, to, rewritten point, distance), whether the fired click passed the gate (a fired click that the gate allowed means the bound is too small or the game's own path bent into the area), `combat swing held` lines.

Expected new log lines: `nav mouth gate: click rewritten` and `nav mouth gate: click held` (INFO), `combat swing held: the click would walk into a cave mouth` (INFO). A high count of held clicks in a run means coverage near the mouth is lost; compare `nav plan` and roam-walk counts with DRILLS18 (U: there is no per-point coverage count).

## The pocket stall at (90,171), area 1 (recorded, not changed)

DRILLS18 m31: in Den, at (88..90,171..172), `nav stalled by monsters, yielding` repeats (near 7 to 9 monsters) for about 20 lines, then the ladder's pinned check escapes by town portal. It is 60 subtiles from the mouth: not a mouth bounce. The character stands in a pack it cannot clear (`near` grows from 7 to 9) and the walk yields to the fight every time. This looks like the fight not finishing the pack (Sorceress without a bound attack skill, swinging the basic attack at 8 Fallen) rather than a nav pocket: the position does not change because the monsters hold it, not because the cells are sealed. So the FIXSTALL pocket escape did what it is for. Not a nav defect on this evidence (U: no fight log lines are in the filtered file); left as is.

## Overlap (declared)

internal/nav/walk.go and internal/nav/nav.go are owned by the open card fixmouth; fixmouth2 and fixmouth3 edited the same functions, internal/combat/fight.go is owned by an earlier card as well. This change adds `internal/nav/mouthgate.go`, `internal/nav/fixmouth4_test.go`, `internal/combat/fixmouth4_test.go` (owned here) and edits: in nav.go the `New` constructor (one line wraps `n.in`) and two Navigator fields; in walk.go the constants block (`mouthTriggerR`), `walkOpts` (one field), the walk's replan branch (step-away re-check), the step-away call and the walk start (gate counter reset); in fight.go one check before the final `cast` and the static opener in `engage`, plus `mouthClickOK` next to `nearMouth`. Merge fixmouth, fixmouth2 and fixmouth3 first.

## Shared-file lines for the merger

- docs/BACKLOG.md: FIXMOUTH4: one click gate for unwanted exits (every nav world click and every movement swing while a stay guard holds; clip, away rewrite, bounded hold), step-away replan re-check; offline tests pass, live DW.m3 x3 + DW.m4 pending (docs/fixmouth4.md).
- docs/INDEX.md: docs/fixmouth4.md.
- CHANGELOG: none until the live proof passes.
