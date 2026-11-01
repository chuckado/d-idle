# Tasks

```yaml
change: skillinv
status: in-progress
agent: skillinv
branch: agent/skillinv
worktree: ~/code/d2lod-wt/skillinv
skip_specs: true
owns:
  - openspec/changes/skillinv/**
  - docs/research/sorc-skill-matrix.md
exclusive: []
gates:
  - "docs/research/sorc-skill-matrix.md has the inventory, the matrix with V/S/I/U tags, the ranked gap decomps, the drill design and the per-act gate; changecheck -cli 0 warnings; leakcheck -mode fail 0; check.sh --hygiene-only no FAIL; openspec validate --all --strict"
merge:
result: ""
live: "not run (docs only; the drill is a proposal)"
```

- [x] 1.1 Step 1: inventory of the 30 Sorceress skills and the plan's use by act
- [x] 1.2 Step 2: coverage matrix with tags and live proof status
- [x] 1.3 Step 3 proposal: ranked gap decomps (SKMISS, SKARMOR, SKTELE, SKLIGHT, SKNOVA, SKMANA, ...)
- [x] 1.4 Step 4 proposal: skill-smoke drill design, missing fork verbs, per-act gate
- [x] 1.5 BACKLOG line, INDEX entry, decomp-queue rows 14.. (proposed)
