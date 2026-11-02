# Tasks

```yaml
change: prconflicts
status: in-progress
agent: prconflicts
branch: agent/prconflicts
worktree: d2lod-wt/prconflicts
skip_specs: true
owns:
  - openspec/changes/prconflicts/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Three PR branches rebased on upstream main under local/prconf-159, -160, -163; 7 suites and the PR differentials pass; not pushed to the PR branches (rebase needs force-push, user decides). Not verified in a live game."
live: "not applicable"
```

- [x] read the maintainer's rules fresh
- [x] rebase the three branches, resolve conflicts
- [x] suites, differentials, parity checks
- [x] log entry
