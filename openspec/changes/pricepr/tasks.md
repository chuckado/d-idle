# Tasks

```yaml
change: pricepr
status: in-progress
skip_specs: true
agent: pricepr
branch: agent/pricepr
worktree: ~/code/d2lod-wt/pricepr
owns:
  - docs/research/pr-price-proof.md
  - docs/research/pr-price-draft.md
  - docs/research/pr-price-evidence-pack.md
  - docs/research/price-trace.md
  - docs/research/pr-bundle-queue.md
  - docs/decomp-queue.md
  - tools/emu/price_*
  - openspec/changes/pricepr/**
exclusive:
  - fork-branch:fix/vendor-prices
gates:
  - "differential against game.exe's own code per claim: mismatches before > 0, after = 0 for the claim's term, over >= 100k cases (tools/emu/price_oracle.py vs test_game --price)"
  - "upstream test_rules, test_combat, test_game, test_protocol green; pr_guard.py 0 FAIL 0 WARN; presubmit --quick 0 FAIL"
  - "independent adversarial verifier report attached verbatim in the proof doc"
  - "nothing pushed, posted or filed on GitHub; our repo not pushed"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: pending (agent/pricepr)
result: "pending"
```

## 1. Work

- [x] 1.1 read the trace, proof, oracle, model, terms, the belt and tome packs, the standard and the queue
- [x] 1.2 tooling commit on a new worktree off upstream/main (oracle, model, terms, `test_game --price`, README row); baseline reproduced (d2d md5 of the 100k run equal to DECOMPREPAIR's)
- [x] 1.3 claim commits D1, D2, D3, D4, D7 with per-claim before/after and upstream unit tests; D5, D6, D8, D9, D10 left out with ponytails
- [x] 1.4 after-run: clean subset 0 mismatches; upstream tests; pr_guard 0/0; presubmit --quick 0 FAIL
- [ ] 1.5 independent verifier (Opus), report verbatim in the proof
- [ ] 1.6 docs: proof, draft, evidence pack, decomp queue item 5, bundle queue row, price-trace PR state
- [ ] 1.7 changecheck, leakcheck, openspec validate
