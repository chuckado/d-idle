# Tasks

```yaml
change: baselinefree
status: archived
agent: baselinefree
branch: agent/baselinefree
worktree: ~/code/d2lod-wt/baselinefree
owns:
  - tools/d2dsoak/baselines.json
  - tools/d2dsoak/baselines-early.json
  - cmd/d2dsoak/**
  - internal/d2dsoakstat/bands.go
  - internal/d2dsoakstat/bands_test.go
  - internal/d2dsoakstat/compare.go
  - internal/d2dsoakstat/gatev2.go
  - docs/research/baseline-audit.md
  - openspec/changes/baselinefree/**
exclusive: []
gates:
  - "baselines.json holds the whole a14 run, g2-8h and provenance on every row; baselines-early.json holds the Fresha chain; every number reproducible with d2dsoak baseline from the raw logs"
  - "d2dsoak bands prints per-band XP/h and a verdict; go test -p 1 ./cmd/d2dsoak ./internal/d2dsoakstat passes"
  - "go run ./cmd/changecheck -cli 0 warnings; go run ./cmd/leakcheck -mode fail 0 hits; tools/ci/check.sh; tools/openspec.sh validate --all --strict"
merge: 9aa58ed
result: "offline from raw logs; only band 11-14 can give a hard verdict (n=4); 2 more fresh runs would harden bands 1-10"
```

## 1. Baselines

- [x] 1.1 Locate the raw logs (worker host soak24 and soak24-i5 runs) and recompute with d2dsoak baseline; verified by the bands/extract numbers in the audit After section
- [x] 1.2 Add a14-full, g2-8h, fresha-1-11-chain, provenance, band tables; verified by TestCommittedBaselineBands

## 2. Gate 1 XP per band

- [x] 2.1 Delta spec soak-gates (bands, sample rule, tolerance); verified by tools/openspec.sh validate --all --strict
- [x] 2.2 d2dsoak bands + baseline chains; verified by go test -p 1 ./cmd/d2dsoak ./internal/d2dsoakstat

## 3. Audit

- [x] 3.1 Append the After section to docs/research/baseline-audit.md
- [ ] 3.2 Follow-up card (not done here): band table in compare and run, ops gate-run record
