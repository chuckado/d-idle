## Why

User go 2026-11-01 ("start the paladin probe"): the second-class probe. It learns what works and what does not for a non-Sorceress class and which Sorceress-fitted pieces break. Part A removes the audit's blocking seams so the probe measures the Paladin and not an unrelated Sorceress fit; part B tests the Paladin Normal path offline.

## What Changes

- Part A (Sorceress behaviour unchanged): the combat build is a profile id string (`combat.Build` is now a string, `LookupProfile` returns an error for an unknown class or id); the leveler table names its combat profile through its `combat` key; `PrefsFor` reads an explicit per-class table with `LookupPrefs` erroring on an unknown class; `panel.classRow` returns an error for an unknown class.
- Part B: Paladin probe tests (leveler plan simulator for both tables, combat ladder, gear, merc, belt, melee resource model) and the report docs/classes/paladin-probe.md.

## Impact

Code: the seam files named in `owns`. Existing tests unchanged except one format verb (`%d` to `%v` for the string build) and one assertion that pinned the silent Sorceress fallback for a Necromancer's gear preferences. No spec, ADR or glossary change.

## Overlap (declared)

Card buildv3 (open) owns `cmd/d2bot/class.go`, `cmd/d2bot/class_test.go` and `internal/combat/pure_test.go`. This change edits them minimally and does not claim them: `class.go` lets the leveler table name its combat profile (the `combat` key buildv3 introduced) and returns the lookup error in the plan; `class_test.go` gains one test; `pure_test.go` changes one format verb (`%d` to `%v`). Merge order: either, the hunks are disjoint from buildv3's table rows.
