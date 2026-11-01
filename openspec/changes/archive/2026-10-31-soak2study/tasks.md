# Tasks

```yaml
change: soak2study
status: archived
agent: soak2study
branch: agent/soak2study
worktree: ~/code/d2lod-wt/soak2study
skip_specs: true
owns:
  - docs/probes/postfix-soak2-analysis.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; ci check --hygiene-only no FAIL"
merge: merged into main (see git log, Change: soak2study)
result: "analysis only, not verified live (offline from the full soak logs); corrections added to postfix-soak2.md"
```

## 1. Work

- [x] 1.1 F4 root cause (code read plus offline repro on the worker host, test not committed)
- [x] 1.2 walking share breakdown against soak 1 (full logs of both soaks)
- [x] 1.3 F1, belt window, F2/F3, F5, F6 root causes
- [x] 1.4 ranked fix-card list, BACKLOG and INDEX lines
