# Tasks

```yaml
change: sorcrest
status: archived
agent: sorcrest
branch: agent/sorcrest
worktree: ~/code/d2lod-wt/sorcrest
skip_specs: true
owns:
  - docs/research/sor-rest-trace.md
  - tools/emu/sk_rest_*
  - tools/emu/sorcrest_*
  - openspec/changes/sorcrest/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and emulator tools only; D1 is a restructure of missile_blow, one claim with SKGROUND D5 plus an independent verifier before any PR"
live: "docs only, nothing run live"
merge: fa03b72
```

- [x] rank the untested rows from the ledger
- [x] Fire Ball and Glacial Spike: hit pipeline, burst radius and roll on game.exe and the real d2d build
- [x] calc differential for the 30 Sorceress rows
- [x] Inferno, Blaze, Enchant read in game.exe, d2d compared by source
- [x] trace doc, ledger lines, matrix cell text, drill recipe, merge lines
