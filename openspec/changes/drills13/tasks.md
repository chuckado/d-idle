# Tasks

```yaml
change: drills13
status: in-progress
agent: drills13
branch: agent/drills13
worktree: ~/code/d2lod-wt/drills13
skip_specs: true
owns:
  - internal/drill/d13.go
  - cmd/d2drill/d13_lab.go
  - docs/probes/drills/2026-11-01-drills13-*.txt
  - openspec/changes/drills13/**
exclusive:
  - instance:slot1
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
result: "A and B run live on a hidden slot (lab-debug-skills binary); see the DRILLS13 section of docs/probes/drills/README.md"
live: "verified live: D-fro1, D-fro2 PASS (with a fresh tables cache), D-m1 recorded twice identically"
```

- [x] D-fro1 / D-fro2 Frozen Armor recast proof
- [x] D-m1 mana timing and interrupted cast
- [x] logs, README section, primitives rows, card lines
