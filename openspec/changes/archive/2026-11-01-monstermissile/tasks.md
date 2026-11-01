# Tasks

```yaml
change: monstermissile
status: archived
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
result: "docs only; game side V run, d2d side read plus port (not a real-build differential); no PR candidate yet"
live: "docs only, nothing run live"
merge: 38f6e36
```

- [x] tables dumped, Act 1 / 2 launch rows listed
- [x] game.exe launch oracle (213 launches) and level-bracket probe
- [x] d2d read and port comparison, deviations D1-D6
- [x] bot findings, drill recipe, ledger line, shared-file lines
