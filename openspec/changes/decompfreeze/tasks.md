# Tasks

```yaml
change: decompfreeze
status: in-progress
agent: decompfreeze
branch: agent/decompfreeze
worktree: ~/code/d2lod-wt/decompfreeze
skip_specs: true
owns:
  - docs/research/freeze-rules-trace.md
  - docs/research/pr-freeze-proof.md
  - docs/research/pr-freeze-draft.md
  - tools/emu/freeze_rules.py
  - tools/emu/freeze_rules_test.cpp
  - docs/decomp-queue.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/decompfreeze/**
exclusive:
  - fork-branch:fix/freeze-rules
gates:
  - "differential on real d2d builds: before > 0, after 0 per claim; pr_guard 0/0; presubmit --quick 0 FAIL; upstream tests rc 0"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: "coordinator merges agent/decompfreeze"
result: "pending"
```

## 1. Work

- [ ] 1.1 Re-derive the freeze rules from the disassembly; decide C3 / C4 / C5
- [ ] 1.2 Read d2d's freeze paths and list the deviations
- [ ] 1.3 Differential oracle (game.exe natively vs real d2d builds), before > 0, after 0 per claim
- [ ] 1.4 Claim commits and tooling commit on fix/freeze-rules (numeric-ID noreply identity)
- [ ] 1.5 CI-class checks (clang-tidy names / includes, gcc sweep, ruff, pr_guard per commit, presubmit, four test binaries, file length, PR size)
- [ ] 1.6 Docs: trace, proof, draft, queue rows, BACKLOG line

## 2. Open

- [ ] 2.1 Independent verifier (a second agent)
- [ ] 2.2 Det state-hash run (needs a d2d slot)
