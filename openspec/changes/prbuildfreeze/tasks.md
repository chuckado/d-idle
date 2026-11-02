# Tasks

```yaml
change: prbuildfreeze
status: in-progress
agent: prbuildfreeze
branch: agent/prbuildfreeze
worktree: d2lod-wt/prbuildfreeze
skip_specs: true
owns:
  - openspec/changes/prbuildfreeze/**
  - docs/research/pr-freeze-rules-draft.md
exclusive:
  - fork-branch:local/freeze-rules-v3
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0"
result: "Offline on the worker host, no game, nothing verified live. Private fork branch local/freeze-rules-v3 head 21010485 pushed to the private remote only; no PR, nothing to the public fork. Independent of PR 159 (clean merge either order, merged tree built and tested). Differential 185,746 -> 0 of 225,748. Det hashes, MSVC and the re-check are U."
```

## 1. Work

- [x] 1.1 Order and shape versus open PR 159 (throwaway merges both ways)
- [x] 1.2 Branch local/freeze-rules-v3 on upstream main, pushed to the private remote
- [x] 1.3 Rebuild, suites, freeze differential on the final head
- [x] 1.4 CI parity (g++, clang++, include-cleaner, pr_guard, presubmit --quick)
- [x] 1.5 Draft docs/research/pr-freeze-rules-draft.md
