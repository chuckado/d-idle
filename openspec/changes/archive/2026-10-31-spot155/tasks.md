# Tasks

```yaml
change: spot155
status: archived
agent: spot155
branch: agent/spot155
merge: 64349e9
worktree: ~/code/d2lod-wt/spot155
skip_specs: true
owns:
  - docs/research/pr-static-evidence-pack.md
  - docs/research/pr-static-proof.md
exclusive:
  - upstream-pr:static-field
gates:
  - "full-path differential: after 0 differ, before > 0 (real d2d builds)"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; check.sh --hygiene-only no FAIL; openspec validate --all --strict"
result: "verified offline on the worker host (real d2d builds against game.exe's code under the emulator); no live game used"
```

## 1. Work

- [x] 1.1 reviewer mode drives Fight::spot (World, two monsters, per-case skill row, difficulty floor, pierce)
- [x] 1.2 oracle prints the same columns; 32-bit seed handling kept
- [x] 1.3 619,328 cases before / after
- [x] 1.4 CI-class checks, one fixup commit, fast-forward push, one PR comment
- [x] 1.5 docs, BACKLOG
