# Verifier brief: fid/smite-r1 (P18), for the independent reviewer (Opus, singly)

Review `fid/smite-r1` (private fork remote, two commits on upstream `d1a10a38`: oracle, then claim) against `docs/research/decomp33-p17-p19.md` section P18 and `docs/upstream/bodies/smite-range-body.md`. Do not trust the trace: re-run.

1. Build `test_smite_damage` and `test_combat` of the oracle commit and the claim head on the NAS builder; `test_smite_damage --smite-damage > f.txt` on each (75,600 lines).
2. Emulator: `smite_damage.py compare f.txt`: oracle build `75600 cells  63709 differ` (exit 1), claim `75600 cells  same`. An empty file and a short S line must exit 1.
3. Read `FUN_0057b420` `0x57b420..0x57b5d0` and `FUN_005ce9f0` around `0x5ceb31` and decide whether the game column is the shield path (item non-null, edx = 0) and whether the hooks (stat getters, `FUN_00629860` / `FUN_006298a0`, `FUN_00645830`, `FUN_0063d340`) hide a rule. Check that the roll is the same stream on both sides (the seed words go to the unit's `+0x20` / `+0x24` and to `Rng::low` / `high`).
4. Judge reach: stat 111 on three uniques and a runeword, Zakarum Shield 46/46. Is a PR about a rule this rare worth posting, and is the `test_combat` change (15 to 16 instead of exactly 15) acceptable?
5. Mutants: remove the maximum rule (`base_high = base_low + 256`) and, separately, the stat 111 term in `make_fighter`; the test must abort each time.
6. `presubmit --quick` 0 FAIL, `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, full ctest equals the parent (5 known failures), each commit builds alone.

Report: verdict per claim, the counts you got, what you could not check.
