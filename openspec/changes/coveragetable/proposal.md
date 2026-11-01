## Why

The user asked (2026-11-01) for a concrete, reproducible measure of how complete d2d is against game.exe 1.14d. Method 1: a generated coverage table with fixed denominators taken from the game's own data tables and a hand-maintained, machine-checked ledger of what was compared and how.

## What Changes

- tools/coverage/coverage.py (stdlib only, `--check` for CI) renders docs/coverage/rule-coverage.md from docs/coverage/denominators.tsv (fixed row universes) and docs/coverage/ledger.tsv (status per row range, evidence doc, case count, seed or date).
- docs/coverage/README.md: how to add a row after a differential, the tag rules, the honest limits.
- tools/coverage/tests: stdlib unit tests with fixtures.
- docs/decomp-queue.md: one pointer line only.

## Impact

Docs and tooling only. No spec, policy, ADR or glossary change. No game, no fork.

Overlap (declared): docs/coverage.md (test coverage of Go packages, a different subject) is untouched; docs/decomp-queue.md is edited by the coordinator, this card adds one pointer line only.
