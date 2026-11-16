# test: seq-rate fails when D2_PATCH_INSTALLER names a file that is not there

Follows up #565 (P17 seq rate, merged at 46b1108e). One review comment from CodeRabbit on that PR was addressed after the PR head stopped following the branch, so it never landed: a `D2_PATCH_INSTALLER` that names a file that does not exist was treated like "no patch set" and the test skipped (exit 0, or 77 with `--seq-rate`), which hides a broken fixture. The other review follow-ups (case-insensitive `d2data.mpq` lookup, skip when there is no 1.14d patch, partial skip returns 0 and 77 only with `--seq-rate`, `catch` binding `error`, a data set that does not load fails, the Double Throw limits note) were already in the merged #565 and are not touched here.

## What changes

One file, `tests/game/test_seq_rate.cpp`, test hardening only. No change to any claim, to `components/`, or to `tools/emu/seq_rate.py`.

- `D2_PATCH_INSTALLER` set and non-empty but not an existing path: print `test_seq_rate: D2_PATCH_INSTALLER <path> does not exist` and return 1 (before: silent skip).
- `D2_PATCH_INSTALLER` unset or empty: unchanged (skip unless `patch_d2.mpq` or `lodpatch_114d.exe` is in the data folder).

## Evidence (container build of this branch on current main b6e6940d, real 1.14d data)

- `test_seq_rate`: `36 cells match game.exe`, rc 0.
- `test_seq_rate --seq-rate` grid then `tools/emu/seq_rate.py compare`: `7600 cells  same`, rc 0 (the claim cells stay at 0 differences; the merged claim is unchanged).
- `D2_PATCH_INSTALLER=/nonexistent`: rc 1 with the message above, in both plain and `--seq-rate` mode. `D2_MPQ_DIR` unset: `SKIP: D2_MPQ_DIR`, rc 0 plain and 77 with `--seq-rate`. Empty `D2_PATCH_INSTALLER`: runs on the data folder, rc 0.
- `test_game` rc 0, `test_combat` rc 0. `clang-tidy-19` on `tests/game/test_seq_rate.cpp`: no warnings.
- `python3 tools/pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN.
- `python3 tools/presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN: build and gcc skipped by `--quick` (built and run by hand above); `data-gating` at line 143 is the existing `load_game_data` call that already asserts `data->patched` two lines later.

## Checklist

- [x] test only; no behaviour change
- [x] commit builds alone

---

Agent: Claude Code
