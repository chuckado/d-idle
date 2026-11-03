# Tasks

```yaml
change: gatetriage
status: in-progress
agent: gatetriage
branch: agent/gatetriage
worktree: d2lod-wt/gatetriage
skip_specs: true
owns:
  - tools/baseline/gate_triage.py
  - tools/baseline/tests/test_gate_triage.py
  - docs/research/gate-triage-examples.md
  - openspec/changes/gatetriage/**
exclusive: []
gates: python3 -I tools/baseline/tests/test_gate_triage.py; changecheck 0 new warnings; leakcheck 0 hits
result: "offline tool over recorded minute archives; outputs checked against the hand findings of 2026-11-02/03; not verified live (no game, no host load)"
credit:
  author: owner
  agents: [Claude Code]
  consent: none
```

- [x] 1. gate_triage.py: verdicts per current rules, windows, XP by band and area, log counts and context, cause hints, exit code
- [x] 2. Fixture tests (poverty clvl 2, funded 332-366, stash 1418 with pack 6, Cold Plains farming)
- [x] 3. Outputs for the five Gate 1 runs in docs/research/gate-triage-examples.md
- [x] 4. gate-run.sh one-line wiring documented (ops repo change is a separate task)
