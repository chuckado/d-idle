# Tasks

```yaml
change: lockstepp3
status: in-progress
agent: lockstepp3
branch: agent/lockstepp3
worktree: d2lod-wt/lockstepp3
skip_specs: true
owns:
  - docs/research/lockstep-p3.md
  - openspec/changes/lockstepp3/**
exclusive:
  - instance:i4
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "verified live on i4 (Booti): unhooked, k=1, k=2, k=4 runs; docs only; scaled bot clock needed before k>1 measured runs"
live: "i4 Booti"
```

- [x] hook periodic STAT line (private repo, local commit)
- [x] unhooked control, 10 min town and 10 min Blood Moor
- [x] k=1 hooked
- [x] k=2 hooked
- [x] k=4 hooked, plus 5 min town
- [x] doc
