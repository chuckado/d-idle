# Evidence pack: waypoint refusal window (agent DECOMPWP, 2026-10-31)

> status: for the user's go / no-go; nothing pushed or posted.

**Claim.** game.exe refuses a waypoint for 10 s after the player's last hostile declaration (never in single player); d2d refuses it for 10 s after every arrival.

**Proof.** Exhaustive scan of game.exe's code for writes to player data +0x160 (one writer, one caller: the hostile path), plus the differential `tools/emu/wp_refuse.py` running the 0x49 dispatcher, handler and the setter natively: [pr-wp-proof.md](pr-wp-proof.md).

| Run | Cases | Differ |
|---|---|---|
| d2d today | 132,964 | 59,309 (47,157 d2d refuses after a level change; 12,152 hostile, absent in d2d) |
| after the patch (rule `--after`) | 132,964 | 0 |

**Branch.** `fix/waypoint-hostile-window` in the upstream read worktree on the worker host, one commit (author shatteredeagle, body "Agent: Claude Code"), 6 files, +10 / -9. Upstream tests test_game, test_combat, test_rules green; pr_guard 0 FAIL 0 WARN; presubmit --quick 0 FAIL, 4 WARN answered in the draft (ruff absent; siblings: players only; "never": single player; build skipped).

**Residual risks.** The after-run checks a Python copy of the patched rule, not d2d's C++ (no test_game mode for 0x49 yet). The pre-existing GCC 16 failure of the d2d app target (apps/d2d/ui.cpp) blocked an app build here. No independent verifier has run yet.

**PROOF: incomplete** (verifier report missing; a test_game case for "waypoint right after arrival travels" would make the after-run C++). **Recommendation:** go after the verifier, or bundle with the next PR.

**Commands (not run):**

    git -C <upstream checkout> push <fork remote> fix/waypoint-hostile-window
    gh pr create -R psi29a/d2decomped --head <fork>:fix/waypoint-hostile-window --title "world: the waypoint refusal counts from hostility, not the level change" --body-file pr-wp-draft-body.md
