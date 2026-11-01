# Tasks

```yaml
change: verifychill
status: archived
agent: verifychill
branch: agent/verifychill
worktree: ~/code/d2lod-wt/verifychill
skip_specs: true
owns:
  - openspec/changes/verifychill/**
  - tools/emu/verify_chill.py
  - tools/emu/verify_chill_mode.inc
  - docs/research/pr-static-chill-proof.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "own differential on real d2d builds: claim mismatches before > 0, after 0, every residual classified"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: 2ff7991
result: "fix/static-field-chill HOLDS WITH LIMITS: own differential 201,757 cases on real d2d builds (6270d162 / 79c9a496), the claim 34,635 -> 0, residuals C3 4,282 / C5 2,836 / roll max <= 0 17,912 / length >= 0x100000 29 classified and unchanged; CI-class checks clean. Offline, nothing verified live, nothing pushed; det state hashes open."
```

- [x] 1.1 Re-derive the rule from the disassembly (FUN_0057c6c0, FUN_0057c1e0, FUN_0057c140, FUN_0057be00, FUN_0057bf80, FUN_0057af80, FUN_0057b230, table 0x732980)
- [x] 1.2 Own generator and harness; real d2d builds before / after; residuals classified
- [x] 1.3 Sweep d2d's other cold-chill paths
- [x] 1.4 CI-class checks, PR size, fast-forward and merge checks
- [x] 1.5 Verifier section, queue row, BACKLOG line
