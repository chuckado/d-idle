# Tasks

```yaml
change: decompwp
status: archived
agent: decompwp
branch: agent/decompwp
merge: 10b4f51
worktree: ~/code/d2lod-wt/decompwp
skip_specs: true
owns:
  - docs/research/waypoint-trace.md
  - docs/research/pr-wp-*.md
  - tools/emu/wp_*.py
  - openspec/changes/decompwp/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, I or U with an address or file:line"
  - "tools/emu/wp_refuse.py runs from upstream's tools/emu over >= 100,000 cases: mismatches before > 0, after the patch 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "game.exe: rows in Levels.txt Waypoint order (9/9/9/3/9, the bot's table matches), panel on the current act's tab, one click travels at once (same act: warp inside the handler), arrival at the target waypoint +3 subtiles, portal pair untouched. The 10 s refusal is real but counts from the player's last hostile declaration (pcdata+0x160, only writer FUN_0055b720), never in single player; d2d counts it from every arrival: proven deviation (132,964 oracle cases, 59,309 differ before, 0 after), local PR candidate fix/waypoint-hostile-window. Soak: 295 nav waypoint failed = 292 unit not found (d2dsrc files the waypoint with class -1: finding W1) + 3 no path; 0 row, refusal or timing failures; 0 successes. Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read P23, the decomp queue, upstream waypoint.md / act1-end.md / objects.md, the PR standard
- [x] 1.2 game.exe: FUN_0049cf90, FUN_0049c490, FUN_0049c510, FUN_0054d750, FUN_0054c5d0, FUN_0055b6c0, FUN_0055b720 (and its one caller FUN_005a5e50), FUN_00549570, FUN_00548ef0, FUN_00584f60, FUN_0053aec0, FUN_00584870 (V); Levels.txt Waypoint column (D)
- [x] 1.3 d2d: commands.cpp:267, setup.cpp:363, panels.cpp:930-950, walk.cpp:450 (V-d2d)
- [x] 1.4 Oracle `tools/emu/wp_refuse.py`: 132,964 cases before / after
- [x] 1.5 Soak lines classified (`tools/emu/wp_soaklines.py`), read-only on the worker host
- [x] 1.6 PR candidate: local branch, patch, after-run 0; upstream tests, pr_guard, presubmit (see evidence pack)
- [x] 1.7 Independent verifier report (next step before any post) (deferred: tracked in docs/research/pr-bundle-queue.md)
- [x] 1.8 Trace doc <= 100 lines, card, OWNERS, changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
