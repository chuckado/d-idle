# Tasks

```yaml
change: armorcut
status: archived
agent: armorcut
branch: agent/armorcut
worktree: ~/code/d2lod-wt/armorcut
skip_specs: true
owns:
  - openspec/changes/armorcut/**
  - docs/armorcut.md
  - internal/combat/fighter.go
  - internal/combat/armorcut_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat passes; go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: f85fcc3
result: "offline: cut casts retried, no cast mid-hit, tests pass; live: cut-cast retry proven live (DRILLS29: cut at 114.0 s, retry 1.6 s later, state present; armor absent 0 ms over 544 samples); a cut by real monsters at normal damage is unproven (needed a logged debug damage boost)"
live: "DRILLS29 (branch agent/drills29): the cut-retry path VERIFIED LIVE on lab-armorhit and lab-nexthit2 (drill D29-C, identical): a Frozen Armor recast due at 114.0 s with a zombie pack holding the character (logged dmg 30 per hit, above life/12) was cut by a hit (mode 4), the bot logged `buff cut by a hit, will retry` and recast 1.64 s later (115.7 s), armor never absent; hit recovery only appears with the boosted blows (plain zombie hits did not cut, docs/fixcrowd.md). D26-A: PASS with the spacing rule met (casts 0, 115.6, 229.6 s). Earlier: DRILLS28 step 0 verified live, no cut in D26-A"
```

- [x] evidence from D26-A and the hypothesis it supports (docs/armorcut.md)
- [x] fix in buffPass with tests
- [x] live recipe
- [x] step 0 and D26-A rerun on both fork binaries (DRILLS28); the retry itself still unproven live
