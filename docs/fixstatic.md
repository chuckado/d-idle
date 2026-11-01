# fixstatic: Static Field and the fresh skill table

> status: current. Card `openspec/changes/fixstatic`. User decision 2026-11-01: drop Static Field for fresh characters (SKILLINV finding F1).

## Why

A fresh Sorceress never learns Static Field: the BUILDV2 table buys Charged Bolt, Frozen Armor, Frost Nova and Telekinesis, and Static Field only at clvl 86 (V: `TestBuildV2SorceressSequence`, in `internal/leveler/early_test.go`). Static Field never kills, only shaves life, and its radius is 5 subtiles at level 1 (`staticRange`, S Kolbot formula) while the Andariel script kites at 10+ subtiles (`internal/boss/act1.go`, `Kite.Range` 10). SME advice agrees: no point on a fresh character.

## Result

- Removed: the assumption in prose. `docs/combat.md`, the Andariel card text and `internal/boss/script.go` (`BeforeAttack` comment) no longer say Andariel opens with Static Field. The Andariel script never had a Static Field step in code (V: grep of `internal/boss`), it only had the card sentence.
- Kept, gated by the skill being learned: `Fighter.staticOpener`, `fallbackSkill`, `attackBlocked` and the knobs `combat.staticCasts` / `combat.staticCastsBR` / floors. All three look the skill up in `KnownSkills` and return nothing when the character does not own it (V by reading, new test `TestStaticFieldNeverPlannedWithoutTheSkill`). So on the fresh table they are already inert and need no code change; deleting them would change the legacy table (`SelectBuild` keeps it for characters that already own Static Field, e.g. Soaky, `TestBuildV2ExistingCharacterKeepsOldTable`) for no gain. Knobs doc is unchanged.
- Frost Nova / Charged Bolt kiting is untouched.
- Skill table name: the bot `minute` line now carries `table=<name>` (`Runner.buildName`: the configured table or the `SelectBuild` choice, `-` when none). `d2dsoakstat.Minute.Table` and `Run.Table` parse it (older logs: empty); `d2dsoak` aggregate prints one `table=...` line above the metric table. Other parsers read the line by key and ignore the new one (V: tests in `cmd/d2soak`, `internal/d2dsoakstat`).

## Open drill idea

Andariel kill time with 1 point in Static Field vs 0 (a legacy-table character against a fresh one) at equal clvl and gear: likely no measurable gain given the range gap; not run (U).
