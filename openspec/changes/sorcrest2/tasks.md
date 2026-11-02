# Tasks

```yaml
change: sorcrest2
status: in-progress
agent: sorcrest2
branch: agent/sorcrest2
worktree: ~/code/d2lod-wt/sorcrest2
skip_specs: true
owns:
  - docs/research/sorcrest2-trace.md
  - tools/emu/sr2_*
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - openspec/changes/sorcrest2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Inferno do 48 cases, Enchant 6, Blaze 3 run on the game side (fake units); d2d burst rim 16,875 cases (549 differ from the integer rule); no PR candidate; docs and emulator scripts only"
live: "docs and emulator tools only; nothing run live"
```

- [ ] D2 rim: d2d side with monsters between subtile centres, game rule as oracle
- [ ] D4 Inferno, Blaze, D5 Enchant: game.exe run where buildable, else dead end written down
- [ ] trace doc, ledger rows proven only, rule-coverage regenerated, merge lines
