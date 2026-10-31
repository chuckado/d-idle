# Tasks

```yaml
change: primreview
status: in-progress
agent: primreview
branch: agent/primreview
worktree: ~/code/d2lod-wt/primreview
owns:
  - docs/primitives-review.md
  - openspec/changes/primreview/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result:
```

## 1. Work

- [ ] 1.1 trace pack-to-belt, pickit potion rule, kit gate, portal use and the loot seam from code
- [ ] 1.2 count every failure family in the navsoak, navsoak2, navsoak3, m2early and goldab logs and tie it to its emitter
- [ ] 1.3 write docs/primitives-review.md (verdict, findings, cleared suspects, log lines to add)
