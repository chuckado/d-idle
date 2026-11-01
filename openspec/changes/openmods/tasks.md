# Tasks

```yaml
change: openmods
status: in-progress
agent: openmods
branch: agent/openmods
worktree: ~/code/d2lod-wt/openmods
skip_specs: true
owns:
  - openspec/changes/openmods/**
  - docs/research/pr-mods-evidence-addendum.md
  - docs/research/pr-bundle-queue.md
  - CHANGELOG.md
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli 0 warnings; go run ./cmd/leakcheck -mode fail 0 hits; tools/check.sh --hygiene-only no FAIL; tools/openspec.sh validate --all --strict passes"
result: "PR #157 opened upstream (head c048a356, fix/boss-stats); gates and differential re-run, counts equal the author's; not verified live (not applicable)"
live: "not applicable: upstream PR, no game or slot used"
```

- [x] 1.1 replay check and re-run of every gate on the head
- [x] 1.2 push to the fork, open the PR
- [x] 1.3 posted lines, queue row, Discord feed
