# Opening the Frozen Armor bundle and the curse-on-immune PR (card openpr)

> status: done 2026-11-01. Upstream PRs #159 and #160 are open; nothing else was posted. CodeRabbit follow-ups belong to a later card.

## Posted

- posted: 2026-11-01 go: the user, popup answer "Open both, bundle first".
- posted: 2026-11-01 fetch upstream: main 52362d52 (#158 merge); both candidates were on 7c3a50af, rebased with `git rebase --onto upstream/main 7c3a50af` without conflicts. The cherry-pick `-x` lines were stripped from the commit messages (identity unchanged: the numeric-ID noreply, `Agent: Claude Code`, no other trailer).
- posted: 2026-11-01 push (plain, new branch) to the fork: fix/frozen-armor-and-resume-final, head d8cc3b47 (5aea44af Frozen Armor claim + tooling, d8cc3b47 freeze-resume claim + tooling).
- posted: 2026-11-01 PR opened: https://github.com/psi29a/d2decomped/pull/159 (13 files, 2 commits; title "Frozen Armor freezes by the freeze rule; a frozen monster thinks again at the freeze end + aidel"; body from pr-bundle-armor-resume-body.md without the internal header, hashes updated, ends `Agent: Claude Code`).
- posted: 2026-11-01 push (plain, new branch) to the fork: fix/curse-immune-final, head 4e2ec701 (9281a06d claim, 4e2ec701 tooling).
- posted: 2026-11-01 PR opened: https://github.com/psi29a/d2decomped/pull/160 (5 files, 2 commits; title "Curses and auras: a lowering counts a fifth on a monster whose base resistance is 100 or more"; body from pr-curseimmune-draft.md at status READY, two wording fixes, ends `Agent: Claude Code`).

## Checks at open time (worker host, upstream main 52362d52)

| | bundle (#159) | curse (#160) |
|---|---|---|
| pr_guard range / each commit | 0 FAIL 0 WARN / 0 and 0 | 0 FAIL 0 WARN / 0 and 0 |
| Release build, test_monsters / rules / combat / protocol / game | all rc 0 | all rc 0 |
| tooling unit tests | 17 + 7 ok (1 skip) | 17 + 7 ok (1 skip) |
| presubmit --quick | 0 FAIL / 3 WARN (answered in the body) | 0 FAIL / 3 WARN |
| differential | armor 0 mismatches of 108,506; resume 0 differ of 12,548 frozen | adj differ 0 of 147,628; dealt differ 1,958 (D2, not claimed) |

Not re-run after the rebase: gcc shadow sweep, clang-tidy, ruff (the bundle body says so). PR page checked with `gh pr view`: state OPEN, 2 commits each, file counts as above.

Body ending convention (copied from merged #157 and #158): the body ends with the line `Agent: Claude Code`; CodeRabbit appends its own release-notes block afterwards. No "Generated with" line, no session link.

## CodeRabbit follow-up on #160 (card crfollow160)

- posted: 2026-11-01 CodeRabbit review (state COMMENTED, 3 actionable) read: three inline findings, all on the tooling and test side: tests/game/test_sk_resist.cpp:77 (Minor: an unopened case file gives an empty run that returns success), tools/emu/sk_resist.py:90 (Minor: parse() silently drops a malformed case line), tools/emu/sk_resist.py:167 (Major: cmp() skips case ids missing from either result file and could report zero differences). Own check (CHECKPRS) added ruff B905 at sk_resist.py:186 (zip without strict=).
- decision: all three valid, each a trivially safe guard in test or tooling code; the claim commit 9281a06d is untouched. Fixes: the test returns 1 with a message when the case file does not open; parse() skips only blank lines and raises ValueError with file:line otherwise; cmp() exits on an empty case set and on any requested id missing from the game or d2d results (the old skip is removed); tbl() zip gets strict=False (same form as freeze_resume.py). +18 -6 lines in two files.
- check: ruff (presubmit rules plus B905) clean, test_tooling 7 ok (1 skip), pr_guard 0 FAIL / 0 WARN for the commit and for upstream/main..head, Release build of test_sk_resist; the d2d side re-run against the stored game-side data: seed 1 adj differ 0 of 147,628, dealt differ 1,958; seed 7 adj differ 0 of 147,628, dealt differ 1,977 (both as before, so behaviour is unchanged). Negative checks: missing case file rc=1, a three-field line names its line, empty case file and a truncated d2d file are rejected.
- posted: 2026-11-01 push (plain, no force) to the fork: fix/curse-immune-final 4e2ec701 -> 802dee68 (one follow-up commit, numeric-ID noreply identity, `Agent: Claude Code`). Branch tip and the PR head were both 4e2ec701 before the push. PR #160 now has 3 commits.
- posted: 2026-11-01 replies to the three CodeRabbit inline threads (comment ids 4437023955, 4437023961, 4437023965; replies 4437044889, 4437044955, 4437045021): each valid, fixed in 802dee68. No re-review was requested (the included allowance was used up by the first review).
- non-inline: walkthrough, finding list and review info repeat the three findings; no merge-risk note, pre-merge check or docstring-coverage item needs action.

## Shared-file lines for the merger

- BACKLOG: upstream #159 (Frozen Armor + freeze resume) and #160 (curse on immune) are open; watch for the maintainer and CodeRabbit (follow-ups need their own card).
- INDEX: docs/research/openpr-log.md: posted log for #159 and #160.
- CHANGELOG: Docs: opened upstream PRs #159 (Frozen Armor and freeze-resume bundle) and #160 (curse lowering on immune monsters); posted log in docs/research/openpr-log.md.

## CodeRabbit follow-up on #159 (card crfollow159)

- posted: 2026-11-01 CodeRabbit review (state COMMENTED, 1 actionable, 0 nitpicks) read; one inline finding on tools/emu/freeze_rules.py line 6 (Minor, maintainability, quick win): the usage text names a `test_freeze_rules --freeze-rules` binary the PR does not add, so `compare` cannot run as documented. Its walkthrough and risk note repeat the same point; no pre-merge check or docstring-coverage item was flagged.
- decision: valid. freeze_rules.py is only a shared helper in this PR (armor_hit.py subclasses its Game runner and has its own d2d side, test_armor_hit). The fix is docstring only: drop the two usage lines and say where the d2d side lives (+4 -2 lines). No claim changes. The bot's suggestion to add the missing target was not taken (it would widen the PR).
- check: the Python code is AST-identical before and after (docstring excluded), tooling test_tooling 7 ok (1 skip), ruff (presubmit rules plus B905) clean; the differential was not re-run because no code changed.
- posted: 2026-11-01 push (plain, no force) to the fork: fix/frozen-armor-and-resume-final d8cc3b47 -> c4e9db97 (one follow-up commit, numeric-ID noreply identity, `Agent: Claude Code`); pr_guard 0 FAIL / 0 WARN for the commit and for upstream/main..head. PR #159 now has 3 commits.
- posted: 2026-11-01 reply to the CodeRabbit inline thread (comment id 4436799936, reply 4436827126): valid, fixed in c4e9db97, docstring only.

## Blizzard / Eruption shard offsets (card openpr-blizzard)

- posted: 2026-11-01 go: the user's popup answer "Open it now" for the Blizzard shard-offset PR (one PR, upstream psi29a/d2decomped).
- posted: 2026-11-01 fetch upstream: main 7d5b3a5c (unchanged since the candidate's base, so no rebase and no new branch was needed; the candidate's tree and identity are as built).
- posted: 2026-11-01 push (plain, new branch name, no force) to the fork: fix/blizzard-shard-offsets-v2 head 3d6f3fe7 (46c55621 tests and tooling, then 3d6f3fe7 claim; numeric-ID noreply identity, `Agent: Claude Code`, no other trailer). The first candidate fix/blizzard-shard-offsets (9a1a90eb, 681d52a9) was not used.
- posted: 2026-11-01 PR opened: https://github.com/psi29a/d2decomped/pull/163 (5 files, 2 commits; title "Blizzard and Eruption shards fall -(R-1)..(R-2) of the centre, not -(R-1)..(R-1) (FUN_005a9820)"). State OPEN, base 7d5b3a5c, head 3d6f3fe7 as checked with gh pr view. Body: the draft's body section with the status line, internal links and internal wording removed; it ends with a blank line and the bare `Agent: Claude Code` line, like #157 and #158 (no "Generated with" line, no session link).
- check (all V, run on the worker host on head 3d6f3fe7 before the push):

| check | result |
|---|---|
| pr_guard, range origin/main..head and each commit | 0 FAIL / 0 WARN, rc 0 each |
| Release build (`-ffp-contract=off`), test_game (with the data dir) and test_rules | build rc 0, both pass |
| presubmit --quick | 0 FAIL, 1 WARN (build skipped by --quick); tooling unit tests ok |
| ruff 0.8.6, repo rules plus B905, py312 | clean |
| test_sk_ground --blizz, base build (tooling only, source of main) | seed 1: 14,926 of 100,000 off the -6..5 grid; seed 7: 14,957 |
| same on the head | 0 and 0 |
| independent harness (verify_blizzard.py gen-d2d 5000 777, two Release probe builds) | base 16,676 of 112,964 shards off the grid (= Blizzard 14,804 + Eruption 1,872 of the claim's split); head 0; shards per cast {25: 3997, 13: 1003}; the per-skill split itself was not re-tabulated |
| git merge-tree (write-tree) head against #160 | clean; against main clean; against #159 conflicts in ai.hpp (a file this PR does not touch, the known #159 conflict) |

- not re-run: the game-side emulator comparisons (64,200 and 120,000 cases), the world-rng-state check, gcc -Wshadow and clang-tidy (clean on the first candidate, identical source), det state hashes, MSVC, a live game. The body says so.
- nothing else posted: no comment, label or reviewer request; CodeRabbit follow-ups are a later task.

## Shared-file lines for the merger

- BACKLOG: #159 got a docstring follow-up (c4e9db97) for CodeRabbit; #160 review still pending after the rate limit (own card).
- BACKLOG: #160 got a tooling and test follow-up (802dee68) for CodeRabbit; watch for the maintainer.
- INDEX, CHANGELOG, decomp-queue, research README: no change.

## CodeRabbit follow-up on #163 (card crfollow163)

- posted: 2026-11-01 CodeRabbit review (state COMMENTED, 1 actionable, 1 nitpick) read: inline finding on tools/emu/sk_ground_blizz.py:147 (Minor, quick win): cmp still accepts -(R-1)..(R-1) as d2d's range, so +(R-1) counts as valid for d2d; bounds, range and share lines and the docstring need the new upper end R-2. Nitpick in the review body (not inline): tests/game/test_sk_ground.cpp lines 66-67 and 86, short loop variable names (x, y, i) against the repo's clang-tidy rule.
- decision: inline finding valid (the claim code in missiles.cpp is right: rng(2 reach) - reach is -6..5 at radius 7; only the tool's idea of d2d's range was stale). Fixed in tooling only: lo/hi, the d2d-only value set, the printed range and the docstring use -(R-1)..(R-2); radius 1 keeps 0..0 (it draws nothing; without that clamp the check flagged 246 radius-1 shards); the old 14.8 percent share is kept as a labelled "before the fix" line. +8 -5 lines in one file.
- decision: nitpick not applied: the brief's standing go covers the one inline thread and keeps the tests unchanged; presubmit's short-names check passes (it only looks at added lines). Left for the author to decide.
- check: ruff (presubmit rules plus B905, py312) clean, tooling tests 31 ok, presubmit --quick 0 FAIL (1 WARN: build skipped by --quick), pr_guard 0 FAIL / 0 WARN for the commit and for origin/main..head. Differential re-run on the stored data: Blizzard 14,926 / 14,957 of 100,000 off-grid on base and 0 / 0 on head (seeds 1 / 7), cmp over 42,100 cases: reading differs 0, game offsets outside d2d's range 0. No game-side or d2d-side code changed, so behaviour is unchanged. No Release build rerun (no C++ touched).
- posted: 2026-11-01 push (plain, no force) to the fork: fix/blizzard-shard-offsets-v2 3d6f3fe7 -> 3ac63fa0 (one follow-up commit, numeric-ID noreply identity, `Agent: Claude Code`). The PR head was 3d6f3fe7 before the push and is 3ac63fa0 now.
- posted: 2026-11-01 reply to the CodeRabbit inline thread (comment id 4437340451, reply 4437363818): valid, fixed in 3ac63fa0. No re-review was requested.
- non-inline: walkthrough repeats the claim and tools; effort 3 (about 25 minutes); possibly related issue 103 (RNG stand-ins), no action. Merge risk Low, with the same note as the inline finding (now fixed). Pre-merge checks: 4 passed, 1 warning: docstring coverage 11.11 percent against 80 (scoped to touched functions with complexity of 2 or more); a style gate on the tooling script, no action taken (not a defect; the author may add docstrings). Status checks: guard success, CodeRabbit neutral; review decision REVIEW_REQUIRED.

## Shared-file lines for the merger

- BACKLOG: #163 got a tooling follow-up (3ac63fa0) for CodeRabbit; watch for the maintainer.
- INDEX, CHANGELOG, decomp-queue, research README: no change.
