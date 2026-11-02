# Tasks

```yaml
change: verifyplayerroll
status: in-progress
agent: verifyplayerroll
branch: agent/verifyplayerroll
worktree: ~/code/d2lod-wt/verifyplayerroll
skip_specs: true
owns:
  - openspec/changes/verifyplayerroll/**
  - docs/research/pr-playerroll-verifier.md
  - docs/research/pr-playerroll-verifier/**
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "pending"
```

## 1. Work

- [ ] 1.1 Diff and clean merge against current upstream main
- [ ] 1.2 Own counts: physical roll and floor, Sorceress and Necromancer element rolls, lo == hi draws
- [ ] 1.3 Caller audit
- [ ] 1.4 Det scenarios base vs head (15, one hidden slot)
- [ ] 1.5 Warnings-as-errors build (g++ and clang++), MSVC hazards by reading
- [ ] 1.6 Suites and ctest base vs head
- [ ] 1.7 Draft claims vs counts; verdict
