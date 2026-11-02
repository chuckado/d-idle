# Tasks

```yaml
change: tpglitch
status: in-progress
agent: tpglitch
branch: agent/soakfixes2
worktree: ~/code/d2lod-wt/tpglitch
skip_specs: true
owns:
  - openspec/changes/tpglitch/**
  - internal/bot/tpdebounce.go
  - internal/bot/tpdebounce_test.go
  - internal/bot/resource.go
  - internal/bot/leveling.go
exclusive: []
gates:
  - go test -p 1 ./internal/d2denv ./internal/combat ./internal/bot passes, including a replay of the 4, no70, 4 sequence of run gate1-20261101-1949 (19:57:17 to 19:57:30)
result: implemented and tested offline (mem.Fake rig); not verified live
```

- [x] 1 Test first: replay of the 4, no70 for 12 s, 4 sequence; the verdict stays true, kit gate and poverty mode do not flip
- [x] 2 Keep the last positive count across a quantity 0 read (reader or portal.go), and one debounce at resource.go:33 for every reader
- [x] 3 A real empty tome (use to 0, then no scroll bought) still reads false within the debounce span
- [x] 4 Fork-side probe checklist in the report (not run)
