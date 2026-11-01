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
result: "pending"
live: "no game used; the build container ran on the NAS, test binaries ran on the worker host offline"
```

- [ ] build container and runner
- [ ] build and parity check of lab2-freezeresume
- [ ] wrapper in the private ops repo
- [ ] docs
