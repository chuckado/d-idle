## Why

docs/process/upstream-pr-standard.md section 5: every upstream candidate needs an independent adversarial verifier before the coordinator may ask the user for a go. Candidate: fix/freeze-resume (a frozen monster thinks again at the freeze end + MonStats aidel; card freezeresume), claim c6cc4f12, tooling efacc7f4, on upstream main 72132b99. The author's oracle hooks the event queue, so the dispatch frame was never run natively (the NEXTHIT lesson: an author's oracle hid a one-frame error).

## What Changes

- Own trace from the disassembly of the event queue (FUN_005416b0 file, FUN_005414d0 frame walk), the state-1 end callback and the freeze-time gate.
- Own oracle (tools/emu/verify_freeze_resume.py) that runs the REAL queue code from game.exe under upstream's emulator, and an own d2d probe (tools/emu/test_verify_freeze_resume.cpp), run against real d2d builds of upstream main, the claim alone and the head.
- Side claims: which unit kinds game.exe freezes, the stun and chill end callbacks, which d2d skills freeze vs chill.
- Verifier record docs/research/pr-freezeresume-verifier.md.

## Impact

Docs and verifier tooling only. No game, no d2d slot. Nothing pushed or posted; the author's branch and worktrees untouched. Overlaps: none expected (new files only); the bundle-queue verdict line goes to the merger.
