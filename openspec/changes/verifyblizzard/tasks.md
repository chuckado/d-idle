# Tasks

```yaml
change: verifyblizzard
status: active
agent: verifyblizzard
branch: agent/verifyblizzard
worktree: ~/code/d2lod-wt/verifyblizzard
skip_specs: true
owns:
  - openspec/changes/verifyblizzard/**
  - tools/emu/verify_blizzard.py
  - tools/emu/verify_blizzard_mode.inc
  - docs/research/pr-blizzard-verifier.md
exclusive: []
gates:
  - "own game-side harness: model vs game.exe FUN_005a9820 0 differences on 100,000+ cases"
  - "own d2d probe on real builds: shards off game.exe's -6..5 grid before > 0, after 0"
  - "changecheck 0 new warnings; leakcheck -mode fail 0"
result: ""
```

- [ ] 1.1 Re-derive FUN_005a9820 from the disassembly; own generator and emulator harness, 100,000+ cases incl. edges
- [ ] 1.2 Own d2d probe; real builds of base and head; counts and histograms
- [ ] 1.3 Diff review, other callers, merge-tree vs upstream main and PRs #159 / #160
- [ ] 1.4 Author's side claims D2 / D3 and whether the PR should carry more than one line
- [ ] 1.5 Verdict doc, shared-file lines for the merger
