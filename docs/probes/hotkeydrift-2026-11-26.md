# Build hotkeys versus profile layout (card hotkeydrift, 2026-11-26)

> status: current. Tags: V verified (how said), S read in code or data, U unverified.

## Finding

Each plan table (`internal/plandata/build-*.json`) may carry a `hotkeys` list (key, skill, hand). Nothing reads it outside the parser (S: grep `Hotkeys`; `leveler.Build.Hotkeys` says "the leveler never presses it"). The keys the fighter uses are the profile `layout` of its combat profile (`internal/combat/classdata/profiles.json`), and on d2d the pseudo keys of `Rig.BindSkills` (card necrokeys). The two lists were written by hand at different times.

V (test `TestBuildHotkeysDoNotDriftFromTheProfileLayout`, `cmd/d2bot/hotkeydrift_test.go`, run on the worker host): rows of a table whose skill has another key, or none, in the table's combat profile (class and combat line resolved the way `planForClass` does):

| tables | drift rows |
|---|---|
| Amazon: bow 1, javelin 2, multishot 3 | 6 |
| Assassin: fireblast 2, sentry 6 | 8 |
| Barbarian: warcry 4 (melee: 0) | 4 |
| Druid: tornado 8 (fire tables: 0) | 8 |
| Necromancer: leveling 5, summoner endgame 7 (summoner normal 30: 0 since card necro1) | 12 |
| Paladin: normal 5, normal v2 4, zealot lite 3, hammerdin 6, hammerdin zeal 7 | 25 |
| Sorceress: blizzard endgame 6 | 6 |

## Decision

The lists stay (they document intent and the game.exe path). The profile layout is the single source of truth; the test is a RATCHET: a table may only lower its number (the test also fails when a number fell, so the baseline map is lowered), and a new table starts at zero. Deleting the keys was not done: 15 tables, no reader, and the player-facing docs quote them (U).

Fixing a table means editing its `hotkeys` rows to the profile keys, or giving the table a `combat` key that names the profile it was written for (the summoner endgame table has no `combat` key, so it resolves the class default profile; setting `"combat": "summoner"` changes which profile plays at endgame and needs a drill, so it is not done here).
