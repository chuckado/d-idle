# Tasks

```yaml
change: lockstepp0
status: archived
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
result: "no game involved; tools/lockstep build and run recipe; game-side behaviour untested"
live: "no game; Wine console probes on the worker host"
merge: 18608e7
```

- [x] toolchain found, probes built
- [x] measurements run (clocks, hook k = 1/2/4, frame gate, CPU)
- [x] docs/research/lockstep-p0.md with verdict
