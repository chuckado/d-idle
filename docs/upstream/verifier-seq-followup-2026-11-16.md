# Verifier: #565 follow-up, seq-rate fails on a missing D2_PATCH_INSTALLER (2026-11-16)

> status: INTERIM, 2026-11-16 (card verifyseqfollow, TASK-559, Opus). Read checks done; the builder runs are NOT done (see "Not run"). Nothing posted. Subject: fork branch `fid/seq-ias-followup` head `67d2262f` on upstream main `b6e6940d`; body `docs/upstream/bodies/seq-ias-followup-body.md` (TASK-557).

Tags: V run (run by this verifier), V read (read by this verifier), U (not verified).

## Provisional verdict: PASS WITH CHANGES (runtime evidence pending)

## Done (V)
- Diff: one file, `tests/game/test_seq_rate.cpp`, +6 -2. No `components/` or tool change, so the claim cannot move (V read). Byte-identical to `310ff43d` (`git diff 310ff43d 67d2262f`: empty) (V run).
- Logic (V read): a set, non-empty, nonexistent `D2_PATCH_INSTALLER` gives a message on stderr and return 1, in both modes. An unset or empty one keeps the old path. Empty was already the same before (`exists("")` false, `fs::path("")`). The check comes after `D2_MPQ_DIR` and `d2data.mpq`, so without data a bad path still skips.
- Identity: author and committer are the user's noreply identity. `Agent: Claude Code` is the only trailer, after a blank line (V run, `git interpret-trailers --parse`).
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick`: 0 FAIL, 3 WARN (data-gating line 143, build, gcc), as the body says (V run).
- clang-tidy-19 on `tests/game/test_seq_rate.cpp` (configured build tree): 0 findings. Positive control `readability-magic-numbers`: 12 hits in the file, so the file is parsed (V run).
- Size: 8 lines, well inside CodeRabbit limits.

## Not run (builder OOM, then the run was stopped at hand-back)
The full build of `67d2262f` on the NAS builder was killed by the OOM killer at -j6 (`test_pet_crowd_merc.cpp`, cc1plus Killed). The -j2 retry was stopped at hand-back with no result. Still open: test_seq_rate 36 cells, `compare` 7,600 same, the nonexistent / valid (`/data/p565-emu/LODPatch_114d.exe`) / unset / empty / no-patch-folder cases, test_game, test_combat, and the "before" behaviour on `b6e6940d`. Scripts: `~/d2d-lab/vseqf/{bld,chain,extra}.sh` (lab only).

## Required body changes
- B1 (medium). "Before: silent skip" is only true when the data folder has no patch file. With `patch_d2.mpq` in the folder (the builder's case, and CI's), the old code passed the missing path to `load_game_data`. Say "skipped, or ran with a path that is not there" (confirm with the `b6e6940d` run).
- B2 (medium). Our own reply on #565 said this was "done (310ff43d)", but it never landed. Say so and link that review thread, so the maintainer sees why a follow-up exists.
- B3 (low). "rc 1 in both modes" holds only when `D2_MPQ_DIR` and `d2data.mpq` are present; add "with the data set".
- B4 (low). Evidence: name the installer used for the "valid" run. The builder's gamedata has no installer `.exe`, so an empty or unset variable is not a valid-installer run.
