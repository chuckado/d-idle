# Tasks

```yaml
change: lockstepp0
status: in-progress
agent: lockstepp0
branch: agent/lockstepp0
worktree: ~/code/d2lod-wt/lockstepp0
skip_specs: true
owns:
  - docs/research/lockstep-p0.md
  - tools/lockstep/**
  - openspec/changes/lockstepp0/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "pending"
live: "no game; Wine console probes on the worker host"
```

- [ ] toolchain found, probes built
- [ ] measurements run (clocks, hook k = 1/2/4, frame gate, CPU)
- [ ] docs/research/lockstep-p0.md with verdict
