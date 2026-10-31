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
