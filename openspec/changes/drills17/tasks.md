# Tasks

```yaml
change: drills17
status: in-progress
agent: drills17
branch: agent/drills17
worktree: ~/code/d2lod-wt/drills17
skip_specs: true
owns:
  - openspec/changes/drills17/**
exclusive: []
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
merge: ""
result: "Verified live on hidden d2d slot 0: fixmouth2 live proof FAIL (DW.m3 x3 and DW.m4 still bounce; callers named), D2 11/11 and D3 12/12 PASS; see the DRILLS17 section of docs/probes/drills/README.md"
live: "verified live: DW.m3 x3 FAIL (identical), DW.m4 FAIL, D2 PASS, D3 PASS"
```

- [x] DW.m3 x3, DW.m4 on current main
- [x] D3 town pass, D2 regression
- [x] README section, card live lines
