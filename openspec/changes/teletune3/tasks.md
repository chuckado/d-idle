# Tasks

```yaml
change: teletune3
status: in-progress
agent: teletune3
branch: agent/teletune3
worktree: ~/code/d2lod-wt/teletune3
skip_specs: true
owns:
  - docs/teletune3.md
  - internal/nav/teletune3_test.go
  - openspec/changes/teletune3/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/drill ./cmd/d2drill passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline: border-leg analysis (the key is not on the walking path; walks split at a replan, cause U), ResetStallMemory, plan reason log field, stall rule tests over noteStuck, drills D22-S and D22-BL written; not verified live, awaits DRILLS26 (docs/teletune3.md)"
```

- [x] border leg: log evidence and a test that a key changes nothing on the walk
- [x] stall rule test over the real stall memory API
- [x] Navigator.ResetStallMemory and its use between drills
- [x] drills D22-S and D22-BL (not run)
- [ ] live proof (DRILLS26)
