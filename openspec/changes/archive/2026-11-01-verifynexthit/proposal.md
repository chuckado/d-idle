## Why

docs/process/upstream-pr-standard.md section 5: every upstream candidate needs an independent adversarial verifier before the coordinator may ask the user for a go. Candidate: fix/nexthit (NextHit justhit state, LastCollide, one unit a frame; card nexthit), claim 8abab0ed, tooling 7901fda2, on upstream main 72132b99.

## What Changes

- Own trace from the disassembly of how FUN_005adb00 sets justhit (state 0x56), how long it lasts and when in the game frame it ends, before reading the author's proof.
- Own oracle (tools/emu/verify_nexthit.py: the stat-list expiry chain run for real) and d2d harness (tools/emu/test_verify_nexthit.cpp: moving monsters, second casts, path-order contacts), run against real d2d builds of upstream main, the tooling alone, the claim alone and the head.
- CI-class checks on the head, det state hashes base vs patched on one hidden d2d slot.
- Verifier section in docs/research/pr-nexthit-verifier.md, a verdict line for docs/research/pr-bundle-queue.md.

## Impact

Docs and verifier tooling only. No game; one hidden d2d slot for the det run. Nothing pushed or posted; the author's branch and worktrees untouched.
