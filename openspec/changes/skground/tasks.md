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
result: "Docs and offline tooling: docs/research/sk-ground-trace.md (Q1-Q16, D1-D8, B1-B5, drill G1-G5); Blizzard and Meteor finished, Frozen Orb and Thunder Storm read at dispatch level, Hydra and Fire Wall not reached; game-side oracle tools/emu/sk_ground_blizz.py run on the worker host (64,200 cases, 0 differ from the reading); no real d2d build comparison, no PR candidate. Not verified live."
live: "docs only, nothing run live"
```

- [x] Blizzard and Meteor traced in game.exe (address tags)
- [x] Frozen Orb, Thunder Storm (read); Hydra, Fire Wall not reached
- [x] d2d comparison (source read; game-side run; no real build)
- [x] bot findings
- [x] trace doc with deviations, drill recipe, matrix cells, shared-file lines
