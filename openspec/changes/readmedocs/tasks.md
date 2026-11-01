# Tasks

```yaml
change: readmedocs
status: in-progress
agent: readmedocs
branch: agent/readmedocs
worktree: ~/code/d2lod-wt/readmedocs
skip_specs: true
owns:
  - README.md
  - docs/capabilities.md
  - docs/running.md
  - openspec/changes/readmedocs/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict; relative links resolve"
result: "Docs only. Offline examples (build, -h, d2dsoak compare on the committed soak log, d2dash status, d2bot --sim stopping on missing tables) were run; nothing that needs a game was run."
live: "docs only, nothing run live"
```

- [x] docs/capabilities.md
- [x] docs/running.md
- [x] README sections
- [x] checks and link check
