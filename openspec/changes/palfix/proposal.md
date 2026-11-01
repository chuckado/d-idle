## Why

PALPROBE (archived) left five ready fix cards for the Paladin and the build-flexibility seams it ran into (docs/classes/paladin-probe.md section 3): PALMANA, PALBAND, SIMCLASS, PALTABLE, PALZEAL. This change implements them offline. The Sorceress path is unchanged (every existing Sorceress test passes untouched).

## What Changes

- PALMANA: a table flag `mana_heavy: true` read by `internal/resource` replaces the Go name list; the four old builds are migrated and the Paladin hammerdin table is now mana-heavy.
- PALBAND: the Paladin Blessed Hammer band starts at clvl 18 (the skill's required level) instead of 24.
- SIMCLASS: the resource simulator takes its main attack and weapon skill from the scenario (`WithLadder` from a profile ladder); Fire Bolt and Staff stay the defaults; a Paladin stock scenario.
- PALTABLE: the legacy table `paladin-normal30` trains Zeal to 4 (rows 13 to 15) and Prayer (row 16) in place of five Holy Fire rows.
- PALZEAL: plandata build tables may name themselves (`selector`, `endgame_for`, `default_for_class`; fill-only), and `build-paladin-zealot-lite.json` is added, selectable with `--build paladin-zealot-lite`, never a default.

## Impact

Code: see `owns`. No spec, ADR or glossary change. Offline only.

## Overlap (declared)

Card buildv3 (open) owns `internal/leveler/build.go` and `internal/combat/classdata/profiles.json`; card skmeleewire (open) owns `internal/resource/sim.go`. This change edits them minimally and does not claim them: `build.go` gains `tableMeta` and `registerDataTables` (an additive block before `TableCombat`), `profiles.json` changes one number (the paladin Blessed Hammer `minLevel`), `sim.go` gains two Scenario fields, `WithLadder`, `BloodMoorPaladin` and two lines in `input`. The hunks are disjoint from those cards' rows; merge order does not matter.
