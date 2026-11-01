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

## Shared-file lines for the merger

- BACKLOG: #159 got a docstring follow-up (c4e9db97) for CodeRabbit; #160 review still pending after the rate limit (own card).
