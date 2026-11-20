# Verifier brief: fid/seq-ias-r1 (P17), for the independent reviewer (Opus, singly)

Review `fid/seq-ias-r1` (private fork remote, two commits on upstream `d1a10a38`: oracle, then claim) against `docs/research/decomp33-p17-p19.md` section P17 and `docs/upstream/bodies/seq-ias-body.md`. Do not trust the trace: re-run.

1. Build `test_seq_rate` of the oracle commit and of the claim head on the NAS builder (`docs/upstream/offline-decomp21-2026-11-14.md` recipe); `test_seq_rate --seq-rate | grep '^D '` on each (17,600 lines each).
2. Emulator: `seq_rate.py compare` on the oracle build's file must say `7600 cells  6941 differ` and exit 1; on the claim's `7600 cells  same` and exit 0. Feed it an empty file and a file with a short D line: both must exit 1 with a message. `seq_rate.py table` must show Jab, Impale, Double Swing, Double Throw, Frenzy as attack, Lightning and Chain Lightning as cast.
3. Decide independently whether the game column is game.exe: read `FUN_00623f50` `0x624290..0x62429b`, `FUN_006216e0`, `FUN_00621580`, `0x62442f..0x62443f`; check which functions the oracle hooks (list in the script's docstring) and whether a hook hides a rule (the stat getters return the grid's stat 93 and 68; `FUN_00620f00` is a no-op and the record's `+0xc` is d2d's speed).
4. Judge the limits: the claim is the rate term only; the AnimData speed for mode 0x12 (the `GH` record, 256 with no weapon; Assassin claws 208 in d2d) and the step count are not claimed. Is a PR that makes Jab 19 ticks at 0 IAS acceptable on this evidence, or does it need the sequence length itself proved (FUN_005539b0 on a unit with `+0x30` set, FUN_006633b0..FUN_006633f0)?
5. Mutants: set `kSequenceRatePenalty` to 0, then drop the call-site term in `start_sequence`; `test_seq_rate` must abort both times.
6. `presubmit --quick` 0 FAIL, `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN, full ctest equals the parent (5 failures both: `test_outdoor`, `smoke_d2d`, `touch_controls_d2d`, `touch_corpse_d2d`, `pad_use_d2d`), each commit builds alone.

Report: verdict per claim, the counts you got, what you could not check.
