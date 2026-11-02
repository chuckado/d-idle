# Tasks

```yaml
change: boardclaim
status: archived
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
result: "docs and tooling only; issue body sync applied separately after review; one posted comment is cut off mid-sentence and needs a fix"
live: "no game"
merge: d33de75
```

- [x] audit and classify
- [x] comments, close, narrow open-work.md
- [x] claim.py with tests (claim, done, watch)
- [x] rules in chunker.md and agent-work.md; groom covered flag
