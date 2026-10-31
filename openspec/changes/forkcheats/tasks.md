# Tasks

```yaml
change: forkcheats
status: in-progress
agent: forkcheats
branch: agent/forkcheats
worktree: ~/code/d2lod-wt/forkcheats
owns:
  - docs/cheats.md
  - docs/forkcheats.md
  - openspec/changes/forkcheats/**
exclusive:
  - fork-branch:lab-cheats
gates:
  - "each new verb exercised live on one hidden slot with before/after state (docs/forkcheats.md)"
  - "cmd/d2drill D0 + D1 + one D2 drill against the new binary: no regression"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: "lab binary d2d.lab-cheats (private fork branch lab-cheats): debug act, hit, merc hire|kill|revive, give pack|belt, durability, belt clear, unequip, all verified live on a hidden slot (docs/forkcheats.md); harness D0, D1, D2.a unchanged. README missing-cheats rows and primitives.md notes edited under the coordinator brief (the m1-liveproof card owns those paths). Drills not rerun."
```

## 1. Work

- [x] 1.1 fork: verbs built on the private fork branch, pushed to the private remote the same day
- [x] 1.2 live proof of each verb and the harness regression run
- [x] 1.3 docs/cheats.md, README missing-cheats table, primitives.md BLOCKED notes
