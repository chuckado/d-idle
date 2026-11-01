# Tasks

```yaml
change: dethash2
status: in-progress
agent: dethash2
branch: agent/dethash2
worktree: ~/code/d2lod-wt/dethash2
skip_specs: true
owns:
  - docs/research/det-hashes-2026-11-01b.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - tools/emu/det_scenarios_lab.py
  - tools/emu/det_scenarios_run.py
  - openspec/changes/dethash2/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: ""
```

## 1. Work

- [ ] 1.1 Build five trees (baseline, #155, freeze v2, walk) with the lab patch set
- [ ] 1.2 New scenarios: cold hits per monster class, merc, pet, player position
- [ ] 1.3 Run all scenarios on all builds; baseline twice
- [ ] 1.4 Verdicts, doc, queue rows, BACKLOG, OWNERS
