# Tasks

```yaml
change: teletune3
status: archived
merge: 022546b
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
result: "stall memory reset, stall escape proven live, border-leg gap is replan spread not the key: verified live (DRILLS27)"
live: "DRILLS27: D22-BL verified live (13 plan lines with why=; replans are why=monsters, in the first no-key leg only; spread no key 630-692, key 677-704; the key is not the cause); D22-S verified live (stall, then 2 hops after a stall: stalledHere); D22-PC then D22-P in one process verified live (3 hops, 229 ticks); regression verified live (D2 11 of 11, D3 12 of 12, Stony Tomb 211, Halls 533, border 717)"
```

- [x] border leg: log evidence and a test that a key changes nothing on the walk
- [x] stall rule test over the real stall memory API
- [x] Navigator.ResetStallMemory and its use between drills
- [x] drills D22-S and D22-BL (not run)
- [x] live proof (DRILLS27, agent/drills27)
