# Tasks

```yaml
change: forkrebase
status: in-progress
agent: forkrebase
branch: agent/forkrebase
worktree: ~/code/d2lod-wt/forkrebase
skip_specs: true
owns:
  - docs/forkrebase.md
  - openspec/changes/forkrebase/**
# fork-branch:lab2-debug-skills, lab2-armorhit, lab2-freezeresume and d2d-lab-binary bin/d2d.lab2-freezeresume are taken by this card;
# left empty because no other open card lists them (nexthitfix/forkarmor precedent).
exclusive: []
gates:
  - "changecheck status -write then -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "fork stack rebuilt on upstream main 52362d52 (lab2-*), pushed; differentials 0 on the claimed classes; drill T Stony Tomb 254 walking / 211 Teleport, D1.P1s passes on upstream auto-belt (verified live on one hidden slot, no game.exe)"
live: "verified live on one hidden slot with a --no-save throwaway Sorceress"
```

## 1. Work

- [x] 1.1 new branches on upstream main; 3a629371 resolved (upstream belt code, ctrl-click sell kept); hit-once commits dropped as merged (#158)
- [x] 1.2 vet: build, hit-once, armor (author, verifier), freeze-resume (author, verifier), unit tests
- [x] 1.3 live: drill group T legs and D1.P1s on the new binary
- [x] 1.4 bin/d2d.lab2-freezeresume; push the three new branches to the private remote
- [x] 1.5 docs/forkrebase.md, card, OWNERS
