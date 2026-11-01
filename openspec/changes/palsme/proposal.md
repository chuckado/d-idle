## Why

The Diablo II SME reviewed PALFIX (Zealot-lite order and the legacy fix) and recommended a leaner Zealot-lite table, plus policy items and doc gaps. The install legality check also had to run on the worker host.

## What Changes

- Zealot-lite data: Prayer 3 becomes 1, the Cleansing and Vigor points go to Might and Holy Fire, Defiance 1 stays.
- docs/palsme.md: before/after levels, the install check result, ready cards for the policy items and doc gaps (record only, nothing implemented).

## Impact

Data and docs only. The table is selectable only with `--build paladin-zealot-lite`, never a default: no existing character is affected. No spec, ADR or glossary change.

## Overlap (declared)

Card palfix (in progress, branch agent/palfix, being merged) owns `internal/plandata/build-paladin-zealot-lite.json`, its copy `docs/plan/data/build-paladin-zealot-lite.json` and `internal/leveler/palzeal_test.go`. This change edits those three files (rows and note of the table; the skill assertions of `TestZealotLiteLegal`) and does not claim them; this branch is based on agent/palfix, so there is no conflict and merge order after palfix is fixed.
