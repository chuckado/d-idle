# Evidence pack: waypoint refusal window (agents DECOMPWP, WPPR, 2026-10-31)

> status: for the user's go / no-go; nothing pushed or posted.

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
