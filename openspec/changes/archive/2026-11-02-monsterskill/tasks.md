# Tasks

```yaml
change: monsterskill
status: archived
agent: monsterskill
branch: agent/monsterskill
worktree: ~/code/d2lod-wt/monsterskill
skip_specs: true
owns:
  - docs/research/monsterskill-trace.md
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - tools/emu/msk_*
  - openspec/changes/monsterskill/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and emulator tools only; see the doc for tags and deviations"
live: "docs and emulator only, nothing run live"
merge: 2085bde
```

- [x] trace slot loading, usability gate, shaman / archer thinks in game.exe (V read, addresses)
- [x] V run: monster skill levels (msk_lvl.py), monster mana (msk_mana.py), Corrupt Archer think vs rules::mon_think (msk_run.sh)
- [x] d2d read (monster_skill_level, usable, MonAct::run) and the bot's threat model (threat.go, fight.go, target.go)
- [x] docs/research/monsterskill-trace.md, ledger row, merge-lines.txt
