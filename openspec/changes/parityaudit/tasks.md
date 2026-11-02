# Tasks

```yaml
change: parityaudit
status: in-progress
agent: parityaudit
branch: agent/parityaudit
worktree: ~/code/d2lod-wt/parityaudit
skip_specs: true
owns:
  - docs/research/backend-parity.md
  - openspec/changes/parityaudit/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only: docs/research/backend-parity.md. Build/test: whole repo go vet + go test -short on the worker host exit 0 (88 packages ok), 4 skips. 26-row matrix, gate tooling gaps, smoke runbook, backlog. Nothing run live (read-only audit)."
live: "docs only, nothing run live"
```

- [x] build and test state (this checkout and the worker host)
- [x] feature parity matrix
- [x] gate tooling gaps and cost estimate
- [x] smoke runbook and ordered backlog
