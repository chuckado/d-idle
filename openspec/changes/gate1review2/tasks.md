# Tasks

```yaml
change: gate1review2
status: in-progress
agent: gate1review2
branch: agent/gate1review2
worktree: ~/code/d2lod-wt/gate1review2
skip_specs: true
owns:
  - docs/soaks/gate1r2-20261101.md
  - docs/probes/soak-logs/gate1r2-20261101-*
  - openspec/changes/gate1review2/**
exclusive: []
gates:
  - changecheck 0 warnings, leakcheck 0 hits
result: "docs and cards only (the three new cards are proposals with their own owns); counts from the run logs on the worker host (read-only), root cause pattern V, fork side U; not verified live"
```

- [x] 1 Read the record, the run logs and the audit
- [x] 2 Numbers per band, stall windows, guard firings, restarts
- [x] 3 HasTP root cause from the inv-tp-dump lines
- [x] 4 Review 2 section, verdict per criterion, cards, rerun recipe
