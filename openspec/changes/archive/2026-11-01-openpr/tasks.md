# Tasks

```yaml
change: openpr
status: archived
merge: 776af38
agent: openpr
branch: agent/openpr
worktree: ~/code/d2lod-wt/openpr
skip_specs: true
owns:
  - openspec/changes/openpr/**
  - docs/research/openpr-log.md
  - docs/research/pr-bundle-queue.md
exclusive:
  - "upstream-pr:frozen-armor-resume-bundle"
  - "upstream-pr:curse-immune"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Opened upstream #159 (bundle: Frozen Armor, then freeze resume; 13 files, 2 commits) and #160 (curse on immune; 5 files, 2 commits), both from the fork, plain pushes, nothing else posted. Verified offline on the worker host, not live."
```

## 1. Work

- [x] 1.1 Rebase both candidates onto upstream main 52362d52 (no conflicts), strip stray cherry-pick lines from the commit messages
- [x] 1.2 Re-vet: pr_guard 0/0 range and per commit, Release build, five test binaries, tooling unit tests, both differentials
- [x] 1.3 Push to the fork and open #159, then #160 (bodies from the verified drafts, ending `Agent: Claude Code`)
- [x] 1.4 Posted log and queue lines
