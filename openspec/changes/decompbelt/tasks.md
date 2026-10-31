# Tasks

```yaml
change: decompbelt
status: in-progress
skip_specs: true
agent: decompbelt
branch: agent/decompbelt
worktree: ~/code/d2lod-wt/decompbelt
owns:
  - docs/research/belt-pickup-trace.md
  - docs/research/pr-belt-proof.md
  - docs/research/pr-belt-draft.md
  - docs/research/pr-belt-evidence-pack.md
  - docs/decomp-queue.md
  - tools/emu/belt_*
exclusive:
  - fork-branch:trace/belt
gates:
  - "every address cited is read in game.exe 1.14d or run in the oracle (V) or tagged D/S/I/U; no U in a claim's causal chain"
  - "the deviation claimed has a differential against game.exe's own code: mismatches before > 0, after = 0, over >= 100k cases"
  - "independent adversarial verifier report attached verbatim before the user is asked"
  - "nothing pushed, posted or filed on GitHub; our repo not pushed"
merge:
result: "Q1 ground potion -> belt first (matching column, else empty column for hp/mp/rv autobelt), pack after; inventory open -> cursor. d2d equal for potions (390,000 cases); D1 isc/tsc proven 3,111 -> 0, branch trace/belt 6c2c7347 + 41bd5a88 local, verifier holds; bot findings B1-B5 for the fix queue; verified offline only (no game, no d2d slot)"
```

## 1. Work

- [x] 1.1 read primitives P1-P3/P7, review F1-F9, the upstream-pr standard, upstream's combat.md "Picking up"
- [x] 1.2 trace the client click (FUN_00480930), the server pickup (FUN_00548b00, FUN_00563560), the belt step (FUN_0062bad0, FUN_00628ba0, FUN_0063c560, FUN_0063c600, FUN_00628a40, FUN_0063afd0)
- [x] 1.3 belt model: belts.txt boxes, hotkeys (FUN_00498a90), use and compaction (FUN_00562390, FUN_0055edc0)
- [x] 1.4 shift/ctrl clicks on pack items (FUN_0048ffe0, FUN_0054d520)
- [x] 1.5 oracle tools/emu/belt_pickup.py (no hooks) and test_game --belt-pickup; 390,000 cases, 3,111 -> 0
- [x] 1.6 d2d fix, test_rules case, combat.md note; pr_guard 0/0; test_rules/test_combat/test_game green
- [x] 1.7 compare our bot (loot Want, pickup.go, BeltOf, freeBeltBox): findings B1-B5
- [x] 1.8 docs: trace, decomp queue, proof/draft/evidence pack
- [x] 1.9 verifier report attached to the proof (holds, with limits)
