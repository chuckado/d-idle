# PR body draft: walk velocity

> status: draft, NOT posted (needs the user's go and the independent verifier first). Card decompwalk.

**Title:** game: monster run speed, chill and velocitypercent follow FUN_00623f50

A moving mode's speed is set once by FUN_00623f50: walk % = faster run/walk (FUN_00621740, diminished) + velocitypercent (stat 0x43, as it is), at least 25; velocity = FUN_00621360's base x walk % / 100. The base is the walk velocity, MonStats Velocity for a monster in every moving mode.

Three claims, one commit each:

1. A running monster moves at MonStats Velocity. Run is only read where an AI turns it into a pace (Run x 100 / Velocity - 100) and by the charges. A Zombie (Velocity 1, Run 3) ran three times too fast.
2. Chill puts MonStats coldeffect on velocitypercent (FUN_0057af80), so it adds to the pace, the auras and the boss speed; the floor is 25, not 10.
3. A player's velocitypercent (Vigor, Burst of Speed, Increased Speed) adds as it is; only faster run/walk is diminished, negative values too.

| Address | What it shows |
|---|---|
| FUN_00623f50 | walk % sum and floor (0x62414e), velocity store (0x6241c9 via FUN_00648690) |
| FUN_00621360 | base: MonStats +0x32 (named Velocity by the loader) for any mode |
| FUN_00621740 | 150 v / (150 + v) for v != 0 |
| FUN_0057af80 | chill's coldeffect on stat 0x43 |

Evidence: an exhaustive differential running game.exe's own code (712,362 cases: 425,665 differ before, 0 after); re-run with `tools/emu/walk_speed.py` (attached script, reads your own MPQs). Upstream tests: test_game, test_rules, test_combat pass; pr_guard 0 FAIL / 0 WARN.

Agent: Claude Code
