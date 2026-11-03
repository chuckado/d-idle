# Tasks

```yaml
change: finala
status: in-progress
agent: finala
branch: agent/finala
worktree: d2lod-wt/finala
skip_specs: true
owns:
  - openspec/changes/finala/**
  - docs/research/pr-guardfix-body.md
  - docs/research/pr-batch2-A-final-body.md
exclusive:
  - "fork-branch:local/guardfix-1"
  - "fork-branch:local/batch2-A-r7"
gates:
  - "changecheck 0 new warnings; leakcheck -mode fail 0; merge_lines_check ok"
  - "fixed pr_guard 0 FAIL on local/batch2-A-r7 and on local/guardfix-1; the new add-and-cite test fails on upstream's guard"
credit:
  component: GUARDFIX
  author: owner
  traced_by: []
  reported_by: []
  verified_by: []
  agents: [Claude Code]
  consent: none
result: "private fork branches local/guardfix-1 and local/batch2-A-r7 pushed; PR texts drafted; nothing posted"
merge:
```

## 1. Work

- [x] 1.1 local/guardfix-1: one commit on upstream main e1eea02a, the pr_guard cited-tool check on the PR head, with two tests
- [x] 1.2 local/batch2-A-r7: the three batch-A commits without the guard commit, on the same upstream main
- [ ] 1.3 Verify r7 on the exact head on the Windows build worker (see the body for the counts)
- [x] 1.4 Push the two branches to the private remote (ghfork pushes after the checks)
- [ ] 1.5 PR texts for the guard fix and for batch A
