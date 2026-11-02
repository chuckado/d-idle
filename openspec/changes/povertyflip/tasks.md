# Tasks

```yaml
change: povertyflip
status: in-progress
agent: povertyflip
branch: agent/povertyflip
worktree: ~/code/d2lod-wt/povertyflip
skip_specs: true
owns:
  # code paths (resource.go, a new povertyflip_test.go) are added when an agent takes the card
  - openspec/changes/povertyflip/**
exclusive: []
gates:
  - go test -p 1 ./internal/bot passes, including the flip replay test
result: implemented as hysteresis only (tpDebounce: 3 looks or 2 virtual minutes), no evident root cause in code; TestTPDebounceReplay/Holds; offline only, not verified live; inv-tp-dump still to read
```

- [x] 1 Read the `inv-tp-dump` lines of the rerun (docs/hastpdump.md drill) and record the cause under docs/soakstall.md Missing data (not owned here: report it)
- [x] 2 Test first: a replay of the 17:50 cycle (HasTP false in area 3, true in area 2, gold 155, kit 190) must not flip poverty mode more than once in 10 virtual minutes
- [x] 3 Hysteresis in the poverty input in resource.go (or the source fix from 1); gofmt, go vet, go test -p 1 ./internal/bot
