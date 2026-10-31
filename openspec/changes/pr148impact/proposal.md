## Why

Upstream PR 148 splits big files by moves only; git cannot follow a split, so the private fork's lab patch set needs a manual port. Measure how big that job is before deciding when to port.

no observable behaviour change: research doc only.

## What Changes

- `docs/research/pr148-impact.md`: what #148 changes, per-area impact (fork hunks, scripted dry run, our citations, PR 147, tome branch), effort, recommended plan
