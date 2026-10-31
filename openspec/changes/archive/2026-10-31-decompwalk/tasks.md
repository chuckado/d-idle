# Tasks

```yaml
change: decompwalk
status: archived
agent: decompwalk
branch: agent/decompwalk
worktree: ~/code/d2lod-wt/decompwalk
skip_specs: true
owns:
  - docs/research/walk-speed-trace.md
  - docs/research/pr-walk-*.md
  - tools/emu/walk_*
  - openspec/changes/decompwalk/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
  - docs/research/pr-bundle-queue.md
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, V-d2d, D, S, I or U with an address or file:line"
  - "tools/emu/walk_speed.py runs game.exe's code on >= 100,000 cases: mismatches before > 0, after the candidate 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 84ee1c4
result: "Player walk/run rules equal for the bot's case (walk 6 / run 9 every class, 0.375 / 0.5625 subtiles a tick, FRW 150v/(150+v) for v != 0, velocitypercent raw, >= 25). d2d deviates on monsters (RN moves by MonStats Run instead of Velocity; chill multiplies instead of adding coldeffect; floor 10 instead of 25) and on player velocitypercent (diminished with FRW): 712,362 exhaustive oracle cases, 425,665 differ before, 0 after; local PR candidate fix/walk-velocity (3 commits on upstream d0ea14b7), not posted, no verifier yet. Walking is 68 % of each kill in the post-fix soak at well under the rule speed: bot-side (B1 ResetRun has no caller, so after a fresh game the bot likely walks; B2 run state unlogged). Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read DECOMPREC (movement-and-recovery-trace.md), DECOMPMON, the risk map, d2d-run.md, the PR standard
- [x] 1.2 game.exe: FUN_00623f50, FUN_00621740, FUN_00621360, FUN_006213d0, FUN_006214a0, FUN_00620e80, FUN_00648690, FUN_0057f240, FUN_0057af80; loader descriptors for Velocity / Run / WalkVelocity / RunVelocity; every MonStats Run read (V)
- [x] 1.3 d2d: ai.cpp:2724..2727 and 643, world/tick.cpp:102..137 and 356..361, fight/strike.cpp:185, combat.hpp:170 and 701..722 (V-d2d)
- [x] 1.4 Oracle `tools/emu/walk_speed.py` + `walk_speed_harness.cpp`: 712,362 cases before / after
- [x] 1.5 Soak attribution (`walk_soak_speed.py`, `walk_soak_backoff.py`), read-only on the worker host
- [x] 1.6 PR candidate: local branch, 3 commits, after-run 0; upstream tests and pr_guard (evidence pack)
- [x] 1.7 Independent verifier (deferred: not run; tracked in docs/research/pr-bundle-queue.md)
- [x] 1.8 Trace doc <= 100 lines, card, OWNERS, changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
