# Tasks

```yaml
change: oriontune
status: in-progress
agent: oriontune
branch: agent/oriontune
worktree: ~/code/d2lod-wt/oriontune
skip_specs: true
owns:
  - docs/process/worker-capacity.md
  - docs/process/leakcheck-allow.txt
  - openspec/changes/oriontune/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Measured on the worker host (read-only 15 min); wrapper, ccache, cached build script and the capacity2 check live in the private ops repo; worker-capacity.md note added. No game used."
live: "worker host measurements only, no game"
```

- [x] Measure the load for 15 minutes (CPU, I/O wait, D-state tasks, mounts)
- [x] d2nice wrapper with CPU quota, tested
- [x] ccache and cached fork build script, cold vs warm measured
- [x] capacity2 with one verdict line, bounded, tested offline and live
- [x] Private tuning doc and public worker-capacity.md note
