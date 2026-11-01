# Tasks

```yaml
change: capdocs
status: archived
agent: capdocs
branch: agent/capdocs
worktree: ~/code/d2lod-wt/capdocs
skip_specs: true
owns:
  - docs/capabilities.md
  - docs/running.md
  - README.md
  - docs/research/sorc-skill-matrix.md
  - openspec/changes/capdocs/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "capabilities, README, running and the sorc matrix refreshed; claims scoped to drill ids; docs only"
live: "not verified live by this card (it cites drill ids)"
merge: 938760d
```

- [x] capabilities.md, README, running.md
- [x] sorc-skill-matrix.md cells
