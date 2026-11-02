# Tasks

```yaml
change: missilefix2
status: archived
agent: missilefix2
branch: agent/missilefix2
worktree: ~/code/d2lod-wt/missilefix2
skip_specs: true
owns:
  - docs/research/pr-missilerecord-draft.md
  - docs/research/pr-missilepoison-draft.md
  - openspec/changes/missilefix2/**
exclusive:
  - fork-branch:local/missilediff-record-up
  - fork-branch:local/missilepoison-skillrows
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs only; fork heads private; narrow third check pending; nothing upstream"
merge: 198d76f
```

- [x] Owner-poison rule in monster_shot_stats + test (fork 1824870b)
- [x] Trap fix through put_missile_element + test (fork 71f98145)
- [x] Re-verifier's differentials re-run on the new heads, his harness unchanged
- [x] Fork suites, pr_guard, presubmit --quick; merge of the two heads + both unit tests
- [x] PR drafts updated (record, poison)
