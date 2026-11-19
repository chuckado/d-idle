# Agency P1a: the capability table, the shadow skill planner, build files

> status: current (card agencyp1a, 2026-11-19; roadmap R7, first slice). Decision: ADR 0020, path D (the player spends every stat and skill point, the hero plays what it actually has, the table says plainly which skills it can use). Design: `docs/design/fable-review-2026-11-19/04-stat-skill-agency.md`. Builds on `docs/points.md` (P0). Tags: V verified (how said), S secondary, U unverified. **Nothing here changes what the hero does**: the planner runs in shadow, the knob is off by default.

## 1. What shipped

| piece | where | state |
|---|---|---|
| capability table | `data/capabilities/capabilities.json` (schema `didle.cap/1`), loader `internal/agency/caps.go` | all 30 Paladin and 30 Sorceress skills, plus the Barbarian and Necromancer skills the plan tables use |
| planner | `internal/agency` (`Derive`, `Judge`, `Planner.Evaluate`, `ViewOf`, `ShadowLine`) | shadow only |
| knob | `agency.planner` (`internal/env/tuning_agency.go`): 0 off (default), 1 shadow | V `go test ./cmd/d2bot -run AgencyPlannerDefaultsOff` |
| log line | `planner shadow:` once a minute (game clock), `cmd/d2bot/agencywire.go` | V `TestAgencyWireShadowLineAndView` |
| API | `GET /api/hero/plan` (`internal/watch/planapi.go`), viewer listener, read-only | V `TestPlanAPIRoute` |
| Points page | one line, "your build: ...; viability: ..." (`web/points.js`, `#buildline`) | hidden when the knob is 0 |
| build files | `internal/buildfile` (`didle.build/1`): strict parser, 17 validation codes, `Expand`, `FromRows`, `ExportPlan`, `ImportPlan`, `LoadFile`, `SaveFile` | V tests |
| scenario | `agency.plan-matches-profile` (`tools/provingground/scenarios/agency.json`) | PENDING, see section 7 |

To enable the shadow on a hero: set the knob `agency.planner` to 1 (`config.json` `"tuning": {"agency.planner": 1}` or `PUT /api/v1/instances/{id}/knobs`). The log gets one `planner shadow:` line per virtual minute and `/api/hero/plan` answers with the view; at 0 the API answers `mode: off` and nothing is logged.

## 2. The capability table (data, not Go)

Per skill: `kind` (damage, aura, buff, curse, summon, utility, movement, passive), `delivery` (melee, projectile, nova, ground, trap, pet, self), `tier` (supported, experimental, unsupported) with a `reason` that is required for every row that is not supported, `element`, `synergies` (null = NOT KNOWN, never "none"; `[]` = none), `reqLevel` and `reqSkills`, and `mana.source` (where the cost per level comes from: `skilldamage.json:mana`, the game's functions run in the emulator, V `internal/power`; or the `skills.txt` columns `mana,lvlmana,manashift,minmana`; `none` for passives). The loader refuses an unknown kind, delivery, tier or class and a non-supported row with no reason. No class or skill name is a Go literal.

| class | supported | experimental | unsupported |
|---|---|---|---|
| Sorceress (30) | 9: Fire Bolt, Ice Bolt, Charged Bolt, Static Field, Frozen Armor, Warmth, the three masteries | 12 | 9: Lightning, Chain Lightning, Telekinesis, Inferno, Blaze, Fire Wall, Hydra, Enchant, Frozen Orb |
| Paladin (30) | 3: Zeal, Might, Holy Fire | 23 | 4: Charge, Vengeance, Conversion, Fist of the Heavens |

- Tier evidence: unchanged from the freebuild prototype. Supported = a live-verified loop named in `internal/combat` comments or a smoke doc; unsupported = on a profile avoid list or no cast loop (Lightning and Chain Lightning are on every Sorceress `avoid` list, broken on the d2d build in use). **U**: tiers are typed, not generated. ADR 0020 wants them GENERATED from Proving Ground matrix rows; that wiring is P2.
- `reqLevel`, `reqSkills`: V for the Sorceress (read from skills.txt by `docs/research/sorc-skill-matrix.md`), S for the Paladin (the SME read, `internal/leveler/palprobe_test.go`, which checks it against a real install on the worker host). Cross-check V (this card): every shipped Paladin and Sorceress plan table except two legacy ones validates against the table and round-trips (section 5), which a wrong requirement would break.
- `synergies`: 8 of the 15 Sorceress and 3 of the 7 Paladin damage skills have no list (U, unknown). `Caps.Overlay(db.SkillByName)` replaces requirements and synergies by the game's own skills.txt expressions (`data.Skill.Synergies()`) when the tables are loaded, which the d2bot wire does at start; offline the planner raises `synergy_unknown` and its damage is a lower bound.
- `planner.priorMargin` (2.0, U) and `mana` (the sustain budget, U) live in the same file.

## 3. The planner

`Derive(allocation, env)` classifies every skill with points, estimates sustained damage per attack (`power.SkillData.Estimate`: V game functions, gear speeds flagged when unread), and picks the main attack, a second element for immunes, the aura, buffs, summons and movement, each with a one-line reason. **The combat profile's pick is the prior**: it is kept unless the estimate beats it by `priorMargin` (design review 04 section 3.2). Warnings: `no_attack`, `tourist`, `single_element`, `unsupported_points`, `weapon_unread`, `mana_starved`, `prereq_only`, `synergy_unknown`. Unknown class or skill is an error naming it; an unsupported skill is never chosen into any slot (must-fail test `TestUnsupportedNeverChosenAndAlwaysFlagged` runs it for every unsupported row of every class in the table).

Viability (`Judge`): the zones of the hero's act at its difficulty run through `power.Model.Zone` with the plan's archetype and attack under the default dial; the line says safe, risky or too hard for the hardest zone, how many zones are in each class and the least certain term ("estimate, defence default"). Nothing new is modelled; the numbers are the survival model's. Without the zone table or a resolved archetype the line says "not judged" and the view's `note` says why.

`GET /api/hero/plan` (`didle.plan/1`): `enabled`, `mode` (off or shadow), `shadowOnly` (always true in P1), `rotation[]` (`slot`, `skill`, `level`, `tier`, `why`), `aura`, `archetype`, `profile` (what the combat profile attacks with today), `agrees`, `unplayed[]`, `warnings[]`, `viability`, `line`, `note`.

The shadow line, once per virtual minute: `planner shadow: clvl=18 class=paladin plan=Zeal/4 profile=Zeal agree=true best=Sacrifice aura=Might auraProfile=Might arch=bruiser unplayed=0 warnings=tourist,prereq_only via=act1:safe`. `best` is the raw highest estimate; `plan` the prior-adjusted pick; `agree` is plan against profile.

How the hero is read: hard points by `leveler.OwnedSkills` (base level, items excluded), the power model's hero from `power.HeroFromPlayer`, the weapon average from stats 21 and 22 (class "1HS" with a flag when the class is unread, U), the combat line from the plan table the leveler would pick (`leveler.SelectBuild`; U that this is the table the player follows once points are free).

## 4. Agreement with the combat profiles (V, `go test ./internal/agency -run AgreementFixture -v`; fixture `internal/agency/testdata/agreement.json`)

Six plan tables times clvl 6, 12, 18, 24, 30 (weapon: a nominal average 20, U). The planner names the same attack as the profile in **30 of 40 checkpoints; 25 of 30 for Paladin and Sorceress** (the prototype without the prior margin: 23 of 30 over its six tables). Every mismatch is documented in the fixture's `why`:

| checkpoint | profile | planner | reason |
|---|---|---|---|
| paladin-zealot-lite and paladin-normal30-v2, clvl 18 | Zeal | Sacrifice | estimate 93 dps against 37 (2.5x, over the 2.0x margin): Zeal counted as one strike, Sacrifice's life cost (8 percent of damage dealt, V calc) not modelled |
| sorceress_blizzard_endgame90, clvl 18, 24 | Ice Bolt | Ice Blast | estimate 103 and 226 dps against 20: the profile's cold band leads with the cheap bolt; a measured fight should decide which is right (U) |
| sorceress_blizzard_endgame90, clvl 30 | Blizzard | Ice Blast | 226 against 111 dps: Blizzard's delay and area are not in the single-target estimate |
| necromancer_summoner_normal30, all five | Teeth | Attack (plain swing) | pets have no damage in the model; the archetype still resolves as summoner and Main is the top summon, `Attack` is only the filler |

Barbarian and Necromancer rows are paper rows: no engine support is needed, the table says which kinds the engine cannot play and the plan flags them (`TestPaperClasses`).

## 5. Build files (`internal/buildfile`)

`didle.build/1` is the goal-list file of the prototype (README there). P1 adds the bridge to the shipped tables: `ExportPlan(name, rules)` turns any `internal/plandata` table into a file; `ImportPlan(file, rules)` validates, expands and writes the plan table JSON the leveler reads; `LoadFile` and `SaveFile` read and write validated files. `agency.Caps.BuildRules()` is the rules source for Paladin and Sorceress without the game tables; for a class the table has no requirements for, `Validate` says `skill_unknown` and never guesses.

V (`TestShippedTablesValidateAgainstCapabilityRules`, `TestPlansRoundTripThroughFiles`): all 10 Paladin and Sorceress tables except `primary-blizzard` validate against the capability rules and expand back to the same cumulative allocation at every row; all 24 tables round-trip under permissive rules. **Findings** (tables untouched): `primary-blizzard` puts 43 points into Static Field (cap 20); `fallback-lightning` and `-v2` take Lightning Mastery at clvl 26 (required 30) and Chain Lightning at 27 without its prerequisite Lightning (the game's rule, d2d `can_learn`, refuses both).

## 6. Tests

`go test -p 1 -race -count=20 ./internal/agency ./internal/buildfile` (V); `go test -p 1 ./internal/watch -run Plan`, `./cmd/d2bot -run 'Agency|ChainSnap'`, `./internal/env`, `./tools/knobdump` (knobs.md in step). Agreement fixture: `UPDATE=1 go test ./internal/agency -run AgreementFixture` rewrites it after a deliberate change.

## 7. Not done, and the live checklist

- Scenario `agency.plan-matches-profile` is PENDING: a scenario asserts game state, the shadow planner only writes a log line. It needs an assertion kind that reads a bot log line (`log_line`), then the matrix runs Paladin and Sorceress at clvl 18. The offline equivalent runs on every merge (section 4).
- Not built (later phases): the planner driving the hero (P2: replace first-known-in-band by the ranked candidates, the table as the unsupported filter), measured kill speed correcting the estimate (`MeasuredDPS`), the allocate API with preview, tiers generated from Proving Ground rows, `POST /api/hero/track` (P3), respec as a player event, synergy lists for the 11 damage skills without one, multi-strike and life-cost terms in `skilldamage.json`.
- Live (U, needs a lab d2d): set `agency.planner` 1 on a Paladin at clvl 18 with points spent by hand, read one `planner shadow:` line within a game minute, open the Points page and see the build line, compare `GET /api/hero/plan` with the line; set the knob to 0 and see both go quiet. Check that a run with 1 equals a run with 0 in quest progress and kills.
