# Tasks

```yaml
change: contribid
status: in-progress
agent: contribid
branch: agent/contribid
worktree: d2lod-wt/contribid
skip_specs: true
owns:
  - AGENTS.md
  - docs/contributing/agent-work.md
  - .github/pull_request_template.md
  - openspec/changes/contribid/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
result: "Docs and PR template only; nothing run live."
live: "docs only, nothing run live"
```

- [x] AGENTS.md identity rule split
- [x] reproducibility rule and checkboxes
