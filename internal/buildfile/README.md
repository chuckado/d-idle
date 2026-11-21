# buildfile: the shareable build file (`didle.build/1`)

> status: current (card agencyp1a, 2026-11-19). Promoted from the prototype branch `agent/fable-proto-buildfile`; the design is in `docs/agency.md` section 5 and `docs/design/fable-review-2026-11-19/04-stat-skill-agency.md` section 3.6. Not wired into the leveler yet (P3 adds `POST /api/hero/track`). Tags: V = ran in a test of this package, U = unverified.

`plans.go` is what P1 adds: `PlanNames`, `ExportPlan` (a shipped plan table of `internal/plandata` to a build file), `ImportPlan` (validate, expand, write the plan table JSON the leveler reads), `LoadFile` and `SaveFile`. The capability table is a rules source (`agency.Caps.BuildRules`) for Paladin and Sorceress, so those files validate without the game tables.

## What it is
JSON, strict decode (unknown fields, a wrong `schema`, trailing data are errors). Skill and class names only, never ids.
- `skills`: ordered goals `{skill, to, when?: {clvl?, after?: {skill, level}}, hold?}`. "Raise skill to N hard points."
  Prerequisites are added automatically, 1 point each, as late as possible (not before the dependent skill itself is spendable).
- `stats`: ordered `{stat: str|dex|vit|ene, to: N | "rest", per?, when?: {clvl}, req_for?}`. `to` is the absolute value on
  the character panel (kolbot style), `per` caps the points per level so "3 Str + 2 Vit of every 5" is two goals; `req_for` is a comment.
- `phases?`: `{name, from_clvl, until_clvl?, skills}`: "Fire Bolt until 17, then Fire Ball" with NO respec (goals open at
  `from_clvl`, unmet ones are dropped after `until_clvl`). `respec_at: {clvl}`: phases are the pre-respec path, `skills` the post-respec build.
- `quest_points?`: `[{clvl, skill, stat}]` extra points the player takes; absent = none assumed (`StdQuests()` has the 12 of the endgame tables).
- NOT in the file: main attack, hotkeys, combat line. The engine derives them (the skill with most points in the current
  phase) so a shared file cannot disagree with its own points. The plan tables carry `hotkeys`/`combat`/`main_attack` today; that stays engine data.

Expansion rule (`Expand`, deterministic, clvl 2..99): each level adds 1 skill point (+ quest) to a bank; points go one at a time
to the FIRST pending step that is spendable (clvl >= when and skill reqlevel, `after` met, prerequisites owned, previous point of
that skill bought). Blocked steps are skipped, `hold` (kolbot `satisfy`) makes a blocked step block everything after it. A point
nothing can take stays banked and is spent the level the next goal unlocks, like the tables (V: `build-paladin_hammerdin_endgame90.json` "schema" text).
Stats do the same with 5 points per level. `Validate` codes: class_unknown, skill_unknown, skill_wrong_class, target_range,
prereq_cycle, skill_unreachable, skills_over_budget, stat_unknown, stat_target, stats_over_budget, rest_not_last, when_invalid, phase_invalid, respec_invalid, quest_invalid, no_goals, stat_after_unsupported.

## Examples (full files in `testdata/`, all validate and expand: `TestExamplesValidateAndExpand`)
- Paladin `paladin-hammerdin.json`: Might, Blessed Hammer 20 from clvl 18 (Holy Bolt comes with it), Holy Shield (Smite, Charge, Defiance, Sacrifice
  come in), Fanaticism from 30, Concentration, Vigor; Str 60, Dex 70 at 2/level, rest Vit; three quest points.
- Sorceress `sorceress-phases.json`: phase "bolt" Fire Bolt to 12 until 17, then Fire Ball 20, then Fire Bolt 20 `after` Fire Ball 20.
- Necromancer `necromancer-respec.json`: phase Teeth 20 + Amplify, `respec_at` 24, post build Bone Spear 20 + Raise Skeleton 5. `Expand` returns
  `Pre` (clvl 2..24), `Pool` (the 23 refunded points respent), `Post` (25..99); nothing is executed (`ToLeveler` maps to `leveler.Build{Respec, RespecPool}`).
- Barbarian `barbarian-after.json`: Frenzy 20, Iron Skin only `after` Frenzy 10.

One-line chat form (`Encode`/`Decode`): `didle1:pal:Zeal20,Holy Shield20@24,Fanaticism20|str=60,dex=70/3,vit=rest`
(`@24` = when.clvl, `/3` = per). No room for name, phases, respec, quest, `after`, `hold`: `Encode` errors instead of dropping them.
`ImportKolbot` maps `[["Blessed Hammer",20,true],...]` (names, not ids; satisfy true/missing -> `hold`) and `[["strength",60],["vitality","all"]]` ("all" -> rest; "block" is an error).

## Prior art (ideas only; clones read-only, untrusted)
| | kolbot AutoSkill/AutoStat | SoloPlay build file | Koolo leveling | our plan table | this format |
|---|---|---|---|---|---|
| shape | `[[id,count,satisfy]]`, `[[stat,target]]` | JS module, `skills:[[id,count,satisfy?]]`, `stats:[[name,target\|"all"]]`, plus gear/charm data | Go `SkillPoints() []skill.ID` (one id per point), `StatPoints()` targets | one row per clvl, points per skill and stat | goals + targets, rows derived |
| prerequisites | auto (`needPreReq`) | via kolbot | none, listed by hand | none, listed by hand | auto, as late as possible |
| respec | none | `respec()` at fixed levels, build "Start/Stepping/Final" | `ShouldResetSkills()` at clvl 24 | `respec` block, not implemented | phases without respec, or `respec_at` computed only |
| stats | absolute targets, "all", "block" | same | absolute targets per level hint | exact points per level | absolute targets, rest, per |
| validation | clamps count to 20, no more | none seen | none seen | `ParseBuild` shape only | 17 problem codes |
| who writes it | a player editing a config | a script author | a Go developer | an agent or the coordinator | a player or a friend |
V sources: `blizzhackers_kolbot/.../libs/config/Paladin.js` (AutoSkill/AutoStat comments), `libs/core/Auto/AutoSkill.js`,
`blizzhackers_kolbot-SoloPlay/r/BuildFiles/paladin/paladin.HammershockBuild.js`, `r/Core/Quest.js` (respec by level), `vietdungdev_Koolo-DiobyteVersion/r/internal/character/paladin_leveling.go`.
Kolbot's `skillCheck` only spends when `hardPoints <= charlvl - reqlevel` (V, AutoSkill.js): one more clvl per extra point. Our tables bank
and dump 5 points at once at the unlock level (V, hammerdin row 18), so one of the two is wrong or kolbot is conservative (U: needs the game).

## What it proves (V = `go test -p 1 ./internal/buildfile/`)
- Compression with exact round trip: `Expand(FromRows(table))` equals the table's cumulative skill and stat allocation at EVERY row, not
  only clvl 10/20/30/last, for 8 tables: sorceress blizzard_endgame90 + normal30, paladin hammerdin_endgame90 + normal30, necromancer
  summoner_endgame90 + leveling_normal30, barbarian warcry_endgame90 + melee_normal30 (`TestRoundTripTables -v`). 89 rows become 17..25 skill goals and 6..13 stat goals (2 to 3 KB JSON).
- Quest points are inferable: the smallest schedule that makes a table affordable gives 8-9 rows for the endgame tables (matches the
  12 points of the "schema" text) and 0-3 for the normal30 tables, which spend ahead of 1 point per level (finding: they quietly assume quest points).
- Banking, late prerequisites, `after`, `hold`, phases without respec, a respec (two row sets), per/rest stat goals, every problem code, strict parse, line and kolbot import all have tests.
## What it does not
- Skill data is a hand-written fixture of 33 skills (U reqlevels and prerequisites, not skills.txt; the repo has no extracted skills.txt). The table round trip uses
  permissive rules, so it proves the format, not the tables' legality. `DataRules` (adapter over `data.Skill`) compiles but never ran on real data.
- `FromRows` output is machine style (every goal carries `when.clvl`, prerequisites explicit): exact, not pleasant to edit by hand.
- Nothing runs `leveler.Plan` on an expanded build, no live game, no stat gear floors evaluated, "block" Dex target absent.

## Open questions
1. Names vs ids: names are readable and survive a skills.txt reorder, but break on a renamed skill; need an alias table and one canonical spelling (case-insensitive today).
2. Localisation: skills.txt keys (English) are the identity; display names are not accepted.
3. Versioning: `schema` is the file grammar; the game data version (1.14d skills.txt, a mod) should be a separate optional `data` field, and reqlevel/prerequisite changes make `Validate` the gate.
4. A player who spends points by hand: the leveler counts what is owned against the table (`Build.Plan`), so a file is then only a SUGGESTION track; the first deviation should not invalidate it, the planner re-expands from the owned state. Not prototyped.
5. Quest points: assume none (safe, under-spends) or the standard 12 (over-spends if skipped)? A file field is the honest answer but players forget it.
6. Stat goals for item requirements (`req_for` is a comment today) and block chance need gear and block formulas, which the file should not own.
