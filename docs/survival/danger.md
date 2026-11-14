# Danger profile (survival slice S-A)

> status: current (card survA). Implements design-2026-11-14.md section 3. Read-only, no behaviour change. Tags: V verified, I inferred, U unverified.

- Code: `internal/danger` (generator `Build`, `bosses.json`, `overrides.json`, handler). Wiring: `cmd/d2bot/dangerwire.go`, route `GET /api/danger[?area=id|name][&difficulty=normal|nightmare|hell]` (no area: list of zones with the bad-pack size).
- Source: the install tables through `data.DB` (Levels, MonStats, MonLvl, SuperUniques); nothing from the game is committed. Overrides are rows with evidence file, tag, expiry; at most five per class; a test fails a missing evidence file or an expired row.
- Per zone and difficulty: monsters (share I, life/defence/to-hit/damage V from tables x MonLvl rows U as in readiness, dps I against a reference defence), encounters typical/bad/elite (I, from group size and density), uniques per clear, immunities (V), pay (XP per kill V, level-gap bands U).
- Bosses (`bosses.json`, evidence boss1-trace.md): superuniques +3 level for damage and to-hit, life x4, experience x5; Extra Strong as `dmgPct` 235 (V run in boss1-trace).
- Units per clear are estimated from MonDen x level size; `Params.UnitsPerDensityGroup` (units = MonDen x mean pack size x k per Levels DrlgType) is fitted to areamon-table.csv: median error outdoor 20%, preset 18%, maze 55%; `TestAreamonDiff` asserts a 40% median over all 36 areas. Levels.txt sizes are not real areas and are not used. Andariel (life 1024, defence 60, level 12), Smith 404, Countess 96, Blood Raven 113 match boss1-trace exactly (V, builder run).
- Not here (later slices): power estimate, margin, ledger, controller. The `readiness` package stays until S-B folds it in.
