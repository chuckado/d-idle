# Tasks

```yaml
change: stateevrun
status: in-progress
agent: stateevrun
branch: agent/stateevrun
worktree: ~/code/d2lod-wt/stateevrun
skip_specs: true
owns:
  - docs/research/stateevrun-trace.md
  - tools/emu/ser_events.py
  - tools/emu/ser_run.sh
  - openspec/changes/stateevrun/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "pending"
live: "docs and emulator scripts only, nothing run live"
```

- [ ] harness ser_events.py (event 11 vs resistance, swing resolution, whole chain)
- [ ] counted runs on the worker host
- [ ] trace doc stateevrun-trace.md
- [ ] merge-lines.txt, OWNERS, checks
