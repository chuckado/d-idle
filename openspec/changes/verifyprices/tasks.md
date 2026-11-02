# Tasks

```yaml
change: verifyprices
status: in-progress
agent: verifyprices
branch: agent/verifyprices
worktree: d2lod-wt/verifyprices
skip_specs: true
owns:
  - openspec/changes/verifyprices/**
  - docs/research/pr-prices-verifier.md
  - tools/emu/vpr_dis.py
  - tools/emu/vpr_oracle.py
  - tools/emu/vpr_probe0.py
  - tools/emu/vpr_probe.inc
  - tools/emu/vpr_cmp.py
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "docs only, not live; NOT READY as is (conflicts with upstream main, divide-first already upstream as a9548d38), READY WITH CHANGES: repair gate alone, rebased"
```

## 1. Work

- [x] 1.1 Rules and addresses against the disassembly
- [x] 1.2 Own oracle and d2d probe, base vs head counts
- [x] 1.3 Readers of the changed code, CI parity, merges
- [x] 1.4 Draft claims vs counted runs; verdict in docs/research/pr-prices-verifier.md
