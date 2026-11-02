# Tasks

```yaml
change: monsterpace
status: in-progress
agent: monsterpace
branch: agent/monsterpace
worktree: ~/code/d2lod-wt/monsterpace
skip_specs: true
owns:
  - docs/research/pr-monsterpace-draft.md
  - tools/emu/mp_run.sh
  - tools/emu/mp_patch.py
  - openspec/changes/monsterpace/**
exclusive:
  - "fork-branch:local/monsterpace"
  - "fork-branch:local/monsterpace-clean"
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "candidate built and counted on the worker host (100,000 cases per think, 122,646 speed cases, 100,000 corpse cases); not verified live (no game); not opened upstream"
live: "no game; the lab d2d binary runs the det scenarios offline"
```

- [x] read upstream think / pace code and PR 156
- [x] real-build probes and fixes on local/monsterpace (commits counted in the draft)
- [x] PR draft
- [ ] independent verification (the coordinator launches it)
