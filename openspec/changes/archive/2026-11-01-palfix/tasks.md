# Tasks

```yaml
change: palfix
status: archived
merge: db9b20e
agent: palfix
branch: agent/palfix
worktree: ~/code/d2lod-wt/palfix
skip_specs: true
owns:
  - docs/plan/data/build-paladin-normal30.json
  - docs/plan/data/build-paladin-zealot-lite.json
  - docs/plan/data/builds-endgame/barbarian_warcry_endgame90.json
  - docs/plan/data/builds-endgame/druid_tornado_endgame90.json
  - docs/plan/data/builds-endgame/necromancer_summoner_endgame90.json
  - docs/plan/data/builds-endgame/paladin_hammerdin_endgame90.json
  - docs/plan/data/builds-endgame/sorceress_blizzard_endgame90.json
  - docs/plan/data/builds-endgame/gen_endgame.py
  - docs/plan/data/potions-belt.json
  - internal/combat/simladder_test.go
  - internal/leveler/palprobe_ladder_test.go
  - internal/leveler/palzeal_test.go
  - internal/plandata/build-barbarian_warcry_endgame90.json
  - internal/plandata/build-druid_tornado_endgame90.json
  - internal/plandata/build-necromancer_summoner_endgame90.json
  - internal/plandata/build-paladin-normal30.json
  - internal/plandata/build-paladin-zealot-lite.json
  - internal/plandata/build-paladin_hammerdin_endgame90.json
  - internal/plandata/build-sorceress_blizzard_endgame90.json
  - internal/resource/manaheavy.go
  - internal/resource/manaheavy_test.go
  - internal/resource/sim_test.go
  - docs/palfix.md
  - openspec/changes/palfix/**
exclusive: []
gates:
  - "gofmt, go vet, go test -p 1 per touched package, whole repo on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Paladin findings fixed offline (mana-heavy flag, Blessed Hammer band 18, simulator class slots, legacy table Zeal and Prayer rows, plandata table seam, Zealot-lite table selectable by --build): unit-tested; not verified live"
live: "offline only"
```

- [x] PALMANA
- [x] PALBAND
- [x] SIMCLASS
- [x] PALTABLE
- [x] PALZEAL
- [x] docs/palfix.md
