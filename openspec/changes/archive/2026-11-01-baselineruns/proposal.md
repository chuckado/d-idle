## Why

The baseline audit (docs/research/baseline-audit.md) found the original-game baselines too thin to gate on; the user decided to add 5 late runs (Sorceress clvl 11-14, profile m1) and 3 fresh-start runs. This card prepares them: the runbook and a launcher in the private ops repo. Nothing is started.

## What Changes

- docs/process/baseline-runs.md: how the existing original-game runs were produced, the recipe per run type, the kit restore, seeds (none), output, rows, quality flags, schedule and capacity, the per-run checklist, what must be supplied and what is unverifiable until a live preflight.
- Ops repo (not this repo): tools/baseline-run.sh, baseline-sup.sh, baseline-collect.sh, baseline-lib.sh and tools/tests/test_baseline_run.sh.

## Impact

Docs only in this repo. No spec, ADR or glossary change.
