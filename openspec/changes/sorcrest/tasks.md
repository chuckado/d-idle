# Tasks

```yaml
change: sorcrest
status: in-progress
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
result: "Fire Ball and Glacial Spike hit pipeline 31,000 cases equal, burst radius 3,375 cases equal, burst roll deviation proven on the real d2d build (0 of 400 trials equal on d2d, 400 of 400 in game.exe), 22,400 calc cases equal; Inferno, Blaze, Enchant read only; no PR candidate; not verified live"
live: "docs only, nothing run live"
```

- [x] rank the untested rows from the ledger
- [x] Fire Ball and Glacial Spike: hit pipeline, burst radius and roll on game.exe and the real d2d build
- [x] calc differential for the 30 Sorceress rows
- [x] Inferno, Blaze, Enchant read in game.exe, d2d compared by source
- [x] trace doc, ledger lines, matrix cell text, drill recipe, merge lines
