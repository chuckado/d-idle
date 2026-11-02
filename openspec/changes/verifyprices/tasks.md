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
result: ""
```

## 1. Work

- [ ] 1.1 Rules and addresses against the disassembly
- [ ] 1.2 Own oracle and d2d probe, base vs head counts
- [ ] 1.3 Readers of the changed code, CI parity, merges
- [ ] 1.4 Draft claims vs counted runs; verdict in docs/research/pr-prices-verifier.md
