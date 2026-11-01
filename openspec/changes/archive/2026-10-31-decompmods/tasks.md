# Tasks

```yaml
change: decompmods
status: archived
agent: decompmods
branch: agent/decompmods
worktree: ~/code/d2lod-wt/decompmods
skip_specs: true
owns:
  - docs/research/monster-mods-trace.md
  - tools/emu/umod_*
  - openspec/changes/decompmods/**
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, I or U with an address or file:line"
  - "tools/emu/umod_stats.py runs game.exe's own inits on >= 100,000 cases against d2d's real rules::boss_stats"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 5a56fbd
result: "game.exe applies the MonUMod inits once at spawn (FUN_005a2120: fixed mods 1..4 then the unit's own, the same inits with flag 0 on its minions). On the same base d2d's rules::boss_stats matches game.exe for champions, uniques and superuniques except ghostly speed (G1, d2d -33 %, game.exe none), berserker level/exp (d2d +0 / x1, game.exe +3 / x5), Cold Enchanted's cold length, two enchants on one unique, the BaseId 0x76 damage halving and stone skin with 2 immunities (157,618 cases, 76,827 differ in some field). Two big deviations: D1 d2d reads boss stats from MonLvl at the boosted level (game.exe: base level, only stat 12 +3 / +2 and exp x5 / x3; population trace 113 of 113 bosses), D2 d2d minions lack +3 level, exp x5 and the boss's minion-flag mods (401 of 401 traced minions). Bot findings B1-B6 (champion damage number, LE not rated, minions of Strong / Fast uniques, AffixResists immunity rule, enchant ids U, an archived AR row wrong). PR candidate (D1 + D2 + G1) not built: no patched d2d build, no det run. Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read AGENTS.md, decomp-queue rules, the PR standard, the risk map, DECOMPMON and earlier traces
- [x] 1.2 game.exe: FUN_005a2120, FUN_005a43e0, FUN_005a4850, FUN_005a48c0, FUN_005a0c00, FUN_005a0930, FUN_00573cb0, the inits at 0x73c008 and the hooks at 0x73c0b8 (V)
- [x] 1.3 d2d: rules/uniques.hpp, ai.cpp make_boss / curse / teleport / multishot / spectral, objgroups.cpp boss tagging, fight/deaths.cpp (V-d2d)
- [x] 1.4 Oracles: umod_spawn.py (6 levels x 12 seeds, 526 boss / minion units), umod_stats.py + harness (157,618 cases)
- [x] 1.5 Bot model: internal/combat affix.go / target.go / resource.go, internal/resource potion.go
- [ ] 1.6 PR candidate on a patched d2d build (not built in the time box; row in pr-bundle-queue.md)
- [x] 1.7 Trace doc, card, OWNERS, decomp-queue row, BACKLOG line, changecheck, leakcheck, openspec validate
