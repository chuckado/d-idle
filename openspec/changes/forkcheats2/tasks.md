# Tasks

```yaml
change: forkcheats2
status: in-progress
agent: forkcheats2
branch: agent/forkcheats2
worktree: ~/code/d2lod-wt/forkcheats2
owns:
  - docs/cheats.md
  - docs/forkcheats.md
  - openspec/changes/forkcheats2/**
exclusive:
  - fork-branch:lab-cheats2
gates:
  - "each new verb exercised live on one hidden slot with before/after state and a negative case (docs/forkcheats.md)"
  - "cmd/d2drill D0 + D1 + one D2 drill against the new binary: no regression"
  - "changecheck -cli has 0 warnings; leakcheck -mode fail finds 0; openspec validate --all --strict passes"
merge:
result: "lab binary d2d.lab-cheats2 (private fork branch lab-cheats2): debug spawn, merc hurt, quest list|set|bit|join, give quality + unid, pin on|off|status, help, all verified live on a hidden slot with negative cases (docs/forkcheats.md); harness D0, D1, D2.a unchanged. README missing-cheats rows and primitives.md notes edited under the coordinator brief (the m1-liveproof card owns those paths). Drills not rerun."
```

## 1. Work

- [x] 1.1 fork: verbs built on a new private fork branch off the cheats branch, pushed to the private remote the same day
- [x] 1.2 live proof of each verb and the harness regression run
- [x] 1.3 docs/cheats.md, docs/forkcheats.md, README missing-cheats table, primitives.md BLOCKED notes
