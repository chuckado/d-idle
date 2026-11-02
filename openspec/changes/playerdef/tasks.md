# Tasks

```yaml
change: playerdef
status: in-progress
agent: playerdef
branch: agent/playerdef
worktree: ~/code/d2lod-wt/playerdef
skip_specs: true
owns:
  - docs/research/playerdef-trace.md
  - tools/emu/pd_dis.sh
  - tools/emu/pd_sh.sh
  - openspec/changes/playerdef/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Trace written from upstream traces and source reads; 8 d2d deviations (none run), 4 bot findings (B1 double resist penalty on d2d, B2 attribute bonus list shape); no PR candidate, no ledger row."
live: "not verified; no game.exe run, no emulator in this session; recipes R1..R7 and differentials in the trace"
```

- [x] read upstream traces, d2d source, bot source
- [x] write docs/research/playerdef-trace.md
- [x] merge-lines.txt, OWNERS
- [ ] differentials of section 6 (needs the emulator; follow-up)
