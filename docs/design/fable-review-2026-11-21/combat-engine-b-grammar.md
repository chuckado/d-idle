# Appendix B. The skill delivery grammar

> status: REVIEW (card fablecombat, docs only). Part of [combat-engine.md](combat-engine.md). A proposal: no schema, loader or data row in this appendix exists in the tree. Field names are suggestions; the closed value sets are the design.
>
> Tags as in the main doc. Every game number in the example rows is **[U]** unless a source is named: the rows show the SHAPE. Real rows are filled from the player's `skills.txt` and from measurements, never from this page.

## 1. Rules of the grammar

1. One row per skill, in the capability table (`data/capabilities/capabilities.json`, proposed schema `didle.cap/2`). The row keeps today's fields (`class`, `kind`, `tier`, `reason`, `element`, `reqLevel`, `reqSkills`, `synergies`, `mana`) and gains one object, `delivery`.
2. Every field has a closed value set, checked at load like roles and tiers are today (`internal/agency/caps.go:101-111`) [V]. An unknown value is a load error that names the skill, never a default.
3. A value that the game's tables hold is a POINTER to the table (`"table"`), not a copy. The loader resolves it through the catalog and fails loudly when the table lacks it.
4. Engine code may branch on any enum value of this grammar. It may not compare a skill name, a skill id, a class, a monster class id or a state id. A CI check greps for the literal lists of appendix A section 3 and counts them (target: zero in `internal/combat`).
5. A row with no `delivery` object is `unsupported` by construction.

## 2. The row

```
"delivery": {
  "purpose": "damage | upkeep | mark | move | utility | passive",
  "use":   { "hand": "left | right",
             "target": "unit | point | corpse | self | none",
             "press": "tap | repeat | channel",
             "sliceMs": 0,                       // repeat and channel: length of one input slice
             "commit": false },                  // cannot be interrupted or drunk through once started
  "aim":   { "rule": "target | cluster | lead | packEdge | offset | corpseBest | corpseNearPack | through | self",
             "radius": 0,                        // effect radius in subtiles, "table" when a calc column holds it
             "leadFrames": "table",              // lead: frames until the effect lands
             "offset": 0, "minDist": 0 },
  "reach": { "range": 0,                         // subtiles; or {"perLevel": [base, num, den]} or "table"
             "los": "missile | walk | none",
             "stance": "any | range | contact | insidePack | behindArmy" },
  "gates": { "cooldown": "table | none",
             "manaReservePct": 0,
             "lifeCost": false,
             "needsState": [], "forbidsState": [],   // state NAMES of states.txt: "wolf", "bear"
             "needs": [],                            // "corpse" | "weapon:dual" | "weapon:shield" | "weapon:bow" | "ammo" | "charges" | "notInTown"
             "visibleRefusal": true },               // the game shows the refusal (cooldown shade, red icon)
  "observe": { "by": "state | minions | aura | traps | clock | none",
               "state": "table",                 // the state the skill switches on, by name
               "group": "",                      // minions: a group name; members come from the summon column
               "want": "table | level | 1",      // how many; "table" = the game's pet cap
               "exclusive": "",                  // replacement group: armor | golem | spirit | aura | curse | form | shout | ...
               "duration": "table | none",       // seconds are derived from the frames column
               "scope": "self | party | target | area",
               "visible": true },                // a player sees it (icon, pets on screen): the hero may act on it
  "effect": { "missile": "table | none", "hits": "one | area | chain | pierce" }
}
```

Value sets, with what each one makes the engine do:

| field | value | engine behaviour |
|---|---|---|
| purpose | `damage` | offered by the attack driver to the pick step |
| | `upkeep` | offered by the upkeep driver when `observe` says it is missing or about to end |
| | `mark` | offered by the mark driver when a chosen target lacks the state |
| | `move` | offered by the move driver (and, for Teleport, to navigation) |
| | `utility`, `passive` | never offered in a fight |
| use.target | `unit` | the input carries the unit id; the point is the unit's position |
| | `point` | the input carries a point chosen by `aim.rule` |
| | `corpse` | the input carries a corpse unit chosen by `aim.rule`; gate `needs: corpse` is implied |
| | `self`, `none` | the input carries the hero's position; `aim` is ignored, `reach.stance` does the positioning |
| use.press | `tap` | one input per cast |
| | `repeat` | the button stays on the unit for `sliceMs`, survival checked between slices; ends when the target dies |
| | `channel` | as `repeat`, aimed at a point that is re-aimed each slice |
| aim.rule | `target` | the ranked target (today's behaviour) |
| | `cluster` | centroid of hostiles within `radius` of the ranked target, clamped to range and `minDist` (the calculation of `trap.go:165-177`) |
| | `lead` | `cluster`, advanced by the pack's mean velocity times `leadFrames` |
| | `packEdge` | the point of the pack nearest the hero, pushed `offset` toward the hero (a wall of fire, a Fissure in front of a pack that walks in) |
| | `offset` | the target plus a bounded random offset (erratic missiles: Tornado) [S review 03 section 3.5] |
| | `corpseBest` | highest monster level corpse in range (today `bestRevivable`, `necro_corpse.go:34`) |
| | `corpseNearPack` | corpse with most hostiles within `radius` (today `pickExplosionCorpse`, `necro_corpse.go:47`) |
| | `through` | a walkable point on the far side of the pack along the hero-to-centroid line (Whirlwind, Charge, Leap Attack) |
| reach.stance | `any` | no positioning beyond range and line of sight |
| | `range` | today's kite and keep-range rules |
| | `contact` | today's melee guard and reach |
| | `insidePack` | stand on a free cell next to the pack centroid, then cast in place |
| | `behindArmy` | keep the minion centroid between hero and pack |
| observe.by | `state` | present on the hero (or target) = up; with a duration, recast `margin` before the end (today `buffDue`, `necro.go:473-487`) |
| | `minions` | count of living own units in `group` against `want`; the cap is also learned by evidence |
| | `aura` | the active aura read (`aura.go:30-68`) |
| | `traps` | census plus pending-cast memory (`trap.go:64-129`) |
| | `clock` | only the recast timer; the row is `experimental` at best until a state is known |

## 3. Worked rows

Shapes only; numbers U unless sourced.

| skill | purpose | target | press | aim | reach | gates | observe |
|---|---|---|---|---|---|---|---|
| Fire Bolt | damage | unit | tap | target | range 20, los missile, stance range | | |
| Zeal | damage | unit | repeat | target | contact | | |
| Static Field | damage | self | tap | self | range perLevel [4,2,3] (today `skills.go:49`), stance any | | |
| Frost Nova | damage | self | tap | self | range 5 | | |
| Blizzard | damage | point | tap | lead, radius table, leadFrames table | range 30, los none (it falls from above, U), stance range | cooldown table | |
| Fissure (Eruption) | damage | point | tap | packEdge | range 20, los walk | cooldown table | |
| Firestorm | damage | point | tap | target | range 14, los walk | cooldown table | |
| Blessed Hammer | damage | none | repeat | self | stance insidePack | | |
| Inferno | damage | point | channel | target | range short, stance range | | |
| Whirlwind | damage | point | channel | through | stance contact | commit true | |
| Charge | move (and damage) | unit | tap | target | range 20, los walk | | |
| Leap Attack | move (and damage) | point | tap | through | range by level | | |
| Teleport | move | point | tap | (the caller's point) | range 18 [S `nav/teleport.go:19`], los none | | |
| Frozen Armor | upkeep | self | tap | self | | notInTown false | state table, exclusive armor, duration table |
| Battle Orders | upkeep | self | tap | self | | | state table, exclusive none, duration table, scope party |
| Might | upkeep | self | tap | self | | | aura, exclusive aura |
| Werewolf | upkeep | self | tap | self | | | state wolf, exclusive form, duration table |
| Fury | damage | unit | repeat | target | contact | needsState [wolf] | |
| Fissure for a shapeshifter | (same row) | | | | | forbidsState [wolf, bear] | |
| Raise Skeleton | upkeep | corpse | tap | corpseBest | range 4 [V live per `necro.go:17-19`] | manaReservePct 10 | minions, group from the summon column, want table |
| Clay Golem | upkeep | self | tap | self | | manaReservePct 10 | minions, exclusive golem, want 1 |
| Oak Sage | upkeep | self | tap | self | | | minions, exclusive spirit, want 1 |
| Valkyrie | upkeep | point | tap | (near the hero) | | cooldown table | minions, want 1 |
| Amplify Damage | mark | point | tap | cluster, radius table | range 20 [U, `necro.go:21`] | | state table, exclusive curse, scope area, duration table |
| Corpse Explosion | damage | corpse | tap | corpseNearPack, radius table | range 20 | | |
| Lightning Sentry | upkeep | point | tap | cluster, minDist 3 | range 15 | | traps, want 5 [S `assassin.go:15`], duration table |
| Death Sentry | upkeep | corpse | tap | corpseNearPack | range 15 | | traps |
| Tiger Strike | damage | unit | tap | target | contact | | (charges: a later axis value `needs: charges` on the finisher row) |

Notes on the rows that correct today's data:
- **Blessed Hammer** is `target: none` with a stance, not `ground` (main doc, section 1). Whether the clicked point matters is U until the live look (question Q8).
- **Curses** are `purpose: mark` with `target: point` and a radius: they are area effects (the code already treats neighbours within 5 as covered, `necro.go:25`). Their table kind today is `buff`, which is why the buff pass needs a name list to skip them (`fighter.go:695`).
- **Summons** name their group through the game's summon column, so the monster class comes from the tables. The spirit skills share `exclusive: spirit` (one spirit at a time [U game rule]); today's code wants "the n-th spirit skill while fewer than n spirits stand" (`druid.go:105-135`), which reads like three spirits can stand together and should be checked before it is copied.
- **Valkyrie** leaves the buff group it sits in today and gets a census.

## 4. Conditions (`when`) of a rotation slot

A slot is `{skill, role, when}`. `when` is a conjunction of these closed predicates. Each replaces a rule that is in Go today.

| predicate | meaning | replaces |
|---|---|---|
| `crowdAtLeast: {n, within}` | n non-immune hostiles within r subtiles of the aim point (or of the hero for `self`) | `novaCrowd`, `novaRadius` (`castfirst.go:21-22`); `needsCrowd` and its knob (`aura.go:104-111`) |
| `targetIs: elite | boss | pack | single` | class of the ranked target | `staticOpener` (`fight.go:810`), `pickCurseTarget` (`necro.go:337-341`) |
| `targetLifeAbovePct: p` | target's life percent | static floor (`fight.go:820`), curse and trap 25 percent lines (`necro.go:327`, `assassin.go:25`) |
| `castsOnTargetBelow: n` | per-target count of this skill | `StaticCasts` (`fight.go:810`) |
| `heroLifeBelowPct: p`, `heroLifeAbovePct: p` | hero life | `RestLifePct` (`aura.go:98-102`), Sacrifice life line (`tuning_combat.go:88`), druid 40 percent line (`druid.go:220`) |
| `resting: true` | no fight, the rest step | `wantAura(resting)` (`aura.go:85`) |
| `targetImmuneTo: [elements]`, `targetNotImmuneTo` | immunity | aura `Elem` skip (`aura.go:113`), `fallbackSkill` (`fight.go:827-833`) |
| `targetResists: {element, atLeast}` | resist after curses | `curseFor` Lower Resist test (`necro.go:358`) |
| `manaAbovePct: p` | hero mana | `NovaManaFloor` (`castfirst.go:105`) |
| `potionsAtLeast: n` | belt count | Sacrifice potion rule (`tuning_combat.go:89`) |
| `minGapMs: n` | own cooldown of the slot | `novaGap` (`castfirst.go:23`) |
| `monsterTag: name` | a tag from the monster rules data | the one-boss name test (`fight.go:813`); the tag lives with the monster rules, not in the rotation |

Example, the Sorceress cold rotation as the profile adapter would emit it (shape):

```
[ {skill: "Frozen Armor",  role: "upkeep"},
  {skill: "Static Field",  role: "opener", when: {targetIs: "elite", targetLifeAbovePct: "difficultyFloor", castsOnTargetBelow: 4}},
  {skill: "Frost Nova",    role: "area",   when: {crowdAtLeast: {n: 3, within: 4}, manaAbovePct: 15, minGapMs: 2500}},
  {skill: "Blizzard",      role: "main"},
  {skill: "Glacial Spike", role: "filler"},
  {skill: "Charged Bolt",  role: "second"},
  {skill: "Static Field",  role: "second", when: {targetImmuneTo: ["cold", "ltng"]}},
  {skill: "Teleport",      role: "move"} ]
```

The order inside a role is the preference order (today's band order). The planner writes this list from `Plan` (`internal/agency/derive.go:109-131`); the adapter writes it from a profile. The Fighter reads only the list.

## 5. Stance records

The melee block of a profile becomes one record per stance, with today's numbers as defaults and a per-build-file override:

```
"stances": {
  "contact":    {reach: 4, sliceMs: 600, retreatLifePct: 40, surround: 4, surroundLifePct: 65, manaReserve: 4},
  "range":      {keep: "skill range minus margin", stepBackDist: 3, stepBackSteps: 2, kiteLifePct: "knob"},
  "insidePack": {standOff: 2, leaveWhenSurrounded: 5, leaveLifePct: 50},
  "behindArmy": {minArmy: 3, keepBehind: 6}
}
```

The `contact` numbers are the Paladin and Barbarian profile values at `5b12ce5e7` [V profiles.json]; `range` names the existing constants (`castfirst.go:24-25`); the other two are U starting points for the Proving Ground to tune. `minArmy` is today's `armyMin` (`necro.go:162`).

## 6. Migration without a flag day

| step | change | proof |
|---|---|---|
| 1 | add `delivery` to every existing capability row by a generator that reads today's `skills.json` (kind, range, hand, hold) and the rules of section 3; hand-correct the rows named above | generated `skills.json` equals the checked-in one except for the listed corrections |
| 2 | `internal/delivery`: row type, loader, validator; `SkillSpec` is built from the row | every `SkillSpec` of every known skill is unchanged (table test over all rows, two builds per rule) |
| 3 | `Rotation` type and `FromProfile`; the Fighter's `attackSkills`, `buffPass`, `wantAura` read slots | golden: the simulator fixtures and the scenario corpus give the same casts in the same order |
| 4 | the special rules move into slots one at a time (Static opener, nova, aura crowd, life price), each deleting its Go | one fixture per rule, knob-off subtest reproduces the old behaviour while the knob exists |
| 5 | drivers replace the class side loops (appendix D) | Proving Ground rows |
| 6 | delete `skills.json`, `skillKinds`, profile bands, the class switch | the literal count of appendix A section 3 reaches zero |

Rows for classes the capability table does not cover yet (Amazon, Assassin; Druid is on a branch) [V] are a precondition of step 1 for those classes. They can be generated as `experimental` from skills.txt and the old table, the way the Barbarian and Necromancer rows were typed on 2026-11-20 [S docs/agency.md section 2].
