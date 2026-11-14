# Danger profile (survival slice S-A)

> status: current (card survA). Implements design-2026-11-14.md section 3. Read-only, no behaviour change. Tags: V verified, I inferred, U unverified.

- Code: `internal/danger` (generator `Build`, `bosses.json`, `overrides.json`, handler). Wiring: `cmd/d2bot/dangerwire.go`, route `GET /api/danger[?area=id|name][&difficulty=normal|nightmare|hell]` (no area: list of zones with the bad-pack size).
- Source: the install tables through `data.DB` (Levels, MonStats, MonLvl, SuperUniques); nothing from the game is committed. Overrides are rows with evidence file, tag, expiry; at most five per class; a test fails a missing evidence file or an expired row.
- Per zone and difficulty: monsters (share I, life/defence/to-hit/damage V from tables x MonLvl rows U as in readiness, dps I against a reference defence), encounters typical/bad/elite (I, from group size and density), uniques per clear, immunities (V), pay (XP per kill V, level-gap bands U).
- Bosses (`bosses.json`, evidence boss1-trace.md): superuniques +3 level for damage and to-hit, life x4, experience x5; Extra Strong as `dmgPct` 235 (V run in boss1-trace).
- Units per clear are estimated from MonDen x level size; `Params.UnitsPerDensityTile` is fitted by `TestAreamonDiff` against areamon-table.csv (see the card result for the tolerance reached).
- Not here (later slices): power estimate, margin, ledger, controller. The `readiness` package stays until S-B folds it in.
