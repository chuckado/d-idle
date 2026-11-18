# Verifier: #565 follow-up, seq-rate fails on a missing D2_PATCH_INSTALLER (2026-11-16)

> status: current, 2026-11-16 (card verifyseqfollow, TASK-559, Opus, singly; independent verifier before posting). Nothing posted. Subject: fork branch `fid/seq-ias-followup` head `67d2262f` on upstream main `b6e6940d`; body `docs/upstream/bodies/seq-ias-followup-body.md` (TASK-557). Checklist from `docs/upstream/verifier-p13-2026-11-16.md`.

Tags: **V run** (run by this verifier today on the NAS builder), **V read** (read by this verifier), **U** (unverified).

## Verdict: PASS WITH CHANGES

The branch is postable as it is. The body needs B1 to B4 below before posting.

## Diff and hygiene
- One file, `tests/game/test_seq_rate.cpp`, +6 -2. No `components/` or `tools/` change, so the merged claim cannot move (V read). Byte-identical to `310ff43d` (`git diff 310ff43d 67d2262f` is empty), same subject line (V run).
- Logic (V read): a set, non-empty `D2_PATCH_INSTALLER` that does not exist gives a stderr message and `return 1` in both modes. An unset or empty one takes the old path; empty already behaved like unset before. The check comes after `D2_MPQ_DIR` and `d2data.mpq`, so with no data a bad path still skips.
- Identity: author and committer are the user's noreply identity. `Agent: Claude Code` is the only trailer, after a blank line (V run, `git interpret-trailers --parse`).
- `pr_guard.py upstream/main HEAD`: 0 FAIL, 0 WARN. `presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN (data-gating line 143, build, gcc), answered in the body (V run).
- clang-tidy-19 on the file in the full build tree: 0 findings. Positive control (`-checks=-*,readability-magic-numbers`): 12 hits in the file, so the file is parsed (V run).
- Size: 8 changed lines, well within CodeRabbit limits. No private code, host names or paths in the diff (V read).

## Runs (V run; full Release build of `67d2262f`, tests and tools on, -j2; the first -j6 attempt was OOM-killed on the shared builder)

Data folder = the builder's gamedata (has `patch_d2.mpq`). "No-patch folder" = the same MPQs without `patch_d2.mpq`. Valid installer = a real `LODPatch_114d.exe`.

| Case | head `67d2262f` | before: upstream main `b6e6940d` |
|---|---|---|
| `D2_PATCH_INSTALLER` nonexistent, plain | **rc 1**, "does not exist" | rc 0, 36 cells (ran on the folder's patch) |
| nonexistent, `--seq-rate` | **rc 1** | rc 0, prints the grid |
| nonexistent, no-patch folder | **rc 1** | rc 0, `SKIP: no 1.14d patch data` |
| valid installer, plain | rc 0, 36 cells | rc 0, 36 cells |
| valid installer, no-patch folder | rc 0, 36 cells | not run |
| unset | rc 0, 36 cells | (unchanged path) |
| empty | rc 0, 36 cells | (unchanged path) |
| unset, no-patch folder | rc 0, `SKIP: no 1.14d patch data` | as head |
| `D2_MPQ_DIR` unset (bad installer set) | rc 0 plain / 77 `--seq-rate`, `SKIP: D2_MPQ_DIR` | as head |

- Claim still proven: `test_seq_rate --seq-rate` grid of 17,600 D lines (folder patch and valid installer give identical grids, `cmp`), then the branch's `tools/emu/seq_rate.py compare`: **7,600 cells same, rc 0**, both grids. `test_seq_rate` plain: **36 cells match game.exe**, rc 0.
- `test_game` rc 0, `test_combat` rc 0.

## Required changes to the body
- **B1 (medium, the "before" is wrong).** "was treated like 'no patch set' and the test skipped" is only half true (V run, table). With a patch file in the data folder, the old test ran on the folder's patch and passed, ignoring the bad path. Without one, it skipped. Write: "was ignored: with `patch_d2.mpq` or `LODPatch_114d.exe` in the data folder the test ran on that and passed, without one it skipped (exit 0, or 77 with `--seq-rate`)". In "What changes" replace "(before: silent skip)" with "(before: ignored, see above)".
- **B2 (medium).** Our reply on #565's CodeRabbit thread (`tests/game/test_seq_rate.cpp` line 141) says it was done in `310ff43d`, but that commit never reached the merged head. Say so in one line and link that thread, so the maintainer sees why this PR exists and that it is the same commit.
- **B3 (low).** "rc 1 ... in both plain and `--seq-rate` mode" holds only when `D2_MPQ_DIR` points at a folder with `d2data.mpq`. Add "with the data set present; without `D2_MPQ_DIR` it still skips first".
- **B4 (low).** In Evidence, name the patch source for each run: the folder's `patch_d2.mpq` for unset and empty, and a real `LODPatch_114d.exe` for the valid-installer run (rc 0, 36 cells, grid identical, compare 7,600 same). Add the no-patch-folder rc 1 case.

## Not checked
- No Windows, macOS or Clang build, no full ctest (the change touches one test's argument handling), no live game. Nothing ran on the worker host.

## Re-check commands (as run)
Lab scripts `vseqf/{bld,chain,extra}.sh` (lab only). On the builder, with the data set:
```sh
D2_PATCH_INSTALLER=/nonexistent test_seq_rate; echo $?            # head 1, b6e6940d 0
D2_PATCH_INSTALLER=<LODPatch_114d.exe> test_seq_rate; echo $?     # 0, 36 cells
test_seq_rate --seq-rate | grep '^D ' > grid.tsv; python3 tools/emu/seq_rate.py compare grid.tsv   # 7600 cells same, 0
```
