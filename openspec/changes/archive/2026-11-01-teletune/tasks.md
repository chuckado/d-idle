# Tasks

```yaml
change: teletune
status: archived
merge: 194e1ab
agent: teletune
branch: agent/teletune
worktree: ~/code/d2lod-wt/teletune
skip_specs: true
owns:
  - docs/teletune.md
  - internal/nav/teletune.go
  - openspec/changes/teletune/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/env ./cmd/d2bot ./tools/knobdump passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "Teleport key wired from the bound skill (production never had it); hops tuned to pay (defaults 36 subtiles/40 percent reserve): live: near walking speed with the defaults, 18 ticks per hop; pockets/escapes and path-aware hops are the next card (TELETUNE2)"
live: "verified live (DRILLS24): default knobs 0 hops, 265 and 605 ticks (walking 254 and 578, DRILLS22 448 and 851), mana untouched; teleMinLeg 0: 3 hops then declined at mana 80 (reserve 60 held, no idle wait), 202 and 755 ticks; per hop 18 ticks (not 33 or 64); no-Teleport control and refused-hop bound unchanged; finding: the leg rule keeps Teleport from leaving a map-walkable pocket"
```

- [x] break-even derivation (docs/teletune.md)
- [x] hopWorth, min leg, mana reserve, no idle wait, skill-data cost
- [x] knobs in NavTuning, docs/knobs.md regenerated
- [x] tests (DRILLS22 geometry, reserve, short leg, no wait, no Teleport unchanged)
- [x] live recipe
- [x] live proof
