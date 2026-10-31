# PR candidate proof: walk velocity (monster RN, chill, floor, velocitypercent)

> status: local candidate, fixed after the verifier (card walkfix, 2026-10-31): branch `fix/walk-velocity-v2`, five claims plus a tooling commit, own reviewer-runnable differential on the real d2d builds. Nothing pushed or posted. The tables below describe the original three claims (card decompwalk) and the verifier's report; the fix pass is the last section. Trace: docs/research/walk-speed-trace.md.

## Claims (one per commit, branch `fix/walk-velocity` on upstream d0ea14b7, worker host `trace/walk` worktree)

| commit | claim | game.exe evidence (V) | d2d before (V-d2d) |
|---|---|---|---|
| 96953bfe | a monster in RN moves at MonStats Velocity x walk %, not Run | `00623f50` stores `00621360` x % / 100 at path +0x7c for every moving mode (`006214a0`: WL, RN); `00621360` reads MonStats +0x32 for a monster whatever the mode; the loader names +0x32 Velocity (descriptor at 0x65161a) and +0x34 Run (0x651640); +0x34 is read only by pace AIs and charges | ai.cpp:2727 `cells_per_sec(type_info.run)` for RN |
| 7e8feb87 | chill adds MonStats coldeffect to velocitypercent; the walk % is at least 25 | `0057af80` puts coldeffect (MonStats byte +0x168 + difficulty; -50 for a player) on stat 0x43 as a stat list; `00623f50` 0x62414e floors the sum at 0x19 | ai.cpp:2724..2726 multiplies by (100 + coldeffect) / 100, floor 10 |
| 97d55732 | a player's velocitypercent adds as it is; only FRW is diminished, for any v != 0 | `00623f50` adds stat 0x43 raw to `00621740` row 4 (stat 96 only); `00621740` skips only v == 0 | strike.cpp:185..186 folds velocitypercent into frw; combat.hpp:170 leaves v < 0 undiminished |

## Differential (author's, decompwalk)

`tools/emu/walk_speed.py gen|game|cmp` + `tools/emu/walk_speed_harness.cpp`: game.exe's code from 0x624113 to 0x6241ce under unicorn (upstream tools/emu/emu.py, data tables built from the patched MonStats / MonStats2), d2d transcribed line for line before and after. Exhaustive grid (see the script header): 712,362 cases, 425,665 differ before (player 208,586, monster 217,079), **0 after**. Per term: M1 34,190; M2 124,301; M3 58,588; P1 207,466; P2 1,120. Tolerance: 1 unit of 1/4096 subtile a tick (float vs integer truncation).

## Limits

- The d2d side is a transcription (as the hit_len template), not d2d's binary; the patched code compiles and upstream's tests pass (evidence pack).
- What feeds stat 0x43 (boss speed +20, pace, auras) is d2d's model, taken as given; the claims are only about how the sum becomes a velocity.
- Pets (pets.cpp:217 `max(run, velocity)`) and mercs (merc.cpp:287) move by their own code; not checked, not changed (the verifier's finding 1; answered in the fix pass below).
- frw == -150 would divide by zero in game.exe; the patch guards it (unreachable).

## Independent verifier (card verifyprs, 2026-10-31)

Verdict: **HOLDS WITH LIMITS** for all three commits (no counter-example inside the claims); the branch is NOT READY to open until limits 1 and 2 are answered.

What was done, without reading the author's oracle first:

- Branch replayed onto current upstream main 72132b99 in a throwaway worktree: the three commits apply cleanly (5 files, +23 / -15).
- Rule re-derived from disasm.py reads: FUN_00623f50's moving-mode branch (FUN_006214a0 true, a path at unit +0x2c): pct = FUN_00621740(unit, row 4) + stat 0x43, at least 0x19 (0x62414e); velocity = FUN_00621360(type, class) x pct / 100 (signed, truncating) to FUN_00648690. FUN_00621740's table at 0x6e8e24 has 12-byte rows {flag, param, stat}: row 4 = {1, 150, 96}, diminished as param x v / (param + v) (idiv) for any v != 0, negative included. FUN_00621360: a monster (type 1) returns MonStats +0x32 << 8 whatever the mode, a player (type 0) CharStats +0x40 (WalkVelocity) << 8 whatever the mode. FUN_0057af80: coldeffect (MonStats +0x168 + difficulty; -50 when the unit is no monster) goes on stats 0x43, 0x44, 0x45 as one stat list, then FUN_00623f50 runs. Same rule as the author's (V, disassembly read).
- Own differential (`tools/emu/verify_walk_velocity.py`, seed 9191): game side = game.exe's own FUN_00623f50 with FUN_00621740 and FUN_00621360 native against fake data tables at [0x744304] (hooks in the script docstring), plus a self-check that the velocity it hands FUN_00648690 equals base x pct / 100 for the pct read from the disassembly (0 failures in range); d2d side = the REAL build: one frame of `monster_update` on a monster in WL or RN with a far spot in an all-open copy of the Blood Moor (the distance moved over Velocity's step gives the walk %), and the real `World::walk_percent(Player, now)` (`verify_walk_velocity_mode.inc`). 121,216 cases (26 traced-AI MonStats rows, both modes, 3 difficulties, sums around the old floor 10 and the new 25, laid, chilled with data and synthetic coldeffect -100..100, Velocity 1..60; all 7 classes, walk / run, chilled, faster run/walk -2000..30000 (not -150), velocitypercent -400..400): 21,216 boundary + 100,000 random. Before (upstream 72132b99 built): **70,171 differ** (monster WL 20,456, monster RN 31,931, player 17,784). After: **0** (tolerance 0.01 percent for d2d's float step).
- CI class (on the replayed branch): pr_guard 0 FAIL 0 WARN for the range and for each commit; test_game / test_rules / test_combat / test_protocol rc 0; presubmit --quick 0 FAIL, 4 WARN (siblings: merc, pet not touched; test-gap; docs-sync; build skipped); gcc -Wshadow -Wunused -Wsign-compare on ai.cpp, strike.cpp, tick.cpp: 0 on added lines; clang-tidy (upstream .clang-tidy, names and includes) on the same TUs: 0 findings. One claim per commit; the commit-2 ponytail (laid creeps at the 25 % floor) is accurate.

Limits and findings:

1. **Same rule, unpatched siblings** (presubmit's siblings WARN): mercs and pets are monster units in game.exe, so FUN_00621360 gives them MonStats Velocity in RN too, but d2d still moves them at Run: `merc.cpp:286-287` (RN at Run, pace floor 10), `pets.cpp:217` (`max(run, velocity)`), `pets.cpp:270-280` (RN at Run; pace floor 10, multiplied; a Shadow in RN takes CharStats RunVelocity as its base, while FUN_00621360 returns WalkVelocity (+0x40) for a player class in any mode and a player's run comes from FUN_00620e80's stat list on 0x43 (RunVelocity x 100 / WalkVelocity - 100): whether a Shadow gets that list is not traced, so the Shadow case is UNVERIFIED). Fix: either extend the claim with its own commits and proof, or add a `// ponytail:` at `const float run = walk;` naming these paths and a line in the PR body.
2. **No reviewer-runnable d2d side in the branch**: the author's after side is a transcription in our repo (walk_speed_harness.cpp), not upstream's binary. Add a tooling commit (a `test_game` mode as for the other PRs, or this verifier's mode) or attach the harness and say so in the body.
3. Skill moves set their own path velocity in game.exe (other FUN_00648690 callers: 0x5cbbb5, 0x5cd47c, 0x5cf86e, 0x5d90a6, 0x5d9ffe); d2d's `ai.cpp:1883-1885` (Diablo's run) and `1937-1939` (the whirl) keep a multiplied chill and a 10 floor. Not traced, outside the claim.
4. `rules::effective_speed` (combat.hpp:171) still leaves a negative value undiminished for the IAS / FHR / FCR / FBR rows of the same FUN_00621740 table (rows 0..3: {1, 120, stat}); the patch only fixed faster run/walk. A separate candidate.
5. game.exe truncates base x pct / 100 to an integer velocity; d2d's float step does not (54,981 of 86,785 monster cases have a remainder; under 1/4096 subtile a tick). Velocity x 256 x pct wraps past 2^31 in game.exe (unreachable with 1.14d data).
6. What feeds stat 0x43 (pace, boss speed, auras, a running player's Run x 100 / Walk) is d2d's model, taken as given on both sides.
7. Det state hashes not run (needs a d2d slot; moving monsters will change them by design).
8. The branch sits on d0ea14b7; rebase onto upstream main before opening (clean on 72132b99).

## Fix pass (card walkfix, 2026-10-31)

Branch `fix/walk-velocity-v2` (worker host, own worktree; the author's `fix/walk-velocity` is untouched) on upstream main 72132b99 (fetched again, still the tip). Six commits, author and committer the user's noreply identity, body `Agent: Claude Code`, no trailer, one claim each:

| commit | claim | d2d before (upstream 72132b99 or the step before) |
|---|---|---|
| d8e63d4a | a running monster moves at MonStats Velocity | ai.cpp RN at Run |
| a5cf6415 | chill adds coldeffect to velocitypercent; floor 25 | multiplied chill, floor 10 |
| 1c3fa4e5 | a player's velocitypercent adds as it is | folded into frw |
| 455e366f | NEW: a merc walks at MonStats Velocity in WL and RN, walk % floor 25 | merc.cpp:286 RN at Run, floor 10 |
| 85d50c78 | NEW: a pet walks at MonStats Velocity, walk % = sum of move, speed_pct and aura velocitypercent, floor 25 | pets.cpp:270 RN at Run, product of two factors, floor 10 |
| 10338345 | tooling: `test_game --walk-velocity`, `tools/emu/walk_velocity.py`, README row | none |

Why the merc and pet claims are proven and not a ponytail: FUN_00621360 returns MonStats +0x32 for any monster unit in any mode and FUN_00623f50 sums and floors, for every moving unit (V, verifier's disassembly reads above); a merc and a pet are monster units in game.exe, and their only d2d-modelled inputs (merc move, pet move / speed_pct / aura velocitypercent) are the same stat 0x43 feed as for monsters, taken as given. The existing differential run on game.exe's own FUN_00623f50 is therefore the oracle for them too: only the d2d side is new.

Not claimed, ponytails in the code and a line in the PR body: (a) a Shadow's base (`pets.cpp` gfxtype 2 branch keeps CharStats Run / Walk velocity: FUN_00621360 gives a player class WalkVelocity in every mode, whether a Shadow's run has the stat list on 0x43 is not traced, UNVERIFIED); the oracle only has non-Shadow pets; (b) `pets.cpp:217`, the chase of an untraced pet, keeps `max(Run, Velocity)` and no floor (no oracle case: its speed is only an approximation of the published behaviour).

### Reviewer-runnable differential (finding 2)

`tests/game/test_game.cpp --walk-velocity CASES` calls the REAL d2d functions: `monster_update` (one frame, monster in WL or RN walking a far spot in an all-open Blood Moor copy), `World::walk_percent` (player), `Fight::merc_turn` (a merc spawned by `World::spawn_merc` from the header), `Fight::pet_walking` (a `Fight::Pet` built in the test). The walk % is the distance moved over Velocity's step. `tools/emu/walk_velocity.py gen|game|compare` is the oracle (game.exe's FUN_00623f50 under unicorn; hooks listed in its docstring); commands are in the script header and in `tools/emu/README.md`. Copies in our repo: `tools/emu/walk_velocity.py`, `tools/emu/walk_velocity_mode.inc`.

Run (worker host, seed 9191, `gen 100000`): 122,646 cases (monster WL 30,867 and RN 30,912; player 29,466; merc WL 7,962 and RN 7,697; pet WL 7,853 and RN 7,889; every case compared, none skipped, velfail 0). Mismatches against game.exe, at tolerance 0.01 percent of the walk %:

| build | monster | player | merc | pet | total |
|---|---|---|---|---|---|
| upstream main 72132b99 (+ the mode) | 38,543 | 15,146 | 8,287 | 7,461 | **69,437** |
| after commits 1 to 3 (d8e63d4a, a5cf6415, 1c3fa4e5) | 0 | 0 | 8,287 | 7,461 | 15,748 |
| after the merc commit 455e366f | 0 | 0 | 0 | 7,461 | 7,461 |
| after the pet commit 85d50c78 = branch head | 0 | 0 | 0 | 0 | **0** |

Per claim, before to after: claims 1 and 2 (monster, WL and RN): 38,543 to 0 (WL 15,371, RN 23,172); claim 3 (player): 15,146 to 0; merc claim: 8,287 to 0 (WL 815, RN 7,472); pet claim: 7,461 to 0 (WL 1,968, RN 5,493). The before builds are upstream main and each earlier commit with the same mode file (the player lines of the mode use `frw + velocity_pct` where the member does not exist yet). The case file is reproducible (`gen` twice, `cmp` equal).

### Findings of the verifier, one by one

1. Siblings: fixed for mercs and pets (commits above); the Shadow base and the untraced pet chase are ponytails and PR-body lines; co-op partners are Players (the same `walk_percent`).
2. Reviewer-runnable d2d side: the tooling commit above.
3. Skill moves (ai.cpp ~1883, ~1937, Diablo's run and the whirl) keep a multiplied chill and a 10 floor: untraced, outside every claim; recorded in the draft limits.
4. `rules::effective_speed` leaves a negative IAS / FHR / FCR / FBR undiminished (FUN_00621740 rows 0..3): a separate candidate, not in this branch; recorded.
5. game.exe truncates velocity to an integer, d2d's float step does not: 49,810 of the cases here have a remainder (54,981 in the verifier's run), under 1/4096 subtile a tick: stated limit.
6. What feeds stat 0x43 is d2d's model, both sides: stated limit.
7. Det state hashes: NOT run (needs a d2d slot, none allowed in this card); moving monsters, mercs and pets change them by design.
8. Rebased: upstream main is still 72132b99; the three original commits replayed with no conflict (new hashes d8e63d4a, a5cf6415, 1c3fa4e5, same content).

### CI-class checks on the branch head (10338345), worker host

- `tools/pr_guard.py upstream/main`: 0 FAIL, 0 WARN; also 0 / 0 for each commit against its parent.
- `tools/presubmit.py --quick`: 0 FAIL, 3 WARN, answered: siblings (merc is patched in commit 455e366f, a co-op partner is a Player; the detector does not see either), error-path at `walk_velocity.py:163` (a fixed 4-byte `struct.unpack` of a 32-bit argument: no length to check), build skipped by `--quick` (done below). The docs-sync warning was answered with a row in `tools/emu/README.md`. ruff (upstream's rule set F,E9,B,BLE001,PLW1510,S602,S604) clean after fixing B023, B007, F841 in the first draft of the script.
- clang-tidy with the repo's `.clang-tidy` (identifier-length, include-cleaner) over the 135 TUs `tools/tidy-pr-select.py` picks for the changed files: no finding in any touched file. The local clang-tidy is newer than CI's and reports structured-binding `x`, `y` names in untouched files (outdoor.hpp, merc.hpp, pets.hpp, setup.cpp, test_monster_ai.cpp, apps/d2d); they are upstream code, not our lines.
- gcc `-Wshadow -Wunused -Wsign-compare -Wextra` over every touched TU (compile_commands flags): one finding on our lines at first, `rng` in the new test mode hid an upstream local of `main` (the MSVC C4456 class that failed two earlier PRs); renamed `case_rng`, now 0 on added lines.
- Built with `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`: `test_game` (with the game data dir), `test_rules`, `test_combat`, `test_protocol` all rc 0; `d2d` links.

