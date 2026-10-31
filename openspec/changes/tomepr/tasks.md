# Tasks

```yaml
change: tomepr
status: in-progress
skip_specs: true
agent: tomepr
branch: agent/tomepr
worktree: ~/code/d2lod-wt/tomepr
owns:
  - docs/research/pr-tome-proof.md
  - docs/research/pr-tome-draft.md
  - docs/research/pr-tome-evidence-pack.md
  - docs/decomp-queue.md
  - tools/emu/tome_*
  - openspec/changes/tomepr/**
exclusive:
  - fork-branch:fix/tome-use
gates:
  - "differential against game.exe's own code: mismatches before > 0, after = 0, over >= 100k cases (tools/emu/tome_use.py vs test_game --tome-use)"
  - "upstream test_game, test_combat, test_rules green; pr_guard.py 0 FAIL 0 WARN"
  - "independent adversarial verifier report attached verbatim in the proof doc"
  - "nothing pushed, posted or filed on GitHub; our repo not pushed"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge:
result:
```

## 1. Work

- [ ] 1.1 read the trace, proof, draft, pack, the belt candidate and the standard
- [ ] 1.2 tooling commit: oracle prints per-case lines, `test_game --tome-use` prints the same; before-run
- [ ] 1.3 fix commit: `read_portal` opens at once; old pair closed; charge only on success; upstream unit test
- [ ] 1.4 after-run 0 mismatches; test_game / test_combat / test_rules green; pr_guard 0/0
- [ ] 1.5 independent verifier (Opus), report verbatim in the proof
- [ ] 1.6 docs: proof, draft, evidence pack, decomp queue item 2; DECOMPTOME's open task 1.6 closed here
- [ ] 1.7 changecheck, leakcheck, openspec validate
