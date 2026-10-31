# Tasks

```yaml
change: wppr
status: in-progress
agent: wppr
branch: agent/wppr
worktree: ~/code/d2lod-wt/wppr
skip_specs: true
owns:
  - docs/research/pr-wp-*.md
  - tools/emu/wp_*.py
  - openspec/changes/wppr/**
  - docs/decomp-queue.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "the C++ differential (test_game --waypoint-refuse against tools/emu/wp_refuse.py game) has mismatches before > 0 and 0 after over >= 100,000 random cases plus the exhaustive grid"
  - "upstream test_rules, test_combat, test_game, test_protocol pass; ctest failures equal to an unmodified upstream build; pr_guard 0 FAIL 0 WARN; presubmit --quick 0 FAIL"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
```

## 1. Work

- [ ] 1.1 Read the decompwp trace, proof, draft, evidence pack and the finished packs (belt, tome, price)
- [ ] 1.2 Replay the local branch onto the current upstream main (tooling commit first, then the claim)
- [ ] 1.3 Oracle in upstream style (`gen` / `game`) and `test_game --waypoint-refuse`: d2d's real C++ against game.exe
- [ ] 1.4 d2d unit test in tests/game/test_game.cpp; upstream tests and ctest against an unmodified build
- [ ] 1.5 pr_guard, presubmit --quick (ruff included)
- [ ] 1.6 Independent Opus verifier; report verbatim in the proof doc
- [ ] 1.7 Evidence pack, draft, proof; decomp-queue, bundle-queue, BACKLOG lines; card, OWNERS, changecheck, leakcheck, openspec validate
