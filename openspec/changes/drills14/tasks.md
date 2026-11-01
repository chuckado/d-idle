# Tasks

```yaml
change: drills14
status: in-progress
agent: drills14
branch: agent/drills14
worktree: ~/code/d2lod-wt/drills14
skip_specs: true
owns:
  - openspec/changes/drills14/**
  - internal/nav/walk.go
  - internal/nav/roam_test.go
exclusive:
  - instance:slot0
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill ./internal/nav pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
result: "DW.m3 and DW.m4 run live on a hidden slot (lab build cheats3); the bounce callers are named in the DRILLS14 section of docs/probes/drills/README.md"
live: "verified live: DW.m3 FAIL x4 (bounce in Navigator.Explore, 4 of 4 entries, identical), DW.m4 FAIL (bounces in Fighter.Clear approach and in Explore)"
```

- [x] DW.m3, DW.m4 harness and mouthStep debug lines
- [x] live runs at debug level, callers named
- [x] proposed fix and its offline test (not implemented)
