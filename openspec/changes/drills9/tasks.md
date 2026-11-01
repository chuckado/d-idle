# Tasks

```yaml
change: drills9
status: archived
agent: drills9
branch: agent/drills9
worktree: ~/code/d2lod-wt/drills9
skip_specs: true
owns:
  - openspec/changes/drills9/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge: 98b8178
result: "verified live: H.pause 6/6 + 6/6, D10 2/2, harness F14 fixes (P1s, chest, shrine, D8.c), regression table in README DRILLS9; open: F17 DW.m1 in group F"
live: "H.pause PASS x3, D10 PASS x2, regression pass; DW.m1 flaky in the full F group (F17)"
```

- [x] 1.1 fixstall4: H.pause x3 live
- [x] 1.2 D10 full drill, two runs
- [x] 1.3 harness: group-run isolation (D1.P1s, H.obj.chest, D-sync)
- [x] 1.4 regression pass, one table

Ownership: docs/OWNERS line for drills9 added; its globs do not overlap m1-liveproof (the drills globs of cmd/d2drill etc. are m1-liveproof's; this card edits them as the harness change DRILLS5 to DRILLS8 did).
