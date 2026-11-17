# Verifier: decomp39, a player's melee reach and line (2026-11-17)

> status: current, 2026-11-17 (card verifydecomp39, TASK-625, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/melee-reach-r1` (oracle `5ff84fb6`, claim `38426e2d`, on upstream `895f8a67`), body `docs/upstream/bodies/decomp39-melee-reach-body.md`, trace `docs/research/decomp39-melee-walls.md` (TASK-624). Checklist from `verifier-decomp35-2026-11-16.md`, `verifier-decomp36-2026-11-16.md`, `verifier-decomp37-2026-11-17.md`, `verifier-decomp38-2026-11-17.md` (their `agent/verifydecomp3*` branches). The fixed `fid/melee-reach-r2` is on the private fork remote only.

Tags: **V run** (run by this verifier today, the builder container), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r1 FAIL (guards, tidy), r2 PASS WITH CHANGES

Post **`fid/melee-reach-r2`**: oracle `60672f4e`, claim `39f9cda6`, on upstream main `895f8a67` (upstream main had not moved at the time of the check); body after B1 to B11 below. The rule is game.exe's and the claim implements it (re-derived below; 0 differ on 131,752 cases x 2 modes, three seeds). r1 cannot be posted as it is:

1. **pr_guard fails (high).** `tools/emu/player_reach.py:14` names `FUN_00549cbf`, which is a label inside `FUN_00549ba0` (V read: `0x549cbf` is the jump target after `ret 0x10`): `pr_guard: 1 FAIL` (V run). r2: `LAB_00549cbf`, `0 FAIL, 0 WARN`.
2. **presubmit and clang-tidy fail (high).** `short-names` FAIL on `px` (`test_player_reach.cpp`, both commits; also `py`, `mx`, `my`, `in`); clang-tidy-19 `misc-include-cleaner`: no header providing `d2d::rules::UnitType` in `fight/tick.cpp:269` (it reaches `unit_size.hpp` only through `melee_reach.hpp`). r2: renamed (`player_x` ..., `fields`), `#include <unit_size.hpp>`; presubmit 0 FAIL, tidy 0 findings on `fight/tick.cpp`, `fight/strike.cpp`, `test_player_reach.cpp` (positive control `readability-magic-numbers`: 3,858 reports, so the units are parsed).
3. **Usage errors (medium).** r1 `player_reach.py` with no argument, `gen` / `game` / `compare` short of arguments: `IndexError`; an unknown verb exits **0** with no output; extra arguments ignored; a missing file a traceback. `test_player_reach --player-reach` without a file runs the pins. r2: `usage:` and exit 1 on all of them, `cannot read PATH` on a missing file, `test_player_reach` refuses any unknown argument. `compare` itself was already right (exit 1 on an empty, unequal or malformed file and ids out of step; V run, hand-made files).
4. **A pin gap (medium).** A mutant that also starts the swing when the units are within 3.5 subtiles centre to centre (`m11`, through a wall at close range) passed r1's 16 pins; the oracle catches it (114 / 120 cases). r2 adds pins 17 and 18 (game.exe's own answers, V run: cases 3744 and 15136 of seed 1); `m11` aborts on pin 17.

Identity and history were already clean in r1 and stay so in r2 (V run): both commits by `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>` as author and committer, oracle first, `Agent: Claude Code` after a blank line, the only parsed trailer (`git interpret-trailers --parse`). r2's `components/` differ from r1's by the one include line only.

## Game side, re-derived (V read by this verifier)

- **`FUN_00622c40`** (`0x622c40..0x622cf1`): a monster target (`[unit] == 1`) whose `FUN_00463860` BaseId is 0x102 or 0x105 is in reach when `range + 8 > apart`; else `apart = FUN_00641530(unit, target)`, `apart <= 0` is in reach **with no line test** (`0x622cb5..0x622cb7`); else `range + extra + 1 >= apart` (`lea ebx, [eax + edx + 1]`, `0x622cac`) and then `FUN_00622aa0(unit, target, 0x804)` must be 0 (`0x622cd2..0x622ce4`). `range` is `FUN_00622870(unit)` of the attacker. Same as `rules::melee_reach`.
- **The blow, `FUN_0057ec10`** (`0x57ec7e..0x57eca6`): `extra = 1 + 2 x (attacker type == 1)`; when the caller's third stack argument `a` is not 0, `extra -= FUN_00622870(attacker) + a` (`0x57ec94..0x57ec9e`), which makes the test `apart <= base + 1 - a` whatever the weapon. **New:** of the 32 call sites (V run, a `.text` scan), 31 pass `a = 0` (30 literally, `0x56e6b3` through `FUN_0056e680`'s sixth argument, 0 from all five of its callers); `FUN_0056e6f0` passes 1, and its one caller is Blade Shield's area hit `FUN_005d7c40` (registered by `FUN_005d7ce0`; `skills.md` row 54). So for a player's Attack and every melee skill d2d routes through `Fight::hit`, extra is 1, as the claim has it. The trace's Q1.7 ("the third stack argument ... only adds to the range") is wrong: it subtracts. r2's `combat.md` note says so.
- **The swing start:** the client's `FUN_00480c80` (`0x480d06..0x480d1f`) walks on while `!FUN_00622c40(player, target, FUN_00622dc0(target))`; the server's `FUN_00549ba0` asks the same at `LAB_00549cbf` (`0x549cbf..0x549cc8`, true jumps to `0x549c9b`, the start). `FUN_00622dc0` -> `FUN_00622d00`: a monster in mode 2 or 0xf, or BaseId 0x6e in 8 / 9 (jump tables `0x622d98`, `0x622da4`, V run), as `rules::unit_moving`.
- **The player's size** is 2: `FUN_00620510`'s jump table `0x62057c` sends type 0 to `mov eax, 2` (V run). The oracle hooks the size; the claim's literal 2 is game.exe's.
- **A monster's blow** asks `FUN_00622c40(monster, target, 3)`: range (MeleeRng) + 4 and the same line (V read, the `sete` above). d2d's monsters, the merc and pets use `hypot` 1.1 / 1.4 cells with no line (`ai.cpp:1943`, `2491`, `2627`, `merc.cpp:242..283`, `pets.cpp:300`, `353`, `1111`, `1182`; V read); none is changed by this PR.

## Differentials (V run)

Game side `player_reach.py game` / `hit` (the emulator); seeds 1 and 7 reproduce the author's case and game files byte for byte (`cmp`); seed 200 is fresh. d2d side `test_player_reach --player-reach` / `--player-hit`, compared with `player_reach.py compare`.

| Tree | seed 1 (30,584) start / blow | seed 7 (70,584) start / blow | seed 200 (30,584, fresh) start / blow |
|---|---|---|---|
| r1 oracle `5ff84fb6` (parent d2d) | 2,799 / 9,869, exit 1 | 7,210 / 24,094, exit 1 | 2,835 / 9,868, exit 1 |
| r2 oracle `60672f4e` | the same, exit 1 | the same, exit 1 | the same, exit 1 |
| r2 claim `39f9cda6` | **0 / 0**, exit 0 | **0 / 0**, exit 0 | **0 / 0**, exit 0 |

Both sides are deterministic per case (the blow loop retries the to-hit roll with armor 0), so the fresh seed is the reseed. r2's test with its 18 pins built against the parent tree fails (`pin 1: swing 1 (game.exe 0)`), on the claim it passes.

## Mutants (V run, local, never pushed)

Each built as `test_player_reach` from the claim with one change; "pins" = the no-argument run (as ctest runs it); oracle = differing cases of 30,584, seed 1 / seed 200.

| Mutant | pins | start | blow |
|---|---|---|---|
| m1 `>` for `>=` in `melee_reach` (off by one) | abort, pin 1 | 1,125 / 1,151 | 1,160 / 1,193 |
| m2 mask 0x805 (walk-only words counted) | abort, pin 6 | 74 / 77 | 98 / 96 |
| m3 mask 0xc04 (object words counted) | abort, pin 7 | 84 / 85 | 95 / 109 |
| m4 the size-overlap test in `sight_blocked` removed | **passes** | **0 / 0** | **0 / 0** |
| m5 the player's size 1 | abort, pin 1 | 1,097 / 1,124 | 1,177 / 1,227 |
| m6 the line's ends not stepped in (sizes 0) | abort, pin 9 | 464 / 493 | 577 / 611 |
| m7 the blow's extra 3 (a monster's) | abort, pin 15 | 0 / 0 | 2,642 / 2,646 |
| m8 `lround` for `floor` (subtile of a position) | abort, pin 2 | 452 / 455 | 552 / 574 |
| m9 the tentacle rule off | **passes** | **0 / 0** | **0 / 0** |
| m10 the start's extra always 1 | abort, pin 1 | 577 / 609 | 0 / 0 |
| m11 also start within 3.5 subtiles centre to centre (through a wall) | r1 passes; **r2 aborts, pin 17** | 114 / 120 | 0 / 0 |

- **m4 is equivalent here, and the body's exception is mislabelled.** The line is asked only when `apart > 0`, and with these sizes no pair with `apart > 0` has `across + down < size1 + size2` (0 of 131,752 cases differ). The close-range exception the body and the trace describe ("sizes overlap, no line to cut"; pin 12 "3 off the units' sizes overlap") is `FUN_00622c40`'s `apart <= 0` branch (`0x622cb7`), which never asks the line. Same behaviour, wrong reason: B6.
- **m9 is not covered by anything in the PR**: the oracle's target is class 0x10 only. The tentacle and vulture wiring in `melee_reach_to` (`mon_bin[0x102]`, `[0x105]`, `[0x6e]`) is unproven for the player's swing: B7.
- With the author's eight (S, the trace section 7) that is 18 mutants; every non-equivalent one but m9 aborts r2's pins.

## Build, ctest and the other oracles (V run)

Builder container, Release, GCC, tests and tools on, `-j2` under the builder's lane lock, ccache; the chain registered with `tools/agentrun.sh`; `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` set for ctest.

| Tree | Build | ctest |
|---|---|---|
| r2 claim before the include fix (components = r1's claim) | full, rc 0 (602 targets) | **359 of 359 pass** (14 oracle-mode tests skip without arguments, as upstream); `test_player_reach`, `test_game`, `test_zeal_chain` Passed |
| r2 claim `39f9cda6` (one include more) | `test_player_reach`, rc 0 | pins rc 0; oracle 0 / 0 on all three seeds; tidy 0 |

Upstream main was not built: no test fails on the claim, so "the same failures as upstream" holds with none (the decomp38 verifier's upstream run on `ed007569`: 355 of 355, S). The oracle-mode tests that skip under ctest and touch combat were run with their oracles on the claim build: `next_hit` (seed 7, 200 runs: 487 lines, `diff` empty) and `armor_hit` (108,506 cases, 0 mismatches). Neither reaches `engage` or `hit` (V read); they pass.

## The big risk: does a fight stall? (V run, a lab-only harness)

A `--stall N SEED` mode added to `test_player_reach` locally (branches `vd39-stall-*`, never pushed): the REAL Blood Moor map with its own collision words, 400 random pairs of a free player spot and a free Fallen spot within 5.5 subtiles whose line a 0x804 word cuts; the Fallen held still (stunned), the player clicks it with Attack every tick through `World::apply` / `World::tick` for 20 s; also a BFS over free subtiles from the player's spot for a spot where game.exe's start test holds.

| | parent (r1 oracle tree) | claim |
|---|---|---|
| first swing started with the line cut (through the wall) | 392 of 400 | **0** |
| a blow landed within 20 s | 400 | 372 |
| ticks to the first blow: median / 90 % / 99 % | 7 / - / - | 13 / 32 / 75 (40 ms ticks) |
| no blow in 20 s | 0 | **28 (7 %)** |

- The 28 are all 4 to 5.4 subtiles apart; in **18 the player never moves**, in 10 a few subtiles; in every one the BFS finds 12 to 33 free spots in reach, the nearest 3 to 40 steps away. So d2d's walk to the unit (`player_walk` type 7 with the unit, `world/tick.cpp:670..700`) gives up where a reachable spot exists: the player stands clicking.
- Whether game.exe's client also stands there (its path to a unit is what d2d models) is **U**: a lab drill (the trace's section 9 drill 1) on game.exe at such a pair decides it. On the parent the problem was hidden: the swing went through the wall.
- **Asymmetry after the PR (I from V read):** d2d's monsters, merc and pets keep their wall-blind `hypot` reach (above), so in these pairs a monster can keep hitting a player who now cannot hit back (game.exe: neither can). That is a fidelity step in the right direction with a visible side effect until D5 lands.
- Charge (I, V read `tick.cpp:243`): within 1.1 cells Charge falls through to `melee_reach_to`; between game.exe's reach and 5.5 subtiles a Charge click now walks instead of charging. Whirlwind and Leap Attack are unchanged.

## Scope answers (the brief's a to f)

- (a) pathing: above; 7 % of close wall-cut pairs stall in d2d after the claim (V run), game.exe's behaviour there U.
- (b) monsters: game.exe applies `FUN_00622c40` to a monster's blow too (extra 3, V read); d2d's monsters are not changed: **the claim covers the player only**.
- (c) ranged: `engage` sends missile and spot skills to `cast_missile` before the reach test, `strike` returns before `hit` for missile skills and throws (V read `tick.cpp:238..240`, `strike.cpp:619..626`); the oracles `next_hit` and `armor_hit` pass on the claim (V run). Missiles and spells are not affected.
- (d) merc, pets: unchanged (`merc.cpp`, `pets.cpp` own `hypot`, V read); a co-op partner's player runs the same `Fight` code (I).
- (e) ctest: 359 of 359 on the claim; no test encoded the old reach.
- (f) `test_game`, `test_zeal_chain` and the combat oracles pass (V run).

## Required changes to the body (`decomp39-melee-reach-body.md`)

- **B1 (high).** Commits: "upstream `895f8a67` plus `60672f4e` (oracle) and `39f9cda6` (claim)"; 18 pinned cases, not 16 (13 in the trace).
- **B2 (high, template).** Add `## How to re-check` with the exact calls (`player_reach.py gen 20000 1 > cases.tsv`, `game` / `hit`, `test_player_reach --player-reach` / `--player-hit` with `D2_MPQ_DIR` and `D2_PATCH_INSTALLER`, `compare`, which exits 1 on any difference, an empty or unequal file, a malformed line or ids out of step; `ctest --test-dir build -R test_player_reach`), the `## Checklist` ticked (test `tests/game/test_player_reach.cpp`; finding `docs/research/re/combat.md`; pr_guard 0 FAIL 0 WARN) and the `Agent: Claude Code` footer.
- **B3 (high, claims limited to proven).** Limits, new first line: "game.exe asks the same test for a monster's blow (`FUN_0057ec10`, extra 3); d2d's monsters, merc and pets keep their `hypot` reach with no wall test, so until that is changed a monster behind a thin wall can still hit a player who cannot hit it back."
- **B4 (high, in play).** "In play": add "On the Blood Moor's own walls, a player 4 to 5.4 subtiles from a monster behind a wall corner sometimes does not walk round: in 28 of 400 such pairs d2d's walk to the unit gives up although a free spot in reach exists (a stall harness, not part of this PR); whether game.exe walks round there was not checked."
- **B5 (medium).** "What changes" 2: Charge is not untouched: within 1.1 cells its start now asks the new reach, so between game.exe's reach and 1.1 cells a Charge click walks first. Say it, or leave Charge's close start on the old test.
- **B6 (medium).** "What stays" and pin 12's comment: the close-range case is `FUN_00622c40`'s "unit distance 0 or less is in reach, the line not asked" (`0x622cb5..0x622cb7`), not the line's size-overlap rule (a mutant removing that rule changes 0 of 131,752 cases).
- **B7 (medium).** Limits: "The tentacle (BaseId 0x102 / 0x105) and vulture (0x6e) rows of `melee_reach_to` are not in the oracle (its target is class 0x10); a mutant turning the tentacle rule off passes every check."
- **B8 (medium).** Evidence row `FUN_0057ec10`: "`extra = 1 + 2 x (monster)`, less the range and the caller's third argument when that is not 0 (only Blade Shield's `FUN_005d7c40` passes one)".
- **B9 (low).** Counted differentials: add the fresh seed 200 row (start 2,835 / blow 9,868 on the parent, 0 / 0 on the claim) and the mutant table of this file (11 more; m4 equivalent, m9 survives).
- **B10 (low).** Answer presubmit's WARNs: `siblings` (unit kinds) with B3's line; `siblings` (hit paths) "missiles, spells and throws never reach the changed code (`engage` casts them first, `strike` returns before `hit`); Whirlwind, Leap Attack and Charge's blow keep their reach"; `doc-claims` "`combat.md` 224 / 227: 'always at 0 or less' and 'every caller but one' are read from the code, the 32 call sites listed"; `build` / `gcc` "built and ctest run (Release, GCC): 359 of 359 pass".
- **B11 (low).** Drop "The whole `ctest` was not run by the author ... the verifier's job".

## For the coordinator (not body changes)

- **Trace fixes** (`decomp39-melee-walls.md`): Q1.7's "only adds to the range" (it subtracts; B8); section 7's "13 pins" (16 in r1, 18 in r2); the D3 / pin 12 reason (B6).
- **Bot (S, opinion):** the stall above means a melee click on d2d after this PR can leave the player standing still, not swinging, while a monster behind the wall keeps hitting (D5). The bot's B1 (`fight.go:721`, click within the skill's range, no line test) needs a "no progress after a click" fallback (pick another target or reposition); worth a bot card once the PR lands.
- **Follow-ups:** D5 (monsters' and the merc's melee: the same test with extra 3) is the natural next PR and removes the asymmetry; a lab drill at one of the 28 stall pairs answers whether game.exe walks round (the pairs: lab file `stall.claim2.tsv`, map seed of the trace).

## Hygiene (V run, r2)

- Identity as above; `pr_guard.py upstream/main HEAD` 0 FAIL 0 WARN; `presubmit.py --quick --base upstream/main` 0 FAIL, 5 WARN (B10); ruff clean.
- Diff: oracle 3 files (`player_reach.py`, `test_player_reach.cpp`, a README row in the checks table), no `components/` change, its test has no failing assert (bisect-safe: no-argument run prints and exits 0); claim 5 files (`fight.hpp`, `fight/strike.cpp`, `fight/tick.cpp`, `combat.md`, the test). No private code, host names or paths (read of the diff).

## Not checked

- No Windows, macOS or Clang build; no upstream-main ctest (none fails on the claim); no live game, no lab drill; nothing on the worker host.
- game.exe's behaviour at the 28 stall pairs (U); the tentacle and vulture rows; Whirlwind, Charge and Leap Attack's own landing (D4); which packets set `FUN_00549ba0`'s flag.

## Re-check commands (as run)

Builder container, the emulator image of `docs/emu-image.md`; lab scripts `verifyd39/{bld,chain2,chain3,chain4,stage,emu,emuruns,fn,xr,pre}.sh`, `py/{pre,imm,dw}.py`, `cpp/stall_block.cpp` (lab only).

```sh
python3 player_reach.py gen 20000 200 > c.tsv
python3 player_reach.py game c.tsv > g.tsv; python3 player_reach.py hit c.tsv > h.tsv
test_player_reach --player-reach c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > d.tsv
test_player_reach --player-hit c.tsv | grep -E '^[0-9]+ -?[0-9]+$' > dh.tsv
python3 player_reach.py compare g.tsv d.tsv; python3 player_reach.py compare h.tsv dh.tsv   # claim: 30584 cases, 0 differ (x2); oracle commit: 2835 / 9868, exit 1
test_player_reach; echo $?                                                                  # claim: OK, 0; r2's test on the parent tree: pin 1, 1
```
