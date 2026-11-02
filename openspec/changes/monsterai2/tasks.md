# Tasks

```yaml
change: monsterai2
status: in-progress
agent: monsterai2
branch: agent/monsterai2
worktree: ~/code/d2lod-wt/monsterai2
skip_specs: true
owns:
  - docs/research/monsterai2-trace.md
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - tools/emu/mai2_*
  - openspec/changes/monsterai2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: ""
live: "docs and emulator only, nothing run live"
```

- [ ] six untraced AIs: acts run in game.exe, Python port checked against it
- [ ] Fallen flee (D11), Fallen Shaman corpse reach (D9), keep y (D7), command word (D8), run base speed (D10)
- [ ] residual mismatches root-caused (Fallen Shaman, Gargoyle Trap, Foul Crow Nest, Blood Raven)
- [ ] docs/research/monsterai2-trace.md, ledger rows, merge-lines.txt
