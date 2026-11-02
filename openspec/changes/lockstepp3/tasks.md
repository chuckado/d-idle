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
  - cmd/lockstepscan/**
  - openspec/changes/lockstepp3/**
exclusive:
  - instance:i4
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: ""
live: "i4 Booti"
```

- [ ] hook periodic STAT line (private repo, local commit)
- [ ] unhooked control, 10 min town and 10 min Blood Moor
- [ ] k=1 hooked
- [ ] k=2 hooked
- [ ] k=4 hooked, plus 5 min town
- [ ] doc
