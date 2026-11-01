# Tasks

```yaml
change: nexthitfix
status: archived
merge: d8e48e0
agent: nexthitfix
branch: agent/nexthitfix
worktree: ~/code/d2lod-wt/nexthitfix
skip_specs: true
owns:
  - openspec/changes/nexthitfix/**
  - docs/research/pr-nexthit-v2-addendum.md
  - docs/research/pr-bundle-queue.md
exclusive: []
gates:
  - "both harnesses 0 differences on the v2 build; det hashes explained; changecheck 0 warnings; leakcheck -mode fail 0"
result: "archived at merge d8e48e0; offline complete, not a live game proof; READY: fix/nexthit-v2 (head 6b3964fa, base 72132b99): justhit ends after the missiles of frame + NextDelay (<=), oracle expires after the contacts; author's harness 0 of 100,198 (old head 8,473 under the corrected oracle), verifier's 0 of 100,624 (old head 25,252); det hashes 9 of 14 identical, 5 explained; fork lab-nexthit 2c6c4a1a651c pushed to the private remote, bin/d2d.lab-nexthit2, readback equal to the old build; offline plus one hidden d2d slot for det hashes and one for the readback, no game.exe window"
# resources held de facto: fork-branch:lab-nexthit and d2d-lab-binary (bin/d2d.lab-nexthit2); not listed in exclusive because the merged but unarchived forknext card still lists them (changecheck overlap)
live: "d2d only: det hashes (one hidden slot) and a probe readback on a --no-save throwaway Sorceress (one hidden slot), no game.exe"
```

## 1. Work

- [x] 1.1 fix/nexthit-v2 in a new worktree (pr/nexthit2): `<=`, ponytail note, oracle after contacts, skills.md line; claim and tooling commits amended in the new branch
- [x] 1.2 re-run the author's harness and the verifier's harness on the v2 build (0 and 0), old head for contrast
- [x] 1.3 det state hashes base vs v2 on one hidden slot, differences classified
- [x] 1.4 CI-class checks on the v2 head
- [x] 1.5 fork: lab-nexthit new commit, bin/d2d.lab-nexthit2, differential 0, live readback, pushed to the private remote
- [x] 1.6 addendum, status banner on the old proof, queue line, card, OWNERS
