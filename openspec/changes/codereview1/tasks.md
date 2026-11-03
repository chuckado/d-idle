# Tasks

```yaml
change: codereview1
status: in-progress
agent: codereview1
branch: agent/codereview1
worktree: d2lod-wt/codereview1
skip_specs: true
owns:
  - docs/research/codereview1.md
  - openspec/changes/codereview1/**
exclusive: []
gates:
  - "changecheck -cli 0 new warnings; leakcheck -mode fail 0"
result: "review delivered in docs/research/codereview1.md: 10 findings with smallest fixes (no code changed); offline code reading only, nothing verified live"
live: "not verified; measurement checklist at the end of the doc"
credit:
  author: owner
  agents: [Claude Code]
```

- [x] read the BELTFIX..BELTFIX5, HIREFIX diffs and the code they touch
- [x] rank findings with file:line, scenario, worst-case cost, smallest fix
- [x] test-coverage review
- [x] card, OWNERS, merge lines
