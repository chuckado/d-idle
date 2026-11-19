# Verifier brief: fid/warp-walk-r2 (P02), for the independent reviewer (Opus, singly)

Review `fid/warp-walk-r2` (private fork remote, head `3b4439b3`, base upstream `4e054f43`) against `docs/research/warp-walk-game-trace.md` and `docs/upstream/drafts/warp-walk-plan.md`. Do not trust the trace: re-run.

1. Build `test_warp_walk` of the oracle commit `18439e09` and of the claim head on the NAS builder; `test_warp_walk --dump 30 FILE` each; `S` line parent `4877 5857 6623`.
2. Emulator image (`docs/emu-image.md`): `warp_walk_path.py game FILE > g.tsv` on the parent's dump; `compare` must say `5857 equal, 0 differ`; `ww_outcome.py` of this repo (`tools/emu/ww_outcome.py`) must say parent 30 differ, claim 0.
3. Check the claim's model independently of the author: read `FUN_00548b00` case 5, `FUN_00548a50`, `FUN_0057f190`, `FUN_0057f090`, `FUN_00649970`, `FUN_006498a0` and decide whether "planned as a unit target" is what game.exe does; check which game functions the oracle hooks (list in the trace, Q2) and whether a hook hides a rule.
4. Decide whether the bits of the grid (d2d's) are an acceptable input (the trace says they are d2d's; the DRLG oracles cover the map), and whether ONE changed combination of 324 justifies a PR at all.
5. `presubmit --quick` 0 FAIL, `pr_guard.py upstream/main` 0 FAIL 0 WARN, ctest equals the parent (4 known data-set failures), each commit builds alone.
Report: verdict per claim, the counts you got, what you could not check.
