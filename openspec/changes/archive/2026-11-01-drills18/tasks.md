# Tasks

```yaml
change: drills18
status: archived
agent: drills18
branch: agent/drills18
worktree: ~/code/d2lod-wt/drills18
skip_specs: true
owns:
  - openspec/changes/drills18/**
exclusive:
  - instance:d2d-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
merge: db32c75
result: "Verified live on hidden d2d slot 0: fixmouth3 live proof PARTIAL (merged as a strict improvement; fixmouth* stay open, fixmouth4 next) (DW.m3 x3 identical: 1 mouth bounce plus 1 pocket-stall town portal escape per run; DW.m4 0 bounces); D2 11/11, D3 12/12 PASS; see the DRILLS18 section of docs/probes/drills/README.md"
live: "verified live: DW.m3 x3 FAIL (identical), DW.m4 PASS on bounces (dirty UI after), D2 PASS, D3 PASS"
```

- [x] DW.m3 x3, DW.m4 on the fixmouth3 build
- [x] D3 town pass, D2 regression
- [x] README section, card live lines
