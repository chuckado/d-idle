# Tasks

```yaml
change: fixrecast
status: archived
merge: 1a4b228
agent: fixrecast
branch: agent/fixrecast
worktree: ~/code/d2lod-wt/fixrecast
skip_specs: true
owns:
  - docs/fixrecast.md
  - docs/knobs.md
  - internal/combat/necro.go
  - internal/combat/buffrecast_test.go
  - internal/data/skills.go
  - internal/data/data_test.go
  - internal/env/tuning_combat.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/data ./internal/env ./tools/knobdump passes including TestFrozenArmorTimerLeavesGap, TestFrozenArmorRecastBeforeExpiryByClock, TestFrozenArmorRecastBeforeExpiryByState, TestAuraFrames"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "clock path proven live (armor drops 120.32 s at level 1 and 132.32 s at level 2 after the cast; recast at 114.0 s and 126.0 s; armor absent in 0 of 819 and 0 of 775 samples); the state path is untested because d2d never fills Player.States (card fixbuff)"
live: "verified live by drills13 (2026-11-01, d2d lab slot): Frozen Armor state on after a cast (8 ticks), real duration 120.0 s at level 1 and 132.0 s at level 2, the bot recast at 114.0 s and 126.0 s with a fresh tables cache, armor never absent in a simulated 600 s; BUT on a stale tables cache (made before this card) the recast stays at 150 s, see DRILLS13 in docs/probes/drills/README.md"
```

- [x] buff recast logic found (Fighter.buffDue, Buff)
- [x] real durations verified from Skills.txt rows
- [x] state read (Frozen Armor 10) and duration-based limit with margin
- [x] tests: gap before, no gap after, levels 1 and 2, margin, fallback
- [x] docs, primitives note, backlog, changelog
- [x] live check (hidden slot)
