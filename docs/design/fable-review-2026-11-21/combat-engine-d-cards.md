# Appendix D. Rollout: sixteen cards and one parked

> status: REVIEW (card fablecombat, docs only). Part of [combat-engine.md](combat-engine.md). These are PROPOSED cards; none exists in `backlog/tasks/`. The coordinator creates them with `tools/backlog/new_task.py` if the owner agrees. Card ids below are free at `5b12ce5e7` (`grep '^card: dlv' backlog/tasks` finds nothing) [V].
>
> Effort is in agent-days of the kind this repo runs (one agent, one PR family), a judgment from comparable merged cards, not a measurement [U]. "Scenario first" means the Proving Ground scenario named is written and RED (or pending with its reason) before the Go change, per ADR 0020 item 1. Scenario ids are those of [appendix C](combat-engine-c-tests.md) section 3.

## 1. Rules for every card in this set

1. Shared data has one owner per wave (docs/agents/rules-short.md rule 2): `data/capabilities/capabilities.json` and `internal/combat/classdata/*.json` are owned by exactly one open card at a time. The order below respects that.
2. Each driver ships behind one knob, `combat.delivery.<driver>`: 0 = old code path, 1 = shadow (the driver proposes and logs, the old path acts), 2 = the driver acts. Rollback is the knob. The old path is deleted by a later cleanup card only after the scenario has been green for a stated number of nightly runs.
3. Every rule is tested with two builds; no card may add a class or skill literal; the literal count of appendix A section 3 is reported in each card's result and may only fall.
4. A card that needs a fork verb names it in `exclusive:` (`fork-branch:<name>`) and the fork work follows AGENTS.md (same-day push to the private remote, no patch in this repo, nothing upstream without the owner's go).
5. No card raises a capability tier by hand. Tiers move when the generator of card dlv-data says so.

## 2. Wave 0: foundation (no behaviour change)

| card | what | owns | gate (scenario first where one applies) | risk and rollback | days |
|---|---|---|---|---|---|
| **dlv-data** | Schema `didle.cap/2`: the `delivery` object of appendix B; generator from today's `skills.json`; hand corrections listed in appendix B section 3; rows for Amazon and Assassin generated as experimental from skills.txt; tier generator with the rule of question Q2 (reporting only until Q2 is answered); CI check that `skills.json` and `skillKinds` equal the generated views | `data/capabilities/`, `internal/delivery/` (new leaf package: row, loader, validator), `tools/capgen/` | L0: every row loads; generated views equal the checked-in files except the listed corrections; `agency` agreement fixture unchanged | risk: three readers of one file during the wave. Rollback: the new object is additive, readers ignore it | 2 to 3 |
| **dlv-rotation** | `Rotation` type and conditions (appendix B section 4); `FromProfile` adapter; the Fighter's attack order, buff groups and aura ladder read slots; `agency` lays its plan out as a Rotation and the shadow line compares slot by slot | `internal/delivery/`, `internal/combat/` (fighter, classdata, aura), `internal/agency/` | golden parity over the simulator tests and the combat scenarios of the corpus; scenario `combat.parity-rotation` | risk: silent reorder of casts. The golden test is the guard; rollback is a revert (one PR) | 3 |
| **dlv-expect** | Expectation ledger (main doc 3.4): record per cast, settle from later snapshots, per `(skill, reason)` counters, demotion with back-off; telemetry line; folds in `capSeen`, curse tries, `leftFail`, buff cut | `internal/combat/` (new file plus the four call sites) | L1 tests with a refusing simulator; scenario `combat.dead-skill-falls-back` (pending until dlv-reads) | risk: a false "no effect" demotes a working skill. Starts log-only (knob 1); demotion is knob 2 | 2 |
| **dlv-reads** | Fork: read verbs `minions`, `skillstate`, `combat events`, corpses and states in `monsters` (appendix C section 4), flag-gated, in `caps`. Bot: typed clients in `internal/d2dctl`; scenario runner assertion kinds of appendix C section 3 | `internal/d2dctl/`, `internal/scenario/`, `tools/provingground/`; exclusive `fork-branch:didle/combat-reads`, `d2d-lab-binary` | fork tests in the style of `input_verbs_d2d`; one scenario per new assertion kind that fails when the assertion is inverted | risk: the fork journal costs frame time; measure ticks per second with the flag on and off. Rollback: the flag | 3 (1 fork, 2 bot) |

Wave 0 total: 10 to 11 days. dlv-data first; dlv-rotation and dlv-reads can run in parallel after it (different owners); dlv-expect after dlv-rotation.

## 3. Wave 1: upkeep and marks (Necromancer Summoner, war cries, pets)

| card | what | owns | scenario first | needs | risk and rollback | days |
|---|---|---|---|---|---|---|
| **dlv-upkeep-state** | Upkeep driver, state-observed: any slot of purpose `upkeep` with `observe.by: state, aura or clock`; exclusive groups; vital flag; the in-fight share and the post-fight pass; state ids resolved from the tables; deletes `buffStates`, the curse skip list in the buff pass, the skill-named aura knob | `internal/combat/` (new `upkeep.go`; `fighter.go` buff section, `aura.go`), `internal/data/` (state column) | `combat.buffs-stay-up` for Sorceress and Barbarian | dlv-rotation; differential rows for the war-cry and Holy Shield states (appendix C section 5) | risk: recast spam when a state id is wrong; the expectation breaker bounds it. Knob `combat.delivery.upkeep` | 3 |
| **dlv-upkeep-minions** | Minion-observed upkeep: groups from the summon column, want from the pet cap, cap by evidence kept, corpse-target summons through the corpse aim rules, golem and spirit as exclusive groups, Valkyrie and Shadow as census rows; deletes `necro.go` summon planning, `druid.go` plans and class ids, `assassin.go` shadow code, the class switch | `internal/combat/` (`upkeep.go`, removes most of `necro.go`, `druid.go`) | `combat.army-holds` (Necromancer), `combat.pets-hold` (Druid) | dlv-upkeep-state, dlv-reads (`minions`, corpses as units); differential A11 for specials | risk: the Necromancer loop is the only part with a live check; the new driver must first reproduce it in shadow (knob 1) on the same scenario. Rollback: knob | 3 to 4 |
| **dlv-mark** | Mark driver: target choice by `when` (elite, pack), area cover from the row's radius, state observed on targets, one mark group per target; all ten curses become data; Taunt, Howl, Battle Cry, Cloak of Shadows as rows | `internal/combat/` (new `mark.go`; removes the curse half of `necro.go`) | `combat.curse-covers-pack` | dlv-upkeep-state; fidelity A2 verdict before any uptime claim | risk: on d2d the put rule deviates, so the scenario's floor is a d2d number. The card says so in its result | 2 to 3 |
| **dlv-stance-army** | Stance `behindArmy` and the army facts for the outer arbiter ("army below want", "upkeep due"); retreat when the census falls under the stance minimum in a boss fight (S8, S11 of the Necromancer note) | `internal/combat/` (position), `internal/arbiter/` (two facts, one source) | `combat.army-holds` extended with a boss row | dlv-upkeep-minions | risk: a hero that hides too long. Bounded by the cast-first watchdog. Knob | 2 |

Wave 1 total: 10 to 12 days. After it the capability rows of roadmap R10 (Summoner) and the upkeep half of R12 (Barbarian) can be judged by scenario.

## 4. Wave 2: point aim, corpse aim, stand in the pack (Druid fire, Blizzard, Hammerdin)

| card | what | owns | scenario first | needs | risk and rollback | days |
|---|---|---|---|---|---|---|
| **dlv-cast-verbs** | Input: caps `input.cast-unit`, `input.cast-point`, `input.cast-hold`; delivery builds an `input.Action` per cast (verb, key, click, verify = the expectation); corpse casts carry the corpse unit; the Fighter stops calling `Click` and `CastAt` directly | `internal/input/`, `internal/d2dcombat/`, `internal/d2dinput/`, `data/input/caps/`, `internal/combat/fighter.go` (cast section); exclusive `fork-branch:didle/cast-verbs` | `input.cast-point-lands` (new, in the style of `input.skill-select-verb`): a point spell lands at the aimed cell | dlv-expect; the differential of question Q3 (fused verb against select then cast) | risk: every cast goes through new plumbing. Shadow is not possible for input; the knob is `input.planner`, which already exists and defaults to 0 | 3 |
| **dlv-aim-point** | Aim rules `cluster`, `lead`, `packEdge`, `offset`; radius and lead frames from the tables; point scoring as a gate (`crowdAtLeast`); line of sight per row (`missile`, `walk`, `none`), also on the watchdog path; traps move onto the same path (their placement is already this) | `internal/combat/` (new `aim.go`; `trap.go`, `castfirst.go`, `reach.go`) | `combat.point-spell-hits` for Druid Fissure and Sorceress Blizzard; `combat.traps-stand` | dlv-cast-verbs; fidelity A5 (delay enforced) and A8 before any tier rises | risk: tuned against a d2d with no delay. The scenario asserts `refused_at_most` so a fixed fork does not break it. Knob `combat.delivery.aim` | 3 to 4 |
| **dlv-aim-corpse** | Corpse aims `corpseBest`, `corpseNearPack` as data; one corpse reservation shared by raise, revive, explode and Death Sentry (S6); Find Potion and Find Item as utility rows after a fight | `internal/combat/` (`aim.go`; removes `necro_corpse.go`) | `combat.corpse-explodes` | dlv-upkeep-minions, dlv-cast-verbs | risk: small. Knob shared with aim | 1 to 2 |
| **dlv-stance-pack** | Stance `insidePack`: free cell next to the centroid, stand, cast in place, leave on the surround rule; Blessed Hammer row corrected after the live look of question Q8; novas and War Cry usable as `main` | `internal/combat/` (position, stance data), `data/capabilities/` (one owner: after dlv-aim-point merges) | `combat.hammer-in-pack` | dlv-rotation, dlv-upkeep-state (the damage aura); H1 live look; A6 before War Cry | risk: standing in a pack is the dangerous stance. The survival model's contact archetype must be the judge before knob 2 (R6) | 3 |

Wave 2 total: 10 to 12 days. This is roadmap R11 (Hammerdin) and the driver half of R16 (Blizzard, Druid, traps).

## 5. Wave 3: hold, channel, movement

| card | what | owns | scenario first | needs | risk and rollback | days |
|---|---|---|---|---|---|---|
| **dlv-press** | `repeat` and `channel` presses in slices with survival between slices, on the verb path; `commit` gate; Zeal on d2d, Strafe, Fend, Inferno, Arctic Blast rows | `internal/combat/` (cast section), `internal/d2dcombat/` | `combat.hold-interruptible` | dlv-cast-verbs (`input.cast-hold`); A1 for attack-sequence rate | risk: a hold that never releases. The verb must have an explicit release and the expectation a timeout. Knob | 2 to 3 |
| **dlv-move-tele** | Move driver, part 1: Teleport as a delivery row; navigation asks the driver for hops (range, cost and refusal from the row and `skillstate`, not from constants); in-fight reposition out of a surround; escape hop offered to Reflex | `internal/combat/` (new `move.go`, removes `telekey.go`), `internal/nav/teleport*.go` | `combat.teleport-reposition`; the existing teleport drill D22 stays green | dlv-expect, dlv-reads; queue row 16 (Teleport rules), A5 (refused Teleport cost) | risk: nav and combat both own movement for a moment. The driver is the only caller of the cast; nav keeps the route. Knob | 3 |
| **dlv-move-melee** | Move driver, part 2: Charge and Leap Attack as gap closers (`aim: target` and `through`), Leap as an escape | `internal/combat/move.go` | a row added to `combat.teleport-reposition` for a Barbarian and a Paladin | dlv-move-tele, dlv-press | risk: moderate; both are optional slots. Knob | 2 |
| **dlv-whirl** (parked) | Whirlwind: `channel` + `through` + `commit`, path choice across the pack, no drink during the spin | `internal/combat/move.go` | `combat.whirlwind-clears` | fidelity A6 (radius) fixed in the fork; dlv-move-melee | parked until A6 has a verdict (question Q9) | 2 to 3 |

Wave 3 total: 7 to 8 days without Whirlwind.

## 6. Cleanup and what is deliberately not a card yet

| card | what | owns | gate | days |
|---|---|---|---|---|
| **dlv-cleanup** | Delete `skills.json`, `skillKinds`, the profile bands, layouts (generated by `internal/hotkeymap`), `classOptions`, the Sorceress default helpers; CI literal check at zero for `internal/combat` | `internal/combat/`, `internal/power/`, `tools/ci/` | every driver knob has been at 2 for the agreed number of green nightly runs | 2 |

Not carded: shapeshift (data on top of dlv-upkeep-state and gates, waits for A8), charge-up and finisher (one more gate value, waits for A7), thrown and arrow quantity (K15, belongs to the town and gear lane), weapon-swap prebuff (B6 of the Barbarian note, needs a swap verb), the class literals in `internal/bot` listed in appendix A section 3 (they belong to the class traits table of the build-flexibility audit, not to combat).

## 7. Totals and order

| wave | cards | agent-days | unlocks |
|---|---|---|---|
| 0 | dlv-data, dlv-rotation, dlv-expect, dlv-reads | 10 to 11 | nothing visible; everything after |
| 1 | dlv-upkeep-state, dlv-upkeep-minions, dlv-mark, dlv-stance-army | 10 to 12 | Summoner, war cries, pets, all armors |
| 2 | dlv-cast-verbs, dlv-aim-point, dlv-aim-corpse, dlv-stance-pack | 10 to 12 | Druid fire, Blizzard, traps placement, Corpse Explosion, Hammerdin, novas |
| 3 | dlv-press, dlv-move-tele, dlv-move-melee (+ dlv-whirl parked) | 7 to 8 | Zeal on verbs, channelled spells, Teleport in fights, Charge, Leap |
| end | dlv-cleanup | 2 | the build-flexibility rule holds for combat |

Sixteen cards without the parked dlv-whirl: 39 to 45 agent-days in sequence (41 to 48 with it). Calendar time is shorter: inside a wave the cards with different owners run in parallel (at most two on `internal/combat` at once, split by file as listed), and wave 0 has two parallel pairs.

Why upkeep before ground-target, although the brief's examples lead with ground: upkeep covers the largest share of rows (71 of the 157 rows of `skills.json` are buff, aura or summon at `5b12ce5e7`, V by count), six of seven classes, it is the next class of ADR 0020 item 4, its failure mode is slow (an unraised skeleton) where a mis-aimed main attack is a dead hero, and it needs no new input verb. Point aim needs the cast verbs and two fidelity verdicts first; those can be asked for while wave 1 runs.

Dependencies outside this set: roadmap R1 (the scenario runner) for every scenario; R6 (the survival model as judge) before `insidePack` acts; R17 input stages I1 and I2 (devp14 switched in) before dlv-cast-verbs; the fidelity rows named per card.
