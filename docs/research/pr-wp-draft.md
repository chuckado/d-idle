# PR draft: waypoint refusal window (agent DECOMPWP, 2026-10-31)

> status: draft for a LOCAL branch; not pushed, not posted. Proof: [pr-wp-proof.md](pr-wp-proof.md). Evidence: [pr-wp-evidence-pack.md](pr-wp-evidence-pack.md).

**Title:** world: the waypoint refusal counts from hostility, not the level change

**Body:**

game.exe 1.14d refuses a waypoint travel (C->S 0x49) while `GetTickCount() < pcdata+0x160 + 10000` (FUN_0054c5d0, the field read by FUN_0055b6c0). d2d took that field for "when the player last changed level" and refused for 10 s after every arrival, so a waypoint used right after a level change did nothing.

| Address | What it shows |
|---|---|
| FUN_0054c5d0 | `+ 0x2710`, GetTickCount, `jbe`: else voice 0x13 and the interaction ends |
| FUN_0055b6c0 | reads player data +0x160 |
| FUN_0055b720 | the only writer of player data +0x160 in .text (0x55b76b): `= GetTickCount()` |
| FUN_005a5e50 | FUN_0055b720's one caller (0x5a5f51): the hostile declaration |
| FUN_00584870 | the town portal's 5 s on the same field (objects.md already says hostility) |

d2d has no hostile declaration, so the test and `arrived_at` go; act1-end.md and waypoint.md now say what the field is; tests/pad_use_d2d.py loses its 10 s sleep.

Evidence (the oracle is attached to this conversation, not in the tree): `PYTHONPATH=tools/emu uv run python wp_refuse.py 100000 7` -> 132,964 cases, 59,309 differ against d2d's rule; `--after` -> 0. It runs FUN_0054d750, FUN_0054c5d0, FUN_0055b6c0, FUN_00554190 and FUN_0055b720 natively.

Callers: World::swap_npcs (no longer stamps arrived_at; nothing else read it); the 0x49 command (travels at once). Siblings: only players declare hostility; mercs, pets, monsters do not use waypoints. "never" in waypoint.md: single player has no hostile declaration (FUN_005a5e50 needs two players).

Tests: test_game, test_combat, test_rules green; pr_guard 0 FAIL 0 WARN. The d2d app target does not build on GCC 16 on upstream/main itself (apps/d2d/ui.cpp:71 -Warray-bounds), unrelated.

Agent: Claude Code
