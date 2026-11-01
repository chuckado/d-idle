# PALSME: SME review of the Paladin tables

Change card `openspec/changes/palsme`. Offline, data and docs only.

## SME verdicts on PALFIX

- Zealot-lite order is sensible and legal: Zeal needs Sacrifice only (reqlvl 12), Smite and Prayer are level 1 without prerequisite, Vigor needs Cleansing (Prayer) plus Defiance. S (SME review).
- Legacy fix (Zeal 4, Prayer 1) is a pure improvement for future points. S.

## Table change (build-paladin-zealot-lite.json)

Prayer 3 becomes 1 (a resting aura only); the two freed points (clvl 6, 9) go to Might. Cleansing and Vigor (4 points at clvl 16 to 19: walk speed and stamina, low value before Blessed Hammer) are dropped; Defiance 1 stays (merc and Paladin; no prerequisite, reqlvl 6). Their points go to Might (clvl 16 slot, 17, 18) and Holy Fire (2 points, clvl 16 and 19; Might is its prerequisite). Rows 20 to 30, stats and hotkeys are unchanged. The table is selectable only with `--build paladin-zealot-lite`; no existing character changes.

Skill levels (with the Den point) before -> after, Prayer / Might / Holy Fire / Defiance / Zeal:

| clvl | before | after |
|---|---|---|
| 5 | 1 / 2 / 0 / 0 / 0 | 1 / 2 / 0 / 0 / 0 |
| 10 | 3 / 5 / 0 / 0 / 0 | 1 / 7 / 0 / 0 / 0 |
| 15 | 3 / 6 / 0 / 0 / 4 | 1 / 8 / 0 / 0 / 4 |
| 20 | 3 / 7 / 0 / 1 / 4 | 1 / 10 / 2 / 1 / 4 |
| 24 | 3 / 7 / 0 / 1 / 4 | 1 / 10 / 2 / 1 / 4 |

Vigor and Cleansing are now 0 (were 2 and 1). Hammer rows are unchanged. V: computed by replaying the table rows.

## Legality

- V: `go test -p 1 ./internal/plandata ./internal/leveler ./internal/resource` pass; `TestZealotLiteLegal` (no illegal pick with and without the Den point, Zeal first at 12 or 13, Sacrifice before Zeal, nothing stuck; assertions updated to Prayer 1, Might 9 or more, Holy Fire 2, Defiance 1, no Vigor).
- V: install check on the worker host: `labrun orion -- go test -count=1 -v -run 'TestPalProbeRulesMatchInstall|TestZealotLiteLegal|Pal' ./internal/leveler` from the worktree. The test reads the real install from `/tmp/d2mpq` (d2data.mpq present on the worker host); it ran (0.08 s, not skipped) and PASSES: the fixture rules (reqlvl and prerequisites of every Paladin skill used) equal the real skills.txt, so the Zealot-lite table and the fixed legacy table, which the simulator proves legal against that fixture, are legal against the real tables. The table rows themselves are not read against skills.txt directly (the check is fixture-to-install plus table-to-fixture).
- Note: the fixture lists Holy Bolt and Holy Fire at reqlvl 6, so the plan's Holy Bolt at clvl 20 is a choice, not a requirement. V (same test).

## Ready cards (record only; new policy features are paused until the primitives have a live PASS)

- PALRETREAT: Paladin retreat threshold 60 to 65 percent when 3 or more hostiles are adjacent at clvl 1 to 8, plus a door or corridor stand (the 40 percent line is too low: one swing cycle can lose 40 percent). S (SME).
- PALBELT: belt of 4 healing columns until about clvl 12; the PALFIX belt plan has 3 healing plus 1 mana, but Zeal and Smite cost 2 mana and the ladder ends in free Attack. S (SME).
- PALBRGATE: Blood Raven gate: clvl 10 or more, Might 3 or more, 4 or more healing potions, merc alive. S (SME).
- PALMERC1: hire the Kashya Rogue as soon as gold allows and keep her alive. S (SME).
- SMEDOCS: research gaps in docs/sme: per-weapon Smite damage and shield numbers; Fallen pack damage numbers for clvl 1 to 8. Queue research.

## Dead ends

- Cleansing/Vigor kept at 1/0 (Vigor 1 needs Cleansing 1 and Defiance 1, 3 points for a speed aura): rejected by the SME, not worth the points before Blessed Hammer.

## Shared-file lines for the merger

- docs/BACKLOG.md: add "PALRETREAT / PALBELT / PALBRGATE / PALMERC1 (from palsme, docs/palsme.md): ready, paused until the primitives have a live PASS; SMEDOCS: research per-weapon Smite and shield numbers, Fallen pack damage at clvl 1 to 8".
- docs/INDEX.md: `docs/palsme.md` (SME review of the Paladin tables, ready cards).
- CHANGELOG.md: "Zealot-lite table: Prayer 1, more Might, Defiance 1 and Holy Fire 2 instead of Cleansing and Vigor (opt-in table only)".
