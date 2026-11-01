# Tasks

```yaml
change: skmelee2
status: in-progress
agent: skmelee2
branch: agent/skmelee2
worktree: ~/code/d2lod-wt/skmelee2
skip_specs: true
owns:
  - docs/research/sk-melee2-trace.md
  - tools/emu/sk_melee2.py
  - tools/emu/sk_melee2_test.cpp
  - openspec/changes/skmelee2/**
exclusive: []
gates:
  - "real-build differential: game.exe oracle vs d2d test_melee on seeds 7 and 1, 100,000 cases a part, counts reported per part"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: ""
```

Overlaps (not edited here, lines for the merger in the trace's last section): docs/decomp-queue.md, docs/BACKLOG.md, docs/INDEX.md, CHANGELOG.md, docs/research/README.md, docs/OWNERS, docs/research/sorc-skill-matrix.md, docs/research/sk-melee-trace.md (status banner).

## 1. Work

- [ ] 1.1 d2d test mode and oracle driver; build on a scratch checkout of upstream main 72132b99
- [ ] 1.2 Run seeds 7 and 1 (100,000 cases a part); per-part differences
- [ ] 1.3 D7 and the shield carrier: read game.exe's statlist aggregation and d2d
- [ ] 1.4 PR decision per proven deviation
- [ ] 1.5 Trace doc: deviations, bot findings, drill recipe, matrix cells, shared-file lines
