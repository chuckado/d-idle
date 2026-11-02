# Tasks

```yaml
change: lockstepp4
status: in-progress
agent: lockstepp4
branch: agent/lockstepp4
worktree: d2lod-wt/lockstepp4
skip_specs: true
owns:
  - docs/research/lockstep-p4.md
  - openspec/changes/lockstepp4/**
exclusive:
  - instance:i4
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "pending"
live: "pending"
```

- [ ] k=2 run with scaled bot clock
- [ ] k=4 run with scaled bot clock
- [ ] doc docs/research/lockstep-p4.md, merge lines
