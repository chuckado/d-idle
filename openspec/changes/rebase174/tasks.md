# Tasks

```yaml
change: rebase174
status: in-progress
agent: rebase174
branch: agent/rebase174
worktree: d2lod-wt/rebase174
skip_specs: true
owns:
  - docs/research/rebase174-20261102.md
  - docs/research/pr-freeze-rules-draft.md
  - docs/research/pr-vendor-repair-gate-draft.md
  - docs/research/pr-bundle-1-draft.md
  - openspec/changes/rebase174/**
exclusive:
  - fork-branch:local/freeze-rules-v5
  - fork-branch:local/vendor-repair-gate-v4
  - fork-branch:local/skmelee3-d1-r174
  - fork-branch:local/skmelee3-d2-r174
  - fork-branch:local/bundle-1
gates:
  - "seven suites green on each new head; counted differential base vs head on the new upstream main"
result: "Four private branches rebased onto upstream main 48908176 (post 174) and pushed to the private fork remote only: local/freeze-rules-v5 32c4943a, local/vendor-repair-gate-v4 ee2d03f9, local/skmelee3-d1-r174 5689c007, local/skmelee3-d2-r174 10e16ad5 (D2 conflict with 174 resolved); plus local/bundle-1 ed087f88 (10 commits, 20 files, +759 -42). Suites, counted differentials, CI parity re-run on each and on the bundle; table in docs/research/rebase174-20261102.md, combined draft docs/research/pr-bundle-1-draft.md. Verified offline on the worker host (emulator against game.exe); not verified live (no game)."
live: "offline: no game; game.exe's own functions run under the emulator on the worker host"
```

- [x] 1 Rebase the four candidates onto the new upstream main (D2 conflict resolved)
- [x] 2 Seven suites and the counted differentials, base vs head on the new main
- [x] 3 CI parity (g++, clang++, tidy, pr_guard, presubmit incl. the attribution check)
- [x] 4 Push the new branches to the private remote only
- [x] 5 Table doc and draft base lines, changecheck, leakcheck
- [x] 6 Bundle branch local/bundle-1, checks on its head, combined draft
