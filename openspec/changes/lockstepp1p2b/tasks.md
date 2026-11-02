# Tasks

```yaml
change: lockstepp1p2b
status: in-progress
agent: lockstepp1p2b
branch: agent/lockstepp1p2b
worktree: d2lod-wt/lockstepp1p2b
skip_specs: true
owns:
  - docs/research/lockstep-p1p2.md
  - cmd/lockstepscan/**
  - openspec/changes/lockstepp1p2b/**
exclusive:
  - instance:i4
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "P1, P2 k=1 control and k=2 verified live on i4 (frames 25.00 and 50.00 per real second); REHASH not captured; bands comparison not done"
live: "i4 Booti: read-only scans; hook run marker lockstep-p2-k1 written"
```

- [x] P1 live read, 25 fps model
- [x] P2 k=1 control
- [x] P2 k=2
