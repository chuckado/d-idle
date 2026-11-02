# Tasks

```yaml
change: memrules
status: archived
agent: memrules
branch: agent/memrules
worktree: ~/code/d2lod-wt/memrules
owns:
  - AGENTS.md
  - docs/cheats.md
  - docs/adr/0014-memory-writes-allowed-no-crutch.md
  - internal/cheat/run.go
  - internal/cheat/run_test.go
  - docs/decisions.md
  - docs/d2-1.14d-memory.md
  - docs/research/d2d-whisper-lan-oracle.md
  - internal/cheat/ledger.go
  - internal/cheat/cheat.go
  - cmd/d2cheat/main.go
  - openspec/changes/memrules/**
exclusive: []
gates:
  - "go test -p 1 ./internal/cheat passes"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "user decision 2026-11-02; offline tests only; the private overlay and gate judge are updated separately"
merge: 8610229
```

- [x] 1.1 rule text in AGENTS.md, docs/cheats.md, ADR, decisions line
- [x] 1.2 delta spec lab-rules
- [x] 1.3 ledger run id, runstart, since-run (go test -p 1 ./internal/cheat)
