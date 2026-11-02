# Tasks

```yaml
change: botsupply
status: archived
agent: botsupply
branch: agent/botsupply
worktree: ~/code/d2lod-wt/botsupply
skip_specs: true
owns:
  - docs/botsupply.md
  - internal/data/supply.go
  - internal/data/supply_test.go
  - internal/resource/supply_test.go
  - openspec/changes/botsupply/**
exclusive: []
gates:
  - "gofmt, go vet, go test -p 1 of data resource loot bot env combat tools/knobdump; changecheck 0 warnings; leakcheck 0"
result: "offline per-package tests; play-affecting defaults on where the traces prove the old values wrong; not verified live"
live: "none"
merge: dd62521
```

- [x] level-aware gold model and supply table (data/supply.go)
- [x] area-aware gold pile bar (goldPileShare) with tests
- [x] area XP table and fallback levels
- [x] ScaledHP fixed (kept: tests reference it), MeanHP
- [x] sim gold and monster life from the model and rows
- [x] docs/botsupply.md, Picks comment, knobs doc
