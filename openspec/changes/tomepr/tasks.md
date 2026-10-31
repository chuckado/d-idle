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
result: "PROOF complete, NOT verified live (oracle + d2d test binary only). Local upstream branch fix/tome-use on 03f42254: 29e75c76 tooling, c5e2a0b1 claim A (item in hand refuses), e35ab5ca claim B (opens on the click tick in any live mode, old pair closed first, charge only on success); 108,000 cases 29,551 -> 23,127 -> 0, seed 31 0; test_game/test_combat/test_rules green, pr_guard 0/0; Opus verifier holds with limits. Closes DECOMPTOME task 1.6 (archived card 2026-10-31-decomptome). Nothing pushed or posted."
```

## 1. Work

- [x] 1.1 read the trace, proof, draft, pack, the belt candidate and the standard
- [x] 1.2 tooling commit: oracle prints per-case lines, `test_game --tome-use` prints the same; before-run
- [x] 1.3 fix commit: `read_portal` opens at once; old pair closed; charge only on success; upstream unit test
- [x] 1.4 after-run 0 mismatches; test_game / test_combat / test_rules green; pr_guard 0/0
- [x] 1.5 independent verifier (Opus), report verbatim in the proof
- [x] 1.6 docs: proof, draft, evidence pack, decomp queue item 2; DECOMPTOME's open task 1.6 closed here
- [x] 1.7 changecheck, leakcheck, openspec validate
