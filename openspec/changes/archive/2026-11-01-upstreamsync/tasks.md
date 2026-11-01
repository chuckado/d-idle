# Tasks

```yaml
change: upstreamsync
status: archived
merge: 19c1daf
agent: upstreamsync
branch: agent/upstreamsync
worktree: ~/code/d2lod-wt/upstreamsync
skip_specs: true
owns:
  - openspec/changes/upstreamsync/**
  - docs/upstreamsync.md
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "no push, no GitHub or upstream post, no fork rebase"
result: "upstream merges recorded; every candidate merges clean (rebased copies built); fork rebase plan written: needs new branches, not a force-push (FORKREBASE). Offline docs, not verified live."
```

## 1. Work

- [x] 1.1 Fetch upstream main (7c3a50af) in the PR checkout and the fork
- [x] 1.2 merge-tree of every open PR branch and local candidate against the new main
- [x] 1.3 Trivial rebases as new branches in new worktrees, pr_guard and tests re-run
- [x] 1.4 Fork branch rebase check and plan (no rebase done)
- [x] 1.5 Superseded local fixes and outdated docs list
- [x] 1.6 docs/upstreamsync.md, card, OWNERS
