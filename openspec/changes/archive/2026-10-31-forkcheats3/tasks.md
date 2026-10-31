# Tasks

```yaml
change: forkcheats3
status: archived
agent: forkcheats3
branch: agent/forkcheats3
worktree: ~/code/d2lod-wt/forkcheats3
skip_specs: true
owns:
  - docs/cheats.md
  - docs/forkcheats.md
  - openspec/changes/forkcheats3/**
exclusive:
  - fork-branch:lab-cheats3
gates:
  - "debug missiles exercised live on one hidden slot: empty list, own bolt (owner type 0), hostile missile (owner type 1), bad args error (docs/forkcheats.md)"
  - "cmd/d2drill D0 + D1 + one D2 drill against the new binary: no regression"
  - "changecheck -cli has 0 warnings; leakcheck -mode fail finds 0; openspec validate --all --strict passes"
merge: f183b682
result: "lab binary d2d.lab-cheats3 (private fork branch lab-cheats3): debug missiles verified live on a hidden slot (empty, own Fire Bolt owner 0, Fallen Shaman fire owner 1, bad args error); harness D0, D1, D2.a unchanged."
```

## 1. Work

- [x] 1.1 fork: verb built on a new private fork branch off lab-cheats2, pushed to the private remote the same day
- [x] 1.2 live proof of each case and the harness regression run
- [x] 1.3 docs/cheats.md, docs/forkcheats.md
