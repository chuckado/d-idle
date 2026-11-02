# Tasks

```yaml
change: skshield
status: in-progress
agent: skshield
branch: agent/skshield
worktree: ~/code/d2lod-wt/skshield
skip_specs: true
owns:
  - docs/research/skshield-trace.md
  - tools/emu/skshield_*
  - openspec/changes/skshield/**
exclusive: []
gates:
  - "equip and hand-toggle run: game.exe's FUN_004c0d20 / FUN_00535d10 / FUN_00627910 on 382 scenarios, 0 differ from the reading; item maker run on 306 weapon and 69 armor rows"
  - "d2d side: SKMELEE2's real-build O part re-compared (cases 1000, seeds 7 and 1)"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "offline, not verified live. Shield never a hand item and attached active; no base damage on a shield (306/306 weapon rows vs 0/69 armor rows); flats and ED reach the player; D7 settled as reading A; D1 off-weapon ED proven by the existing O part, D2 broken item stats buildable; no PR candidate built (no branch possible here)"
```

## 1. Work

- [x] 1.1 Read the hand-item functions, the equip attach and the stat merge in game.exe
- [x] 1.2 Run them natively on 382 scenarios (skshield_run.py) and the item maker on every weapon and armor row (skshield_base.py)
- [x] 1.3 Compare with d2d (fight.cpp, combat.hpp) and re-compare SKMELEE2's O part (skshield_cmp.py)
- [x] 1.4 Bot findings with file:line, PR decision, drill recipe, matrix text, shared-file lines
- [x] 1.5 Checks: changecheck (0 warnings), leakcheck (0 hits), openspec validate --all --strict (57 passed), tools/ci/check.sh ok; gofmt and go test not applicable (no Go changed)
