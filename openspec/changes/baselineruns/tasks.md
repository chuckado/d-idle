# Tasks

```yaml
change: baselineruns
status: in-progress
agent: baselineruns
branch: agent/baselineruns
worktree: ~/code/d2lod-wt/baselineruns
skip_specs: true
owns:
  - docs/process/baseline-runs.md
  - openspec/changes/baselineruns/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only here: docs/process/baseline-runs.md. Launcher in the ops repo with 64 stub tests passing. Nothing run live: no game started, only read-only worker-host looks (saves, ledger, capacity probe)."
live: "not run live; read-only worker-host looks only"
```

- [x] 1. Find how the old original-game runs were produced
- [x] 2. Runbook docs/process/baseline-runs.md
- [x] 3. Launcher in the ops repo with stub tests
- [x] 4. Report prerequisites and unverifiable items
