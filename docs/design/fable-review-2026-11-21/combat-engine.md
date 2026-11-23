# Combat engine and skill delivery layer: what exists, and a data-driven design

> status: REVIEW for the owner (design review 2026-11-21, card fablecombat, docs only). Nothing here is decided and nothing here changes Go or behaviour. Consumer: roadmap packages R10, R11, R12 and R16 of [06-roadmap](../fable-review-2026-11-19/06-roadmap.md) and the cards proposed in appendix D.
>
> Tags: **[V]** read by me in this tree at `5b12ce5e7` (file and line given; "V read" means the code says so, NOT that it was seen in a game); **[S]** read in another doc of this repo, named; **[U]** unverified, from memory or judgment. No game was started and no worker job was run for this review. Game facts tagged U must be read from the player's `skills.txt` or measured before they become data. Target is classic 1.14d.
>
> Appendices: [A: the engine today, per kind](combat-engine-a-today.md), [B: the delivery grammar](combat-engine-b-grammar.md), [C: tests, read verbs, differentials](combat-engine-c-tests.md), [D: rollout cards](combat-engine-d-cards.md).

## 0. Answer in twelve lines

1. The engine is one loop that picks the first affordable attack of a hand-written level band and aims it at the current tile of one ranked monster (`internal/combat/fight.go:600-605`, `:646`, `:799`) [V]. That loop is live-proven for a targeted missile and a melee swing, nothing else.
2. The data has seven skill kinds, but only `melee`, `self` and `buff` change what the loop does. `ground` and `bolt` are never read after parsing: a ground spell is a bolt with another label (appendix A, section 1) [V].
3. Everything that is not "attack one monster" lives in three class-switched side loops with skill names in Go: Necromancer (summons, curses, Corpse Explosion), Druid (pets), Assassin (traps), dispatched by `classTick` (`internal/combat/druid.go:230-240`) [V]. Two more rules are Sorceress specials by name (Static Field opener, Frost Nova on a crowd).
4. `hold` is honoured only for a left-hand pixel click. On the product backend (a swinger) and for every right-hand skill it is ignored (`internal/combat/fighter.go:780-796`) [V]. Channelled skills, Strafe, Fend and Whirlwind therefore have no driver at all.
5. No movement skill exists in combat. Teleport is a navigation feature keyed by a hotkey (`internal/nav/teleport.go:11-25`); Charge, Leap, Whirlwind have rows and no code. Shapeshift and charge-up have nothing (`KindCharged` is parsed and never used) [V].
6. A cast is "the input was sent" (`fighter.go:800-803`). Only summons (census) and one curse state are ever checked afterwards. Cooldown is the bot's own clock (`fighter.go:862-868`) [V].
7. Three tables classify the same skills and do not agree by construction: `internal/combat/classdata/skills.json` (7 kinds), `data/capabilities/capabilities.json` (kind, delivery, tier) and `internal/power/profiles.json` `skillKinds` (4 kinds). The capability tier is typed by hand and gates nothing in combat; the planner's rotation is a log line (`internal/agency/planner.go:66-89`), the Fighter still reads the profile bands (`fighter.go:834`) [V].
8. **Design:** one row per skill in the capability table carries a small **delivery grammar** on five independent axes: purpose, target, aim, press, gates, plus an upkeep block. The engine has four drivers (attack, upkeep, mark, move) that branch on those enum values and never on a class or a skill name (section 2, appendix B).
9. **One cast order:** the planner emits a `Rotation` (ordered slots with conditions). The Fighter consumes only a Rotation. Until the planner drives, an adapter compiles today's profile bands into the same type, so the switch is a data change per hero, not a code path (section 2.4).
10. **Every cast gets an expectation** (mana left the pool, a missile of mine appeared, a state is present, the census grew, the target lost life) that later ticks settle. Failures demote the skill for a while and are counted. This generalises four separate ad-hoc memories that exist today (section 3.4).
11. **Tests:** a logic-only simulator for the engine's choices, Proving Ground behaviour scenarios with new combat assertions, three new READ verbs in the fork (minions, skill state, combat events; ADR 0022 makes reads free), and a per-rule differential wherever the number matters (appendix C).
12. **Order:** foundation with no behaviour change, then upkeep (the largest group of rows and the safest failure mode), marks, point-aimed spells, stand-in-pack, hold and channel, movement, forms. Sixteen cards plus one parked, about 39 to 45 agent-days in sequence, each scenario-first with a knob to roll back (section 5, appendix D).

## 1. What the engine supports today

The full map with file and line is appendix A. The short form, per skill kind of the brief:

| skill kind | state today | why (short) |
|---|---|---|
| melee | **works, V live for Paladin** [S docs/classes/smoke-paladin.md]; code-only for other classes | left-hand loop, melee guard, `swinger` verb on d2d (`fight.go:662-666`, `:721`, `fighter.go:783-784`) |
| single-target missile | **works, V live for Fire Bolt** [S classdata.go:20-21]; code-only otherwise | the main loop |
| self-centred nova as main attack | generic path exists, never listed in a band; two skills are special rules by name | `castfirst.go:98-126`, `fight.go:808-824` |
| aura upkeep | **works, data-driven, V live for Might and Holy Fire** | ladder in the profile (`aura.go:85-126`); one knob is named after a skill |
| timed buff upkeep | works by clock for any buff in a profile group; by state for five armors only | `fighter.go:685-731`, `necro.go:42-43`, `:473-487` |
| summons upkeep | Necromancer: code with a short live check [S docs/summon.md]; Druid and Assassin: code never run live; Amazon Valkyrie is recast by a clock as if it were a buff | `necro.go:182-317`, `druid.go:105-227`, `assassin.go:88-110`, profiles.json amazon `buffs` |
| curses | Necromancer only, three curses chosen in Go, one state number proven | `necro.go:323-446` |
| ground-targeted AoE | **no driver**: aimed at the monster's tile like a bolt; no lead, no cluster choice, no effect check | `fight.go:646`, `:799`; traps have their own placement (`trap.go:165-182`) |
| corpse-targeted | Necromancer by name; on d2d the corpse is sent as a bare point, never as a unit | `necro_corpse.go:73-95`, `internal/d2dcombat/input.go:68-75` |
| movement skills | Teleport: navigation only; Charge, Leap, Leap Attack, Whirlwind: none | `nav/teleport.go`, `telekey.go:18` |
| shapeshift | none (form skills have melee rows; nothing reads or sets the form) | no `Werewolf` or `Werebear` row in skills.json |
| channelled, hold | left-hand pixel hold only | `fighter.go:785-791` |
| charge-up and finisher | none | `skills.go:35`, `classdata.go:164` |

Three structural findings matter more than any single row:

- **F1. The kinds are labels, not drivers.** `grep 'Kind(Bolt|Ground|Charged)'` over the non-test Go of `internal/combat` finds only the parser. A new delivery kind added to the JSON changes nothing until someone writes a branch for it in `engage`, which is how the class side loops came to exist.
- **F2. Build flexibility is broken in about 40 places in `internal/combat` alone** (class checks, skill names, monster class ids, state ids; inventory in appendix A, section 3). The rule in AGENTS.md says none of these may be in Go. The aura ladder (`AuraPref`, `classdata.go:47-56`) shows the repo already knows how to do it right: conditions as data fields, one generic chooser.
- **F3. The engine cannot tell that a skill does nothing.** Two incidents in the code comments are exactly this: a second Raise Skeleton cast on a used corpse "spent no mana" (`necro.go:135`), and a curse per unit "burned a cast every 1.2 s" (`necro.go:25`). Each got its own memory. Every unsupported skill the player may now buy (ADR 0020) can fail the same way.

Corrections to merged notes, so later cards do not inherit them [V read]:
- The Druid note says no ground skill checks line of sight. The main loop does, for every non-melee kind, before it casts (`fight.go:726`); the watchdog path does not (`castfirst.go:157-179`).
- The same note calls the Blizzard cooldown a special case. It is already generic: any skill with a `delay` in skills.txt gets it (`skills.go:118-120`, `fighter.go:862-868`). What is missing is reading the game's own cooldown state instead of trusting our clock (fidelity gap A5 makes the two differ on d2d [S review 02]).
- The capability row of Blessed Hammer says `delivery: ground`. The hammer is spawned at the caster and spirals outward; the clicked point does not place it [U, game knowledge; H1 of docs/process/hammerdin-support.md asks for the live look]. It is a self-centred skill with a POSITION rule, not a point-aimed one. This changes where it sits in the rollout.

## 2. The data-driven design

### 2.1 Principle

A skill is described on **five independent axes**. The engine branches only on the values of these axes. Appendix B has the schema, the closed value sets and worked rows.

| axis | question it answers | values (closed set) |
|---|---|---|
| purpose | why would the hero cast this now | `damage`, `upkeep` (something of ours must be present), `mark` (something must be present on THEM), `move`, `utility`, `passive` |
| target | what the game message carries | `unit`, `point`, `corpse`, `self`, `none` |
| aim | how the unit or point is chosen | `target`, `cluster`, `lead`, `packEdge`, `offset`, `corpseBest`, `corpseNearPack`, `through`, `self` |
| press | how the button is used | `tap`, `repeat` (held on a unit, the game chains: Zeal), `channel` (held stream, interruptible) |
| gates | what must be true first | cooldown, mana and reserve, life cost, required state (form), forbidden state, corpse present, weapon shape, crowd size, target class (elite, boss), not in town |

Plus two small blocks: **reach** (range, the line-of-sight rule, and a `stance` the skill needs: `any`, `range`, `contact`, `insidePack`) and, for upkeep and marks, **observe** (how the engine sees that the thing is up: a state number, a minion group count, the active aura, a trap census, or only a clock) with `exclusive` (the replacement group: armor, golem, spirit, aura, curse, form).

Why axes instead of sixteen kinds: review 03 lists K1 to K16 [S 03-classes-plan section 2]. Those are real, but most are one axis value apart. Blizzard and a Lightning Sentry differ only in purpose and observe. Corpse Explosion, Revive and Death Sentry share `target: corpse`. Whirlwind is `damage` + `press: channel` + `aim: through`. Werewolf is an upkeep whose observe is a state and whose exclusive group is `form`; Fury is a melee attack with a gate `state: wolf`. With axes, the sixteen kinds need **four drivers** and the capability row can drive behaviour directly.

### 2.2 Where the data lives

- **One table.** `data/capabilities/capabilities.json` gets a `delivery` object per row (schema `didle.cap/2`). `internal/combat/classdata/skills.json` and `power/profiles.json` `skillKinds` become GENERATED views of it, checked in CI, then deleted when their last reader is gone. Until then a parity test pins that every one of today's `SkillSpec` values comes out unchanged.
- **Game facts come from the game's tables, not from typing.** Mana, cooldown, required level and prerequisites already do (`skills.go:102-120`). Add: the state a skill switches on, the monster a summon skill creates, and the pet cap. skills.txt carries these as columns [U: `aurastate`, `auratargetstate`, `summon`, `petmax`; `internal/data/skills.go:13-58` reads `Summon`, `Param` and `Calc` but no state column today, V]. That removes the hand-kept `buffStates`, `curseStates`, the monster class constants of `druid.go:16-24` (several tagged U there) and the guessed `wantMinions`.
- **The tier becomes a consequence.** `unsupported`: the row has no valid delivery or its driver does not exist. `experimental`: delivery valid and driver exists. `supported`: additionally the driver's Proving Ground row for that skill is green on the backend in use. ADR 0020 already asks for generated tiers [S docs/agency.md section 2]; this gives the rule.
- **A leaf package** `internal/delivery` holds the row type, the loader and the `Rotation` type. `internal/agency` imports `internal/combat` today (`agency/caps.go:14`), so the shared types cannot live in either without a cycle.

### 2.3 What is left of a class profile

| today in `profiles.json` | becomes |
|---|---|
| `bands[].attack`, `requires`, `minLevel`, `avoid`, `lifePrice` | rotation slots (2.4); `avoid` is a tier; `lifePrice` is a gate on the skill row |
| `aura`, `bands[].aura` with `restLifePct`, `needsCrowd`, `elem` | rotation slots of purpose `upkeep`, exclusive `aura`, with the same conditions as generic `when` fields |
| `buffs` groups | rotation slots of purpose `upkeep`; the group is the row's `exclusive` |
| `layout` (F-keys) | generated per hero by `internal/hotkeymap` from the rotation (ADR 0021; the package exists and is U live) |
| `melee` block (reach, hold, retreat, surround) | a `stance` record chosen by the rotation's main attack; the numbers stay data, per stance and overridable per build file |
| class-switched Go (`isNecro`, `golemLadder`, `druidPlans`, `trapLadder`, `curseFor`) | upkeep and mark slots in rotation order; the ladder is the slot order |

So a "class profile" shrinks to: a default guide track (a build file) and optional stance overrides. Nothing else is per class.

### 2.4 The rotation: one source of the cast order

`Rotation` (`didle.rotation/1`) is an ordered list of slots: `{skill, role, when}`. Roles: `opener`, `main`, `area`, `second` (other element), `filler` (free or cheap), `finisher`, `upkeep`, `mark`, `move`, `rest`. `when` is a small closed condition set (crowd at least N within R, target is elite or boss, target life above P, hero life below P, resting, target immune to E, mana above P, casts on this target below N). Appendix B lists them with the code they replace.

- `internal/agency.Derive` already computes main, secondary, aura, buffs, summons, movement (`agency/derive.go:109-131`) [V]. It gains one function that lays them out as a Rotation.
- A `FromProfile` adapter turns a profile's bands into the same Rotation. With it the Fighter can switch to "consume a Rotation" with **no behaviour change**, proven by a golden test over the existing simulator fixtures.
- The planner's existing shadow line compares plan with profile (`planner.go:66-89`). It becomes a comparison of two Rotations, slot by slot, which is the evidence for the P2 switch of docs/agency.md section 7.
- Static Field on elites becomes an `opener` slot with `when: {elite, targetLifeAbovePct from the difficulty floor, castsOnTarget < 4}`; Frost Nova becomes an `area` slot with `when: {crowd >= 3 within 4, manaAbove}`. Both rules leave Go; the knobs `staticCasts` and `novaManaFloor` become slot defaults.

## 3. Engine architecture

### 3.1 The tick

The order of today's loop was paid for with incidents (the cast-first rule, the pin, the cave-mouth guards) and stays. What changes is steps 4 and 5.

```
1 sense      snapshot -> view: hostiles, own units by group, corpses, missiles, my states, cooldowns
2 reflex     Safety (drink, portal, chicken, crowd retreat): unchanged, always first        fight.go:257-286
3 house      merc potion, hard-mode rules, lost fight, caller interlude: unchanged          fight.go:289-330
4 propose    each driver offers at most ONE action: upkeep, mark, attack, move
5 choose     fixed class order with two budget rules (3.2); one action per tick
6 deliver    aim -> reach -> gates -> input (verb > key > verified click) -> expectation
7 settle     expectations of earlier casts are checked against this tick's view
```

### 3.2 Choosing between drivers

A second arbiter is not needed inside a fight. The outer arbiter works in virtual seconds with 120 s leases (`internal/arbiter/arbiter.go:98-100`) [V]; a fight needs a decision every poll. A fixed order with two budgets is enough and is explainable in one log line:

1. `upkeep` whose loss is dangerous (a slot flagged `vital`: the armor of a low-life caster, the form of a shapeshifter, the golem of a summoner) goes first.
2. `attack` when a cast is possible. The **cast-first invariant** stays as written: a non-cast may not replace casts for longer than `evadeGap` (`castfirst.go:38-42`), and the watchdog forces one (`castfirst.go:86-93`).
3. Other `upkeep` and `mark`, limited to a share of casts per fight (data, default one in four) and by the mana reserve, and skipped when the crowd would cut the cast (the existing `crowdAdjacentMax` rule, `fighter.go:758`).
4. `move` (approach, standoff, kite, reposition), under the existing evade and pin rules.

After a fight the same upkeep driver runs with no share limit; that replaces `classUpkeep` (`fight.go:136`, `druid.go:243-256`).

### 3.3 Target selection and the threat model

- **Unit ranking stays** (`RankFor`, `target.go:322`) for `aim: target`.
- **Point aims** add a small step: candidate points are the ranked target, the pack centroid around it (the calculation exists for traps, `trap.go:165-177`), and for `lead` the target's position advanced by its velocity times the skill's delay. A point is scored by hostiles inside the row's effect radius that are not immune. First version: the score only gates (`crowd >= N`); it does not reorder skills.
- **Stance** replaces `MeleeClass()` as the switch for positioning. `contact` is today's melee guard. `range` is today's kite and keep-range. `insidePack` is new and small: walk to a free cell adjacent to the pack centroid, stand, cast; step out on the surround rule (Blessed Hammer, novas as a main attack, War Cry). `behindArmy` is new: keep the minion centroid between hero and pack (Necromancer S8, [S docs/process/necromancer-support.md]).
- **Threat inputs** are unchanged: pack danger, the missile threat table (`threat.go:40-52`), the per-monster rules (`rules.go:94`). One cleanup: `classOptions` (`rules.go:228-253`) is a class-keyed map of skill names used for immune workarounds; it becomes a query on the Rotation ("which slots have an element this target is not immune to").

### 3.4 Expectations: how the engine learns that a cast did nothing

Each delivered action records what should be observable and by when:

| purpose | expectation | settles as failed when |
|---|---|---|
| any paid skill | mana dropped by about the cost | no drop within 2 polls: the game refused (cooldown, state, no target, busy) |
| damage, unit or point | a missile owned by me of the row's missile class exists, or the target (or a unit in the radius) lost life | neither within the cast length plus travel time |
| upkeep, state | the state is present on me | absent after a grace period |
| upkeep, minions | the group count grew, or mana was paid and the count is at the cap | paid and not grown: record the cap (today's `capSeen`, `necro.go:309-314`) |
| mark | the state is present on the target or a unit near the point | absent: count a try (today's `curseRec.tries`, `necro.go:395`) |
| move | the hero's position changed toward the point | unchanged (today's teleport bad-hop count, `nav/teleport.go:24`) |

Three consecutive failures of one `(skill, reason)` demote that slot for a back-off time and the next slot plays. Counters go to telemetry as `delivery ok/refused/noEffect` per skill. This is the single mechanism behind the "unsupported, labelled" promise of ADR 0020: an experimental skill that does nothing costs a few casts, is reported to the player, and the hero falls back instead of standing in a pack casting air.

### 3.5 Survival and chicken

- Reflex is first and unchanged. Nothing in this design may delay `Safety`.
- **Holds are sliced.** A `repeat` or `channel` press is issued in slices no longer than one poll interval, with `Safety` between slices. Today a held Zeal blocks for `holdMs` (600 ms in the profile) inside the input call (`fighter.go:787-788`) [V].
- **Commit gates.** A skill the game does not let you interrupt or drink during (Whirlwind: potions are locked during the spin [S docs/process/barbarian-support.md B4]) carries `commit: true`; the engine starts it only when life is above the drink line plus a margin. Data per row.
- **The survival model reads the Rotation.** `power` derives the archetype from the combat profile bands today (`power/archetype.go:87-101`) and the planner already carries `Spec` and `Arch` in its plan (`agency/derive.go:123-126`) [V]. With one Rotation the judge and the fighter cannot disagree about what the hero plays.
- **Escape by skill** is later: the power profiles already name `escape.mobilitySkill: from_build` [V power/profiles.json]. When the move driver exists, Reflex may ask it for an escape hop. Not in the first rollout.

### 3.6 The input layer

ADR 0021: verb, then key, then verified click; the fork's verbs are the game's own client messages. For casting that gives this mapping:

| bot intent | client messages (1.14d) | fork today | proposal |
|---|---|---|---|
| put skill on a hand | select skill, 0x3C [S docs/process/sandbox-devp14.md] | `cmd select` fixed on devp14, cap `input.select` | as is |
| cast at a unit | right or left skill on unit [U ids] | `cmd skill id x y unit [left]`: select and cast fused in one verb [S docs/d2d-backend.md] | cap `input.cast-unit` |
| cast at a point | right or left skill on location [U ids] | `cmd skill id x y` with no unit: U whether every ground skill accepts it [S druid-support section 5 item 3] | cap `input.cast-point`, drilled per skill family |
| cast at a corpse | skill on unit, the unit is the dead monster [U] | bot sends a bare point: `CastAt` only looks up LIVING monsters (`d2dcombat/input.go:68-75`) [V] | corpse unit id through `input.cast-unit`; needs corpses in the monsters read |
| hold or channel | the hold variants of the same messages, repeated while held [U] | none; `Swing` sends one command | cap `input.cast-hold` with an explicit release, or repeated casts per slice |

So "ground-target cast = select verb + cast-at verb" is right in message terms. `cmd skill` fuses the two. Whether the fused verb equals select-then-cast one tick later is a differential, not an opinion (question Q3). On the bot side `SemanticInput.Cast(skill, point, unit)` already has the right shape (`internal/input/semantic.go:16`) [V]; delivery builds one `input.Action` per cast with the three paths and the expectation as `Verify` (`internal/input/planner.go:22-29`). The Fighter stops calling `In.Click` and `In.CastAt` directly.

### 3.7 The outer arbiter

The split is: the arbiter decides WHETHER and WHERE the hero fights; the engine decides HOW. Three contact points:

1. `Clear` and `Kill` stay the executor of an Objective-class intent and keep their typed outcomes (`ErrTownPortal`, `ErrChicken`, `ErrDead`, `types.go:51-62`), which the arbiter records through `Report`.
2. Out-of-fight upkeep becomes an Upkeep-class source fed by two facts the engine publishes: "upkeep due" (count of unmet upkeep slots that can be met here) and "army below want". That replaces the buff edges in the goal chain and the implicit `classUpkeep` call.
3. Reflex stays inside the engine during a fight; the arbiter's `survive` source (`arbiter/sources.go:52-59`) covers the time outside fights. Both read the same thresholds.

## 4. Test strategy (summary; appendix C)

| layer | proves | does not prove |
|---|---|---|
| unit tests of pure pieces (aim, gates, rotation layout, expectations) | the rule as written | anything about the game |
| logic simulator (`internal/combat/sim_test.go` grown by a table-driven effect model: radius, delay, cooldown refusal, corpses, minion groups, states) | the engine CHOOSES right: point, order, upkeep, fallback after a refused cast; build-flexibility (every rule tested with two builds) | damage, timing, whether the game does what the row says |
| Proving Ground behaviour scenarios on lockstep d2d, new assertion kinds | the hero DOES it with `writes after start: 0`: summons stay at or above N for T minutes, hits per hammer, kills per mana, a Whirlwind path clears a pack | that d2d equals game.exe |
| per-rule differentials (ADR 0023) through the decomp queue | the numbers: mana per level, cast and cooldown frames, effect radius, pet caps, state put rules | nothing about the bot |

New READ verbs asked of the fork (reads are free, ADR 0022): `minions` (my summons with owner, group, source skill, life), `skillstate` (per skill: usable now, refusal reason, cooldown frames left), `combat events since <tick>` (cast accepted or refused with reason, damage by source skill, kill credit, minion spawn and death). The missile list with owner already exists (`internal/d2dctl/parse.go:632-644`) [V]. The event read is the important one: it turns "hammer hits per mana" from an inference into a count and is the same record a differential needs.

## 5. Rollout (summary; appendix D has the cards)

Ordering rule: after a foundation that changes no behaviour, take the driver that unlocks the most rows with the mildest failure mode first.

| step | driver | unlocks (skills.json rows by kind at `5b12ce5e7`, V by count) | classes | agent-days |
|---|---|---|---|---|
| 0 | foundation: grammar, Rotation, expectations, read verbs, assertions | none (parity) | all | 10 to 11 |
| 1 | **upkeep** (buff 32, aura 20, summon 19 rows) | Necromancer Summoner army, Druid spirits and pets, Barbarian war cries, Holy Shield, Valkyrie, Shadow, every armor | 6 of 7 | 8 to 9 |
| 2 | **mark** | curses (Necromancer), Cloak of Shadows, Taunt, Howl, Battle Cry | 3 | 2 to 3 |
| 3 | **point and corpse aim** (ground 16 rows, corpse skills, traps onto the shared path) | Fissure, Firestorm, Volcano, Blizzard, Meteor, Fire Wall, Corpse Explosion, sentries | 4 | 7 to 9 (3 of them the cast verbs) |
| 4 | **insidePack stance** | Blessed Hammer, Nova, Poison Nova, War Cry, Shock Wave | 4 | 3 |
| 5 | **hold and channel** on verbs | Zeal chain on d2d, Strafe, Fend, Inferno, Arctic Blast | 4 | 2 to 3 |
| 6 | **move** | Teleport in fights and hops planned by the walker, Charge, Leap, Leap Attack; Whirlwind last and parked (fidelity A6) | 3 | 5 (+2 to 3 parked) |
| 7 | forms and charge-up (data on top of 1 and 5) | Werewolf, Werebear lines, martial arts | 2 | not carded: parked until fidelity A7, A8 |
| end | cleanup: delete the old tables and the class switch | the build-flexibility rule holds for combat | all | 2 |

This order agrees with the class order of ADR 0020 item 4 (Sorceress second build needs nothing here; Summoner = steps 1 to 3; Hammerdin = step 4 plus 1; Frenzy = step 1; ground builds = step 3) and moves the Hammerdin out of the "ground" group.

## 6. Open questions for the owner

Each has my recommendation. Nothing below is decided; where an answer would change an ADR or `docs/decisions.md`, proposed text is given and not written there.

1. **Q1. One table.** Make `data/capabilities/capabilities.json` the only hand-edited skill table, with `skills.json` and `skillKinds` generated from it and then removed? *Recommend yes.* Three tables is how Blessed Hammer came to be `ground` in two of them. Proposed decision line: "Skill delivery data has one source, the capability table; other skill tables are generated views."
2. **Q2. Generated tiers.** Adopt the tier rule of 2.2 (unsupported = no valid delivery or no driver; experimental = valid and driver exists; supported = plus a green Proving Ground row on the backend in use)? *Recommend yes*; it is the P2 item ADR 0020 already implies. Consequence to accept: some of today's hand-typed "supported" rows drop to experimental until their scenario exists.
3. **Q3. Fused cast verb.** Is `cmd skill` (select and cast in one verb) acceptable as INPUT under ADR 0021 "one verb per message", if a differential shows it equals select-then-cast? *Recommend yes, conditional on that differential*, and split it only if the differential fails. A fused verb saves one tick per skill switch.
4. **Q4. Read verbs.** Ask the fork for `minions`, `skillstate` and `combat events` as flag-gated READ verbs, recorded in the run record? *Recommend yes.* Without `combat events` every combat scenario asserts through life deltas and guesses. Proposed addition to the ADR 0022 consequences: "Combat effect reads (minion list, skill state, combat event journal) are reads in the sense of this ADR."
5. **Q5. May the hero ACT on those reads?** Verification with a read is clearly free. Using `skillstate` to decide (cast when the game says the cooldown is over) is the hero knowing what a player sees on the skill icon. *Recommend: allowed for facts a player can see on screen (cooldown shade, minion count, a curse icon over a monster), not for hidden numbers (exact monster resist after a curse).* The engine marks each read as `visible` or `hidden` in the row; hidden reads are for tests only.
6. **Q6. Experimental skills in the rotation.** When the player buys an experimental attack, does the hero use it? *Recommend: yes, behind the expectation breaker of 3.4, with the capability reason shown; unsupported skills are never put in the rotation.* This is the engine half of ADR 0020 item 2 ("unsupported skills are allowed with a clear label").
7. **Q7. Upkeep share.** Default budget for non-vital upkeep and marks inside a fight: one cast in four, mana reserve 10 percent of the pool (today's summon rule, `necro.go:191`). *Recommend these as data defaults per archetype in the survival params,* tuned by the Proving Ground, not as knobs named after classes.
8. **Q8. Hammerdin before ground spells.** Treat Blessed Hammer as self-centred with the `insidePack` stance (step 4) and not as part of ground-target delivery? *Recommend yes, after one live look* (H1 of the Hammerdin note) confirms the clicked point does not place the hammer.
9. **Q9. Whirlwind and forms.** Park Whirlwind (step 6 tail) and shapeshift (step 7) until fidelity items A6, A7, A8 have verdicts? *Recommend yes.* The grammar covers them, so nothing is lost by waiting, and a driver tuned on a wrong radius would be tuned twice.
10. **Q10. The frozen Game.exe backend.** New delivery paths get a verb path first and the pixel path only where it already exists (left hold, F-key cast). *Recommend: no new pixel work*, consistent with review 07 section 4.4.
11. **Q11. Card size.** The foundation cards (D0 to D3 in appendix D) change no behaviour and so have no Proving Ground scenario of their own; ADR 0020 item 1 lifts the freeze only for work with a green scenario. *Recommend: accept a golden parity test over the existing simulator and scenario corpus as their gate,* stated on each card.

## 7. What I could not verify

- Anything live. Every "works" above is V by reading code plus the S live notes named; the review started no game.
- skills.txt column names for state, summon and pet cap, and the client message ids for cast-on-unit, cast-on-location and their hold variants [U, memory].
- Whether d2d's `cmd skill` with no unit is accepted for every point-aimed skill, and whether it accepts a dead unit id for corpse skills [U].
- That the hammer ignores the clicked point (Q8) [U].
- The agent-day estimates are judgment from the size of comparable merged cards, not a measurement.
