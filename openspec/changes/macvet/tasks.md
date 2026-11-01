# Tasks

```yaml
change: macvet
status: in-progress
agent: macvet
branch: agent/macvet
worktree: ~/code/d2lod-wt/macvet
skip_specs: true
owns:
  - openspec/changes/macvet/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Second-machine vetting built and run on three branches (not the live game). Parity on the stored differentials: armor 0 of 108,506, freeze-resume 0 of 12,548 / 12,516, armor verifier 0 of 134,400, curse adj 0 of 147,628 (dealt 1,958 unclaimed), hit-once 0 of 100,198 after one parity bug (fused multiply-add on arm64, fixed with -ffp-contract=off). Private ops repo holds the wrapper and doc."
live: "not verified live: no game run; offline differentials on stored data and game.exe under the emulator"
```

- [x] toolchain on the second machine, checkout, d2d test builds
- [x] differentials against the stored data, compared with the worker host
- [x] ruff, clang-tidy, merge check on the two PR heads
- [x] wrapper and doc in the private ops repo
- [x] public section in docs/review-strategy.md, card, OWNERS
