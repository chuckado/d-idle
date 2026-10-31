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
result: "D9 and D10 confirmed in game.exe (V): npc.txt and quest mults divide by 1024 first above 0xffff (0x62faac / 0x62fae5 / 0x62fb1d, quest loop); FUN_0062e660 needs ItemTypes Repair and restock or FUN_00629930 durability unless charges are missing (bows, crossbows, throwing potions). Real d2d builds, 220,000 cases (price_edge.py 120,000 + #152 set 100,000): D9 55,962 / 699 / 15,964 -> 0, D10 repair 20,682 -> 0, clean subset 0 / 0 / 0. Local upstream branch fix/vendor-prices-followon on #152 head 8a45cfc9: c3bab73e tooling, e297025a D9, 312272a7 D10; tests, clang-tidy (our lines), shadow sweep, pr_guard 0/0, presubmit --quick 0 FAIL. Not verified live; no verifier yet; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read the queue rules, the PR standard, the price docs and #152's tooling
- [x] 1.2 game.exe: re-derive D9 (NPC and quest mults) and D10 (FUN_0062e660 / FUN_00629930) from the disassembly
- [x] 1.3 Edge-case generator and per-claim counter; before (upstream rules, #152 head) and after on real d2d builds
- [x] 1.4 Local branch fix/vendor-prices-followon on #152's head: tooling, D9, D10 commits
- [x] 1.5 CI-class checks: clang-tidy, shadow sweep, pr_guard, presubmit, the four test binaries
- [x] 1.6 Proof, draft, evidence pack; decomp-queue rows, bundle-queue row, BACKLOG line, README / INDEX lines
- [x] 1.7 changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
