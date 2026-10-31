# Tasks

```yaml
change: decompprice2
status: in-progress
agent: decompprice2
branch: agent/decompprice2
worktree: ~/code/d2lod-wt/decompprice2
skip_specs: true
owns:
  - docs/research/pr-price2-*.md
  - tools/emu/price_edge.py
  - tools/emu/price_claims.py
  - openspec/changes/decompprice2/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "every rule is tagged V, D, S, I or U with an address"
  - "per claim: >= 100,000 cases on real d2d builds, mismatches before > 0, after 0"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
```

## 1. Work

- [ ] 1.1 Read the queue rules, the PR standard, the price docs and #152's tooling
- [ ] 1.2 game.exe: re-derive D9 (NPC and quest mults) and D10 (FUN_0062e660 / FUN_00629930) from the disassembly
- [ ] 1.3 Edge-case generator and per-claim counter; before (upstream rules, #152 head) and after on real d2d builds
- [ ] 1.4 Local branch fix/vendor-prices-followon on #152's head: tooling, D9, D10 commits
- [ ] 1.5 CI-class checks: clang-tidy, shadow sweep, pr_guard, presubmit, the four test binaries
- [ ] 1.6 Proof, draft, evidence pack; decomp-queue rows, bundle-queue row, BACKLOG line, README / INDEX lines
- [ ] 1.7 changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
