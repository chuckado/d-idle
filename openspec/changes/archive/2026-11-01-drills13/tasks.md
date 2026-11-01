# Tasks

```yaml
change: drills13
status: archived
merge: 89d6a22
agent: drills13
branch: agent/drills13
worktree: ~/code/d2lod-wt/drills13
skip_specs: true
owns:
  - openspec/changes/drills13/**
exclusive: []
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
result: "verified live: D-fro1, D-fro2 PASS (fresh tables cache), D-m1 recorded twice identically; see DRILLS13 in docs/probes/drills/README.md"
live: "verified live: D-fro1, D-fro2 PASS (with a fresh tables cache), D-m1 recorded twice identically"
```

- [x] D-fro1 / D-fro2 Frozen Armor recast proof
- [x] D-m1 mana timing and interrupted cast
- [x] logs, README section, primitives rows, card lines

Ownership: no docs/OWNERS line added. The files this card edits (internal/drill/d13.go, cmd/d2drill/d13_lab.go and a three-line hook in cmd/d2drill/main.go, docs/probes/drills, the README section, primitives rows) are covered by the standing drills globs of m1-liveproof, as for DRILLS5 to DRILLS11; it also sets the `live:` line of card fixrecast.
