# Tasks

```yaml
change: drills27
status: archived
merge: 022546b
agent: drills27
branch: agent/drills27
worktree: ~/code/d2lod-wt/drills27
skip_specs: true
owns:
  - openspec/changes/drills27/**
exclusive:
  - instance:drills27-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "stall memory reset, stall escape proven live, border-leg gap is replan spread not the key: verified live (DRILLS27)"
live: "slot 1, groups T, D2, D3 (DRILLS27)"
```

- [x] 1.1 D22-BL, D22-S, D22-PC then D22-P
- [x] 1.2 regression D2, D3, group T legs
- [x] 1.3 logs, README section, card lines
