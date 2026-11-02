# Tasks

```yaml
change: snaplib
status: archived
agent: snaplib
branch: agent/snaplib
worktree: ~/code/d2lod-wt/snaplib
owns:
  - internal/saves/snaplib.go
  - internal/saves/snaplib_test.go
  - internal/cheat/d2s/tail.go
  - cmd/d2restore/**
  - docs/snaplib.md
  - openspec/changes/snaplib/**
skip_specs: true
exclusive: []
gates:
  - "go test -p 1 ./internal/saves ./internal/cheat/... ./cmd/d2restore; gofmt -l empty; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
merge: cf575ae
result: "offline on synthetic fixtures; the game loading a renamed character and the map seed question are untested; private tools still key by name"
```

## 1. Work

- [x] 1.1 manifest, snapshot, verify, list (table tests in internal/saves/snaplib_test.go)
- [x] 1.2 rename with checksum, place with refusals and rollback (same tests)
- [x] 1.3 cmd/d2restore with preflight and ledger line (cmd/d2restore/main_test.go, fake preflight, real flock test)
- [x] 1.4 docs/snaplib.md with the live checklist
- [ ] 1.5 live: first restore into a free hidden instance (coordinator, checklist in docs/snaplib.md)
