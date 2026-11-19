# Verifier: Crushing Blow D6, the crush is not in the get-hit's size (2026-11-19)

> status: current, 2026-11-19 (card verifycrushd6, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/crush-d6-r1` (oracle `e177207f`, claim `a35db81c`, on upstream `da98b7f8`), body `docs/upstream/bodies/crush-d6-body.md`, trace `docs/research/decomp42-crush-d6-trace.md` section 1.

Tags: **V run** (run by this verifier today, the builder container), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r1 FAIL (stale base, escapes, an untested call site), r3 PASS WITH CHANGES

The rule is game.exe's, for every path through the damage handler (re-derived below), and r1's code change implements it for every player-side blow (melee, missile, kick, hireling, pet: all reach `Fight::land`). r1 should not be posted as it is:

1. **Stale base (high).** r1 is on `da98b7f8`; upstream main is `d5de50ca` (15 commits on). It rebases with no conflict; the posted commits must be the rebased ones, built and tested there.
2. **The oracle and the pins miss three plausible mutants (high).** r1's own m8 (a missile blow carries no crush) and m9 (`marked` dropped) pass its pins and its compare, and so does putting the old inline `soak` + `hurt` back in `Fight::land` (the d2d side called `blow_hurt`, one step short of the player-visible path: the decomp37/38 and D5 lesson again). A fourth, m11 (the crush taken off the life too, to keep the size right), passes too: r1's compare never reads the monster's life. r2 measures through `Fight::land`, adds missile cases (FUN_0057c6c0's second argument 1, the record flag 0x20, event 6 and its doubled divisor, the sum run by game.exe itself at `0x57c7a0`), fire cases (an element marks the blow: FUN_0057ce30 run natively) and the life after the blow, in the oracle commit; the claim pins sixteen game.exe decisions (get-hit, seed and life) through `Fight::land`. All 13 non-equivalent mutants abort r2's pins and differ on its compare.
3. **The trace's "missile path not run" (D4) and "element not run" (D3) limits are closed by r2**, and the controls (no crush, both trees) are exact on the new missile and fire cases: 0 of 103,520 on three seeds.
4. **Sibling paths (medium, body).** A monster's MonProp crush on a monster or on a pet is still its own `hurt` call whose whole size is the crush (`fight/tick.cpp:459`, `:517`): game.exe never sizes a crush. Not this PR's to fix; the body must say so (the presubmit `siblings` WARN). Life Tap heals from `blow.damage` (crush and elements in it); game.exe's func 5 reads the record's physical only: a separate finding.

**Post `fid/crush-d6-r3`**: oracle `f2a7007c`, claim `8867eb7f`, on upstream main `d5de50ca` (pushed to the private fork remote today, V run `git ls-remote`); author and committer `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>`, oracle first, `Agent: Claude Code` the only parsed trailer after a blank line in both (V run, `git interpret-trailers --parse`). The code change (`ai.cpp`, `ai.hpp`, `strike.cpp`, `combat.hpp`, `combat.md`) is r1's byte for byte; only the oracle, the test and the two commit messages changed. `fid/crush-d6-r2` (`629e3126`, pushed earlier today) is the same tree before four clang-tidy fixes in the test (a short name, three includes); it is superseded, not to be posted (no force-push, so r3 is a new branch). The tables below were measured on r2's trees (same code; the test differs only in names and includes); on r3 the parent and the claim were re-run on seed 7 with the same counts (65,145 and 0, controls 0), the pins pass and ctest's `test_crush_get_hit` passes (V run). r3 also applies cleanly on top of D5 (`fid/crush-d5-r2`) (V run; combined build below). The body then needs C1 to C8.

## Game side, re-derived (V read)

- **The crush handler writes no record total.** `FUN_005bffc0` (`0x5bffc0..0x5c0127`, MonProp/item event func 16, table `0x7325f0`): the seed roll against stat 136, the divisor (10 player, 10 hireling, 4, 8 for a MonStats boss row or `FUN_005a0180(unit, 2)`, plus the player share), doubled for event 6 (`0x5c00a2..0x5c00a8`), `stat 6 / divisor` (`idiv`, 256ths), less `MulDiv(crush, min(stat 0x24, 100), 100)` when stat 0x24 > 0, stat 6 written by `FUN_00627260` (`0x5c00f6`), the record's kill bit `or word [rec+4], 2` when the life is gone (`0x5c0106`), `FUN_00621e40(unit, 0x93, 0)` when it took anything. Its only record write is the kill bit: never `+0x4c`.
- **The total.** `FUN_0057c1e0` `0x57c3a5..0x57c3c7`: `+0x4c = +8 + 0x10 + 0x1c + 0x20 + 0x24 + 0x28`, plus `+0x38` only when the attacker is a type-1 unit that is not a hireling (`0x57c214..0x57c228`); the resistances (`0x57c378..0x57c3a3`, rows at `0x732980`) and event 0xb (`0x57c2fc`) before it.
- **Every path through the damage handler sums before the strike events.** `FUN_0057c6c0` has two callers. `0x56ba73` (`FUN_0056b9c0`, the record path: `FUN_0057dfb0`'s hit result into the record flags, then `FUN_0057c6c0(..., 1, rec)`, then `FUN_0057cee0`): with the second argument 1, `FUN_0057c1e0` runs at `0x57c7a0`, then events 6 and 2 only when the record's flag dword has 0x20 (and 6 only without 0x80). `0x57d5c7` (`FUN_0057d4f0`): copies a record stored on the attacker (the list at unit `+0xac`, matched by both units' type and id), sets its flag dword to 0x20 and calls with 0: no sum in the handler, the stored record carries its total, then events 5 and 1 (`0x57c7d8..0x57c7f3`). Then, in both, the physical cut to the life left (`0x57c7f8..0x57c820`, after the total), the leech (`FUN_0057c420`), and at `0x57c852` the total off stat 6 (under `0x100`: 0). So whatever unit carries the Crushing Blow (player, hireling, pet, monster MonProp) and whatever the blow (melee record, missile record), the crush is out of `+0x4c`, which is what `FUN_0057cb00` sizes. Which code stores the melee record on the attacker was not read (U); it cannot hold a crush that has not run yet.
- **Life Tap reads the physical only.** Func 5 (`0x5c61e0`, table `0x7325f0`) reads the record's `+8` (`0x5c6227`, `0x5c62d2`) at event 1 or 2, before the cut at `0x57c7f8`: not the crush and not the elements (see Findings beside the PR).

## d2d side: every hit path (the brief's a)

V read on r1 / r2 (`git grep` of `player_blow(`, `hurt(`, `crushed`):

| Path | How the blow reaches the monster | After the claim |
|---|---|---|
| player melee, kick, Telekinesis (`fight/strike.cpp:707`, `:951`, `:959`) | `player_blow` -> `Fight::land` -> `blow_hurt` | fixed |
| player missile (`fight/missiles.cpp:891`) | `player_blow` with `swing.missile` (divisor doubled), then `missile_blow` adds to the same `Blow` (`crush_fine` kept) -> `land` | fixed (r1 not tested; r2 tests it) |
| hireling melee and missile (`fight/deaths.cpp:484`, `fight/tick.cpp:365`) | `player_blow` -> `land` | fixed |
| pets (`fight/pets.cpp:304`, `:1019`, `:1104`, `:1125`, `:1184`) | `player_blow` -> `pet_strike` -> `land` | fixed |
| a monster's MonProp crush on a monster (`fight/tick.cpp:458..459`) | the crush is its OWN `hurt(..., crush)` call after `take_blows`: class 0, unmarked, size = the crush | **still sized by the crush alone**: a separate get-hit test whose whole size is the crush (game.exe: none, the crush is never in a total). Not in this PR |
| a monster's MonProp crush on a pet (`fight/tick.cpp:516..517`) | `pet_hurt(pet, crush)` -> `hurt` | **same as the row above** |
| a monster's crush on the hireling (`fight/tick.cpp:690..693`) | added to `side[1].damage`; the size test is `struck_big(side[1].blows)` per blow | not sized (right) |
| a monster's crush on the player (`fight/tick.cpp:831..838`) | life only; the size test runs on `foe.blows` | not sized (right) |
| thorns, traps, burning ground (`ai.cpp:2578`, `world/objects.cpp:775`, `fight/deaths.cpp:73`) | `hurt` without a crush | unchanged |

Other users of the size: knockback (`Fight::knock_back`) and `struck_event` do not read it; the stun (`stun_until`) does not; `rules::struck_big` (player, hireling) is per monster blow. Nothing else read the old crush-inclusive size (V read). **Life Tap** (`fight/strike.cpp:549..551`) heals `blow.damage` x calc1 %, and `blow.damage` carries the crush's whole points and the elements; game.exe's func 5 heals from the record's physical `+8` only (V read above): a sibling deviation the claim neither fixes nor worsens.

## Rounding, 256ths, the max-life reference (the brief's b)

- The size stays in 256ths end to end: `player_blow` adds `crushing_fine` (256ths) to `frac`/`damage` and keeps the same value in `crush_fine`; `blow_hurt` hands `hurt` `taken` (256ths, Bone Armor's pool off the physical only) and `crush_fine`; `hurt` takes it off `(damage << 8) + fraction` before the carry from earlier blows joins. So `taken - crush` is the record's `+8 + 0x10 + ...` after event 0xb, and the crush's part under one point does not leak into the size (pin 8: a crush of 12.5 points with a blow one 256th under max / 16 does not roll; m05, the crush cut to whole points, differs on 8,928 of 103,520) (V run).
- The floor at 0 is never reached: `taken >= crush`, because soak takes at most the physical (and a pet's Weaken cut keeps the damage at or above the crush, `fight/deaths.cpp:508..509`, V read); m06 builds to the claim's binary byte for byte (equivalent).
- The max-life reference and the divisor by class are `rules::get_hit`'s, unchanged (`monster.max_life_fine()`, 256ths; S: `combat.md`, `hit_rule.py`); the oracle runs classes 0 to 13 and random ones up to `0x5f`, marked and unmarked (V run).

## Interactions (the brief's c)

- **Leech #587:** unchanged: the leech still cuts the physical to the life the crush left (`after_crush`, `combat.hpp`), game.exe's `0x57c7f8..0x57c826`, after the total (V read). No leech in the cases (U for a leech-and-crush case; the leech has its own oracle).
- **Gear absorb #606, merc absorb #632 (open):** monster blows on the player and the hireling; this PR touches neither path (`Fight::hurt_side`, `struck_big` per monster blow) and shares no lines with them (V read the diffs).
- **Crush D5 (`fid/crush-d5-r2`):** both touch `components/game/ai.hpp` (D5 `Monster::crush_divisor`, D6 the `hurt` declaration) and `tools/emu/README.md` (adjacent rows). r1 and r2 cherry-pick on top of D5 r2 with no conflict (V run). The D6 cases are plain (4) or superunique (8), divisors D5 does not change. Combined tree: below.

## Differential (V run, builder, r2 trees)

`crush_hit.py gen 100000 SEED` (103,520 cases with the boundary rows; about 30 percent missile, 30 percent with fire), `game` once per seed (and `--no-crush`), the d2d side from each tree's `test_crush_get_hit --crush-hit`, `compare` on get-hit, seed and life:

| Tree | seed 7 | seed 5001 | seed 9091 | control (no crush), each seed |
|---|---|---|---|---|
| parent (r2 oracle commit `8fb5d1e4`) | 65,145 differ (get-hit 38,601, seed 52,109, life 0), exit 1 | 65,544 (38,613 / 52,555 / 0) | 65,316 (38,696 / 52,353 / 0) | 0, exit 0 |
| r2 claim `629e3126` | 0, exit 0 | 0 | 0 | 0 |

Every parent get-hit difference is d2d staggering where game.exe does not. Play grid (the author's 4,890 surviving procs, melee, no fire): game.exe staggers 1,921, the parent 4,513, the claim 1,921 with 0 differ: the author's numbers reproduced. The sixteen pins are `crush_hit.py game` on their cases. The compare's exits on empty, short, malformed and missing files: the author's runs (S); the new width check exits 1 when the d2d side is empty (V run: the first d2d run here printed nothing and every compare exited 1).

## Mutants (V run, builder, local, never pushed)

One text change each on the r2 claim (`mutb.py`): a `.cpp` mutant is rebuilt as the test target; a header mutant recompiles the test's unit and links it first. Pins = `test_crush_get_hit` with no argument (16 pins); compare = seed 7 (103,520). r1 column: the author's run (S) or by construction (I: r1's d2d side calls `blow_hurt`, prints no life, and has no missile or fire case).

| Mutant | r2 pins | r2 compare: get-hit / seed / life / total | r1 |
|---|---|---|---|
| m01 the parent: crush in the size | abort | 38,601 / 52,109 / 0 / 65,145 | caught |
| m02 `crush_fine` never set | abort | 38,601 / 52,109 / 0 / 65,145 | caught |
| m03 `blow_hurt` passes 0 | abort | 38,601 / 52,109 / 0 / 65,145 | caught |
| m04 the crush taken off twice | abort | 13,948 / 29,223 / 0 / 30,965 | caught |
| m05 the crush cut to whole points | abort | 3,582 / 8,438 / 0 / 8,928 | compare only (its pin was added after) |
| m06 no floor at 0 | passes | 0 (binary equal to the claim's) | equivalent |
| m07 `crush_fine` halved | abort | 22,067 / 39,211 / 0 / 43,975 | caught |
| m08 a missile blow carries no crush | abort | 7,931 / 13,690 / 0 / 15,479 | **escapes** |
| m09 `marked` dropped | abort | 3,922 / 8,631 / 0 / 9,523 | **escapes** |
| m10 `Fight::land` back to the inline soak and hurt | abort | 38,601 / 52,109 / 0 / 65,145 | **escapes** (I) |
| m11 the crush taken off the life too (size kept) | abort | 0 / 0 / 103,520 / 103,520 | **escapes** (I) |
| m12 only the crush's part under a point in `crush_fine` | abort | 35,078 / 48,568 / 0 / 59,542 | caught (I) |
| m13 a missile's crush halved in `crush_fine` | abort | 4,138 / 8,679 / 0 / 9,380 | **escapes** (I) |
| m14 a crush marks the blow | abort | 5,414 / 12,931 / 0 / 13,794 | **escapes** (I) |

The first mutant pass linked each header mutant with the object the previous `.cpp` mutant left (a restored source rebuilds only at the next build); the header rows are a clean second pass.

## Build, ctest, tidy (V run)

Builder container, Release, GCC, tests and tools on, `-j2`; three jobs, each one long-lane lockrun registered with `tools/agentrun.sh`; `D2_MPQ_DIR` and `D2_PATCH_INSTALLER` set.

| Tree | Build | `ctest -j2` |
|---|---|---|
| parent (the oracle commit's code; its test had a wrong assert, fixed before r2) | full, rc 0 | 407 of 408 pass; fails `smoke_d2d` |
| r2 claim `629e3126` | full, rc 0 | 407 of 408 pass; fails `smoke_d2d` |

`smoke_d2d` fails on both trees, as on upstream `d5de50ca` (`tests/smoke_d2d.py:444`, S: `verifier-crushd5`): not this PR's.

clang-tidy-19 (the repo's config, warnings as errors): `fight/strike.cpp` 0, `ai.cpp` 0; `test_crush_get_hit.cpp` on r2 had 4 errors (`hc` too short, no direct include for `Boss`, `kMaxLife`, `kLife`; the first two were in r1's test too), fixed in r3's oracle commit: 0. Positive control: an unused `#include <map>` injected makes tidy fail (rc 1, `misc-include-cleaner`), restored after.

**D5 + D6 together** (r2 cherry-picked on `fid/crush-d5-r2`, `e2e6886f`, local only): the four crush and combat tests build and pass (`ctest -R` 4 of 4); the D6 compare on seed 7: 0 of 103,520 (control 0); the D5 oracle (`crush_div.py`, the D5 verifier's 29,360 cases): 0 differ (V run).

## Hygiene (V run)

- `pr_guard.py upstream/main HEAD` on r3: 0 FAIL, 0 WARN. `presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN: `siblings` (monster, not player, merc, pet, co-op partner: answered by C5), build and gcc skipped by `--quick`.
- Diff r3 against upstream: 8 files (`ai.cpp`, `ai.hpp`, `fight/strike.cpp`, `rules/combat.hpp`, `combat.md`, `test_crush_get_hit.cpp`, `crush_hit.py`, `tools/emu/README.md`); no private code, host names or paths (read). No `todo.md` / `unverified.md` row names it (V read).

## Required changes to the body (`crush-d6-body.md`, becomes the r3 body)

- **C1 (high).** Branch `fid/crush-d6-r3`; "Two commits on upstream `d5de50ca`: the oracle (`f2a7007c`), then the claim (`8867eb7f`)".
- **C2 (high).** What changes: "`test_crush_get_hit` pins sixteen game.exe decisions (get-hit, seed and life) through `Fight::land`: melee, missile, fire". Counted differentials: cases are melee or missile, with or without fire; for a missile game.exe's side runs `FUN_0057c6c0` with its second argument 1 (it sums the total itself at `0x57c7a0`, then events 6 and 2); the d2d side is `player_blow` then `Fight::land`; `compare` checks the life too. Replace the table with this file's (65,145 / 65,544 / 65,316 of 103,520; claim 0; control 0).
- **C3 (high).** Mutants: "14 mutants (the verifier's table): 13 abort `test_crush_get_hit` and differ on the compare; the one that does not (no floor at 0) builds to the same binary". Drop the two "not caught" mutants and "Both test binaries print ok; ... the whole-tree ctest has not run".
- **C4 (high, template).** Add the Checklist: a test in `tests/` (`test_crush_get_hit`); the finding in `docs/research/re/combat.md`; `pr_guard` passes; "`ctest` with the game data: 407 passed, 1 failed; fails on main too: `smoke_d2d`"; labels `oracle`, `test`, `trace`.
- **C5 (medium, Limits).** Replace the bullets with: "A monster's MonProp Crushing Blow on another monster or on a pet is still its own `hurt` call sized by the crush alone (`fight/tick.cpp:459`, `:517`); game.exe sizes no crush in any path; not changed here." "Which code stores the melee record on the attacker before `FUN_0057d4f0` hands it to `FUN_0057c6c0` was not read; the harness sums a melee record with `FUN_0057c1e0` first (a missile's `FUN_0057c6c0` sums itself, as run)." "Life Tap heals from the blow's damage in d2d (crush and elements in it), from the record's physical in game.exe (`0x5c6227`); not changed here."
- **C6 (medium, evidence).** Add rows: "`FUN_0057d4f0` `0x57d563..0x57d5c7`: the melee caller copies a record stored on the attacker, sets its flag dword to 0x20, calls `FUN_0057c6c0` with 0 (events 5 and 1)"; "`FUN_0056b9c0` `0x56ba5f..0x56ba7f`: the record caller, `FUN_0057c6c0` with 1, then `FUN_0057cee0`"; the `FUN_005bffc0` row: "writes stat 6, the record's kill bit (`+4 |= 2`) and nothing else of the record".
- **C7 (low).** In play: add "the same for a missile (the crush an eighth, a sixteenth for a boss row) and a blow with fire".
- **C8 (low).** How to re-check: the d2d grep becomes `grep -E '^[0-9]+ [01] [0-9]+ [0-9]+ -?[0-9]+$'`; add the control (`game ... --no-crush`, the test with a third argument `--no-crush`), "parent: 65,145 differ, exit 1; claim 0, exit 0", and `ctest --test-dir build -R test_crush_get_hit --output-on-failure`.

## Findings beside the PR (not blockers)

- **F1 (medium).** A monster's MonProp crush on a monster or a pet is sized by the crush alone (`fight/tick.cpp:459`, `:517`): a follow-up candidate; the oracle would run the same chain with a type-1 attacker carrying stat 136.
- **F2 (low).** Life Tap amount: d2d `blow.damage` x calc1 %, game.exe the record's `+8` (the physical before the cut) (`0x5c61e0`, V read; the MulDiv tail not read: U).
- **F3 (low, trace).** decomp42 section 1.3's m5 row ("not rebuilt against the ninth pin") is settled: m05 aborts r2's pins.

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (the trace's 1.8 recipe stands); nothing on the worker host.
- Which code stores a melee record on the attacker (U); a leech-and-crush case; the full ctest of the D5+D6 tree (only the four crush and combat tests).
