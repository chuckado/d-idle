# Tasks

```yaml
change: palprobe
status: in-progress
agent: palprobe
branch: agent/palprobe
worktree: ~/code/d2lod-wt/palprobe
skip_specs: true
owns:
  - internal/combat/skills.go
  - internal/combat/classdata.go
  - internal/combat/profileid_test.go
  - internal/combat/palprobe_test.go
  - internal/leveler/prefs.go
  - internal/leveler/palprobe_test.go
  - internal/leveler/palprobe_ladder_test.go
  - internal/town/panel/panel.go
  - internal/town/panel/panel_class_test.go
  - internal/town/palprobe_test.go
  - internal/town/palprobe_merc_test.go
  - internal/leveler/class_test.go
  - internal/resource/palprobe_test.go
  - docs/classes/paladin-probe.md
  - openspec/changes/palprobe/**
exclusive: []
gates:
  - "gofmt, go vet, go test -p 1 per touched package, whole repo on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "pending"
live: "offline only; PALHOLD, PALHAND, PALAURA, PALDUR still owed live"
```

- [x] part A seams: profile id (rows 1 to 3), PrefsFor (row 6), classRow (row 29); row 5 checked
- [x] part B probes (a) to (f)
- [x] (g) Zealot-lite table: not added, S4 seam absent (documented as a ready card)
- [ ] docs/classes/paladin-probe.md
