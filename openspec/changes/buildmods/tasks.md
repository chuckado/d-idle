# Tasks

```yaml
change: buildmods
status: in-progress
agent: buildmods
branch: agent/buildmods
worktree: ~/code/d2lod-wt/buildmods
skip_specs: true
owns:
  - docs/research/pr-mods-proof.md
  - docs/research/pr-mods-draft.md
  - tools/emu/boss_level.py
  - tools/emu/boss_level_test.cpp
  - openspec/changes/buildmods/**
exclusive: []
gates:
  - "each claim commit: mismatches before > 0 and after 0 on >= 100,000 cases, game.exe's own code vs real d2d builds"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
```

## 1. Work

- [ ] 1.1 Re-derive D1 / D2 / D3 from the disassembly (FUN_00573cb0, FUN_005a4f50, FUN_005a0e40, FUN_005a0e80, FUN_005a2120)
- [ ] 1.2 Upstream notes on boss levels (docs/research/re, git history)
- [ ] 1.3 Oracle `boss_level.py` and test mode `test_boss_level --boss-level`
- [ ] 1.4 Claim commits on upstream main, differential per commit on real builds
- [ ] 1.5 CI-class checks
- [ ] 1.6 Proof, draft, queue rows, backlog line
