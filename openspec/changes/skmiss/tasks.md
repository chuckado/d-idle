# Tasks

```yaml
change: skmiss
status: in-progress
agent: skmiss
branch: agent/skmiss
worktree: ~/code/d2lod-wt/skmiss
skip_specs: true
owns:
  - docs/research/sk-missile-trace.md
  - tools/emu/sk_missile_hit.py
  - tools/emu/sk_missile_hit_test.cpp
  - docs/decomp-queue.md
  - docs/BACKLOG.md
  - docs/INDEX.md
  - openspec/changes/skmiss/**
exclusive:
  - fork-branch:fix/missile-roll-top
gates:
  - "differential on real d2d builds: before > 0, after 0 for any claim; pr_guard 0/0; presubmit --quick 0 FAIL; upstream tests rc 0"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: ""
```

## 1. Work

- [ ] 1.1 Trace spawn, mana payment, flight and hit in game.exe 1.14d (tags)
- [ ] 1.2 Read d2d's missile creation, flight and damage; the bot's casting model
- [ ] 1.3 Differential oracle (game.exe natively vs the real d2d build), before and after
- [ ] 1.4 Local upstream candidate (only if a deviation is proven on the real build)
- [ ] 1.5 Docs: trace (deviations table, drill recipe), queue row 14, BACKLOG, INDEX

## 2. Open

- [ ] 2.1 Independent verifier (a second agent)
- [ ] 2.2 Det state-hash run (needs a d2d slot)
