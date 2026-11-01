# Tasks

```yaml
change: drills11
status: in-progress
agent: drills11
branch: agent/drills11
worktree: ~/code/d2lod-wt/drills11
skip_specs: true
owns:
  - openspec/changes/drills11/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: "see README DRILLS11"
```

- [ ] 1.1 fixemptyplan live: D-emptyplan x3 plus the pre-fix control
- [ ] 1.2 fixportalmelee live: D-portalmelee, control, x3
- [ ] 1.3 cut-cast data: D-castcut
- [ ] 1.4 regression: D2, D5, D6, DW, F once

Ownership: no docs/OWNERS line added. The files this card edits (cmd/d2drill, internal/drill, docs/probes/drills, the README section, primitives rows) are covered by the standing drills globs of m1-liveproof; this card edits them as the harness change DRILLS5 to DRILLS10 did, and writes docs/probes/drills/*drills11* only.
