# Tasks

```yaml
change: ledgerkey
status: in-progress
agent: snaplib
branch: agent/ledgerkey
worktree: ~/code/d2lod-wt/ledgerkey
owns:
  - internal/cheat/ledger.go
  - internal/cheat/ledgerkey.go
  - internal/cheat/ledgerkey_test.go
  - cmd/d2cheat/main.go
  - openspec/changes/ledgerkey/**
skip_specs: true
exclusive: []
gates:
  - "go test -p 1 ./internal/cheat/...; go vet ./cmd/d2cheat; changecheck -cli 0 warnings; leakcheck -mode fail 0"
merge:
result: "Offline unit tests only; no live run needed for the key itself."
```

## 1. Work

- [x] 1.1 key, filter, CheatedKey, summary (internal/cheat/ledgerkey_test.go, legacy and renamed cases)
- [x] 1.2 d2cheat ledger flags
- [ ] 1.3 private tools (soak24.sh preflight, ledgercheck.sh) key by instance/name (coordinator)
