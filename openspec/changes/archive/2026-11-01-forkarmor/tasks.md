# Tasks

```yaml
change: forkarmor
status: archived
merge: fb4c078
agent: forkarmor
branch: agent/forkarmor
worktree: ~/code/d2lod-wt/forkarmor
skip_specs: true
owns:
  - docs/forkarmor.md
  - openspec/changes/forkarmor/**
# fork-branch:lab-armorhit and d2d-lab-binary bin/d2d.lab-armorhit are taken by this card;
# left empty because the open card forknext lists the same d2d-lab-binary (nexthitfix precedent).
exclusive: []
gates:
  - "changecheck status -write then -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "private fork carries the Frozen Armor fix (lab-armorhit); verified by differential and one live readback"
live: "verified live on one hidden slot with a --no-save throwaway Sorceress (no game.exe)"
```

## 1. Work

- [x] 1.1 fork branch lab-armorhit from lab-nexthit, claim and tooling cherry-picked
- [x] 1.2 vet: build, both armor differentials and the hit-once differential on the fork build, live readback
- [x] 1.3 push lab-armorhit to the private remote
- [x] 1.4 docs/forkarmor.md, card, OWNERS
