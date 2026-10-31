# Tasks

```yaml
change: decompstatic
status: archived
agent: decompstatic
branch: agent/decompstatic
worktree: ~/code/d2lod-wt/decompstatic
skip_specs: true
owns:
  - docs/research/static-field-chill-trace.md
  - docs/research/pr-static-*.md
  - tools/emu/static_field.py
  - tools/emu/cold_length.py
  - openspec/changes/decompstatic/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, D, S, I or U with an address or file:line"
  - "tools/emu/static_field.py runs game.exe's code on >= 100,000 cases: mismatches before > 0, after the candidate 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: f0c9f25
result: "Static Field = MulDiv(life >> 8, calc1 25, 100) whole points of current life, at most life - 1, none at or under StaticFieldMin 0/33/50 % (expansion), negative lightning resistance cancelled, then 256ths with pierce; hits monsters and hostile players in 5 + (lvl-1) subtiles with line of sight. d2d deviates (S1-S6): 309,664 oracle cases, 115,103 differ before, 0 after the local candidate fix/static-field (2 commits on upstream 72132b99, pr_guard 0/0, upstream tests rc 0, no verifier yet, not posted). Cold length: zeroed by stat 0x99, halved by 0x76, resisted like damage, chill applied whatever the damage; d2d needs a whole resisted point and ignores the stats: 28,739 of 101,170 cases (12,804 of 81,559 with neither stat); no patch. Bot: no damage or chill model; Normal floor 25 % is tuning; drill recipe in the trace. Not verified live."
```

## 1. Work

- [x] 1.1 Read AGENTS.md, decomp-queue rules, the PR standard, skill-damage-trace, walk-speed-trace
- [x] 1.2 game.exe: FUN_005c9800, FUN_005c96a0, FUN_0056b7e0, FUN_0056b3e0, FUN_006492a0, FUN_0056c8e0, FUN_0057c6c0, FUN_0057c1e0, FUN_0057c140, FUN_0057bf80, FUN_0057be00, FUN_0057af80, FUN_0057b230, table 0x732980, loader descriptors (V)
- [x] 1.3 d2d: kStaticField in Fight::spot, missile_blow, land's chill block (S)
- [x] 1.4 Oracles `tools/emu/static_field.py` (309,664 cases) and `tools/emu/cold_length.py` (101,170 cases)
- [x] 1.5 PR candidate: local branch fix/static-field, after-run 0; upstream tests and pr_guard (evidence pack)
- [x] 1.6 Bot model read (internal/combat, env tuning), findings B1-B5, DRILL RECIPE
- [ ] 1.7 Independent verifier (not run: a later step needing the coordinator)
- [x] 1.8 Trace doc, card, OWNERS, changecheck, leakcheck, openspec validate
