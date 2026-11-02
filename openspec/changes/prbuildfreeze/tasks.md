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
  - fork-branch:local/freeze-rules-v4
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0"
result: "Offline on the worker host, no game, nothing verified live. Private fork branch local/freeze-rules-v4 head 9caa1b9b (base cb716f10) pushed to the private remote only; no PR, nothing to the public fork. PR 159 is merged in the base; clean merge with PR 174 checked. Differential 185,746 -> 0 of 225,748 (also on v4 + PR 174 merged tree). Det hashes, MSVC and the re-check are U."
```

## 1. Work

- [x] 1.1 Order versus PR 159 (merged meanwhile; v3 and v4 shapes, throwaway merges)
- [x] 1.2 Branch local/freeze-rules-v3 on upstream main, pushed to the private remote
- [x] 1.3 Rebuild, suites, freeze differential on the final head
- [x] 1.4 CI parity (g++, clang++, include-cleaner, pr_guard, presubmit --quick)
- [x] 1.5 Draft docs/research/pr-freeze-rules-draft.md
