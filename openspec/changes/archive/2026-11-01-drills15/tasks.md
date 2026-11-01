# Tasks

```yaml
change: drills15
status: archived
merge: c53dc1d
agent: drills15
branch: agent/drills15
worktree: ~/code/d2lod-wt/drills15
skip_specs: true
owns:
  - openspec/changes/drills15/**
exclusive: []
gates:
  - "gofmt, go vet and go test -p 1 ./internal/drill ./cmd/d2drill pass; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
result: "verified live on d2d (game.exe not run: comparison is to the traces): 13 drills, see DRILLS15 in docs/probes/drills/README.md"
live: "verified live on d2d: T1 PASS, T2 T3 T4 T5 T7 FAIL (deviations), M2 M3 M6 E1 FAIL (deviations), M4 K1 K2 PASS"
```

- [x] Teleport T1 to T5, T7
- [x] Mana M2, M3, M4, M6 (M5 not run)
- [x] Missile K1 reach, K2 Charged Bolt count (Lightning pass-through, roll not run)
- [x] Armor E1 (A2 to A5 not run)
- [x] logs, README section, primitives rows, card lines
