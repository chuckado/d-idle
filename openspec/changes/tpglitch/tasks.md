# Tasks

```yaml
change: tpglitch
status: proposed
agent: tpglitch
branch: agent/tpglitch
worktree: ~/code/d2lod-wt/tpglitch
skip_specs: true
owns:
  # code paths (reader.go, portal.go, resource.go, tpdebounce.go, tests) are added when an agent takes the card
  - openspec/changes/tpglitch/**
exclusive: []
gates:
  - go test -p 1 ./internal/d2denv ./internal/combat ./internal/bot passes, including a replay of the 4, no70, 4 sequence of run gate1-20261101-1949 (19:57:17 to 19:57:30)
result: not started
```

- [ ] 1 Test first: replay of the 4, no70 for 12 s, 4 sequence; the verdict stays true, kit gate and poverty mode do not flip
- [ ] 2 Keep the last positive count across a quantity 0 read (reader or portal.go), and one debounce at resource.go:33 for every reader
- [ ] 3 A real empty tome (use to 0, then no scroll bought) still reads false within the debounce span
- [ ] 4 Fork-side probe checklist in the report (not run)
