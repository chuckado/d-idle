# Tasks

```yaml
change: mercfix
status: in-progress
agent: mercfix
branch: agent/mercfix
worktree: d2lod-wt/mercfix
skip_specs: true
owns:
  - cmd/d2bot/mercwire.go
  - cmd/d2bot/mercwire_test.go
  - cmd/d2bot/d2dwire.go
  - docs/botlogs.md
  - openspec/changes/mercfix/**
exclusive: []
gates:
  - "go test -p 1 ./cmd/d2bot ./internal/merc ./internal/d2dmerc ./internal/task; changecheck 0 warnings; leakcheck 0 hits"
result: "Guard reads the dump merc also with the policy off; the claim is read back (claim verified / not seen after N tries, bounded retry); offline tests only, NOT verified live (a probe on a d2d slot: claim, walk out of town, read merc.present and the unit table)"
live: "not run"
credit:
  component: mercfix
  author: owner
  agents: [Claude Code]
  verified_by: ["Claude Code (card mercfix, own reading of the Gate 2 log)"]
  consent: none
```

- [x] Reconstruct the claim timeline from the Gate 2 bot log
- [x] Find why the claim and the guard disagree (hook nil with the policy off)
- [x] guardMercHook and claimDep read-back with tests
- [x] docs/botlogs.md line, card, shared-file lines
