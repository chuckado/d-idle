# Tasks

```yaml
change: missilefix3
status: in-progress
agent: missilefix3
branch: agent/missilefix3
worktree: ~/code/d2lod-wt/missilefix3
skip_specs: true
owns:
  - docs/research/pr-missilerecord-reverify.md
  - docs/research/pr-missilerecord-draft.md
  - docs/research/pr-missilepoison-draft.md
  - openspec/changes/missilefix3/**
exclusive:
  - fork-branch:local/missilediff-record-up
  - fork-branch:local/missilepoison-skillrows
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "fork: record-up 59d319c8, skillrows f52df8d4 pushed to the private remote; all verifier sets 0 differences, merge 8/8; drafts revised; verified offline on the worker host, not live; nothing upstream"
```

- [x] 1 SrcDamage -1 on the owner poison: decide against game.exe, implement, unit test, re-run suites and the verifier's sets
- [x] 2 PR drafts: the verifier's overclaim list
- [x] 3 reverify doc updated with the third-check section
