# Tasks

```yaml
change: beltfix
status: in-progress
agent: beltfix
branch: agent/beltfix
worktree: ~/code/d2lod-wt/beltfix
skip_specs: true
owns:
  - docs/research/beltfix-analysis.md
  - internal/bot/kitgate.go
  - internal/bot/goal.go
  - internal/bot/beltfix_test.go
  - openspec/changes/beltfix/**
exclusive: []
gates:
  - "gofmt; go vet and go test -p 1 of internal/bot, resource, town; whole repo on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Analysis of the Gate 1 BELT window (vmin 446 to 457, real, not an artefact) and a bot fix (kit trip priced at one potion when potions are missing, emergency cooldown with no potion); unit tests replay the window state. Not verified live."
live: "not run live; Gate 1 re-run after merge is the check"
```

- [x] find and explain the window from all-minutes.log and bot-0.log
- [x] decide artefact vs defect (defect)
- [x] fix + replay tests
- [x] docs/research/beltfix-analysis.md
