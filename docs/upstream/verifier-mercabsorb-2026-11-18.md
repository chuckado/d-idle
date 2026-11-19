# Verifier: merc gear absorb r1 (2026-11-18)

> status: current, 2026-11-19 (card verifymerc, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/merc-absorb-r1` (oracle `15490e40`, claim `0595e2d1`, on upstream `9b64c11a`), body `docs/upstream/bodies/merc-absorb-body.md`, trace `docs/research/decomp41-merc-crush-trace.md` (agent decomp41, read from `agent/decomp41` `0c31be15b`). Checklist from `verifier-decomp35..39`, `verifier-gearabsorb3-2026-11-17.md`, `verifier-crush3-2026-11-17.md`. The steps that need a build ran on 2026-11-19 on the builder container's long lane (one registered job, `-j2`), after the coordinator's wait.

Tags: **V run** (run by this verifier today), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r1 FAIL (stale base, test gap, a wrong RE sentence, body), r2 PASS WITH CHANGES

The rule is game.exe's and r1 implements it for a merc whose stat list carries absorb stats (V run, below). r1 cannot be posted as it is:

1. **Stale base (high).** Upstream main moved from `9b64c11a` to `563f9257` (#624 aura pulse, #625 curse slot; `fight/` touched). r1 cherry-picks without conflict; the posted commits must be the rebased ones, built and tested there.
2. **The cap from the merc's whole max life is not pinned (medium).** Every pin and every oracle case puts the whole `max` on the merc's level-up life (`merc_st.life`), so a heal capped at `merc_st.life` instead of `merc_max_life()` (the gear's +life left out of the cap) passes r1's tests and its compare. game.exe caps at `FUN_00625d10`, which reads the unit's stat list `[unit+0x5c]` (V read), the list the item stats join (I). r2 adds two pins: pins 0 and 7 again with 20 and 50 of the 100 points from the armor's stat 7 (same lives). Mutant v3 below.
3. **The RE finding overclaims (medium, claims).** r1's `combat.md` and the `merc.cpp` comment say game.exe's functions "test no unit type" and "a mercenary's absorb gear acts the same". `FUN_0057c6c0` does test the type (a type-1 defender needs MonStats flag 0xf of its row, `0x57c76a..0x57c777`, V read), and the equip path is inferred. r2 says "the type-1 test is a MonStats flag the hireling rows pass" and "that a hireling's worn items put the stats on its stat list is inferred (as its gear resistances), not run"; the claim's commit message the same.
4. **Body (medium, template and claims).** No `## Checklist`; "ctest not run"; the re-check block generates 20,000 cases while the counted table is 6,000; the Limits miss the merc's crush now reading the healed life; the presubmit WARNs are not answered. C1 to C8 below.

Not a problem: **dead code.** d2d's merc does carry absorb stats in play: the merc equip command (`world/commands.cpp:527`, packet 0x61) moves the held item into `character.merc_items`, the save loads them (`world/setup.cpp:208`), `Fight::merc_gear` sums every worn item's props, sockets and set bonuses (`fight/merc.cpp:115..`), and `make_fighter` reads 142..149 from that sum (`combat.hpp:176..177`); the parent zeroed them after (`fight/merc.cpp`), which the claim removes. V read. The oracle's d2d side runs that path (the absorb on a worn torso piece's props), so the d2d half is V run.

**Post `fid/merc-absorb-r2`** (on the private fork remote): oracle `e724e726`, claim `fc7818e9`, on upstream main `563f9257`; author and committer `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>`, oracle first, `Agent: Claude Code` the only trailer after a blank line (V run, `git interpret-trailers --parse`). Against r1: rebased; the test's `life_after` takes a `gear_life` and two cap pins use it; the `combat.md` sentence, the `merc.cpp` comment and the commit message as in 3.

## Game side, re-derived

- **The gate (V read).** `FUN_0057c6c0` `0x57c6c0..0x57c786`: `FUN_00554200`, `FUN_00620bb0` / `FUN_0061ab00` (with the attacker's class flag at `0x57c71b..0x57c730`), then for a defender of type 1 `FUN_00457490(class, 0xf)` (zero ends the call), then the dead test `FUN_005541b0`. Nothing names a player for the absorb (`FUN_0057c1e0` -> `FUN_0057bf80` -> `FUN_0057bf10`, S: `verifier-gearabsorb3`) or the heal.
- **The heal `FUN_0057a980` (V read).** Amount at most 0, `FUN_005541b0` (dead) or state `0x5c` (`FUN_00639df0`): nothing. Else `stat 6 + amount`; at least `FUN_00625d10` (max life) gives the max (`0x57a9c2..0x57a9ca`), stored with `FUN_00627260`: a life above max is set DOWN to max.
- **`FUN_00625d10` (V read)** reads the unit's stat list at `[unit+0x5c]`: whatever adds to stat 7 (the merc's gear's +life too: I).
- **The merc's life under the oracle (V run).** `merc_absorb.py game` (FUN_0057c6c0 on a type-1 unit of the hireling's MonStats class, the MonStats readers real): seed 7 reproduces decomp41's case and game files byte for byte (`cmp`); fresh seed 4242, 6,000 cases. Player against merc: the first 300 seed-4242 cases with the class column set to 0, through `gear_absorb.py game` (a type-0 unit): equal to the merc's lives in 300 of 300 (`cmp`).
- **The equip path (I).** That a worn merc item puts its stats on the merc's stat list in game.exe is not run; the oracle starts from the stat list, as the merc's gear resistances (which d2d already models this way). The trace's drill R1 settles it live.

## Scope answers (the brief's a to e)

- **(a) The stat transfer.** game.exe: I (above); d2d: V read and V run (not dead code). The claim must be worded as C3.
- **(b) Pets, summons, golems.** Excluded and unchanged: `Fight::pet_foe` builds `simple_fighter` with resistances, thorns and damage resist only (`fight/pets.cpp:81..88`, V read), so a pet's absorb stays 0 and it has no heal-back; monsters' foes the same (`fight/tick.cpp:425..`). game.exe reads absorb on any unit's stat list; whether an Iron Golem keeps its item's absorb is a pre-existing question (U), not this PR's.
- **(c) Ordering.** Heal per blow before that blow's total, capped per blow, stopping at the death (`rules::absorb_heal`, shared with #606; pins 0, 3, 4, 7, 10). In `Fight::hurt_side` the merc's heal runs after the poison filing and BEFORE its Crushing Blows, which take `hit_points - damage`: on the parent the crush read the life after the hits, now the healed life after the hits; game.exe reads the life before the hits (D7 of the trace, pre-existing). The merc's leech (its own swings), belt potions and regeneration are applied elsewhere in the frame; their order against the heal matters only at the cap (U, low; as #606 for the player). C4.
- **(d) ctest.** r2 claim `fc7818e9`, full build, `ctest -j2`: 400 of 401 pass, 13 skipped for data; the failure is `smoke_d2d` (`tests/smoke_d2d.py:444`, the created character's class, name and hardcore flag read back), which fails again alone (`-j1`); SMOKE_PARENT. The merc and absorb tests all pass (V run). The decomp41 subset left `/data/d41/ctest-merc.txt` on the builder (read only): 40 of 40 selected tests pass on r1's claim at `9b64c11a` (5 skipped for data; `test_stand_attack` passes); no parent subset file exists.
- **(e) Checks.** `pr_guard.py upstream/main HEAD` on r2: 0 FAIL, 0 WARN (V run). `presubmit.py --quick --base upstream/main`: 0 FAIL, 5 WARN (V run): siblings (merc, not player, pet, monster, co-op partner), doc-claims on `combat.md` and `tools/emu/README.md` ("every table loaded": true, `drlg.boot`), build and gcc skipped by `--quick`. clang-tidy-19 by file on r2 (`fight/tick.cpp`, `fight/merc.cpp`, `test_merc_absorb.cpp`, `test_gear_absorb.cpp`): 0 findings each; positive control `readability-magic-numbers` on `test_merc_absorb.cpp`: 3,897 reports (as errors: the repo sets warnings-as-errors), so the units are parsed (V run).

## Differentials

| Tree | seed 7 | seed 5001 | seed 9091 | seed 4242 (fresh) |
|---|---|---|---|---|
| parent (oracle commit `e724e726`, V run) | 3,075, exit 1 | 3,056, exit 1 | 3,057, exit 1 | 3,077, exit 1 |
| r1 claim | 0 (S) | 0 (S) | 0 (S) | |
| r2 claim `fc7818e9` (V run) | 0, exit 0 | 0, exit 0 | 0, exit 0 | 0, exit 0 |

6,000 cases a seed; seeds 5001 and 9091 reuse decomp41's case and game files (read only), 7 and 4242 are this verifier's (7 equal to decomp41's byte for byte). `compare` exits 1 on an empty, a short (5 lines) and a malformed d2d file (V run).

## Mutants (V run; spec `mut.json` in the lab dir, local, never pushed)

Each one text change of r2's claim, `test_merc_absorb` rebuilt in the full build tree; the pins (no argument) and the compare on seed 7.

| Mutant | r2 pins | compare seed 7 (of 6,000) |
|---|---|---|
| v1 no heal | abort | 2,357 |
| v2 heal after the totals | abort | 1,932 |
| v3 cap at the level-up life (gear +life left out) | abort, **only the new cap pins** (line 163) | **0** |
| v4 cap at the player's max life | abort | 2,242 |
| v5 uncapped | abort | 1,461 |
| v6 merc absorb zeroed (the parent) | abort | 3,075 |
| v7 percent and flat swapped | abort | 2,915 |
| v8 cold and magic swapped | abort | 2,175 |
| v9 fire and lightning flat swapped | abort | 1,770 |
| v10 heal from the player's blows | abort | 2,357 |
| v11 heal in whole points | abort | 1,788 |
| v12 heal counted twice | abort | 2,874 |
| v13 a life above max kept | abort | **0** (pin 7 only) |
| v14 flat dropped | abort | 2,485 |

14 of 14 abort r2's pins. v3 passes r1's twelve pins and every oracle case: r1's gap, closed by r2.

## Required changes to the body (`merc-absorb-body.md`)

- **C1 (high).** Commits: "Two commits on upstream `563f9257`: the oracle (`e724e726`), the claim (`fc7818e9`)". Counts: add the fresh seed 4242 row (parent 3,077, claim 0).
- **C2 (high, template).** Add `## Checklist`, ticked: tests (`tests/game/test_merc_absorb.cpp`, `tests/game/test_gear_absorb.cpp`), the finding (`docs/research/re/combat.md`), `pr_guard` 0 FAIL 0 WARN, the ctest line with this file's numbers, labels `oracle`, `port`, `test`.
- **C3 (high, claims).** "What changes", first sentence: "A mercenary whose stats carry gear absorb (stats 142 to 149, from its worn items as its resistances) now has it take part of every elemental or magic field of a monster blow and heal it back, as a player (#606)". Evidence row `FUN_0057c6c0`: "for a type-1 defender one MonStats flag of its class row (`FUN_00457490`, 0xf), which the hireling rows pass; no test names a player". Limits: keep the equip path as I and add "d2d's side of the oracle puts the stats on a worn item (`merc_gear`), so that half is run".
- **C4 (medium, Limits).** Add: "The merc's Crushing Blow (a monster's MonProp, after the hits in d2d) now reads the life after the heal and the hits; game.exe reads the life before the hits (pre-existing)". And: "The merc's leech, potions and regeneration are applied elsewhere in d2d's frame; their order against the heal matters only at the cap".
- **C5 (medium).** WARN answers, one line each: siblings ("the player has it since #606; pets and monsters build `simple_fighter`, no gear, absorb 0; a co-op partner is a player"); doc-claims ("every table loaded": `drlg.boot`).
- **C6 (medium).** "How to re-check": `gen 6000` (the counted size), seeds 7, 5001, 9091, 4242.
- **C7 (low).** Pins: "twelve cases, and pins 0 and 7 again with part of the max life from the merc's armor (the cap is its whole max life)"; mutants: "14 of 14 abort" with this file's table (v3 and v13 only through the pins).
- **C8 (low).** Build and tests: "full build, ctest 400 of 401, 13 skipped for data; `smoke_d2d` fails on upstream `563f9257` too (same assert); clang-tidy 0 on the four changed C++ files".

## For the coordinator (not body changes)

- Trace (decomp41) 1.3: "ctest NOT run" and the lost tidy re-run are superseded by this file; Q1's "nothing names a player" stands, "tests no unit type" (combat.md) does not.
- Follow-ups (no card from here): the Iron Golem's item absorb (U); D7, the merc's crush base (trace).

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill (R1 of the trace settles the equip path); nothing on the worker host.
