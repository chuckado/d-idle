# Tasks

```yaml
change: actgate
status: in-progress
agent: actgate
branch: agent/actgate
worktree: ~/code/d2lod-wt/actgate
skip_specs: true
owns:
  - docs/research/sorc-act-gate.md
  - openspec/changes/actgate/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only: docs/research/sorc-act-gate.md (19 rows: GREEN 2, YELLOW 12, RED 3, UNKNOWN 2), gate rule, ordered cards, proposed spec delta. Nothing run live."
live: "docs only, nothing run live"
```

- [x] per-act verdict tables
- [x] gate rule per act and soak report fields
- [x] card list ordered, proposed spec delta, shared-file lines
