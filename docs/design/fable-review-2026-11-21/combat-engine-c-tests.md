# Appendix C. Test strategy: simulator, Proving Ground, read verbs, differentials

> status: REVIEW (card fablecombat, docs only). Part of [combat-engine.md](combat-engine.md). Proposal; no assertion kind, verb or scenario named "new" below exists in the tree.
>
> Tags as in the main doc. File and line references are to this tree at `5b12ce5e7`.

## 1. Four layers, and what each one may claim

| layer | runs where | may claim | may never claim |
|---|---|---|---|
| L0 pure unit tests | every merge | "the rule as written": aim point, gate, slot order, expectation settling | anything about the game |
| L1 logic simulator | every merge | "the engine chooses X in situation Y" | damage, timing, that the game accepts the cast |
| L2 Proving Ground behaviour scenario, lockstep d2d | merge gate for the touched family (ADR 0020 item 1) | "the hero does it, with `writes after start: 0`" on d2d | that d2d equals game.exe |
| L3 per-rule differential (ADR 0023) | decomp queue | "d2d equals game.exe for this rule", or a listed shim | anything about the bot |

A capability tier of `supported` needs L2 green for that skill and, where the scenario's verdict depends on a number, the L3 rows named in section 5.

## 2. The simulator: what it is and where it stops

Today's combat simulator is a test helper: monsters walk to the player and hit when adjacent, a cast damages the monster nearest the click by a fixed amount, potions heal, time moves only on sleeps (`internal/combat/sim_test.go:22-73`) [V]. It has player state ids, a class and skills; it has no missiles, no area effects, no cooldown refusal, no corpses that can be consumed, no own minions, no states on monsters. `mem.Fake` (`internal/mem/source.go:51`) is a byte image: it can hold any struct the tests write into it and no behaviour.

Proposal: grow the simulator by ONE table-driven effect model that reads the same delivery rows as the engine. For a delivered cast it applies, by the row's axes:

| row says | simulator does |
|---|---|
| `gates.cooldown` | refuses (no mana spent, no effect) while the delay runs |
| `use.target: corpse` | refuses without a corpse unit at the point; consumes it |
| `needsState`, `forbidsState` | refuses when the state rule fails |
| `effect.hits: area`, `aim.radius`, `leadFrames` | damages every monster inside the radius at the point, `leadFrames` later |
| `observe.by: state` | sets the state on the hero or on monsters in the radius for `duration` |
| `observe.by: minions` | adds a friendly unit of the group up to a cap the TEST sets (so the cap-by-evidence path is exercised) |
| `purpose: move` | moves the hero to the point when the test's grid allows |

What this buys, and its honest limit: it tests the engine's CHOICES against the grammar (does it lead a moving pack, does it stop casting a refused skill and fall to the next slot, does it keep five traps, does it hold the upkeep share). It is circular about the game: the simulator believes the row, so a wrong row passes. That is why a row is only `experimental` on L1 alone.

Every L1 rule test runs for two builds (AGENTS.md build-flexibility rule): the same test body parameterised by rotation, for example the upkeep-by-state test with a Sorceress armor and a Barbarian shout, the point-aim test with Blizzard and Fissure.

Golden parity for the foundation cards: record, for each existing simulator test and each combat scenario of the corpus (`docs/scenarios.md`), the sequence `(tick, action, skill, target or point)`; the refactor must reproduce it exactly.

## 3. Proving Ground combat scenarios

The scenario format exists (`didle.scenario/1`, `tools/provingground/scenarios/*.json`): `pre`, `trigger` with goals and knobs, `budget`, `matrix` of classes and seeds, `provenance`, `assert`, and a `needs` list of backend capabilities [V input-scenarios.json]. Assertion kinds in use today: `quest_bits`, `no_writes`, `stat_between`, `wealth_between`, `stash_between`, `right_skill_is`, `skill_points`, `item_in_pack`, `item_absent`, `portal_open`, `area_changes_le`, `life_pct_at_least`, `log_line` [V, listed from the shipped files]. None of them can state a combat effect.

New assertion kinds (each backed by a read of section 4):

| kind | reads | meaning |
|---|---|---|
| `minions_between {group, min, max, from_min, to_min, pct}` | `minions` | the group count is inside the band for at least pct percent of the samples in the window |
| `state_uptime_at_least {state, on: hero or target, pct, from_min}` | player states, monster states | upkeep and mark coverage |
| `casts_between {skill, min, max}` | combat events | the skill was cast (accepted) this often |
| `refused_at_most {skill, n, reason}` | combat events | casts the game refused (cooldown, no corpse, wrong form) |
| `hits_per_cast_at_least {skill, x}` | combat events | area skills actually hit |
| `kills_per_100_mana_at_least {x}` | combat events, mana | efficiency, the anti "casting air" check |
| `kills_at_least {n}`, `kills_by_source {skill or group, share}` | combat events | the army or the trap did the killing |
| `deaths_at_most {n}`, `chickens_at_most {n}` | bot record | survival side of every combat scenario |
| `moved_through {pack: true}` | positions | a movement attack crossed the pack |
| `input_path {action, path}` | the planner's `input path` line (`internal/input/planner.go:113-115`) | the cast went by verb |

Scenario set, one or two per driver. All are `kind: behaviour`, lockstep, `no_writes`, classes and build as matrix parameters; setup writes (level, points) are CHEATED setup and are mechanics evidence only, as the existing scenarios already state in `provenance` [V input-scenarios.json].

| id (proposed) | setup | asserts | driver |
|---|---|---|---|
| `combat.parity-rotation` | Paladin and Sorceress at clvl 12, field start, planner off | same kill and cast counts within the seed band as the run before the refactor; `log_line` shows the rotation came from the profile adapter | foundation |
| `combat.buffs-stay-up` | Sorceress with an armor, Barbarian with a shout, 15 virtual minutes of clearing | `state_uptime_at_least` 90 percent; `casts_between` bounded (no recast spam) | upkeep |
| `combat.army-holds` | Necromancer summoner at clvl 12 with the guide track, Cold Plains | `minions_between {skeleton, min 3}` for 80 percent of minutes 3 to 15; golem present 90 percent; `kills_by_source` army share above half; `deaths_at_most 0` | upkeep |
| `combat.pets-hold` | Druid with a spirit and wolves | `minions_between` per group; a spirit is never recast while one stands (`casts_between`) | upkeep |
| `combat.curse-covers-pack` | Necromancer, packs of four or more | `state_uptime_at_least` on the pack's units from the first cast to the pack's death; `casts_between` at most 2 per pack | mark |
| `combat.point-spell-hits` | Druid Fissure, Sorceress Blizzard, clvl by skill | `hits_per_cast_at_least 2` on packs; `refused_at_most` cooldown 2; `kills_per_100_mana_at_least` a floor | point aim |
| `combat.corpse-explodes` | Necromancer with Corpse Explosion | `casts_between`, `hits_per_cast_at_least 2`, no cast with `refused: no corpse` above 2 | corpse |
| `combat.traps-stand` | Assassin with one sentry skill | `minions_between {trap, max 5}`, at least 2 standing near an engaged pack | point aim |
| `combat.hammer-in-pack` | Hammerdin at clvl 20 | `hits_per_cast_at_least 1.5`, `kills_per_100_mana_at_least` a floor, `state_uptime` of the damage aura while casting | insidePack |
| `combat.hold-interruptible` | Paladin Zeal on a pack with a scripted life drop | a potion is drunk within one poll of the life line during a hold | hold |
| `combat.teleport-reposition` | Sorceress with Teleport, a pinned start | the hero leaves a surround by a hop, `deaths_at_most 0` | move |
| `combat.whirlwind-clears` | Barbarian with Whirlwind | `moved_through`, `kills_at_least`, no drink attempted during a spin | move, parked |
| `combat.dead-skill-falls-back` | any class, a rotation whose main is forced to a skill the backend refuses | `refused_at_most` 6 then the next slot plays; `kills_at_least` proves the fight went on | expectations |

The last one is the scenario for the ADR 0020 promise about experimental skills; it should exist before any experimental attack is allowed into a rotation.

## 4. Reads the fork must expose

ADR 0022: any engine fact exposed as a read never makes a run CHEATED; the verb is flag-gated and recorded in the run record. What exists: `monsters` (class, cell, life, mode, boss kind, mods, unit id; `internal/d2dctl/parse.go:185-199`), `missiles` with owner type, owner id, a friendly column and time to live (`parse.go:632-644`), `state`, `caps`, `debug levelmap` [V]. Missing for combat:

| verb (proposed cap name) | returns | why the existing reads are not enough |
|---|---|---|
| `minions` (`read.minions`) | for each unit owned by the hero: unit id, monster class, owner id, the skill id that made it, life, mode | the bot infers ownership from class and AI because owner words read 0 for Necromancer minions (`internal/combat/friendly.go:83-87`); a revived monster keeps its own class and needs a state never seen live [V comment] |
| `skillstate` (`read.skillstate`) | for each learned skill: effective level, mana cost now, usable now, refusal reason (cooldown frames left, state, no charge, in town), hand it is on | cooldown is the bot's own clock today (`fighter.go:862-868`); the refusal reason is otherwise invisible |
| `combat events since <tick>` (`read.combat-events`) | journal lines: cast accepted or refused (skill, reason), damage applied (source unit, source skill or missile, target, amount before and after resist), kill credit, minion spawn and death, state put and removed (unit, state, by) | scenarios otherwise infer hits from life deltas between samples, which cannot attribute damage to a skill or to the army |
| corpses in `monsters` | dead units stay listed with a `raisable` flag while the corpse exists | corpse casts need a unit id (`internal/d2dcombat/input.go:68-75` only finds living monsters) |
| states in `monsters` and `state` | the state ids set on each unit | upkeep and mark observation by state; today the bot reads states from the memory image where that exists |

Two rules for these reads:
- **Verification is always allowed.** A scenario or an expectation may use any of them.
- **Deciding on them is a product question (Q5 of the main doc).** The proposal: a read is marked `visible` when a player sees the same fact on screen (cooldown shade on the icon, pets standing, a curse icon over a monster, the corpse on the ground). The hero may decide on visible reads only. Damage numbers and exact resist values are test-only.

The event journal is also the cheapest fidelity instrument: the same lines from game.exe (through the existing emulator harnesses of the decomp lane) and from d2d are the two sides of a differential.

## 5. Differentials: where fidelity decides the verdict

Per ADR 0023 a rule is accepted with a zero-difference differential or a listed shim. For each driver, the rules whose error would make a green L2 scenario mean nothing. Gap ids are those of review 02 [S docs/design/fable-review-2026-11-19/02-fidelity-and-gaps.md]; queue rows are `docs/decomp-queue.md`.

| driver | rule that must hold | known state | consequence if wrong |
|---|---|---|---|
| all | mana cost per level: the bot rounds the formula up (`skills.go:110`) | queue row 14 covers Sorceress missiles; other classes U | budgets and the mana gate are off by one per cast |
| all | cast length per class and skill: modelled for the Sorceress only (`castgap.go:17-46`) | Lightning sequence shim R01 open | cast gap wrong: lost casts or dropped inputs |
| upkeep | buff state numbers and lengths (queue row 15 did the Sorceress armors) | war cries, Holy Shield, Cyclone Armor, Fade: U | recast by clock, uptime scenario meaningless |
| upkeep | pet caps and which skill makes which monster | Necromancer partly V live; Druid ids U (`druid.go:16-24`); summon specials A11 | census wrong |
| upkeep | auras reach the mercenary and pets (A3) | deviates | the summoner's damage is understated on d2d |
| mark | curse put rule: slot, refresh, length (A2) | deviates in 54,685 of 100,000 sequences [S] | "is it up" and recast cadence differ from game.exe |
| point aim | skill delay enforced (A5) | failed at the last check, current state unknown [S] | Blizzard and Meteor spam on d2d: inflated kills |
| point aim | effect radius and delay of each area spell; Druid fire (A8) | no verdict | hits-per-cast floors are tuned to the wrong shape |
| corpse | which corpses can be used, and by which skills | U (`necro.go:99-101` says bosses "U") | wasted casts |
| traps | everything (A7: 0 of 30 Assassin skills run-tested) | untested | no claim possible; the driver can be built, the tier cannot rise |
| insidePack | hammer spawn point and path; War Cry hit rate (A6: about 8 times too often) | U; deviates | the Hammerdin position rule and the Singer |
| hold | what one held input does per frame on d2d against game.exe (Zeal chain, Strafe count) | attack-sequence rate 30 points slow (A1) | damage per hold |
| move | Teleport refusals and cost (queue row 16; A5 "refused Teleport still costs mana"); Whirlwind radius (A6: 5 cells instead of 5 subtiles); Leap range | deviates | movement attacks tuned twice |
| forms | shapeshift rules (A8) | not built | parked |

Order of asking: A5 (delay) and the corpse and point-cast input checks first, because step 3 of the rollout leans on them; A2 and A3 next, for the Summoner; A6 before any Barbarian area claim. Each rollout card in appendix D names its rows.

## 6. What stays unmeasured

- Kill speed against game.exe as a whole. ADR 0023 dropped the A:B verdict; these scenarios compare d2d with d2d.
- Watchability (does the fight look sensible). That is a human look at the watch video per driver, listed as a checklist line on each card.
- Hell. Every scenario above is Normal. Immunity handling changes with the rotation's `second` slots and needs its own rows when the difficulty transition (R9) is real.
