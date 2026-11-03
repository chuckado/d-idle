# Tasks

```yaml
change: msvcc2026
status: in-progress
agent: msvcc2026
branch: agent/msvcc2026
worktree: ~/code/d2lod-wt/msvcc2026
owns:
  - docs/research/msvcc2026-trace.md
  - tools/emu/msvcc_len.py
  - tools/emu/msvcc_split.py
  - tools/emu/msvcc_run.sh
  - openspec/changes/msvcc2026/**
skip_specs: true
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
merge:
result: "Docs and emulator-host scripts only. Literal sizes and the split proof are V run (worker host, g++ probe); MSVC itself and upstream CI status are NOT verified here (no Windows access, gh denied)."
```

## 1. Work

- [x] 1.1 Is the target built on the upstream Windows job (ci.yml, CMake): yes, read.
- [x] 1.2 Measure the literals (msvcc_len.py): 4 raw literals over 16380 bytes in 2 files.
- [x] 1.3 Minimal fix and byte-identity proof (msvcc_split.py, msvcc_run.sh): 4 constants, 0 mismatches.
- [ ] 1.4 Branch local/batch2-msvc2026 and MSVC build through the Windows worker: NOT done, not possible from the headless runner (see trace, "Limits").
