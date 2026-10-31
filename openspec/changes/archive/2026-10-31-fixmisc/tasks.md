# Tasks

```yaml
change: fixmisc
status: archived
merge: be5e7d0
agent: fixmisc
branch: agent/fixmisc
worktree: ~/code/d2lod-wt/fixmisc
skip_specs: true
owns:
  - internal/nav/sync.go
  - internal/nav/fixmisc_test.go
  - internal/combat/types.go
  - internal/combat/fight.go
  - internal/combat/fixmisc_test.go
  - internal/resource/policy_test.go
  - docs/fixmisc.md
exclusive: []
gates:
  - "go test per package (nav, combat, resource) and whole-repo vet/test -short on the worker host pass; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof: debug goto then Fighter.Clear without a warm-up walk (docs/fixmisc.md)"
result: "verified live (DRILLS6 D-sync x3): first Clear after debug goto killed 3 of 3 zombies, 0 unreachable lines; town portal and waypoint arrival variants not run"
live: "DRILLS6 2026-10-31: D-sync PASS x3 (first Fighter.Clear after debug goto, no warm-up walk: 3 of 3 zombies killed, 0 unreachable lines); town portal and waypoint arrivals not run"
```

- [x] 1.1 F3 decided from code: intended (potion.go doc comment, existing policy tests), pinned by a test
- [x] 1.2 F4 root cause: PathTo is pure planning, the tracker was never updated after the jump; Navigator.Sync plus Syncer in Clear/Kill
- [x] 1.3 tests for same-level jump, level-change jump (portal/waypoint arrival), loading then ready (restart)
- [x] 1.4 live proof: D-sync PASS x3 (DRILLS6)
