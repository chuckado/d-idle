## Why

BUILDFLEX audit row 5 (seam S3): `Navigator.TeleportKey` has no production assignment (only tools/rigs set it), so Teleport hops in navigation, the task layer's `CanTeleport`, boss `NeedsTeleport` stops and the farm table's Teleport flag never fire in production, although the Fighter already binds Teleport to its layout key. The project must support more than the one lightning Sorceress build, so the capability is derived from what the character knows and has bound. docs/fixtelekey.md holds the call graph.

## What Changes

- internal/nav: `TeleportKeyFn` (key supplied at use time) and `TeleportOK` (level rule) on the Navigator; `HasTeleport()`; hops and explore legs skip a level the rule refuses. A fixed `TeleportKey` still wins (rigs, tests).
- internal/combat: `Fighter.SkillKey(name)` and `TeleportKey()` read the verified bind map; no skill, no key.
- cmd/d2bot: `wireTeleport` installs both for the Game.exe and d2d stacks; adapters and the boss and farm code ask `HasTeleport()`.
- internal/boss: `Opts.CanTeleportFn`, asked at each use (Teleport can be learned after the runner is built).
- Which areas are visited and when the bot hops is unchanged; a character without Teleport behaves exactly as before.

## Impact

Offline only. No spec names Teleport use (grep openspec/specs), so no delta. Live proof pending (recipe in docs/fixtelekey.md).

Overlap (declared): internal/nav, internal/boss, internal/combat and cmd/d2bot are standing defaults of main or files of other open cards; this card adds new files (internal/nav/telecap*.go, internal/combat/telekey*.go, cmd/d2bot/telewire*.go, internal/boss/boss_opts_test.go) and edits a few lines in internal/nav/{nav,teleport}.go, internal/nav/sim_test.go, internal/boss/run.go, cmd/d2bot/{adapters,boss,behaviors,d2dwire}.go without listing them in owns.
