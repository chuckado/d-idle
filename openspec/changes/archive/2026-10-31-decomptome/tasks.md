# Tasks

```yaml
change: decomptome
status: archived
agent: decomptome
branch: agent/decomptome
worktree: ~/code/d2lod-wt/decomptome
owns:
  - docs/research/town-portal-trace.md
  - docs/research/pr-tome-*.md
  - tools/emu/tome_*.py
  - openspec/changes/decomptome/**
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, U or I with an address or file:line"
  - "tools/emu/tome_use.py runs from upstream's tools/emu over >= 100,000 cases and the empty-tome class has 0 differ"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: eae3b82
result: "game.exe: an empty tome is a silent no-op (same in d2d, 5,743 oracle cases 0 differ); a pack TP opens at the click with no cast in any live mode (refused dead, state 54, cursor, town); one pair per player (old closed first). d2d deviates: SC cast, use dropped while pmode >= 0, charge spent at the click (108,000 cases, 22,395 differ). Soak lines: 778 no-portal-opened = 525 empty tome, 134 dead context, 119 charged (busy drop or cut cast, I); not timing. PR candidate not ready (no patch). Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read primitives P10-P14, primitives-review F4/F5, research README, PR standard and spec
- [x] 1.2 game.exe: FUN_0054d750, FUN_0057eec0, FUN_0054b1e0, FUN_0055e170, FUN_005bf240, FUN_005be290, FUN_00535430, FUN_0056d130, FUN_005bf3d0, FUN_0055e050 (disasm.py, V); misc / books / skills / objects / states rows (D)
- [x] 1.3 d2d: read_portal, cast_scroll, open_portal_at, portal_spot, use_portal, the SC action frame, the GH override (V-d2d, file:line)
- [x] 1.4 Oracle `tools/emu/tome_use.py`: 108,000 cases, classes in the proof doc
- [x] 1.5 Soak lines classified (`tools/emu/tome_soaklines.py`), read-only streaming on the worker host
- [ ] 1.6 PR candidate: d2d patch, test mode, after-run 0, verifier (next card)
- [x] 1.7 Trace doc <= 100 lines, card, OWNERS, changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
