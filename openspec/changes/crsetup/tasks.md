# Tasks

```yaml
change: crsetup
status: in-progress
agent: crsetup
branch: agent/crsetup
worktree: d2lod-wt/crsetup
skip_specs: true
owns:
  - .coderabbit.yaml
  - docs/process/coderabbit-prescan.md
  - openspec/changes/crsetup/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
result: "Docs and config only. Both YAML files validate against the published schema (jsonschema, 0 errors). Nothing run against GitHub or CodeRabbit; quota test and CLI sign-in unrun."
live: "docs only, nothing run live"
```

- [x] .coderabbit.yaml for d-idle, schema-validated
- [x] fork-coderabbit.yaml in the card dir
- [x] runbook with install, fork config, quota test, workflow, CLI
