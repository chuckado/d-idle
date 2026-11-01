# Tasks

```yaml
change: roamleg
status: archived
merge: e11da5e
agent: roamleg
branch: agent/roamleg
worktree: ~/code/d2lod-wt/roamleg
skip_specs: true
owns:
  - docs/roamleg.md
  - internal/nav/roamleg_test.go
  - openspec/changes/roamleg/**
exclusive: []
gates:
  - "go test -p 1 ./internal/nav ./internal/combat passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "held mouth-gate walk no longer condemns roam targets: failed legs 50 -> 0, step-aways 62 -> 11, bounces stay 0 (verified live, DRILLS23); sweep completeness unverified because the drill character kills nothing"
live: "verified live (DRILLS23): DW.m3 x3 0 bounces, 0 roam leg failed (was 50), held line 1 per run, step-aways 62 to 11, D3 12/12, D2 11/11; Den clear completeness not shown (no attack skill)"
```

- [x] design note (docs/roamleg.md)
- [x] roam.go: ErrMouthGate ends the episode, ErrRoamHeld, counter, debug line
- [x] tests (held leg, no-path leg unchanged)
- [x] live proof recipe
- [x] live proof (DW.m3 x3, DRILLS23)
