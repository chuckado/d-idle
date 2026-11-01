# Tasks

```yaml
change: buildmods
status: archived
agent: buildmods
branch: agent/buildmods
worktree: ~/code/d2lod-wt/buildmods
skip_specs: true
owns:
  - docs/research/pr-mods-proof.md
  - docs/research/pr-mods-draft.md
  - tools/emu/boss_level.py
  - tools/emu/boss_level_test.cpp
  - docs/research/monster-mods-trace.md
  - docs/decomp-queue.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/buildmods/**
exclusive:
  - fork-branch:fix/boss-stats
gates:
  - "each claim commit: mismatches before > 0 and after 0 on >= 100,000 cases, game.exe's own code vs real d2d builds"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: a3a56f1
result: "D1 re-derived and NARROWED: game.exe reads a boss's life, defence and experience once at the MonStats level (FUN_00573cb0, mask 7) but AR and damage per hit at the raised level (FUN_005a4f50, stat 12), so d2d's AR and damage were right and only life / defence / XP change; D2 (minion leveladd +3 / x5), D3 (ghostly keeps its speed; the -33 store is dead code) and B (berserker +3 / x5) confirmed. Built as local fix/boss-stats on upstream main 72132b99 (4 claim commits + tooling, head c048a356, 6 files): game.exe's own code vs real d2d builds, 136,624 cases: D1 286,774, D2 109,003, D3 5,939, B 26,868 field mismatches -> 0 each; residuals S1 (803), H1 (1,198), life rounding and lifeless rows classified, not claimed. pr_guard 0/0 range and per commit, presubmit --quick 0 FAIL / 2 WARN, ruff clean, tidy and gcc shadow clean on touched files, test_game / rules / combat / protocol / monsters rc 0. Gameplay-visible: Act 1 Normal bosses 1.3-2.1x less life, 1.2-2x less XP; unique packs about 2x XP. Not verified live; no verifier, no det hashes (d2d slot busy); nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Re-derive D1 / D2 / D3 from the disassembly (FUN_00573cb0, FUN_005a4f50, FUN_005a0e40, FUN_005a0e80, FUN_005a2120)
- [x] 1.2 Upstream notes on boss levels (docs/research/re, git history)
- [x] 1.3 Oracle `boss_level.py` and test mode `test_boss_level --boss-level`
- [x] 1.4 Claim commits on upstream main, differential per commit on real builds
- [x] 1.5 CI-class checks
- [x] 1.6 Proof, draft, queue rows, backlog line
