# fixrecast: recast armors by their real duration and state (card fixrecast)

> status: current. Offline fix for finding F2 of [research/sorc-skill-matrix.md](research/sorc-skill-matrix.md). Tags: VERIFIED OFFLINE = a test in this change; UNVERIFIED LIVE = needs the recipe below; S = read in code/data; I = inferred.

## Evidence
- Skills.txt rows (1.14d, read from the loaded table; the row text is the evidence, S): Frozen Armor `auralencalc` = `ln34 + (Shiver + Chilling points) x par7` with Param3 3000, Param4 300, Param7 250. `lnAB` is Param A + Param B x (level-1) frames, 25 frames per second (the same form gives the armor percent, `ln12` = 30 + 5 (L-1)). So Frozen Armor lasts **120 s at level 1, 132 s at level 2** (3000 and 3300 frames), plus 10 s per synergy point. Shiver Armor: same Params (3000/300). Chilling Armor: Param3/4 = 3600/150, 144 s at level 1. Energy Shield `ln12`: 3600 + 1500 (L-1) frames = 144 s at level 1, 204 s at level 2. Bone Armor and Cyclone Armor have no `auralencalc` (they end when the absorb is used up).
- Old behaviour (S, `Fighter.buffDue`): every armor except Bone Armor was recast `BuffRecastSec` (150 s) after the last cast. Frozen Armor was therefore down 30 s of 150 (level 1) or 18 s (level 2): 20 % and 12 %.
- The state is readable: `Player.States` (the state bit array of the player's stat list). Bone Armor state 14 is VERIFIED LIVE there (docs/summon.md, docs/classes/smoke-necromancer.md). Frozen Armor is states.txt `frozenarmor`, id 10 (S: `d2.StateFrozenArmor`, stock row order; the probe `--probe-states` was written to find the array by casting it). That bit 10 switches on after a Frozen Armor cast is UNVERIFIED LIVE.
- The state is a bit, not a timer: it says up or down, never how long is left. So "about to expire" can only come from the cast time plus the real duration.

## What changed
- `data.Skill.AuraLenCalc`, `Skill.AuraFrames(level)`: the leading `lnAB` term (the synergy term is left out on purpose: it only lengthens the duration, so ignoring it recasts early, never late).
- `Fighter.buffDue` (internal/combat/necro.go), for every armor:
  1. State readable (Bone Armor 14, Frozen Armor 10): present = up. Bone Armor, no duration, stays "not due while present". An armor with a duration is due at `min(buffRecastSec, duration - buffMarginSec)` after the cast. Once seen present the state is learned; absent afterwards (4 s grace after a cast) = recast at once (killed or dispelled).
  2. Otherwise (state not learned, other armors, Shiver / Chilling / Energy Shield): the same limit by the clock. No duration known (no `auralencalc`, no catalog): the old `buffRecastSec` timer.
- New knob `combat.buffMarginSec` (6 s, 0..30); `combat.buffRecastSec` stays as the upper bound and fallback, so old configs keep working. Level used is the effective skill level of the spec (points plus item bonus). The limit never drops below 4 s.
- Clock: `Fighter.now()`, the bot's clock. Under lockstep the bot's clock runs at least as fast as game time, so an error only recasts earlier. No numeric spec gate changed (no spec names the recast interval; early-game-policy mentions armor only for buying). The cast rules are unchanged: `Buff` is still called where it was (before fights, in town); no new adjacency policy.

## Verified offline
`go test -p 1 ./internal/combat ./internal/data ./internal/env ./tools/knobdump`:
- `TestFrozenArmorTimerLeavesGap` (the old behaviour, a catalog without auralencalc: the armor is down for over 100 s of 600 s), `TestFrozenArmorRecastBeforeExpiryByClock` (levels 1 and 2: 0 s down, cast count matches duration - 6 s), `TestFrozenArmorRecastBeforeExpiryByState` (state readable: 0 s down), `TestFrozenArmorStateGoneRecastsEarly`, `TestBuffLimit` (114 s, 126 s, knob upper bound, margin 0), `TestBuffLimitFallbackNoDuration`, `TestAuraFrames` (Frozen Armor and Energy Shield by level), `TestBuffDueByState` (Bone Armor unchanged).
- Before the change the same 600 s run left the armor down 30 s per cycle.

## Unverified live
Frozen Armor bit 10 in `Player.States` after a cast; that a recast while up refreshes the full duration (D2 does; d2d's behaviour is U); the in-game timer equals the table (120 s at level 1) in d2d.

## Live recipe (hidden slot, throwaway Sorceress, every cheat ledgered)
1. `debug points` to learn Frozen Armor (one point), bind it, `cmd skill 40 x y` on the player; read the player state list (`dump`, the player's states) for state 10 and note the virtual tick.
2. Pause the lockstep (`d2dctl` pause) and `step` in 25-tick units to 100 virtual s, then 114 s: the bot's `buff skill=Frozen Armor` log line must appear by about 114 s, before the state drops at 120 s (check by stepping to 121 s with the bot idle: state 10 gone).
3. Level 2 (`debug points`, one more point): expect the recast at about 126 s and the state to last 132 s.
4. Leave the character in town, no panels.

## Primitives
Primitive P10-adjacent (buffs): see the row in [primitives.md](primitives.md).
