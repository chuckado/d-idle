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
result: "docs/research/stateevrun-trace.md: first-pass claims run under the emulator (V run: 104,158 + 101,386 + 16,000 + 3 x 21,386 cases). Event 11 before resistances holds but fires at queue time; event 3 on a miss holds; event 7 owner and reach mode 1/3 were wrong in the first pass; d2d already models Shiver on a miss; d2d ES absorbs after resistances (deviation, no real-build differential, no PR candidate). Not verified live; no ledger row changed."
live: "docs and emulator scripts only, nothing run live"
```

- [x] harness ser_events.py (event 11 vs resistance, swing resolution, whole chain)
- [x] counted runs on the worker host
- [x] trace doc stateevrun-trace.md
- [x] merge-lines.txt, OWNERS, checks
