# Tasks

```yaml
change: upstreamsync
status: in-progress
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
result: "Offline, nothing verified live, no game, nothing pushed or posted, fork not rebased. Upstream main 7c3a50af: merge preview clean for #155, #157, #158 and the armor, resume, bundle, curse-immune and price follow-on candidates; #156 conflicts in one tools/emu/README.md row; four rebased copies built and checked (pr_guard 0/0, test_rules, test_game); fork stack conflicts in update.cpp and rules.hpp from the lab click-modifier commit (same against the old main); plan and outdated-docs list in docs/upstreamsync.md."
```

## 1. Work

- [x] 1.1 Fetch upstream main (7c3a50af) in the PR checkout and the fork
- [x] 1.2 merge-tree of every open PR branch and local candidate against the new main
- [x] 1.3 Trivial rebases as new branches in new worktrees, pr_guard and tests re-run
- [x] 1.4 Fork branch rebase check and plan (no rebase done)
- [x] 1.5 Superseded local fixes and outdated docs list
- [x] 1.6 docs/upstreamsync.md, card, OWNERS
