# Tasks

```yaml
change: linuxbuild
status: archived
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
result: "builder container up (limits set, runner registered); lab2-freezeresume build parity with the game host builds verified (armor, hit-once, freeze-resume counts equal); drill leg with a builder-built binary not run"
live: "no game used; the build container ran on the NAS, test binaries ran on the worker host offline"
merge: a8c49d9
```

- [x] build container and runner
- [x] build and parity check of lab2-freezeresume
- [x] wrapper in the private ops repo
- [x] docs
