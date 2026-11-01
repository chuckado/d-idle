# PALFIX: the Paladin fix cards

Offline fixes of the five cards PALPROBE left (docs/classes/paladin-probe.md section 3). Nothing here was run in a game. Tags: V = verified by a test in this change or a file read, U = unverified.

## PALMANA (F5, audit row 26)

- Before: `manaHeavyBuilds` was a Go map of four names, and `paladin_hammerdin_endgame90` was missing, so a hammerdin kept the single mana column in Nightmare and Hell.
- Now: a build table says `"mana_heavy": true`; `internal/resource` scans the embedded `build-*.json` once (`loadManaHeavy`). The four old tables carry the flag, plus the Paladin hammerdin. The generator `docs/plan/data/builds-endgame/gen_endgame.py` writes the flag (`MANA_HEAVY`) so a regenerate keeps it. `potions-belt.json` "builds" lists five names and `TestManaHeavyMatchesJSON` still compares count and names. V: `TestManaHeavyFromTableFlag` (5 builds; the zeal hammerdin, v2 and others are not), `TestPlanForBuildManaHeavy` (its old "paladin is not mana-heavy" row now uses `paladin_hammerdin_zeal_endgame90`: the one deliberate behaviour change).
- U: the 35 percent drink threshold and the two-column belt for a hammerdin need a live Hell check, like the other four (W7 of docs/builds-endgame.md).

## PALBAND (F2, audit row 10)

- Before: the Paladin Blessed Hammer band had `minLevel` 24; the tables train the hammer at 20 (legacy) and 22 (v2), so the points idled for 2 to 4 levels (replay: first hammer point at clvl 20 / 22 with the Den point).
- Now: `minLevel` 18, the skill's own required level (skills.txt reqlevel 18, fixture of `TestPalProbeRulesMatchInstall`, which skips without an install: S/U for the live file); `requires: Blessed Hammer` still gates it. Replay ladders now start with Blessed Hammer the level the table trains it (legacy clvl 21, v2 clvl 23 in the no-Den replay). V: `TestPalProbeHammerBandStartsAtTraining` (formerly a skipped finding) asserts it for both normal tables and both endgame hammerdin tables.

## SIMCLASS (F7, audit row 33)

- Before: `world.input` hard-coded a Fire Bolt slot and a Staff skill.
- Now: `Scenario.MainAttack` / `WeaponSkill` (empty = Fire Bolt / Staff), `Scenario.WithLadder(names)` takes the first and last rung of a profile ladder, and `BloodMoorPaladin()` is the stock melee scenario (clvl 12, Zeal 2 mana as main attack, Attack as weapon rung; life, mana, damage U). V: Sorceress stock replay pinned by `TestSorceressStockReplayPinned` (22 kills, 28 swings, 66 casts: identical); Paladin stock: 34 kills, 65 Zeal casts, 17 swings, 0 deaths, 1 trip (`TestBloodMoorPaladinReplay`, shape only); `TestPaladinScenarioFollowsProfileLadder` (internal/combat) checks the scenario names equal the profile's clvl 12 ladder.
- U: the model has no Zeal multi-hit or hold, so the Paladin numbers are the policy's shape, not a prediction.

## PALTABLE (F3, F4)

Decision: fix the legacy rows, conservatively. A character already on `paladin-normal30` owns skills the v2 table dropped, so only its future points change.

- Before (no-Den replay to clvl 30): Zeal 1, Prayer 0, Holy Fire 11.
- Now: rows 13, 14, 15 train Zeal instead of Holy Fire and row 16 trains Prayer instead of Holy Fire. Replay: Zeal 4 (4 by clvl 16, same curve as v2), Prayer 1 (the planner pulls it into the idle clvl 10 point), Holy Fire 7. Skill points per row (one) and stats are unchanged; Sacrifice precedes Zeal, Zeal first at 12. V: `TestPalProbeLegacyTableHasNoPrayer`, `TestPalProbeLegacyZealStaysAtOne` (both formerly skipped findings), `TestPalProbeTablesSimulate` (legality with the probe's skills.txt fixture).
- U: whether Zeal 4 beats four more Holy Fire levels for a legacy character in Hell; the SME made that call for v2 and the legacy rows now follow it.

## PALZEAL (the S4 seam and the Zealot-lite table)

- The seam: a `build-*.json` table in `internal/plandata` may say `"selector": "<name>"` (the `--build` / `levelerBuild` name), optionally `"endgame_for": "<class>"` and `"default_for_class": "<class>"`. `registerDataTables` (internal/leveler/build.go, runs in `init`) adds them to `buildFiles` / `endgameBuild` / `defaultBuild` fill-only: a taken selector, a class with an endgame table, or a class with a default (Go map or `gear.DefaultBuild`) is never overridden. V: `TestRegisterDataTablesFillsGapsOnly`, and `TestSelectionUnchangedForAllClasses` pins `DefaultBuildName`, `EndgameBuildName`, `ResolveBuild(endgame)` and `SelectBuild` (fresh and existing character) for all 7 classes; the same test passes unchanged on the commit before the seam (run in a scratch worktree).
- The table: `build-paladin-zealot-lite.json` (selector `paladin-zealot-lite`) is the recon section 3.2 table (Might, Sacrifice, Smite at 4, Prayer 3, Zeal 4 at 12 to 15, Defiance and Cleansing at 16, Vigor 18 to 19), rows 20 to 30 from v2, 29 rows, `respec` off, hotkeys Might / Zeal / Blessed Hammer / Concentration. It carries no endgame or default metadata: `--build paladin-zealot-lite` only. V: `TestZealotLiteLegal` (no illegal pick with or without the Den point, Zeal 4 first at 12 or 13, Sacrifice before Zeal, nothing stuck, 29 or 30 points spent), `TestZealotLiteSelectableByName`.
- U: the SME review of the row choices, the legality against the real skills.txt (no install here: `TestPalProbeRulesMatchInstall` checks the fixture where an install exists), and every live effect.

## Dead ends

- Not extended to the other six classes' per-class literals; this change only removes the five findings.
- `docs/plan/data/builds-endgame/*` has no `build-` prefix, so `TestMatchesDocs` skips them (it stops at the first doc it cannot find); the embedded copies were edited by hand to match.

## Shared-file lines for the merger

- docs/BACKLOG.md: replace the line "PALBAND/PALTABLE/PALMANA/SIMCLASS/PALZEAL (from palprobe): ready fix cards" by "PALBAND/PALTABLE/PALMANA/SIMCLASS/PALZEAL: done offline (palfix, docs/palfix.md); still owed live: Zealot-lite and the legacy rows in a game, SME review of the Zealot-lite rows".
- docs/INDEX.md: `docs/palfix.md` (Paladin fix cards, offline).
- CHANGELOG.md: "Paladin fixes: mana-heavy flag in the build table (hammerdin included), Blessed Hammer band from clvl 18, simulator takes the class's attack names, legacy Paladin table trains Zeal 4 and Prayer, data tables can name themselves and `--build paladin-zealot-lite` exists".
