# BUILDV3: the fresh Sorceress skill table, rewritten to the order the leveler follows

> status: table shipped OPT-IN by name (`fallback-lightning-v3`); the default of a fresh Sorceress stays `fallback-lightning-v2` until the live soak below passes.

User decision (2026-11-01, popup "Apply the proposed rows"): apply the SME's rows. Why: the SKPASS replay (`docs/research/sk-passives-trace.md`, B1) showed the leveler spends a point it may not spend on the NEXT legal row, so v2 effectively ran Charged Bolt to 20 by clvl 24, Lightning 1 at clvl 25, Chain Lightning from 26, Lightning Mastery from 30 (the rows said 40, 27 and 26). The spend-on-next-legal-row behaviour is kept (unattended safety, no saving of points); the ROWS are fixed so each is legal when its point arrives.

## Data and selector

- Table: `internal/plandata/build-fallback-lightning-v3.json` (copy in `docs/plan/data/`, a test keeps them equal). The v2 file and the legacy `fallback` file rows are unchanged; both v2 and v3 gained one metadata key, `"combat": "lightning"` (V: `go test ./internal/leveler`).
- Selector: the existing generic one. `leveler.buildFiles` maps any table name to a plandata file; config `levelerBuild` / `--build <name>` wins over `SelectBuild` (V `cmd/d2bot/class_test.go`: v3 by name gives leveler `fallback-lightning-v3` and the lightning combat line). Nothing new was needed for selection; adding another table is one JSON file plus one map line.
- Combat line from data: `cmd/d2bot/class.go` used to compare the table name to `"fallback"` / `FallbackV2`. It now asks `leveler.TableCombat(name)` (the table's own `"combat"` key), so any table can declare the lightning line. The legacy `fallback` name check stays (its file is unchanged on purpose).
- Default flip (live check owner, after the soak): `freshBuild["sorceress"] = FallbackV3` in `internal/leveler/build.go` and `cmd/d2bot/class_test.go` expectations. `SelectBuild` compares only rows up to `v2Rows` (15) for the Static Field legacy guard; v3 rows 2-10 equal v2, rows 11-15 differ only in Lightning at 12 (Soaky keeps `fallback` either way, since Static Field is not in v3).

## Approved rows (quest replay: row N is spent at clvl N)

| clvl | skill | why (V: reqlevel / reqskills from patched Skills.txt) |
|---|---|---|
| 2-10 | as v2: Charged Bolt to 8, Frozen Armor 1 (Den point), Frost Nova 1 | unchanged |
| 11 | Telekinesis 1 | reqlvl 6, Teleport prerequisite |
| 12 | Lightning 1 | reqlvl 12, Charged Bolt met; +synergy to Charged Bolt and Chain Lightning |
| 13-17 | Charged Bolt to 13 | |
| 18 | Teleport 1 | reqlvl 18 |
| 19-30 | Chain Lightning one point per level (12 points, Radament's point included) | reqlvl 18, Lightning met |
| 31-35 (clvl 30-33) | Lightning Mastery 1-5 (Izual's 2 points at 33 go to LM 5 and Chain Lightning) | reqlvl 30 |
| 36-43 | Chain Lightning to 20 | |
| 44-58 | Lightning Mastery to 20 | |
| 59-77 | Lightning to 20 | |
| 78-84 | Charged Bolt to 20 | |
| 85-99 | filler: Teleport, Frost Nova (U, past any soak) | |

Removed vs v2: the second Frozen Armor point (nothing it feeds is cast), Frozen Armor at 22. No Warmth, Energy Shield or Static Field (decided). Stat columns are v2's (5 Vitality per level after the Str floor; the Str rows at 25-32 etc. are v2's).

## Replay (`python3 tools/emu/sk_pass_plan.py internal/plandata/build-fallback-lightning-v3.json quests|none`)

I (replay of the bot's own rule). Effective order, quest points Den 7, Radament 20, Izual 33:
Charged Bolt 1-4 (clvl 2-5); Frozen Armor (6); Frost Nova (7) and Charged Bolt 5-8 (7-10); Telekinesis (11); Lightning (12); Charged Bolt 9-13 (13-17); Teleport (18); Chain Lightning 1 at 19, 3 at 20 (Radament), then 4-12 at 21-29; Lightning Mastery 1-4 (30-33) and 5 (33); Chain Lightning 13-20 (33-40); Lightning Mastery 6-20 (41-55); Lightning 2-6 (56-60).

Without quest points (everything about one to two levels later, same order): Telekinesis 12, Lightning 13, Teleport 19, Chain Lightning 1-12 at 20-31, Lightning Mastery 1-5 at 32-36, Chain Lightning to 20 at 44, Lightning Mastery to 20 at 59.

No row is refused: `TestBuildV3SpentInTableOrder` (`internal/leveler/buildv3_test.go`) replays `Build.Plan` over clvl 2-60 with and without the quests, with a rule that records every refusal, and asserts (1) zero refusals, (2) every spent point equals the next table row in order, (3) the approved milestones at their clvl (quests case), (4) no Static Field, Warmth or Energy Shield, Frozen Armor stays 1. The reqlevel/reqskills inside the test are copied from the replay tool so the test needs no game data. The old v2 replay refuses Lightning Mastery (26, 28-31) and Chain Lightning (27, 32-39) rows, which is what moved them.

## Damage per band (COMPUTED, replay's own formula, plus or minus 15 percent; best of Lightning and Chain Lightning, average hit in whole points, no items or resistance; v2 to v3)

| clvl | v2 (quests) | v3 (quests) | ratio | ratio, no quests |
|---|---|---|---|---|
| 25 | 52 (Lightning 1) | 92 (CL 8) | 1.75x | v2 has no lightning skill yet |
| 29 | 68 (CL 4) | 132 (CL 12) | 1.96x | 2.01x |
| 33 | 154 | 282 | 1.83x | 2.07x |
| 37 | 238 (CL 9) | 365 (CL 17) | 1.54x | 1.83x |
| 45 | 312 | 567 | 1.82x | 1.79x |

The SME's about 1.8x / 1.9x / 1.4x for 25 / 29 / 37 reproduces (1.75 / 1.96 / 1.54; the clvl 37 figure is the one that moved). Charged Bolt is one bolt per cast and several fire, so a per-bolt comparison against Chain Lightning's hit understates Charged Bolt (open question 1). v3 gives up Charged Bolt 14-20 until clvl 78+: its level 13 stays the filler attack before Chain Lightning.

## Combat code at clvl 19-25 (what the attack order does)

Before: `internal/combat/classdata/profiles.json` lightning profile had one band `minLevel 26` (Chain Lightning, Lightning, Charged Bolt, Fire Bolt) over a base band (Charged Bolt, Fire Bolt, Ice Bolt). A skill that is not learned is skipped (`fighter.go` builds the order from the known skills), so with v3 a Chain Lightning learned at 19 would have sat unused until clvl 26 while Charged Bolt kept being fired. Change (smallest, data only): a new first band `requires: "Chain Lightning"`, `minLevel 0`, the same attack list. The switch is therefore derived from the learned skill, not a clvl literal; the `minLevel 26` band stays below it, so v2 and the legacy table (Chain Lightning learned at 26 / 27) behave exactly as before. `TestClassDataLoads` now lets a `requires` band sit above lower-level bands (the skill gates it); `TestSorceressProfileKeepsOldBehaviour` was changed deliberately: "knows everything at clvl 25 uses Charged Bolt" became "has not learned Chain Lightning at clvl 25 uses Charged Bolt" (the old test's `has()` said every skill is learned). New pure tests: clvl 19 with Chain Lightning learned attacks with it, clvl 22 without it keeps Charged Bolt, clvl 26 keeps the older level band. Mana: Chain Lightning at 1-3 points costs more mana than Charged Bolt 13 (U, no data: open question 2). Other combat policy untouched.

## Live check recipe (owner flips the default after it)

1. Fresh Sorceress, config `levelerBuild: fallback-lightning-v3` (start log shows `levelerBuild=fallback-lightning-v3`, lightning combat line).
2. Soak clvl 19 to 30 in one area (same area and gear rule for both runs), twice: current (v2) and new (v3). Log per run: kills per minute, mana drinks per 10 min, deaths, rests per pack, time per clvl, skill points per level from the `Spending stat and skill points` lines (they must show Lightning at 12, Teleport 18, Chain Lightning each level from 19, Lightning Mastery from 30).
3. Pass: no more deaths than v2, kills per minute not lower from clvl 22, mana drinks per 10 min not above v2 by more than a third. Then flip `freshBuild` and the d2bot test expectations; the matrix and `docs/earlygame.md` get one line.
4. Not run here (no game): everything above is UNVERIFIED live.

## Open questions (SME UNSURE items)

1. Lightning at clvl 12 vs another Charged Bolt: the replay's Charged Bolt figure is per bolt and Charged Bolt fires several bolts, so the early gain is the synergy, not the 1-point Lightning hit.
2. Mana drinks per band: no data until the soak (Chain Lightning at 1-3 points).
3. Rows 85-99 filler is a guess (U).

## Build-specific assumptions found (for BUILDFLEX; not fixed here)

- `internal/leveler/build.go:131-140` `buildFiles`, `:203-210` `FallbackV2`/`PaladinV2` constants and `freshBuild` map: one fresh table per class, named in Go; default flips are code edits.
- `internal/leveler/build.go` `v2Rows = 15` and `SelectBuild`: the "owns a skill the new table dropped" check only looks at rows up to clvl 15 and assumes legacy vs v2 pairs.
- `internal/leveler/build.go` `defaultBuild` map (`sorceress: "fallback"`, `paladin: "paladin-normal30"`) and `LoadBuild` `name == "fallback"` energy shift (`sorcEnergyPerRow`).
- `cmd/d2bot/class.go:35` `c == d2.Sorceress && lb == "fallback"`: the legacy name still decides the lightning combat line (v2 and v3 now use the table's `combat` key); `c == d2.Sorceress` and `combat.Build` as a two-value enum (`BuildBlizzard`, `BuildLightning`) in `internal/combat/skills.go:17`, `ProfileFor` maps only Sorceress builds (`internal/combat/classdata.go`).
- `internal/combat/classdata/profiles.json` lightning profile: `minLevel 26` band and the F-key layout (`Static Field` f1, `Chain Lightning` f5) are fixed per profile, not per table; Energy Shield buff is listed though no table learns it.
- `internal/combat/skills.go:125` `attackOrder` always uses the Sorceress profile.
- `docs/combat.md` and the boss scripts' lightning-specific thresholds were not audited here.

## Shared-file lines for the merger

- `docs/decisions.md` (coordinator adds; I do not edit it): "2026-11-01: fresh Sorceress skill rows rewritten to the SME's order (Lightning 12, Teleport 18, Chain Lightning from 19, Lightning Mastery from 30); shipped as table `fallback-lightning-v3`, opt-in by name, default stays `fallback-lightning-v2` until a live soak of clvl 19-30 passes (user popup 'Apply the proposed rows'). docs/buildv3.md".
- `docs/plan/BACKLOG.md`: "BUILDV3-LIVE: soak clvl 19-30 v2 vs v3 (recipe in docs/buildv3.md), then flip `freshBuild` to `FallbackV3`."; "BUILDFLEX: take the list 'Build-specific assumptions found' in docs/buildv3.md."
- `CHANGELOG.md` (Unreleased): "Added the opt-in Sorceress table `fallback-lightning-v3` (Chain Lightning from clvl 19) and a Chain Lightning attack band gated by the learned skill."
- `openspec/changes/INDEX`: buildv3 (in-progress then archived at merge).
