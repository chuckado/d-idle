# PR candidate evidence pack: Static Field

> status: local candidate, verifier NOT run, nothing posted (card decompstatic, 2026-10-31). Proof: docs/research/pr-static-proof.md; body: docs/research/pr-static-draft.md.

- Claim: d2d's Static Field damage per unit = game.exe FUN_005c96a0 (whole-point life, StaticFieldMin floor, life - 1 cap, calc2 minimum, negative resistance cancelled, 256ths with pierce).
- Branch: `fix/static-field` in the worker host's `static` worktree (PR checkout dir), base upstream/main 72132b99; d9d64d53 claim (5 files, +37 / -3), 6270d162 tooling (oracle + `test_game --static-field`). Author shatteredeagle noreply, body `Agent: Claude Code`, no Claude trailer. Not pushed.
- Numbers: 309,664 cases; before 115,103 differ (line-for-line copy of the old code); after 0 (real build). Slices: Normal-like res 0..99 10,607 / 34,380; res < 0 9,747 / 29,055; floor skips 25,490 / 113,553.
- Checks: test_game / test_rules / test_combat rc 0; pr_guard 0 FAIL 0 WARN; shadowing sweep clean; presubmit and det hashes not run.
- Residual risks: the old side is a Python copy, not a build of upstream/main; one commit holds five rules of one function (S1..S6); the area walk is out of scope.
- Verifier: not run. PROOF: incomplete (verifier).
- Recommendation: run the verifier, then bundle with the next open bundle on the user's go.
- Commands (not run): `git -C <static worktree> push origin fix/static-field`; `gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/static-field --body-file pr-static-draft.md`.

posted: 2026-10-31 branch pushed to the user's fork as fix/static-field (local fix/static-field-v2 on the worker host, the author's two commits unchanged: replay onto upstream main 72132b99 was a no-op, author and committer already the numeric-ID noreply form, no Claude trailer). Re-run before opening on that tree: test_game / test_rules / test_combat / test_protocol rc 0, pr_guard 0 FAIL 0 WARN against upstream/main, presubmit --quick 0 FAIL 3 WARN (answered in the body), `--static-field` 309,664 cases 0 differ (md5 5583e13b), gcc -Wshadow -Wunused -Wsign-compare 0 warnings, clang-tidy only the existing outdoor.hpp names.
posted: 2026-10-31 PR opened: https://github.com/psi29a/d2decomped/pull/155 (7 files, 2 commits; user go: open the next ready bundle while #153 has no CodeRabbit review; no review-request comment). The body carries the verifier's limits; the cold chill deviations are not mentioned.
posted: 2026-10-31 CI of #155 on the pushed head: guard, Linux (GCC, incl. Names and includes), Windows (MSVC), macOS (Clang), Flatpak x86_64 and aarch64 all pass; no fixup needed. CodeRabbit posted "Review limit reached" (next included review in about 48 minutes from 23:13Z) and its check shows skipping; no trigger comment posted (coordinator decides).
