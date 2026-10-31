# Tasks

```yaml
change: fixmisc
status: in-progress
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
merge:
result: ""
live: ""
```

- [x] 1.1 F3 decided from code: intended (potion.go doc comment, existing policy tests), pinned by a test
- [x] 1.2 F4 root cause: PathTo is pure planning, the tracker was never updated after the jump; Navigator.Sync plus Syncer in Clear/Kill
- [x] 1.3 tests for same-level jump, level-change jump (portal/waypoint arrival), loading then ready (restart)
- [ ] 1.4 live proof: drill rerun after merge (not run here)
