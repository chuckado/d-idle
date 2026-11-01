# armorcut: a Frozen Armor recast cut in a fight (card armorcut)

> status: current, offline fix, awaits the live proof. Tags: V = verified (read in a log or test), S = read in code or secondary, U = unverified.

## Evidence (DRILLS26 D26-A, docs/probes/drills/2026-11-01-drills26-a-armorhit.txt, on both lab binaries)
- Setup (V, from the log header): Blood Moor, Frozen Armor 1, Fire Bolt 3, zombie pack topped up, the bot's own `Fighter.Buff` then `Fighter.Clear` rounds for 300 game s.
- Casts at 0.0, 119.3, 232.4, 240.2, 242.8 s (V). The 119.3 s recast refreshed (state `40:1` at its read). The 232.4 s cast came 113.1 s after the previous one, so the first armor was still up: the state read at the end of that round is `40:1` (V) but that is the OLD armor; it ended at about 239.6 s (119.3 + 0.3 + 120.0, V by the drill's own sums) and the sample at 240.2 s shows `buffs=""` (V). So the 232.4 s cast did not extend it.
- Mana at the reads: 25 (after 119.3), 34 (after 232.4), 26 (after 240.2), 20 (after 242.8) (V). Frozen Armor costs far less than 34, and every other cast in the run was paid, so a short-mana fizzle is not supported: hypothesis (b) is out (V).
- The log is filtered to buff, kill, fight and level lines: it has NO hit events, no player mode and no per-cast action-frame read (V). So it cannot show that a hit cut the cast. It shows only: the cast was issued and counted by the bot (`combat buff` line), the state did not extend, and the bot's own state check fixed it 0.8 s after the lapse (recast at 240.2 s, and again at 242.8 s: the first of those two also seems not to have been seen present at once, the 4 s grace of `buffDue`).
- What the log supports: (c) the bot's bookkeeping marked the buff refreshed on the key press (`f.buffAt[...] = now` straight after `cast`, before any check) is the proven weak point (V, code: `buffPass`). It turned an unverified cast into "done for 114 s", and the state bit cannot say otherwise while the old armor is still up. Why the cast did not take is U: (a) a hit that cut the cast is the lead (a fight with a pack in reach, DRILLS13 M1: a cut cast is paid, not refunded); a fourth candidate (d), d2d not extending a state that is already set, is also open (U) and the log cannot tell it from (a).
- The state is a bit, not a timer (S, docs/fixrecast.md): while the old armor is up a refresh cannot be confirmed from `Player.States`. The only observable of a cut is the player's mode (hit recovery 4, V-d2d DRILLS15 M2, `modeGetHit` in internal/combat/hard.go).

## What changed
- `Fighter.buffPass` (internal/combat/fighter.go):
  1. Casts nothing while the player is in hit recovery (mode 4) or still casting (mode 10; S, U live): the next tick retries.
  2. After the key press `watchCast` polls the player mode for the cast gap plus 400 ms. If hit recovery was seen the cast counts as cut: the previous cast time is restored (no previous: deleted), so `buffDue` says due again and the fight loop retries at once. Bounded by `buffTriesMax` (3 consecutive cuts per skill); the cast itself still needs `sp.Cost <= mana` every time, so a dry bar does not spam. After 3 cuts the cast is accepted as done: the state check (absent for 4 s after a cast = recast) and the clock take over.
  3. The recast still starts `buffMarginSec` (6 s, data-driven by the skill's `auralencalc`, no literal) before the real duration, so a cut attempt has that lead to be retried in. No knob changed (docs/knobs.md untouched).
- Unchanged: Bone Armor and the other buffs go through the same pass; with no hit recovery the added code is a few snapshot reads inside the sleep the pass already did. Characters with no armor skip the pass entirely.

## Verified offline
`go test -p 1 -count=1 ./internal/combat`: `TestBuffCutByHitRetriesAtOnce` (cut cast at the limit is retried at once, a completed cast does not repeat), `TestBuffNoCastMidHitAndBoundedRetries` (no cast in hit recovery or cast mode; an always-cut cast is tried 1 + buffTriesMax times then left to the state and clock), `TestBuffCutRetryNeedsMana`. All older combat tests (fixrecast, fixbuff, necro) pass.

## Unverified live
That hit recovery is what cut the cast at 232.4 s (or whether d2d refuses to refresh a set state: candidate d); that mode 10 is the cast mode on d2d.

## Live recipe (hidden fork slot, the DRILLS26 D26-A harness, ledgered cheats as in DRILLS26)
0. Discriminate (d) first, no monsters: cast Frozen Armor at 60 s and again at 114 s on an idle character; `debug states` must still show `40:1` at 121 s to 125 s. If it is gone, d2d does not refresh a set state: the fix then is to recast only when the state is absent (a follow-up card), not this retry.
1. Rerun D26-A (same seed 11, 300 game s) on `d2d.lab-armorhit` and on `d2d.lab-nexthit2`, with the player mode added to the filtered log (`combat buff cut by a hit, will retry` lines are the new evidence, plus the existing `combat buff`).
2. PASS: the armor is never absent for more than one poll interval over 300 s on both binaries, and every `buff cut by a hit` line is followed by a `combat buff` within one round. FAIL with no cut line and an absent armor means candidate (d) or an unseen cause: report the mode column at the cast.
3. Leave the character in town, no panels, nothing running (`tools/agentreap.sh <agent>`).

## Shared-file lines for the merger
- BACKLOG: armorcut live proof (D26-A rerun, step 0 discriminates a refresh refused by d2d from a cut cast).
- INDEX: docs/armorcut.md, Frozen Armor recast cut by a hit.
- CHANGELOG: the bot retries an armor recast that a monster hit cut, and casts no buff during hit recovery.
