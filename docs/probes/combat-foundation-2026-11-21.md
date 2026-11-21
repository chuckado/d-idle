# Combat engine foundation group, 2026-11-21

> status: record of the foundation group of the approved combat engine plan (docs/design/fable-review-2026-11-21/combat-engine.md, appendix D). Offline only: no game, no live check. Tags: V verified offline by the named test, U not verified.

## Cards and merges

| card | task | branch | merged head (merge commit) | what |
|---|---|---|---|---|
| (all 17 cards) | TASK-718..736 | agent/combatcards | 35960ca36 | cards dlv-data .. dlv-cleanup created, dlv-whirl parked |
| dlv-data | TASK-718 | agent/cmb1 | a698c969b | cap/2 delivery object, generated views and tiers |
| dlv-rotation | TASK-719 | agent/cmb2 | 87636c8c1 | Rotation type, Profile.Rotation adapter, Fighter reads slots |
| dlv-expect | TASK-720 | agent/cmb3 | b442e454f | cast expectation ledger, knob combat.deliveryExpect |

Not done: dlv-reads (TASK-721). It needs fork read verbs (fork branch didle/combat-reads, lab binary) and a live measurement of ticks per second; it waits for a brief that names the fork and an instance.

## Parity evidence (V)

- dlv-data: `internal/delivery` TestGoldenParityWithTheOldTables compares the generated combat view and power skillKinds, entry by entry, with copies of the two hand-typed tables as they were at the card base (testdata/skills.golden.json, skillkinds.golden.json): 0 differences, 0 listed corrections. TestGeneratedViewsAreCheckedIn is the CI check (`go run ./tools/capgen -check`).
- dlv-rotation: `internal/combat` TestRotationEqualsTheProfileLists compares attack names, buff groups and aura ladder of every profile at clvl 1..99 (hero knows everything, hero knows nothing) with the lists the Fighter read before; TestRotationParityFailsOnAReorder shows it can fail (Sorceress and Paladin).
- dlv-expect: TestExpectLogOnlyChangesNoDecision: knob 0 against 1 gives the same keys, moves, casts and virtual time on the simulator; mode 2 acts only on a refusing simulator.
- Every card: `go test -race -count=20` on the touched packages, `go test` of internal/bot and cmd/d2bot, tools/knobdump (dlv-expect), the merge gate (go-whole, check.sh, leakcheck).

## Tier changes of dlv-data (rule of Q2, proven.json is empty today)

supported to experimental (no green Proving Ground row yet): Fire Bolt, Ice Bolt, Charged Bolt, Static Field, Frozen Armor, Zeal, Might, Holy Fire, Bone Armor, Raise Skeleton, Clay Golem, Raise Skeletal Mage. Experimental to supported (passives, nothing to deliver): Golem Mastery, Summon Resist. Shape Shifting stays unsupported (driver none). The agency planner is shadow-only; the one visible effect is the aura tie-break order of the shadow plan (supported auras no longer sort first). Engine decisions are unchanged on the corpus.

## Interface decisions others must know

- capabilities.json is schema didle.cap/2; the old `delivery` string is now `shape`; the new `delivery` object has purpose, use, reach, observe, power, driver. No `tier` key: it is computed at load. Edit the table, then `go run ./tools/capgen`.
- Rows of Amazon, Assassin and extra Barbarian/Necromancer skills were generated from the old combat table; their class is S (game knowledge, no skills.txt on the build box) and their driver follows the old kind.
- Knob `combat.deliveryExpect` (0 off default, 1 log only, 2 demote after three failed casts, back-off 20 s doubling to 5 min). The four ad-hoc memories (capSeen, curse tries, leftFail, buff cut) are not folded in yet (U).
- Pending, not owned here: Proving Ground scenarios combat.parity-rotation and combat.dead-skill-falls-back; wiring agency.ShadowDiff into the shadow log line (cmd/d2bot); a `go run ./tools/capgen -check` line in tools/ci.

## Proposed decision text (not written)

"Skill delivery data has one source, the capability table (data/capabilities/capabilities.json); other skill tables are generated views and tiers are computed, never typed."
