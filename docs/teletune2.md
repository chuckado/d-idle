# teletune2: path-aware hops, stall escape, break-even at 18 ticks

> status: offline, awaits the live legs below. Builds on docs/teletune.md (whose 64 tick break-even this replaces) and DRILLS24 (docs/probes/drills/README.md, DRILLS24 section; logs `docs/probes/drills/2026-11-01-drills24-*.txt`). Tags: V = measured live, S = secondary, U = unverified, UNSURE = a guess that a knob moves.

## Revised break-even (V, DRILLS24)

Per-hop time `T` = 18 ticks on all 9 landed hops (V; neither the nominal 33 nor the 64 of DRILLS22, which was mostly mana waiting). Running speed `v` = 0.5625 subtiles a tick (V). So `v x T` = 10.1 subtiles: a hop beats walking in TIME whenever it skips more than 10 subtiles of walk.

Rule (code: `hopWorth`, `pathHop`): take a hop only if the landing cell is on the planned walking path and skips more than `nav.teleMinLeg` of that path, default `13 = v x T + margin 3`.

Margin (UNSURE, derived from the data): T had no spread (18 of 18 ticks), so the margin does not cover hop-time jitter. It covers what the model leaves out: the walk resumes from the landing cell after a stop and a fresh plan (two polls of 2.5 ticks is about 3 subtiles, S), and the teleport stop radius. DRILLS24 supports it from both sides: the declined leg (36 rule) cost 265 ticks against 254 walking, i.e. a leg that is only a few subtiles better than walking is not worth the risk; Stony Tomb with hops of about 17 saved 52 ticks over 3 hops (about 17 ticks, 9.8 subtiles per hop, against the 7 predicted by 17 minus 10). Raise the knob for more safety, lower it to 10 for the pure time break-even. The margin is a knob default, no skill or level literal.

Mana reserve is kept: `nav.teleReservePct` 40 (unchanged, 60 of 150; the 4th hop at mana 80 is declined, V DRILLS24).

## Path-aware hop cell

The old cell was the walkable cell nearest to the exit in a straight line. DRILLS24 Halls of the Dead 1: hops went straight toward the exit (189,186), (189,168), (189,150) and the walk from there detoured around walls: 755 ticks against 578 walking.

Now (`Navigator.pathHop`, teletune.go): plan the walking path from the character to the target (exit gate or frontier cell); among the path cells within the cast range (and at least `teleportMinHop` away, not a cell that already refused) take the one farthest along the path. The hop skips exactly that much path (`saved`, the length along the path to the cell). It is taken only when `saved > teleMinLeg`; a hop therefore never lands where the walk from there is longer than from here (test `TestPathHopNeverLengthensThePath`). Where a wall bends the path, a path cell beyond the wall is within range and the hop crosses it; on straight ground the hop is the usual 17 to 18 subtiles. Walls between here and the cell do not matter (d2d has no room limit; game.exe refuses hops outside the caster's room, SKTELE, so the range stays at the navigator's 18, U for game.exe).

When no complete walk to the target is known (island, void between) the plan is partial: the old greedy cell is used and the leg is its walk length (+Inf when none, so the hop passes, as in teletune). Unchanged: `teleStep` for islands.

## Stall escape

A walk stall is recorded by the stuck-spot memory (`noteStuck`, stuckmem.go: a stall at this spot within 10 minutes and 10 subtiles). `stalledHere` reads it. At such a spot a hop

- ignores the minimum leg (the map says the walk is short and it does not work: the real walk length is unknown, DRILLS24 pocket at Rocky Waste (165,220)), and
- uses `nav.teleEscapeReservePct` (default 15, about 23 of 150 mana, enough for three casts of a 5 to 8 mana attack, UNSURE) instead of the reserve; never above `teleReservePct`.

The DRILLS24 pocket hopped to Stony Tomb with 3 hops, mana 150 to 88.5, so 15 percent is far below what it needed; a bigger escape never has to drain the pool because the hop count is bounded: `teleportBadHops` 3 for refused hops (casts that do not move the character), unchanged. Level rules, towns, border exits, boss teleport stops and characters without a Teleport key are unchanged.

## Refused-hop bound

After the 36 leg rule, the DRILLS24 town leg no longer cast at all, so the live bound was not exercised. `TestRefusedHopBoundIsExercised` (leg rule and reserve off, casts do nothing) proves exactly `teleportBadHops` casts and then walking.

## Tests

`go test -p 1 ./internal/nav ./internal/env`: `TestPathHopNeverLengthensThePath` (Halls shape: wall with a far gap; the straight cell leaves a longer path than the path cell), `TestStonyTombHopsThenManaReserveStopsThem` (open 380 subtile leg: 3 hops, mana 150 to 78, the 4th declined, no wait), `TestShortLegsWalk`, `TestStallAllowsHopBelowNormalFloor` (stall: short leg hops, 60 mana ok, 40 mana refused), `TestRefusedHopBoundIsExercised`, `TestHopWorth` (escape rows), `TestMinLegDefaultMatchesBreakEven`.

## Live recipe

Same slot, cheats and command line as DRILLS24 (group T, `-label NAME -v`, `-only D22-C,D22-A,D22-B,D22-T,D22-R,D22-P`), built from this branch, DEFAULT knobs (do not pass `-tele-min-leg`):

1. Legs Rocky Waste to Stony Tomb, Dry Hills to Halls of the Dead 1 and the border leg, Teleport learned. PASS: ticks <= walking (254, 578, 708) on all three.
2. Pocket start Rocky Waste (165,220) with Teleport (D22-P). PASS: the character reaches Stony Tomb by hop (the stall escape, `nav teleport after a stall, leg rule off` in the log), mana never below 15 percent of 150 (23) after a hop.
3. Controls: no-Teleport legs (254, 578, 708), the town leg 0 casts; run the drills alone via `-only` (the navigator keeps its stall memory between drills in one process).
Also read the per-hop `hopticks` lines to confirm 18.

Note: with the default 13 the first leg (open ground) now takes hops until the 60 mana reserve (expect 3 hops, about 202 ticks). If Halls still loses, the log shows which hops were taken (`aimed` cell against the path).

## Shared-file lines for the merger

- INDEX: docs/teletune2.md: Teleport hops along the planned path, minimum skipped path 13 subtiles (v x 18 ticks + 3), stall escape, knob nav.teleEscapeReservePct.
- CHANGELOG: Teleport hops now land on the planned walking path and are taken when they skip more than 13 subtiles of it (was 36 subtiles of walk to a straight-line cell); after a walk stall the leg rule is off and mana may fall to 15 percent (new knob nav.teleEscapeReservePct).
- BACKLOG: live proof of teletune2 (legs and pocket, recipe in docs/teletune2.md).
