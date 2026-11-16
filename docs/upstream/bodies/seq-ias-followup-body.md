# test: seq-rate fails when D2_PATCH_INSTALLER names a file that is not there

Follows up #565 (P17 seq rate, merged at 46b1108e). One review comment from CodeRabbit on that PR was addressed after the PR head stopped following the branch, so it never landed: a `D2_PATCH_INSTALLER` that names a file that does not exist was ignored: with `patch_d2.mpq` or `LODPatch_114d.exe` in the data folder the test ran on that and passed, without one it skipped (exit 0, or 77 with `--seq-rate`). Either way a broken fixture path went unnoticed. Our reply on [that review thread](https://github.com/psi29a/d2decomped/pull/565#discussion_r4587634015) says it was done in `310ff43d`; that commit never reached the merged head, and this PR carries the same change. The other review follow-ups (case-insensitive `d2data.mpq` lookup, skip when there is no 1.14d patch, partial skip returns 0 and 77 only with `--seq-rate`, `catch` binding `error`, a data set that does not load fails, the Double Throw limits note) were already in the merged #565 and are not touched here.

## What changes

One file, `tests/game/test_seq_rate.cpp`, test hardening only. No change to any claim, to `components/`, or to `tools/emu/seq_rate.py`.

- `D2_PATCH_INSTALLER` set and non-empty but not an existing path: print `test_seq_rate: D2_PATCH_INSTALLER <path> does not exist` and return 1 (before: ignored, see above).
- `D2_PATCH_INSTALLER` unset or empty: unchanged (skip unless `patch_d2.mpq` or `lodpatch_114d.exe` is in the data folder).

## Evidence (container build of this branch on current main c3dc51c3, real 1.14d data, data folder has `patch_d2.mpq`)

- `D2_PATCH_INSTALLER` unset, and set but empty (patch source: the folder's `patch_d2.mpq`): `test_seq_rate` prints `36 cells match game.exe`, rc 0.
- `test_seq_rate --seq-rate` grid (17,600 D lines) then `tools/emu/seq_rate.py compare`: `7600 cells  same`, rc 0 (the claim cells stay at 0 differences; the merged claim is unchanged).
- `D2_PATCH_INSTALLER=/nonexistent/LODPatch_114d.exe`, with `D2_MPQ_DIR` pointing at the data set: rc 1 with the message above, in both plain and `--seq-rate` mode. The same on a data folder without `patch_d2.mpq`: rc 1 (before: `SKIP: no 1.14d patch data`, rc 0).
- Without `D2_MPQ_DIR` the data check still skips first: `SKIP: D2_MPQ_DIR`, rc 0 plain and 77 with `--seq-rate`. A data folder without a patch and no installer set: `SKIP: no 1.14d patch data`, rc 0.
- A real `LODPatch_114d.exe` as `D2_PATCH_INSTALLER` was run on the previous base (b6e6940d, same test file): rc 0, 36 cells, grid identical to the folder-patch grid. It was not repeated on c3dc51c3.
- `test_game` rc 0, `test_combat` rc 0. `clang-tidy-19` on `tests/game/test_seq_rate.cpp`: no warnings (a `readability-magic-numbers` control run on the same file reports 12, so the file is parsed).
- `python3 tools/pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `python3 tools/presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN: build and gcc skipped by `--quick` (built and run by hand above); `data-gating` at line 143 is the existing `load_game_data` call that already asserts `data->patched` two lines later.

## Checklist

- [x] test only; no behaviour change
- [x] commit builds alone

---

Agent: Claude Code
