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
  - tools/emu/ss_run.sh
  - openspec/changes/seedspawn/**
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: ""
credit:
  component: SEEDSPAWN
  author: owner
  agents: [Claude Code]
```

- [ ] 1. Read seed-study.md and the upstream oracle tools
- [ ] 2. Differential: levels 1, 2, 3, 4, 8, units / seeds / monsters, seeds 1..20+
- [ ] 3. d2d source: where the world RNG is shared
- [ ] 4. Write docs/research/seedspawn-trace.md
- [ ] 5. Ledger rows for what is proven; checks; merge-lines
