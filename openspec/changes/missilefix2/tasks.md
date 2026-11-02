# Tasks

```yaml
change: missilefix2
status: in-progress
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
result: "fork heads pushed to the private remote: local/missilediff-record-up 1824870b (owner poison), local/missilepoison-skillrows 71f98145 (trap fix); the re-verifier's differentials re-run with his harness unchanged: 0 differences (record C / D, owner-poison set, skill set, hits, dot); ctest of 6 suites + unit tests pass, pr_guard 0/0, presubmit --quick 0 FAIL; merge of the two heads clean and both unit tests pass; NOT verified live (no game); a narrow third check is owed"
```

- [x] Owner-poison rule in monster_shot_stats + test (fork 1824870b)
- [x] Trap fix through put_missile_element + test (fork 71f98145)
- [x] Re-verifier's differentials re-run on the new heads, his harness unchanged
- [x] Fork suites, pr_guard, presubmit --quick; merge of the two heads + both unit tests
- [x] PR drafts updated (record, poison)
