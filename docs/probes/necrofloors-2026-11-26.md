# Necromancer life floors (card necrofloors, 2026-11-26)

> status: current. Source of every number: the d2-sme comparison in docs/probes/necro-readiness-2026-11-26.md (S, advice only, U). Governance: docs/process/threshold-governance.md (the new numbers are `evidence: 1`, a seed, not a measurement).

## What changed (data: internal/plandata/admission.json, copy docs/plan/data/admission.json)

| row | area | number | effect |
|---|---|---|---|
| `act2_sewers` (existing, one field added) | 47, 48, 49 | `minMaxLife.caster` 130 | a caster under 130 max life is refused (`too_weak`) in the sewers; melee has no life rule there |
| `act2_duriel_lair` (new) | 73 | `minMaxLife` caster 220, melee 220; level and kit copy `act2_tombs` | Duriel's lair needs 220 max life for a caster (the tombs row said 150 for casters); melee stays at its tombs number |

The rows are keyed by ARCHETYPE (`melee` or `caster`), not by class: a Sorceress is a caster and gets the same life floors (build flexibility: no class literal). Test `TestCasterLifeFloorsSewersAndDuriel` (internal/admit) checks the edges for both archetypes (129 and 130, 219 and 220, melee unchanged, the tomb floor of 150 still open for a caster).

## Not done: Andariel

The SME delta (life 180 to 200 plus 6 healing potions) is NOT a row. The Andariel level floor is the owner's calibration (clvl 11 with the merc, 10 without, user 2026-11-24, internal/plandata/boss_levels.go) and stays. A life row at area 37 would override it in practice: a clvl 11 Necromancer reaches about 110 to 130 max life (S, estimate from the class life table, U), so a 180 floor would hold the hero back for several levels. Proposed row, ready for the owner's word (not committed):

`{"id": "act1_andariel_lair", "areas": [37], "minLevel": {"melee": 0, "caster": 0}, "kit": {"healing": 6, "portal": 1}, "minMaxLife": {"caster": 180}}` (needs `TestRowsSane` to accept a zero level floor, or copy the boss floors 11 and 10 into it).

## Not verified live

No Necromancer walked these areas under the rows. Checklist: a lab Necromancer (CHEATED stats allowed for mechanics only) at max life 129 and 130 asks `ops status --why` for the sewers road; expected `too_weak` with "life 130 needed for the act2_sewers road".
