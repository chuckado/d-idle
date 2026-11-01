# Tasks

```yaml
change: capdocs
status: in-progress
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
result: "capability docs refreshed against DRILLS24 to 29 and the offline cards; docs only, nothing run in a game"
live: "not verified live by this card (it cites drill ids)"
```

- [x] capabilities.md, README, running.md
- [x] sorc-skill-matrix.md cells
