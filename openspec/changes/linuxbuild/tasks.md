# Tasks

```yaml
change: linuxbuild
status: in-progress
agent: linuxbuild
branch: agent/linuxbuild
worktree: ~/code/d2lod-wt/linuxbuild
skip_specs: true
owns:
  - docs/builder.md
  - openspec/changes/linuxbuild/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Builder container, runner (online, a test workflow ran on it) and the private wrapper done. lab2-freezeresume made: d2d 304 s cold, tests 46 s; the test binaries gave armor 0 of 108,506, hit-once 0 of 100,198, freeze-resume 0 of 12,548 and 0 of 12,516, identical to worker-made outputs. NAS memory low point 7.4 GB. A drill leg with the builder-made d2d was not run. Verified on the real hosts, not live in the game."
live: "no game used; the build container ran on the NAS, test binaries ran on the worker host offline"
```

- [x] build container and runner
- [x] build and parity check of lab2-freezeresume
- [x] wrapper in the private ops repo
- [x] docs
