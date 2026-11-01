# PR candidate evidence pack: walk velocity

> status: local candidate, nothing pushed or posted (card walkfix, 2026-10-31; earlier: decompwalk). Proof: docs/research/pr-walk-proof.md (author's differential, the verifier's report, the fix pass); body: docs/research/pr-walk-draft.md.
>
> status: #156 MERGED upstream 2026-11-01T17:37Z (maintainer); the README table row conflict against the new main was resolved by keeping both rows. The posted lines below are history (card upstreamsync).

PROOF: complete for the five claims. READY pending the user's go (and the det state hashes, see Residual risks).

## The claim

A unit in a moving mode walks at FUN_00623f50's rate: base MonStats Velocity (a monster unit, a merc and a pet in WL and RN), walk % = diminished faster run/walk + velocitypercent, at least 25; chill adds coldeffect to velocitypercent.

## Branch

`fix/walk-velocity-v3` in the worker host's `~/d2d-lab/pr/walk3` worktree (card prfixes: v2 plus two ponytail notes; v2, head 10338345, is untouched), base upstream/main 72132b99 (the tip when checked), head 8db11206. Commits: d8e63d4a monster RN at Velocity, a5cf6415 chill adds coldeffect + floor 25, 1c3fa4e5 player velocitypercent raw, 455e366f merc, 8e5ff4cd pet (v2's 85d50c78 plus the two notes), 8db11206 tooling (test_game `--walk-velocity`, `tools/emu/walk_velocity.py`, README row). Author and committer the user's noreply identity, body `Agent: Claude Code`, no trailer. 10 files, +411 / -20. Local only, not pushed. The author's `fix/walk-velocity` (97d55732 on d0ea14b7) is superseded and untouched.

## Numbers (122,646 cases, seed 9191, game.exe md5 2b3600c1...)

| claim | mismatches before | after |
|---|---|---|
| monster WL / RN (claims 1, 2) | 15,371 / 23,172 | 0 |
| player (claim 3) | 15,146 | 0 |
| merc WL / RN | 815 / 7,472 | 0 |
| pet WL / RN | 1,968 / 5,493 | 0 |
| total | 69,437 | 0 |

The author's earlier exhaustive run (712,362 cases, 425,665 differ, 0 after) is in the proof doc; it used a transcription of d2d, this one the real build.

## Commands (worker host; reviewers use the body's)

```sh
cd <walk3>/tools/emu   # D2_MPQ_DIR, D2_PATCH_INSTALLER, D2_BUILD set; venv with unicorn, capstone, pefile
python walk_velocity.py gen 100000 9191 > cases.tsv
python walk_velocity.py game cases.tsv <game.exe> > game.tsv
<walk3>/../walk3-out/build/tests/test_game --walk-velocity cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
python walk_velocity.py compare cases.tsv game.tsv d2d.tsv     # mismatches 0
```

## Upstream checks (branch head 8db11206, re-run for v3)

The tree differs from v2's head 10338345 by four comment lines (pets.cpp, fight/tick.cpp). pr_guard 0 FAIL / 0 WARN (range and each commit); presubmit --quick with ruff on PATH 0 FAIL, 3 WARN (unchanged); built with `-Wshadow -Wunused -Wsign-compare`: 0 warnings on added lines; test_rules, test_combat, test_protocol, test_game (data dir) rc 0; clang-tidy (tidy-pr-select over 135 TUs): 0 findings in the touched files (the local clang-tidy flags pre-existing lines in untouched files, e.g. `drlg/outdoor.hpp`); `test_game --walk-velocity` on the same 122,646 cases: 0 mismatches, output byte-identical to v2's head (before, on upstream main: 69,437 per the author, not re-run).

Earlier (v2 head):

pr_guard 0 FAIL / 0 WARN (range and each commit); presubmit --quick 0 FAIL, 3 WARN answered in the body; clang-tidy names and includes: none in touched files; gcc -Wshadow / -Wunused / -Wsign-compare: 0 on added lines (one `rng` shadow in the first draft of the test mode found and fixed); ruff (upstream rule set) clean; test_game (data dir set), test_rules, test_combat, test_protocol rc 0.

## Residual risks

- Det state hashes (`tools/d2ddet`) not run: need a d2d slot. Movement of monsters, mercs and pets changes the hashes by design; run before posting if the user wants the proof of "nothing else moved".
- A Shadow pet's base and an untraced pet's chase are not claimed (ponytails in the code, a line in the body); the two pet follow paths at RunVelocity x 1.1 (`pets.cpp` untraced pet, `fight/tick.cpp` pets in town) carry a ponytail since v3; merc chill, item faster run/walk and pet chill are not modelled; a leftover pace stat list in game.exe is untraced; skill moves keep a multiplied chill; negative IAS/FHR/FCR/FBR is a separate candidate.
- Stat 0x43 inputs are d2d's model on both sides; game.exe's integer truncation vs d2d's float step (under 1/4096 subtile a tick).
- The merc and pet claims rely on "a merc and a pet are monster units" (the verifier's disassembly reads of FUN_00621360); no live trace of a merc's path velocity.
- Opus verifier's verdict (HOLDS WITH LIMITS) is on the three original commits; the merc and pet commits and the tooling have only their own differential, not a second reviewer.

## Recommendation

Go for the bundle slot after the det state hashes (a d2d slot) and the user's go; the second verifier's verdict (card verifyprs2) is HOLDS WITH LIMITS, limits in the body. Not run, for the user to approve:

```sh
git -C ~/d2d-lab/pr/walk3 push origin fix/walk-velocity-v3
gh pr create --repo psi29a/d2decomped --base main --head shatteredeagle:fix/walk-velocity-v3   # body: docs/research/pr-walk-draft.md
```

## Posted log

- posted: 2026-11-01 replay: upstream main unchanged at 72132b99, #152 #153 #155 still open; own branch fix/walk-velocity-v4 from v3 head 8db11206 (card walkopen). Review lessons applied before opening: the reviewer mode moved out of test_game.cpp (1,609 lines with it, upstream's limit about 1,500) into tests/game/test_walk_velocity.cpp (reviewer command now `test_walk_velocity --walk-velocity cases.tsv`; skips rc 77 without the argument or patched data), oracle seed masked to 32 bits (`& 0xFFFFFFFF`, seed 9191 gives the same cases), include-cleaner fixes (clang-tidy clean on the new file). Folded into the tooling commit; the tree differs from 8db11206 only in tests/game/test_game.cpp, tests/game/test_walk_velocity.cpp, tools/emu/README.md, tools/emu/walk_velocity.py.
- posted: 2026-11-01 gates on head e4a13181: pr_guard 0 FAIL / 0 WARN (range and each commit), presubmit --quick 0 FAIL / 3 WARN (answered in the body), test_game (data dir), test_rules, test_combat, test_protocol rc 0, gcc -Wshadow -Wunused -Wsign-compare 0 on added lines, reviewer mode 122,646 cases 0 mismatches, output identical to v2 and v3. The verifier's own harnesses were not re-run: the component sources are byte-identical to the verified head.
- posted: 2026-11-01 branch pushed to the fork as fix/walk-velocity (head e4a13181, six commits on upstream main 72132b99).
- posted: 2026-11-01 PR #156 opened https://github.com/psi29a/d2decomped/pull/156 (title 'game: monsters, mercs and pets walk at Velocity, chill adds coldeffect (differential proof vs game.exe)'; body: the draft with the precise claims, counts, reviewer command, all limits and the determinism note). go: user, 2026-11-01 ('queue the walk velocity pr behind the spot155').
- CI on e4a13181 (read-only polls): guard pass 7s; Linux (GCC) pass 20m25s; Windows (MSVC) pass 16m5s; macOS (Clang) pass 18m23s; Flatpak x86_64 pass 18m10s; Flatpak aarch64 pass 13m12s. No fixup needed.
- CodeRabbit: 'Review limit reached' notice on #156 at about 6 minutes after opening ("Next included review available in 29 minutes", i.e. about 03:50Z on 2026-11-01); no review comments yet; nothing answered. A later agent handles findings under the standing go.

- posted: 2026-11-01 CodeRabbit posted Review limit reached on #156 at open; next included review about 03:50Z 2026-11-01; a manual trigger needs the user go.
- posted: 2026-11-01 CodeRabbit first review on #156 (03:57Z), 2 inline MINOR findings, judged by card crfix156. (1) components/game/fight/merc.cpp, run at the foe sets `move_pct = 0`, suggested `run * 100 / velocity - 100`: INVALID. game.exe's run at a foe (FUN_005ded00 -> FUN_005deb60) sets no pace and a unit walks at MonStats Velocity x the walk % in RN too (proof doc, merc and pet section: disassembly read; merc differential 0 mismatches); the suggested formula is the pet model. (2) tools/emu/walk_velocity.py `compare` exit status ignored absent d2d rows (an empty d2d file compared 0 cases and exited 0): VALID, reproduced (empty d2d file rc 0 before, rc 1 after; 5 rows cut rc 1), fixed.
- posted: 2026-11-01 reply in the finding 1 thread (comment 4434322274, reply https://github.com/psi29a/d2decomped/pull/156#discussion_r4434330438): not changing, evidence as above. Nothing else posted; no `@coderabbitai review`.
- posted: 2026-11-01 finding 2 pushed to the fork branch fix/walk-velocity as a plain fast-forward (no force-push, as the #153 CodeRabbit fixes): old head e4a13181, new head 3f418788, one fixup-style commit on top (tools: walk_velocity.py compare fails on a case d2d did not answer; numeric-ID noreply identity, Agent: Claude Code, no trailer; decision: coordinator, option b). b7ccff90 (the autosquash rewrite) is kept only as local ref/walk-b7ccff90 on the worker host. Rechecked on 3f418788: compare on the 122,646 cases 0 mismatches rc 0, empty d2d file rc 1, 5 rows cut rc 1; pr_guard 0 FAIL / 0 WARN (range against upstream/main and the new commit); ruff B905 clean; presubmit --quick 0 FAIL, 3 WARN (unchanged). Only walk_velocity.py changed, so the C++ checks are unaffected.
- posted: 2026-11-01 reply in the finding 2 thread (comment 4434322277), naming 3f418788. No `@coderabbitai review` posted.
