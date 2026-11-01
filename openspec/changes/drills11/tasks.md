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
result: "verified live: fixemptyplan PASS x3 with a failing pre-fix control (6 empty path of 47 back-offs); fixportalmelee control PASS, the 3-adjacent drill FAIL as a product finding F18 (rule unreachable at live distances); cut-cast data recorded (hits do not cut, debug hit does); regression pass"
live: "D-emptyplan PASS x3 (0 empty path of 58 back-offs; control 6 of 47), D-portalmelee FAIL x3 (F18), D-portalmelee.ctl PASS x3, D-castcut recorded, D2 D5 D6 DW F pass"
```

- [x] 1.1 fixemptyplan live: D-emptyplan x3 plus the pre-fix control
- [x] 1.2 fixportalmelee live: D-portalmelee, control, x3
- [x] 1.3 cut-cast data: D-castcut
- [x] 1.4 regression: D2, D5, D6, DW, F once

Ownership: no docs/OWNERS line added. The files this card edits (cmd/d2drill, internal/drill, docs/probes/drills, the README section, primitives rows) are covered by the standing drills globs of m1-liveproof; this card edits them as the harness change DRILLS5 to DRILLS10 did, and writes docs/probes/drills/*drills11* only.
