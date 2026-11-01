# Tasks

```yaml
change: forkdbg
status: in-progress
agent: forkdbg
branch: agent/forkdbg
worktree: ~/code/d2lod-wt/forkdbg
skip_specs: true
owns:
  - internal/d2dctl/debug_skills.go
  - internal/d2dctl/debug_skills_test.go
  - docs/forkdbg.md
exclusive:
  - d2d-lab-binary
  - fork-branch:lab-debug-skills
gates:
  - "gofmt, go vet and go test -p 1 for internal/d2dctl; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh --hygiene-only; openspec validate --all --strict"
result: pending
```

## 1. Work

- [ ] 1.1 fork: debug mana full, states set|clear, mon set, cast on branch lab-debug-skills; pushed to the private remote
- [ ] 1.2 internal/d2dctl: SkillDebug client with ledger-first calls, tests against a fake socket
- [ ] 1.3 live proof of each verb once on one hidden slot (before/after through dump), cleanup
- [ ] 1.4 docs/forkdbg.md with the M1-M5 recipe and the fork commit hashes
