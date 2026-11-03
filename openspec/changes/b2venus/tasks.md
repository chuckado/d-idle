# Tasks

```yaml
change: b2venus
status: in-progress
agent: b2venus
branch: agent/b2venus
worktree: d2lod-wt/b2venus
skip_specs: true
owns:
  - docs/research/b2-venus-rerun.md
  - openspec/changes/b2venus/**
exclusive: []
gates:
  - "MSVC 14.51 and WSL results for local/batch2-bundle-r3, local/batch2-mana-teleport, local/batch2-monsterdmg-pr recorded with the exact heads"
  - "changecheck 0 warnings, leakcheck 0 hits"
result: "PARTIAL, not verified live on the game: MSVC 14.51 (0 warnings, ctest 0 failed) and WSL parity stages rc 0 for the three branches, damage-roll mdmg2 and v3 0 differences on r3. Mana, Teleport seed 9 and v4 not run (OOM and queueing on the shared Venus); see docs/research/b2-venus-rerun.md"
credit:
  component: Batch 2 re-run on the Venus build worker
  author: owner
  traced_by: []
  agents: [Claude Code]
  verified_by: []
  consent: none
```

## 1. Work

- [x] 1.1 MSVC 14.51 lane for base 005c7cab and the three branches (build, warnings, ctest)
- [x] 1.2 WSL Linux lane (tools/parity.sh stages) for the three branches
- [ ] 1.3 Counted differentials in WSL: damage-roll sets done (mdmg2, v3); mana, Teleport seed 9, v4 NOT run
- [x] 1.4 docs/research/b2-venus-rerun.md
