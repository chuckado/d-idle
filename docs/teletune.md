# teletune: when a Teleport hop is worth it

> status: card teletune, offline, NOT verified live (recipe at the end). Evidence: DRILLS22 (docs/probes/drills/README.md on branch agent/drills22, logs `docs/probes/drills/2026-11-01-drills22-*.txt`), docs/research/sk-teleport-trace.md (SKTELE), docs/fixmanamodel.md. Tags: V = measured live on d2d (DRILLS22) or read from game.exe, S = secondary, U = unverified, UNSURE = a guess that a knob lets you move.

## The finding (V, DRILLS22)

With Teleport (level 1, clvl 18, 150 max mana) the navigator hopped on tile exits and lost to walking: Rocky Waste to Stony Tomb 7 hops, 448 ticks, mana 150.0 to 3.0, against 254 ticks walking; Dry Hills to Halls of the Dead 1 5 hops, 851 ticks, mana 150 to 69.9, against 578. A hop covers about 17 subtiles (the navigator's range 18) and pays 21 to 24 mana (23.3 measured; 24 at level 1, one less per level, V read SKTELE). Border exits never hop (`exit.go`: `!g.border`).

## Break-even

Let `v` = running speed, 0.5625 subtiles a tick (V live, DRILLS), `T` = ticks one hop takes (settle 700 ms = 17.5 ticks, the cast animation, one observation), `W` = the A* walking length from the character to the hop's landing cell, `m` = mana now, `M` = max mana, `c` = cost of one cast (`data.Skill.ManaCost` at the learned level), `r` = natural regeneration (internal/resource `ManaPerSec`; 150 max gives 12 256ths a frame = 1.17 per second, 45 max gives 0.29).

A hop pays only if both hold:

1. Time: `W / v > T + wait`, where `wait = max(0, c - m) / r`. The old navigator accepted `wait` (it slept for mana up to 40 s). The new one never waits: with no mana for the hop it walks, so `wait = 0` and the rule is `W > v * T`.
2. Mana: `m - c >= reserve`, `reserve = ceil(teleReservePct / 100 * M)`, so the fights after the hop still have their casts.

`T`: nominal 33 ticks (17.5 settle + about 13 cast frames, S + 2.5 poll, U) gives `v * T` = 19 subtiles; DRILLS22 measured 448 ticks / 7 hops = 64 per hop including the mana waits (V). The default leg uses 64 (the measured figure, not the nominal one): `teleMinLeg = 0.5625 x 64 = 36 subtiles`. Reading: a hop lands at most 18 subtiles away, so it is worth it only when the walk to that cell is at least twice as long, i.e. across a wall, a river or a loop in the cave, not along open ground.

### Why 7 hops on 150 mana lose (V arithmetic on DRILLS22)

- Stony Tomb: 254 ticks = 143 subtiles. Seven hops cost 7 x 23.3 = 163 mana against a 150 pool: 13 mana short at 1.17 per second is 11 s = about 280 ticks of waiting, which with the hop time is the 448 measured. The mana ran to 3.0, so a fight at the end had nothing.
- Halls of the Dead 1: 5 hops cost 116 mana, the pool covered it (150 to 69.9 with regeneration), so mana was not the limit there: 851 against 578 is hop overhead and detours (UNSURE: the split between settle time, hop-then-walk-back and the 10 subtile stop of `teleportToward` was not logged per hop). Both legs are straight open ground: every hop's walk to its landing cell is the 17 subtiles it jumps, below 36.
- The model `TeleLegTicks` (internal/nav/teletune.go, test `TestBreakEvenDrills22Legs`) reproduces the direction for Stony Tomb with 64 ticks a hop. The new code takes 0 hops on both legs (straight, `W` = 17 < 36), so the leg is the walk (254 / 578 ticks).

## What changed (internal/nav/teletune.go, teleport.go)

- `hop()` asks `hopDeclined` before casting: `hopWorth` (pure) returns false for "leg" (`W` below `teleMinLeg`) or "mana" (below the reserve). `W` is +Inf when no known walk reaches the cell (island, void): those hops always pass the leg rule (and the old `teleStep` path for islands is unchanged, it may still wait for mana: islands cannot be walked).
- A declined hop returns "walk this leg": `teleportToward` ends and the walk goes on; `teleLeg` returns not handled without marking the frontier failed and without the blind-cast fallback. No idle wait for hop mana in `hop()` (the `Refill` hook is only reached by `teleStep`).
- Cost: `Navigator.TeleportCost` (wired in `cmd/d2bot/telewire.go` from `data.Skill.ManaCost` at the player's effective level), fallback `TeleportMana`, fallback 24.
- Knobs (`env.NavTuning`, `nav.teleReservePct` default 40, `nav.teleMinLeg` default 36; docs/knobs.md regenerated): a build or profile overrides them. No class, level or skill literal: speed and hop time are the constants above, cost and learned level come from the skill data and the player.
- `teleReservePct` 40 (UNSURE): a short fight is about eight casts of an attack costing 5 to 8 mana (Lightning or Fire Bolt at mid levels, `castsPerMonster` 1..6 in docs/knobs.md), which is 40 to 64 mana, 27 to 43 percent of 150. A build with a bigger pool or a cheaper attack can lower it; a melee build has no mana to protect and should set 0.
- Unchanged: which exits are used, border exits (never hop), boss teleport stops, refused-hop bound `teleportBadHops` 3, the level rules (`TeleportOK`), characters without Teleport (no key: every caller walks, test `TestNoTeleportKeyWalksWithoutError`).
- Existing tests that assumed hops on open ground now call `oldHopRules` (both knobs 0); `TestTeleportWaitsForMana` and `TestTeleportRefillHookIsCalledWhenDry` became `TestTeleportWithoutManaWalksInsteadOfWaiting`.

## What Teleport is for (do not claim more)

With a hop range of 18 subtiles and a 21 to 24 mana price at level 1, Teleport on the bot's own legs helps only where the walk is more than twice the jump: across walls, cut-off pockets and islands (`teleStep`), escaping, and the stand points of bosses that need it (`NeedsTeleport`, for example the Act 2 Summoner). It does not speed up open ground. Larger pools, higher skill levels (cost 24 minus level plus one) and a longer hop range on d2d (DRILLS15 T2: 15 to 24 cells landed; the cap is the navigator's `defaultTeleportRange`) would move the break-even; raising the range is NOT done here (game.exe refuses hops outside the caster's room, SKTELE: the room limit is 40 to 80 subtiles, but an unrefused 18 is the only measured safe value).

## Live recipe (not run)

Group T of d2drill, same slot and cheats as DRILLS22 (`debug stat 12 18`, `debug stat 9` = 150, `debug points 54 1`, `debug mana full`; command line in the DRILLS22 section of docs/probes/drills/README.md), built from branch agent/teletune:

1. Legs Rocky Waste to Stony Tomb and Dry Hills to Halls of the Dead 1 with Teleport learned, default knobs. PASS: ticks <= the walking control (254 and 578) OR fewer hops than DRILLS22 (7 and 5) with the mana never below `teleReservePct` of 150 (60), and 0 refused hops, 0 errors. Expected: 0 hops and the walking ticks (the legs are straight).
2. Repeat with `nav.teleMinLeg` 0 and `nav.teleReservePct` 40: expect hops until mana falls to 60 (3 hops at 23.3 on 150: 150, 127, 103, 80, 57 refuses the 4th), then walking with no idle wait; PASS: mana >= 60 at every hop, no `nav teleport declined` followed by a sleep.
3. Control: the no-Teleport character walks (254 / 578) as D22-C.
4. Open question for the run: log `T` per hop (command to the next observation) to settle the 33 against 64 ticks (UNSURE above).

## Shared-file lines for the merger

- INDEX: docs/teletune.md: when a Teleport hop pays (minimum leg, mana reserve), knobs nav.teleMinLeg and nav.teleReservePct.
- CHANGELOG: Teleport hops are skipped when the walk to the landing cell is shorter than 36 subtiles or the mana after the hop would fall below 40 percent of the pool; the navigator no longer waits for mana before a hop (it walks).
- BACKLOG: live proof of teletune (group T legs, recipe in docs/teletune.md); per-hop time log to replace the 64 tick estimate.
