# Tasks

```yaml
change: palrecon
status: in-progress
agent: palrecon
branch: agent/palrecon
worktree: ~/code/d2lod-wt/palrecon
skip_specs: true
owns:
  - docs/research/paladin-act1-recon.md
  - openspec/changes/palrecon/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only: docs/research/paladin-act1-recon.md (what exists, Act 1 melee needs, Zealot-lite table proposal, generality metric with predicted file list, 8 SME questions). Nothing run live."
live: "docs only, nothing run live"
```

- [x] read what exists (plandata, combat profiles, leveler, merc, belt, drills, SME)
- [x] write docs/research/paladin-act1-recon.md
- [x] checks and commit
