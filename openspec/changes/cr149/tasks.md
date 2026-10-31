# Tasks

```yaml
change: cr149
status: in-progress
agent: CR149
branch: agent/cr149
worktree: ~/code/d2lod-wt/cr149
skip_specs: true
owns:
  - docs/research/pr-tome-evidence-pack.md
exclusive:
  - upstream-pr:tome
gates:
  - "pr_guard 0 FAIL 0 WARN, upstream tests, both oracles unchanged before the push"
result: pushed c8a9dca8, 6a98903a, 0dc3e0d3 to the fork branch, two thread replies posted; pr_guard 0/0, presubmit --quick 0 FAIL, oracles 0 mismatches; verified on the worker host, no game
```

## 1. Work

- [x] wait for the BUNDLECLICK push, work from the pushed head
- [x] test_game: case_id rename, portal checks before the patch-only return
- [x] take_portal finding: investigate against game.exe, reply
- [x] prove, push, reply in the threads, posted: lines
