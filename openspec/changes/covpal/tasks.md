# Tasks

```yaml
change: covpal
status: in-progress
agent: covpal
branch: agent/covpal
worktree: ~/code/d2lod-wt/covpal
skip_specs: true
owns:
  - docs/research/covpal-trace.md
  - tools/emu/covpal.py
  - tools/emu/covpal_build.sh
  - tools/emu/covpal_run.sh
  - tools/emu/covpal_rows.py
  - tools/emu/covpal_tab.py
  - docs/coverage/ledger.tsv
  - docs/coverage/gaps.tsv
  - docs/coverage/rule-coverage.md
  - openspec/changes/covpal/**
exclusive: []
gates:
  - "differential on real d2d builds: 0 mismatches counted per row in the doc; ledger case counts appear in the doc"
  - "python3 tools/coverage/coverage.py --check; changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Offline, nothing verified live. 15 Paladin rows run (game.exe calc machine vs d2d, 14,400 cases per aspect, 0 differ); Redemption callback and Vengeance read; D1 (Redemption one corpse a pulse) and D2 (Vengeance mastery term) read on both sides, no PR candidate; bot B1 element table; ledger rows 100-125 minus the 15 already covered now all tested."
```

## 1. Work

- [x] 1.1 Run the calc / aura / element differential for the 15 rows (two seeds)
- [x] 1.2 Read Redemption (do 82 and callback), the area search, Vengeance and Fist in game.exe
- [x] 1.3 Compare with d2d and the bot, write the trace
- [x] 1.4 Ledger lines, gap 6 removed, rule-coverage.md regenerated

## 2. Open

- [ ] 2.1 Game-side differential for D1 (Redemption) and D2 (Vengeance mastery) as described in the trace
- [ ] 2.2 Decode FUN_006456d0 and the aura filters 0x1102 / 0x12003 / 0xa783
