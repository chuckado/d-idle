# Tasks

```yaml
change: drills16
status: in-progress
agent: drills16
branch: agent/drills16
worktree: ~/code/d2lod-wt/drills16
skip_specs: true
owns:
  - openspec/changes/drills16/**
exclusive: []
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
result: "Verified live on hidden d2d slots 0 and 1: fixmouth FAIL (bounces remain, callers named), fixbuff PASS on all four live items; see the DRILLS16 section of docs/probes/drills/README.md"
live: "verified live: DW.m3 x3 (identical) and DW.m4 FAIL; R-bs1, R-bs3, R-bs4 PASS"
```

- [x] DW.m3 x3, DW.m4 rerun on current main
- [x] group R drills (debug states line, Player.States, long fight, armor ids)
- [x] README section, card live lines

Ownership: no docs/OWNERS line added. internal/drill/d16.go, cmd/d2drill/d16_lab.go and the two-line hook in cmd/d2drill/main.go are covered by the standing drills globs of m1-liveproof, as for DRILLS13; it also sets the `live:` lines of cards fixmouth and fixbuff. Slots used: 0 and 1 (hidden, via d2dslot).
