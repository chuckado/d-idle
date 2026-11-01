# Tasks

```yaml
change: opennexthit
status: archived
merge: 6bbd958
agent: opennexthit
branch: agent/opennexthit
worktree: ~/code/d2lod-wt/opennexthit
skip_specs: true
owns:
  - openspec/changes/opennexthit/**
  - docs/research/pr-nexthit-posted.md
  - docs/research/pr-bundle-queue.md
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli 0 warnings; go run ./cmd/leakcheck -mode fail 0 hits; tools/ci/check.sh; tools/openspec.sh validate --all --strict passes"
result: "PR #158 opened upstream (head 6b3964fa, fork branch fix/nexthit-justhit); both harnesses 0 on the pushed head; not verified live (not applicable)"
live: "not applicable: upstream PR, no game or slot used"
```

- [x] 1.1 replay check and re-run of every gate on the head
- [x] 1.2 push to the fork, open the PR
- [x] 1.3 posted lines, queue row, Discord feed
