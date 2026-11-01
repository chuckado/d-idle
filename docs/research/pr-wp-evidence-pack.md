# Evidence pack: waypoint refusal window (agents DECOMPWP, WPPR, 2026-10-31)

> status: POSTED 2026-10-31 as psi29a/d2decomped#153 (user go); see the posted lines at the end.
>
> status: MERGED 2026-11-01 (maintainer, #153; card upstreamsync).

**Claim.** game.exe refuses a waypoint travel for 10 s only after the player's last hostile declaration (never in single player); d2d refused it for 10 s after every arrival. The patch drops d2d's test and its `arrived_at` stamp.

**Proof.** Whole-section scans (author and verifier) find one writer of player data +0x160, called only from the hostile declaration; the differential runs game.exe's 0x49 dispatcher and handler natively (`tools/emu/wp_refuse.py game`) against d2d's REAL C++ (`test_game --waypoint-refuse`: `World::apply(cmd::Waypoint)` after real `World::arrive` level changes): [pr-wp-proof.md](pr-wp-proof.md).

| Run (C++ harness) | Cases | Before | After |
|---|---|---|---|
| single player, grid + 100,000 random (`gen 100000 7 --solo`) | 140,002 | 66,969 | 0 (md5 equal) |
| with hostile declarations (`gen 100000 7`, DECOMPWP's set) | 132,964 | 59,309 | 30,763, all hostile-window cases (d2d has no PvP: out of scope) |
| verifier, own generator (tick >= 10,000) + `gen 50000 271828 --solo` | 19,231 + 90,002 | 9,696 + 38,549 | 0 + 0 |

**Branch.** `fix/waypoint-hostile-window` on the worker host, on upstream/main d0ea14b7: 74199f4f (tooling) and 4724e138 (claim), author shatteredeagle, body "Agent: Claude Code", no trailer. 9 files, +199 / -9. Checks: test_game (19 OK incl. the new waypoint case), test_combat, test_rules, test_protocol rc 0; ctest failures test_mpq, smoke_d2d, quests_act1, quests_act4 identical on unmodified upstream; pad_use_d2d passes without its 10 s sleep; pr_guard 0 FAIL 0 WARN; presubmit --quick 0 FAIL (ruff clean, 3 WARN answered in the draft). Merges cleanly with the #149 and #152 heads.

**Verifier.** Independent Opus agent: **holds, with limits** (verbatim in the proof).

**Residual risks.** game.exe refuses in the first 10 s of machine uptime (zeroed field), d2d does not; PvP / joined path not modelled; a crafted self-targeted 0x5d might stamp the field (not run); linear sweeps cannot see code in data islands; existing d2d gaps (range, same-act, dead/busy drop) unchanged.

**PROOF: complete.** **Recommendation:** go when the bundle queue allows (after #149 / #152 merge, or bundled with the user's word). Bot follow-up after the merge: `WpCooldownSec` to 0 (BACKLOG).

**Commands (not run):**

    git -C <upstream checkout> push <fork remote> fix/waypoint-hostile-window
    gh pr create -R psi29a/d2decomped --head <fork>:fix/waypoint-hostile-window --title "world: the waypoint refusal counts from hostility, not the level change" --body-file pr-wp-draft-body.md

posted: 2026-10-31 branch fix/waypoint-hostile-window pushed to the user's fork; replayed onto upstream main 72132b99 (tooling 890adf1d, claim 44353293; the claim's unit test renames `camp` to `encampment` so it no longer hides main's `camp`, an MSVC C4456 error).
posted: 2026-10-31 PR opened: https://github.com/psi29a/d2decomped/pull/153 (9 files, 2 commits; user go 'push the next pr'; no review-request comment, as #147/#149/#152).
posted: 2026-10-31 fixup pushed to the PR branch as a fast-forward (old head 44353293, new head 51c960e5, no force-push, nothing posted on the PR): the Linux job's "Names and includes" step (clang-tidy readability-identifier-length) flagged `at` in tests/game/test_game.cpp (tooling commit 890adf1d); renamed to `when`, no behaviour change. Note: the fixup commit 51c960e5 was authored with the email shatteredeagle@users.noreply.github.com instead of the numeric-ID noreply form that docs/process/upstream-pr-standard.md specifies, so GitHub may not link it to the account; correcting it needs a force-push of the PR branch: decision pending with the user.
posted: 2026-10-31 force-push of the PR branch (pinned lease): old head 51c960e5, new head 1141b66a; reason: author attribution of the fixup (the earlier posted line), now folded into the tooling commit so the branch is two commits, all with the numeric-ID noreply form as author and committer; tree unchanged (git diff 51c960e5 HEAD empty, same tree hash); rechecked on the worker host: pr_guard against upstream/main 0 FAIL 0 WARN, test_game, test_rules, test_combat, test_protocol rc 0; nothing posted on the PR.
posted: 2026-10-31 23:01Z comment `@coderabbitai review` on #153 (by the user's account; the earlier "no review-request comment" line is superseded); CodeRabbit answered at 23:08Z with 2 inline findings and 1 nitpick.
posted: 2026-10-31 fixup pushed to the PR branch as a fast-forward (old head 1141b66a, new head 5315e66a, no force-push), standing go for CodeRabbit findings: (1) a zero-arrival row would inherit the previous case's level: `test_game --waypoint-refuse` now rejects it (the generated cases never have one: 0 of 140,002 rows with count 0); (2) the documented `gen` example in tools/emu/wp_refuse.py gets `--solo` as in tools/emu/README.md; (3) nitpick: `id`/`i` of the new case loop renamed `case_id`/`arrival_index`. Rechecked on the worker host: the single-player differential (140,002 cases) is 66,969 mismatches before and 0 after, the new output byte-identical to the earlier after run; pr_guard against upstream/main 0 FAIL 0 WARN; test_game, test_rules, test_combat, test_protocol rc 0; clang-tidy on test_game.cpp clean on the changed lines (only older findings in drlg/outdoor.hpp from this host's newer clang-tidy).
posted: 2026-10-31 replies on #153 CodeRabbit threads: comment 4433633373 (zero-arrival rows, answered by reply 4433657626) and comment 4433633374 (`--solo`, answered by reply 4433657704), each naming 5315e66a. The nitpick has no thread. No `@coderabbitai review` re-trigger posted.
posted: 2026-10-31 CI on 5315e66a: Linux (GCC), macOS (Clang), Windows (MSVC), Flatpak x86_64 and aarch64, guard all pass (CodeRabbit check skipping, release publish skipping). CodeRabbit replied on both threads confirming the fixes; no new findings, no review-limit notice.
