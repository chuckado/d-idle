# Tasks

```yaml
change: seedspawn
status: in-progress
agent: seedspawn
branch: agent/seedspawn
worktree: d2lod-wt/seedspawn
skip_specs: true
owns:
  - docs/research/seedspawn-trace.md
  - tools/emu/ss_*
  - openspec/changes/seedspawn/**
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "docs and emulator scripts only; V run on the worker host (game.exe in the emulator against d2d's real code), not verified on a live game; no PR candidate"
credit:
  component: SEEDSPAWN
  author: owner
  agents: [Claude Code]
```

- [x] 1. Read seed-study.md and the upstream oracle tools
- [x] 2. Differential: levels 1, 2, 3, 4, 8, units / seeds / monsters, seeds 1..30 plus two 32-bit ranges and edge seeds
- [x] 3. d2d source: where the world RNG is shared
- [x] 4. Write docs/research/seedspawn-trace.md
- [x] 5. Ledger rows for what is proven (levels.layout, coverage files regenerated); checks; merge-lines
