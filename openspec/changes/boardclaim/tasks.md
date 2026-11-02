# Tasks

```yaml
change: boardclaim
status: in-progress
agent: boardclaim
branch: agent/boardclaim
worktree: ~/code/d2lod-wt/boardclaim
skip_specs: true
owns:
  - tools/contrib/claim.py
  - tools/contrib/tests/test_claim.py
  - tools/contrib/chunker.py
  - docs/community/**
  - docs/process/chunker.md
  - docs/contributing/agent-work.md
  - openspec/changes/boardclaim/**
exclusive: []
gates:
  - "contrib tests; board.py --check; changecheck -cli 0 warnings; leakcheck 0; tools/ci/check.sh"
result: "Audit of the 16 open work-item issues done, comments posted and #9 closed on the project's own repo; claim.py and rules written; unit tests only. Not verified live beyond the real gh posts."
live: "no game"
```

- [x] audit and classify
- [x] comments, close, narrow open-work.md
- [x] claim.py with tests (claim, done, watch)
- [x] rules in chunker.md and agent-work.md; groom covered flag
