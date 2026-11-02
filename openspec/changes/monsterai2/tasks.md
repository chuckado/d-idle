# Tasks

```yaml
change: monsterai2
status: in-progress
agent: monsterai2
branch: agent/monsterai2
worktree: ~/code/d2lod-wt/monsterai2
skip_specs: true
owns:
  - docs/research/monsterai2-trace.md
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - tools/emu/mai2_*
  - openspec/changes/monsterai2/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and emulator tools only. V run: six think ports equal game.exe on 36,000 cases; d2d andariel_think 3,830 cases 0 differ; Fallen flee pace (416 of 416), Blood Raven (pace, target home distance), Fallen Shaman corpse reach proven; four mismatch families root-caused; run-speed chain read (base of stat 0x43 still U). No PR candidate."
live: "docs and emulator only, nothing run live"
```

- [x] six untraced AIs: acts run in game.exe, Python port checked against it (36,000 cases, 0 differ)
- [x] Fallen flee (D1), Fallen Shaman corpse reach (D3), keep y (closed), command word (D2), run speed chain (D7, D8)
- [x] residual mismatches root-caused (Fallen Shaman, Gargoyle Trap, Foul Crow Nest, Blood Raven)
- [x] docs/research/monsterai2-trace.md, ledger rows, merge-lines.txt
