# botthreat: missile threat model, shooter priority, runner guard

> status: current. Offline only (mem.Fake / table-driven tests); no live run. Findings come from docs/research/monster-missile-trace.md (B1-B4) and monster-ai-trace.md (B1-B3). Tags: V verified (how), S secondary, I inference, U unverified.

## What exists (internal/combat/threat.go)

- `ThreatTable` keyed by missiles.txt id from `db.Missiles` (`SetTables` builds it; nil without a DB). Speed = Vel x 0.046875 x 25 subtiles/s (S: trace Q6; reproduces 0.14 s arrows and 0.43 s shaman fire at 4 subtiles, tested). Blockable is I: ToHit or MaxDamage > 0 (the trace's V list: all ToHit rows are blockable; shaman, mage, spit rows are not).
- `Classify(dist, reactSec, shooterKnown)`: time to reach >= reaction budget -> dodge; else kill-shooter if the owner is in view, else ignore. Unknown row or Vel 0 keeps the old rule (dodge).
- `RankIncoming` returns dodgeable positions (highest row damage first) and owner ids of undodgeable missiles. Owner type 1 = monster (S, as `hostileMissiles` already assumes).
- `ShooterPeak(mon, difficulty)`: largest of A1/A2/S1 max and the El slots; unblockable (I) = ranged, no physical column, an elemental one. Caveat: El slots are the Normal column set, so Hell peaks are understated; the threshold knob is the lever.
- `Runs` / `Stands`: AI column sets. Runners zombie, brute, bloodhawk, corruptrogue (V run, AI trace D1-D4), corruptlancer, arach (spelling U, D4/D5), plus Flying. Stander: fallen (S, mon-wake-trace).

## Knobs (docs/knobs.md)

| knob | default | why |
|---|---|---|
| combat.runnerGuard | 1 | no kite, step-back or curse retreat while a runner is within range: V run shows these AIs run after the player (effect on speed U) |
| combat.threatDodge | 0 | filter the radius dodge through the table and make owners of undodgeable missiles kill-first; U until drilled |
| combat.dodgeReactMs | 400 | poll plus click budget (I, 100-300 ms measured nowhere) |
| combat.shooterKillMin | 0 | kill-first for unblockable ranged monsters with peak >= value; 0 off (the derivation is I) |
| combat.standOK | 0 | a standing Fallen is not dropped as `no-shot` (S); approaches stay bounded by maxApproaches |

## Not done

- Movement primitives support only evade-to-point and approach; there is no LOS-break move, so "fast shooter" means kill-first (the target gets closed on by the normal approach), not hiding.
- B1 retreat on d2d: the d2d backend feeds no `Snapshot.Missiles` (missile trace B6), so the dodge half is moot there.

## Live drill checklist (needs the game; not run)

1. Log `snap.Missiles` classes in flight near a Skeleton Archer, Quill Rat, Fallen Shaman: confirm class ids equal missiles.txt ids and OwnerType 1 for monsters (U: client table unverified while a missile is in flight).
2. Per shooter, time from the missile first seen to hit; compare with the table's speed; set `dodgeReactMs` from measured poll plus click latency.
3. With `combat.threatDodge 1`: a Skeleton Mage / archer pack at NM; count dodges and hits taken versus 0.
4. Zombie and Brute (Burial Grounds, Cave): record distance over time while the bot would step back; confirm contact is not broken, then keep `runnerGuard 1`.
5. A Fallen standing at 10+ subtiles: with `standOK 1` check the approach, no `no-shot` ignore.
6. Skeleton Mage in NM/Hell: read monstats ranged flag and El columns to pick `shooterKillMin`.
