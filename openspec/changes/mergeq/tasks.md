# Tasks

```yaml
change: mergeq
status: in-progress
agent: mergeq
branch: agent/mergeq
worktree: d2lod-wt/mergeq
skip_specs: true
owns:
  - tools/merge_lines_check.py
  - tools/ci/check.sh
  - tools/ci/tests/test_merge_lines_check.py
  - AGENTS.md
  - openspec/changes/mergeq/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict; python3 -m unittest tools/ci/tests"
result: "merge-lines linter plus check.sh wiring; tests offline. Not run against a real merge."
live: "tooling only, nothing run live"
credit:
  author: owner
  agents: [Claude Code]
```

- [x] merge_lines_check.py and tests
- [x] wire into tools/ci/check.sh (--changed)
- [x] shared-file lines
