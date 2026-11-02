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
result: "Wine IAT time hook PASS: k=1/2/4 ratios within 0.04 percent, gate 25.0/49.9/100.0 fps, CPU 1.8 percent of a core; Q5 not run (no libfaketime). Verified on the worker host with Wine console probes; no game."
live: "no game; Wine console probes on the worker host"
```

- [x] toolchain found, probes built
- [x] measurements run (clocks, hook k = 1/2/4, frame gate, CPU)
- [x] docs/research/lockstep-p0.md with verdict
