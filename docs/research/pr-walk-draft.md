# PR body draft: walk velocity

> status: draft, NOT posted (needs the user's go). Branch `fix/walk-velocity-v3` (v2 plus two ponytail notes), card prfixes (the verifier's findings are fixed or listed under Limits). Earlier draft: card decompwalk.

**Title:** game: monster, merc and pet walk speed, chill and velocitypercent follow FUN_00623f50

A moving mode's speed is set once by FUN_00623f50: walk % = faster run/walk (FUN_00621740, diminished) + velocitypercent (stat 0x43, as it is), at least 25; velocity = FUN_00621360's base x walk % / 100. The base is the walk velocity, MonStats Velocity for a monster unit in every moving mode (a merc and a pet are monster units).

Five claims, one commit each:

1. A running monster moves at MonStats Velocity. Run is only read where an AI turns it into a pace (Run x 100 / Velocity - 100) and by the charges. A Zombie (Velocity 1, Run 3) ran three times too fast.
2. Chill puts MonStats coldeffect on velocitypercent (FUN_0057af80), so it adds to the pace, the auras and the boss speed; the floor is 25, not 10.
3. A player's velocitypercent (Vigor, Burst of Speed, Increased Speed) adds as it is; only faster run/walk is diminished, negative values too.
4. A merc walks at MonStats Velocity in WL and RN, floor 25 (merc.cpp).
5. A pet walks at MonStats Velocity, its walk % the sum of its pace, speed and aura velocitypercent, floor 25 (pets.cpp `pet_walking`).

| Address | What it shows |
|---|---|
| FUN_00623f50 | walk % sum and floor (0x62414e), velocity store (0x6241c9 via FUN_00648690) |
| FUN_00621360 | base: MonStats +0x32 (named Velocity by the loader) for any mode |
| FUN_00621740 | 150 v / (150 + v) for v != 0 |
| FUN_0057af80 | chill's coldeffect on stat 0x43 |

Evidence (re-run on your own copy; the d2d side is the real build): `test_game --walk-velocity` calls the real `monster_update`, `World::walk_percent`, `Fight::merc_turn` and `Fight::pet_walking`; `tools/emu/walk_velocity.py` runs game.exe's own FUN_00623f50. 122,646 cases (boundaries plus 100,000 random, seed 9191): 69,437 differ on current main, 0 with this branch (per claim before to after: monster 38,543, player 15,146, merc 8,287, pet 7,461). Commands:

```sh
cd tools/emu
uv run python walk_velocity.py gen 100000 9191 > cases.tsv
uv run python walk_velocity.py game cases.tsv > g.txt
../../build/tests/test_game --walk-velocity cases.tsv | grep -E '^[0-9]+ ' > d.txt
uv run python walk_velocity.py compare cases.tsv g.txt d.txt     # mismatches 0
```

`tools/emu/walk_velocity.py` is under `tools/`, so it needs CODEOWNERS review; it reads your own game data. Upstream tests: test_game, test_rules, test_combat, test_protocol pass; pr_guard 0 FAIL / 0 WARN; presubmit --quick 0 FAIL.

Limits: what feeds stat 0x43 (boss speed +20, pace, auras, a running player's Run x 100 / Walk) is d2d's model, taken as given on both sides. game.exe truncates the velocity to an integer and d2d's float step does not (under 1/4096 subtile a tick).

Not claimed (each has a `// ponytail:` where d2d's code stands):
- Two pet paths still follow the owner at RunVelocity x 1.1: an untraced pet (`pets.cpp`, no target) and every pet in town (`fight/tick.cpp`). game.exe's rule for those paths is untraced; they are outside the claim and behave as before.
- Merc chill, an item's faster run/walk on a merc, and pet chill are not modelled (stat 0x43 / 96 inputs are d2d's own).
- A leftover pace stat list in game.exe is untraced.
- Shadow pets are excluded: FUN_00621360 gives a player class WalkVelocity in every mode, and whether a Shadow's run has the stat list is not traced (a ponytail in `pets.cpp`); the chase of an untraced pet (`max(Run, Velocity)`, no floor) stays.
- Skill moves (`ai.cpp` Diablo's run and the whirl, about lines 1883 and 1937) keep a multiplied chill and a 10 floor.
- `rules::effective_speed` leaves negative IAS / FHR / FCR / FBR undiminished (FUN_00621740 rows 0..3): a separate candidate.
- Integer truncation (game.exe) against d2d's float step, as above.
- The scenario state hashes were not re-run.

Presubmit answers: siblings (merc and pet are claims 4 and 5, a co-op partner is a Player: the same code); error-path at `walk_velocity.py` (a fixed 4-byte `struct.unpack` of a 32-bit argument).

Agent: Claude Code
