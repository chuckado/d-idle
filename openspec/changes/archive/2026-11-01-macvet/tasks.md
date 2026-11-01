# Tasks

```yaml
change: macvet
status: archived
agent: macvet
branch: agent/macvet
worktree: ~/code/d2lod-wt/macvet
skip_specs: true
owns:
  - openspec/changes/macvet/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "second-machine vetting wrapper works; parity with the game host byte-identical on armor, freeze-resume, curse and hit-once (with -ffp-contract=off); one arm64 fused multiply-add parity bug found and fixed in the wrapper; not for Linux binaries or live runs"
live: "not verified live: no game run; offline differentials on stored data and game.exe under the emulator"
merge: 7c4fd4b
```

- [x] toolchain on the second machine, checkout, d2d test builds
- [x] differentials against the stored data, compared with the worker host
- [x] ruff, clang-tidy, merge check on the two PR heads
- [x] wrapper and doc in the private ops repo
- [x] public section in docs/review-strategy.md, card, OWNERS
