# Tasks

```yaml
change: fixtelekey
status: archived
merge: 194e1ab
agent: fixtelekey
branch: agent/fixtelekey
worktree: ~/code/d2lod-wt/fixtelekey
skip_specs: true
owns:
  - docs/fixtelekey.md
  - internal/nav/telecap.go
  - internal/nav/telecap_test.go
  - internal/combat/telekey.go
  - internal/combat/telekey_test.go
  - cmd/d2bot/telewire.go
  - cmd/d2bot/telewire_test.go
  - internal/boss/boss_opts_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/combat ./internal/boss ./cmd/d2bot passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "Teleport key wired from the bound skill (production never had it); hops tuned to pay (defaults 36 subtiles/40 percent reserve): live: near walking speed with the defaults, 18 ticks per hop; pockets/escapes and path-aware hops are the next card (TELETUNE2)"
live: "verified live on d2d (DRILLS22): hops fire on Act 2 tile-exit legs at clvl 18 (12 hops, 0 refused, key sk54), no hop in town, refused hops bounded at 3, no-Teleport control walks; border exits never hop and the hops are slower than walking (finding); boss stand point and a second class not run; DRILLS24 re-ran the legs on the teletune build: key sk54 and the bind still hold, no hop in town, goto 74 reaches the Arcane Sanctuary with HasTeleport and TeleportOK true"
```

- [x] call graph of the finding (docs/fixtelekey.md)
- [x] Navigator capability hooks, Fighter.SkillKey, wireTeleport, boss CanTeleportFn
- [x] tests (key from bind, none without the skill, other key, refused level, bounded refused hops)
- [x] live proof recipe
- [x] live proof (DRILLS22)
