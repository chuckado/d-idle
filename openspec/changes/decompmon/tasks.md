# Tasks

```yaml
change: decompmon
status: in-progress
agent: decompmon
branch: agent/decompmon
worktree: ~/code/d2lod-wt/decompmon
owns:
  - docs/research/monster-attack-trace.md
  - docs/archive/monster-attack-trace-2026-10-30.md
  - tools/emu/mon_*.py
  - tools/emu/monster_hit.py
  - openspec/changes/decompmon/**
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, I or U with an address or file:line"
  - "the oracles run from upstream's tools/emu: melee 0 differ on 300,672 cases; the missile block gate and damage roll 0 differ from the reading on 100,000 cases each"
result: "melee clean on upstream 0dc4c62a (to-hit / damage / get-hit 100,000 / 100,000 / 100,672, 0 differ); missile deviations (a) d2d blocks pure-element missiles (10,503 of 100,000 block cases), (b) d2d MissA rolls to-hit without the row's ToHit (94 of 142 rows), none early; monster rules explain ~0 of the XP/h gap; damage-driven exits and trips 20-45 % of it are bot exposure (B1 no missiles in the d2d snapshot, B2 running into packs). PR candidate not built. Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read the baseline, risk map, first-pass trace, mon-wake-trace, per-kill-time-split, gap-casts-per-kill
- [x] 1.2 game.exe: FUN_005adf10, FUN_005ad730, FUN_0057dfb0, FUN_0057dd60, FUN_005a89a0, FUN_005a4f50, FUN_0059fa30, FUN_0059f900 (disasm, V); Missiles / MonStats / Skills (D)
- [x] 1.3 d2d on upstream main 0dc4c62a: monster_blow, get_hit, the MissA launch and hit, andariel_missile / missile_blow (V-d2d)
- [x] 1.4 Oracles: monster_hit.py 300,672 cases 0 differ; mon_missile.py block and roll 100,000 each 0 differ from the reading, tables exhaustive
- [x] 1.5 Q4 activity split from the m2early and Fresha logs (read-only); Q5 bot findings with file:line
- [ ] 1.6 PR candidate for the two missile gates (not built: time box, no early-game effect)
- [x] 1.7 Trace doc <= 100 lines, queue / README / INDEX lines, card, OWNERS, changecheck, leakcheck, openspec validate
