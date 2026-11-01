## Why

docs/process/upstream-pr-standard.md section 5: every upstream candidate needs an independent adversarial verifier before the coordinator may ask the user for a go. Candidate: fix/armor-hit (Frozen Armor's striker effect by the freeze rule; card armorhit), claim 36cbe015, tooling d5bd19c1, on upstream main 72132b99.

## What Changes

- Own trace from the disassembly of FUN_005cac50 -> FUN_0057c6c0 -> FUN_0057c1e0 / FUN_0057b230 / FUN_0057af80, including the freeze's end and the frame the monster acts again.
- Own oracle (tools/emu/verify_armor_hit.py: timers recorded and replayed, the freeze-end callback run for real) and d2d harness (tools/emu/test_verify_armor_hit.cpp), run against base, tooling alone, claim alone and head.
- CI-class checks on the head, det state hashes base vs patched on one hidden d2d slot.
- Verifier section in docs/research/pr-armorhit-verifier.md, a verdict line for docs/research/pr-bundle-queue.md.

## Impact

Docs and verifier tooling only. No game; one hidden d2d slot for the det run. Nothing pushed or posted; the author's branch and worktrees untouched.
