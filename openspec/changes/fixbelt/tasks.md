# Tasks

```yaml
change: fixbelt
status: in-progress
agent: fixbelt
branch: agent/fixbelt
worktree: ~/code/d2lod-wt/fixbelt
skip_specs: true
owns:
  - docs/fixbelt.md
  - internal/resource/recovery.go
  - internal/resource/recovery_test.go
  - internal/bot/broke.go
  - internal/bot/area_ban_test.go
  - openspec/changes/fixbelt/**
exclusive: []
gates:
  - "go test -p 1 internal/resource internal/bot; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Unit tests only (mem.Fake bot rig); NOT verified live: needs a Gate 1 soak, recipe in docs/fixbelt.md."
live: "not run; recipe in docs/fixbelt.md"
```

- [x] locate the window in the minute log
- [x] fix: low-kit override of the recovery cooldown
- [x] unit tests
- [ ] live proof (Gate 1 soak)
