# Tasks

```yaml
change: fixmouth
status: in-progress
agent: fixmouth
branch: agent/fixmouth
worktree: ~/code/d2lod-wt/fixmouth
skip_specs: true
owns:
  - internal/nav/walk.go
  - internal/nav/nav.go
  - internal/nav/fixmouth_test.go
  - internal/combat/combatspin_test.go
  - docs/fixmouth.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, no bounce (docs/fixmouth.md)"
merge: ""
result: "offline only: unit tests pass; live proof pending"
```

- [x] decide the option (first-walk arrival radius) from the logged geometry
- [x] arrival state in nav walk, mouthStep radius, NearMouth after arrival
- [x] table test, roam test, Fighter.move test
- [ ] live proof (DW.m3 x3, DW.m4)
