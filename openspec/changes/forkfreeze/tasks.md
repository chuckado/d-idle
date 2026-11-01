# Tasks

```yaml
change: forkfreeze
status: in-progress
agent: forkfreeze
branch: agent/forkfreeze
worktree: ~/code/d2lod-wt/forkfreeze
skip_specs: true
owns:
  - docs/forkfreeze.md
  - docs/research/pr-freezeresume-draft.md
  - openspec/changes/forkfreeze/**
# fork-branch:lab-freezeresume and d2d-lab-binary bin/d2d.lab-freezeresume are taken by this card;
# left empty because no other open card lists them (nexthitfix/forkarmor precedent).
exclusive: []
gates:
  - "changecheck status -write then -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "private fork carries the freeze-end resume fix (lab-freezeresume); verified by four differentials on the fork build; live readback on one hidden slot (monster mode, no game.exe)"
live: "verified live on one hidden slot with a --no-save throwaway Sorceress (no game.exe)"
```

## 1. Work

- [x] 1.1 fork branch lab-freezeresume from lab-armorhit, claim and tooling cherry-picked (no conflict)
- [x] 1.2 vet: build, author and verifier freeze differentials, both armor differentials, hit-once, live readback
- [x] 1.3 push lab-freezeresume to the private remote
- [x] 1.4 revise the upstream PR body draft per the verifier (held for the user's go)
- [x] 1.5 docs/forkfreeze.md, card, OWNERS
