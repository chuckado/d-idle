# Tasks

```yaml
change: monstermissile
status: in-progress
agent: monstermissile
branch: agent/monstermissile
worktree: ~/code/d2lod-wt/monstermissile
skip_specs: true
owns:
  - docs/research/monster-missile-trace.md
  - tools/emu/mm_*
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - openspec/changes/monstermissile/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict; coverage.py --check"
result: "Docs and emulator scripts only: docs/research/monster-missile-trace.md (Q1-Q6, D1-D6, B1-B6, no PR candidate), 213 game.exe launches under the emulator (V run on the game side; the d2d side is a read and a port), one READ-ONLY ledger line for 48 missile rows. Nothing run live, no game, no d2d slot."
live: "docs only, nothing run live"
```

- [x] tables dumped, Act 1 / 2 launch rows listed
- [x] game.exe launch oracle (213 launches) and level-bracket probe
- [x] d2d read and port comparison, deviations D1-D6
- [x] bot findings, drill recipe, ledger line, shared-file lines
