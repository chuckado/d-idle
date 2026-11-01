# Tasks

```yaml
change: skground
status: in-progress
agent: skground
branch: agent/skground
worktree: ~/code/d2lod-wt/skground
skip_specs: true
owns:
  - docs/research/sk-ground-trace.md
  - tools/emu/sk_ground*
  - openspec/changes/skground/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: ""
live: "docs only, nothing run live"
```

- [ ] Blizzard and Meteor traced in game.exe (address tags)
- [ ] Frozen Orb, Thunder Storm, Hydra, Fire Wall (as far as the time box allows)
- [ ] d2d comparison (source read; real build where the tooling allows)
- [ ] bot findings
- [ ] trace doc with deviations, drill recipe, matrix cells, shared-file lines
