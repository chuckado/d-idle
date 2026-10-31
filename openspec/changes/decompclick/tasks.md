# Tasks

```yaml
change: decompclick
status: in-progress
skip_specs: true
agent: decompclick
branch: agent/decompclick
worktree: ~/code/d2lod-wt/decompclick
owns:
  - docs/research/click-handlers-trace.md
  - docs/research/pr-click-proof.md
  - docs/research/pr-click-draft.md
  - docs/research/pr-click-evidence-pack.md
  - tools/emu/click_*
exclusive:
  - fork-branch:trace/click
gates:
  - "every address cited is read in game.exe 1.14d or run in the oracle (V) or tagged D/S/I/U; no U in a claim's causal chain"
  - "a claimed deviation has a differential against game.exe's own code: mismatches before > 0, after = 0, over >= 100k cases"
  - "independent adversarial verifier report attached verbatim before the user is asked"
  - "nothing pushed, posted or filed on GitHub; our repo not pushed"
result: "D3 shift-click: proven world-level gap (game.exe 0x63 FUN_0054d520 -> FUN_0063c600; d2d no command); PR candidate trace/click c4038d15 + ff0d8b2f on upstream 0dc4c62a, 13 files, 390,000 cases 129,283 -> 0, verifier holds with limits, PROOF complete, local only. D4 ctrl-click sell and D2 inventory-open pickup: client-only, no PR. Verified offline only (no game, no slot)"
```

## 1. Work

- [x] 1.1 trace Q1-Q3 in game.exe 1.14d (client FUN_0048ffe0 / 0x498870 / FUN_00480930; server FUN_0054d520 / FUN_00579510 / FUN_0055cf50)
- [x] 1.2 Q4 d2d upstream main: which layer each gap is in
- [x] 1.3 Q5 bot relevance and recommendation
- [x] 1.4 D3 candidate: branch trace/click (claim + tooling commits), oracle differential, upstream tests, pr_guard
- [x] 1.5 independent verifier (Opus) report attached to pr-click-proof.md
- [x] 1.6 trace doc, queue item 9, bundle row, index lines, OWNERS
- [x] 1.7 changecheck, leakcheck, openspec validate; commit
