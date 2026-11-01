# TELETUNE3: the border leg, the stall escape, the stall memory between drills

> status: offline work of card teletune3; the live proof is the DRILLS26 recipe below (not run). Follows docs/teletune2.md and the DRILLS25 section of docs/probes/drills/README.md.

Tags: V = verified (log or test of this branch), S = read from code, U = unverified.

## 1. The border leg: 717 ticks with the key against 708 without, 0 hops

Question: where does a Teleport-capable character with 0 hops pay 3 to 9 ticks more than the control on Rocky Waste to Dry Hills?

Findings.
- S: no Teleport code runs on a border leg that walks. `hasTeleport()`/`teleKey()` are called in `exit.go:107` (`n.hasTeleport() && !g.border`: the key is looked up once per gate attempt, a closure call), `explore.go:69` (only when exploring: on d2d the whole map is known, so `ExitTo` finds the gate without exploring) and `probe.go:44`. `walk.go` never reads the key; `hop`, `pathHop` and `hopDeclined` are reached only through `teleportToward` (tile exits) and `teleLeg` (explore). So there is no per-step poll, plan or key lookup to remove. V: `TestBorderLegIsTheSameWithAKeyBound` (internal/nav, two keys): `ExitTo` over a fully known two-level border gate gives the same clicks, the same elapsed time and the same number of world polls with a key bound as without, 0 casts, at most 2 key lookups.
- V (DRILLS25 `t-default`, border leg of D22-C and D22-A, same start (171,229), same gate (399,77), border=true, both lines 'nav exit ... border=true' then 'nav plan' of 51 waypoints): the two walks click identical positions for 22 waypoints. At waypoint 23 the character stands at (296,197) in the control and at (295,197) with the key: one subtile of position noise. The control then replans at (315,193) (`nav plan ... wps=24 cells=148`, then 8 more replans on its way up x 320 to 380: 10 plans in all), the keyed walk makes no replan and follows the original 51 waypoint plan along x 357. Different routes, 708 against 717 ticks. The replan trigger is a change of the known cell count or of the monster signature (S walk.go, plan block); which of the two fired is not in the log.
- V: the spread over runs is as large as the gap. Keyed border leg: 687 (DRILLS22, 64 tick hops era), 714 (DRILLS24 default), 714 (minLeg 0), 717 (DRILLS25). The no-Teleport control was 708 three times, but each control ran after two walking legs and each keyed run after two hop legs, so the game state at the start of the leg differs (monsters, the d2d room state), which is where a replan trigger can come from.

Cause (U, best explanation): not the key. The border walk is a route choice that depends on a mid-walk replan; the keyed runs hit a different replan history than the controls. What would settle it: the new `why=` field of the `nav plan` debug line (first, goal, map, monsters, deviation, partial) and drill D22-BL (below), which walks the border leg three times in each arm from a fresh prep and prints every plan line with its tick stamp. If the spread inside one arm is about 9 ticks, it is walking variance; if the no-key arm is always 708 and the keyed arm always 714 or more, read the replan reasons of the keyed plans.

No fix applied: there is no poll or plan on the walking path to cache, so skipping a plan would be a change without a measured cost. The only change in nav for this item is the `why` field (log only).

## 2. The stall escape on one level

`stalledHere` (teletune.go) is true when a stall was recorded on the same level within `stuckRadius` (10 subtiles) of the character in the last `stuckMemory` (10 minutes). Then `hopWorth` ignores the leg rule and lowers the mana floor to `nav.teleEscapeReservePct` (15 percent).

V (internal/nav, `TestStalledHereThroughNoteStuck`, the real `noteStuck` API, two keys): no stall 0 casts on a 12 subtile leg; a stall at the spot or 9 subtiles away 1 cast; 30 subtiles away 0; on another level 0; older than the window 0, just inside it 1; with a stall, 60 mana (36 left, floor 23) 1 cast and 45 mana (21 left) 0.

Live drill D22-S (written, not run). Geometry, from the DRILLS25 pocket logs: the character stands at (165,220) after `debug warp 33 44`; the walk to the Stony Tomb gate clicks (162,221) and stays at (165,220) for the whole stall ladder (rungs 0 to 5), then ExitTo escapes by town portal. D22-S switches the escape off (`nv.Escape = nil` for the drill), sets `nav.teleMinLeg` to 200 (no hop can pass the leg rule, so a hop can only happen through the stall rule), learns Teleport and:
1. Call 1 `ExitTo(55)`: expects `nav teleport declined ... why=leg`, 0 hops, `nav stuck` lines, ExitTo returns the stall (the spot is noted).
2. Waits 3300 ticks (132 s: the stalled-gate mark of 2 minutes expires, the 10 minute stall memory does not).
3. Call 2 `ExitTo(55)` from the same cell: expects `nav teleport after a stall, leg rule off` and hops. PASS only if all of that holds; FAIL reasons are named (setup does not discriminate, no stall, no decline, stall did not fire).

## 3. Stall memory between drills

`Navigator.ResetStallMemory()` (stuckmem.go) clears the stall spots, hard blocks, stalled gates, declined-exit marks, refused hop cells, hop decline reason and the search memory (roam coverage, visited and failed frontier buckets, mouth bounces). The bot never calls it (V by grep: only cmd/d2drill). `D22()` calls it before every drill, so D22-PC then D22-P in one process no longer inherits the escape of the first (DRILLS25: 0 hops, 7560 ticks). V: `TestResetStallMemory` (memory empty after the call; a 12 subtile leg declined again; the same stall without the reset allows the hop).

## Live recipe (DRILLS26)

Same slot, cheats and binary as DRILLS25, built from this branch, default knobs except where stated. Run each alone with `-only` (D22-BL needs the character unlearned at its start, so it must be the first learning step of its process; it refuses otherwise):
1. `d2drill -group T -repeat 1 -label d22bl -v -only D22-BL`: three keyless and three keyed border legs. Read the `border-nokey-N plan:` and `border-key-N plan:` lines: tick stamp, from, plan size and `why=`. PASS criterion (measurement): keyed spread and keyless spread both under 10 ticks, or the gap explained by a plan reason.
2. `d2drill -group T -repeat 1 -label d22s -v -only D22-S`: PASS as in section 2. Capture the `stall-1` and `stall-2` lines.
3. `d2drill -group T -repeat 1 -label d22pp -v -only D22-PC,D22-P`: both pocket drills in ONE process; D22-P must hop (3 hops, about 233 ticks) now that the memory is reset between drills.
4. Regression `-only D22-C,D22-A,D22-B,D22-T,D22-R`: ticks 254/578 beaten on the tile legs, border leg unchanged within the spread of step 1.

## Shared-file lines for the merger

- BACKLOG: TELETUNE3 live proof (docs/teletune3.md recipe: D22-BL, D22-S, D22-PC then D22-P in one process).
- INDEX: docs/teletune3.md: border-leg analysis (key not on the walking path), stall escape tests and drill D22-S, `Navigator.ResetStallMemory`.
- CHANGELOG: none for the bot (nothing changes on the bot path; one debug log field `why` on `nav plan`).
