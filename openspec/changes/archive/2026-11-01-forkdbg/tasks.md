# Tasks

```yaml
change: forkdbg
status: archived
agent: forkdbg
branch: agent/forkdbg
worktree: ~/code/d2lod-wt/forkdbg
merge: 1842cce
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
result: "verbs proven once live on a hidden slot (each before/after through dump and monsters; docs/forkdbg.md); the M1-M5 drills not run yet; fork commit d9b96c2b on the private remote branch lab-debug-skills"
```

## 1. Work

- [x] 1.1 fork: debug mana full, states set|clear, mon set, cast on branch lab-debug-skills; pushed to the private remote
- [x] 1.2 internal/d2dctl: SkillDebug client with ledger-first calls, tests against a fake socket
- [x] 1.3 live proof of each verb once on one hidden slot (before/after through dump), cleanup
- [x] 1.4 docs/forkdbg.md with the M1-M5 recipe and the fork commit hashes
