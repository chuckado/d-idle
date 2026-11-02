# PR draft: a Nightmare or Hell monster starts at its area's MonLvl, not its own Level(N) / Level(H) (NMLEVEL)

> status: DRAFT held for the user's go and the independent verifier (nothing pushed upstream, opened or posted). Candidate on the private fork remote only: `local/nmlevel`, head `9672fc6c`, on upstream main `6935cf64` (read-only fetch, 2026-11-02): `7023cc73` (the rule), `1682f73f` (oracle and test mode), `9672fc6c` (one include the CI tidy step wants; fold it into `1682f73f` before the PR). Evidence: [levelspawn-trace.md](levelspawn-trace.md) D1 (branch agent/levelspawn) and [monster-stats-trace.md](monster-stats-trace.md) D1. Tags: V = run or read by us (how given), S = secondary, U = unverified.

Title (rule commit subject): `Nightmare and Hell monsters start at the area MonLvl, not their own Level(N) / Level(H)`

## What changes

game.exe's unit maker FUN_00573cb0 (V read, 0x573e15..0x573e70) takes a monster's level from the class MonStats `Level` / `Level(N)` / `Level(H)`. In an expansion game on Nightmare or Hell it takes the AREA's Levels.txt `MonLvl2Ex` / `MonLvl3Ex` instead (FUN_0061a1b0 gives the level id, FUN_0061dca0 the column), unless the class row is noRatio (MonStats flag byte +0xc & 4, mask 0x6ce270) or boss (& 0x40, mask 0x6ce280). Life, defence and experience then read the MonLvl row at that level (V read, FUN_006538a0 mask 7); to-hit and damage at stat 12 per hit (V read, FUN_005a4f50), which includes the mods' leveladd.

d2d read `Level(N)` / `Level(H)` for every monster (`rules::monster_stats`; the comment on `MonStats` said the area rule "waits for those difficulties"). The change adds `rules::start_level(type, difficulty, area_level)` (data driven: no_ratio and boss_column are columns of the row, no monster names) and an `area_level` parameter on `monster_stats` (-1: area unknown, the MonStats level as before). `make_monster` / `spawn_monsters` take the Levels.txt id (`Monster::level_id`), `make_boss` reads it, and every unit maker that knows its level passes it: the room population (`Fight::rooms_up`, the chest-risen spawn in `world/objects.cpp`), monster summons / clones inside `monster_update` (`level.id`), the worm of Impregnate, the Act 3 fetish pack, the Act 4 seal bosses and Diablo, Baal's stair copy. Unchanged on purpose: pets (`summons.cpp`, `pets.cpp`) and the co-op client's local copy (`apps/d2d/town/net.cpp`; the host's level arrives in the replicated unit, `replication.hpp`), both pass no id.

## Proof

Differential (V run, the worker host): game.exe's own FUN_00573cb0, FUN_005a2120 (mods) and FUN_005a4f50 under the emulator against d2d's real `make_monster` / `make_boss` (`tests/game/test_boss_level.cpp --boss-level`), the existing `tools/emu/boss_level.py` extended: `gen-area N SEED` makes cases with a Levels.txt id and a new kind `p` (plain monster); in an area case FUN_0061a1b0 returns the case's id and FUN_0061dca0 runs for real (hooks otherwise as the file's docstring lists). Cases: every MonStats row alive in the difficulty x kind (plain, champion, unique, superunique, minion) with two ids drawn from the levels that have a MonLvl in that difficulty (plus ids 0 and count + 1, off the table: FUN_0061dca0 answers 1), then 100,000 random cases.

| build | seed | cases | Normal differ | Nightmare differ | Hell differ |
|---|---|---|---|---|---|
| base (upstream + test mode only) | 11 | 117,440 | 0 of 39,001 | 36,350 of 39,152 | 34,761 of 39,287 |
| base | 12 | 117,440 | 0 of 39,033 | 36,845 of 39,561 | 34,278 of 38,846 |
| head `7023cc73` | 11 | 117,440 | 0 | 0 | 0 |
| head | 12 | 117,440 | 0 | 0 | 0 |

"Differ" = the level field, or a field that is not one of PR #157's named residuals. Level field mismatches: base 71,111 / 71,123, head 0 / 0. The named residuals (identical in kind and about the same count before and after; they are not this claim): life within 2 points (256ths against whole points), classes game.exe gives no life, S1 (stone skin, game.exe's defence twice d2d's) and H1 (the willowisp family's damage). Head seed 11: 29,091 of them (25,872 / 1,689 / 673 / 857), 0 outside the named classes; seed 12: 28,811, 0 outside.

Real populations (V run): `tools/emu/ls_area.py` (card levelspawn: upstream's room population of each level under game.exe's code, a unit's class and level) over 34 levels of all acts the census covers, seeds 1 and 2, all three difficulties, against the same `test_boss_level` with the level the unit was made in (`tools/emu/nm_pop.py cases` / `compare`). Plain units (units of a unique's party, +3 level and x5 experience from the mods pass, and one level-0 class, are counted apart):

| difficulty | plain units | base: level differs | head: level differs | experience, mean per unit: game.exe / base / head |
|---|---|---|---|---|
| Normal | 7,731 | 0 | 0 | 130.6 / 130.6 / 130.6 |
| Nightmare | 7,360 | 2,166 (29.4 %) | 0 | 4,119.6 / 4,003.3 (-2.8 %) / 4,119.6 |
| Hell | 7,114 | 3,133 (44.0 %) | 0 | 37,437.5 / 35,065.9 (-6.3 %) / 37,437.5 |

The classes most often at another level on base (class = MonStats row): Nightmare rows 20, 131, 3, 94, 74, 22; Hell rows 20, 19, 6, 0, 131, 94 (`nm_pop.py compare` prints the top six). Experience differs in 2,150 / 3,119 units on base and in 0 on head.

Unit test (`tests/monsters/test_monsters.cpp`, no game data): three hand-made rows (plain, noRatio, boss) x three difficulties x the leveladd of plain / champion / unique / superunique / minion taken from `boss_stats` (0 / 2 / 3 / 3 / 3): the level, the life / defence / experience row at the start level, to-hit and damage at the raised one, noRatio rows as written, Normal never reading the area, an unknown area (-1) changing nothing.

Single source (grep, V): every consumer of a monster's level reads `monster.stats.level` (XP penalty `kill_award`, drops `roll_drops`, poison / mana-burn / enchant rows in `make_boss`, `missile_blow`, auras, curses); the only other readers of `MonType::level` are `monster_ranges` (skill tooltips, the type's own level, unchanged) and the upgrade ladder `gamedata.cpp:377` (`level[0]`, Normal, unchanged). The co-op client takes the host's `stats.level` from the replicated unit.

Checks (V, run on the worker host): `ctest` `test_game`, `test_combat`, `test_rules`, `test_game_rules`, `test_monsters`, `test_monster_ai`, `test_monster_thinks` pass on `7023cc73` (8 of 8 with `test_boss_level`, which skips without arguments); `pr_guard` 0 FAIL / 0 WARN against `6935cf64`; `presubmit --quick` 0 FAIL, 4 WARN (answered below), ruff clean; clang-tidy (the CI step's invocation `clang-tidy -p build --quiet --extra-arg=-Wno-unknown-warning-option` over the 8 changed translation units): nothing in our files after `9672fc6c` (one missing `<monsters.hpp>` include, the first run), the `readability-identifier-length` findings of `outdoor.hpp` and `merc.hpp` are upstream's; every changed .cpp compiled with g++ and clang++ at `-Wall -Wextra -Wpedantic -Wshadow -Wrange-loop-construct -Wconversion -Wsign-compare -Wunused`: no warning on a line of this diff; MSVC `/W4 /WX` not run (hand review of the diff: no shadowed local, no narrowing, no new unused name).

Det state hashes (V run, the 14 scenarios, hidden slot): base = upstream + the lab verbs (7 commits plus a read-only `debug monstate`), head = base + `7023cc73`. 13 of 14 scenarios identical in the monster hash and the state hash (passive, fight, cold, static, cold2, hit_plain, hit_champion, hit_unique, hit_boss, hit_immune, hit_minion, merc, follow). `pet`: monster hash identical; the state hash takes one of two values (`65e4691440ea`, `0de0e52954f6`) on base and on head alike (base: the first of its five runs the first value, the other four the second; head: three of four runs the second, the last the first): a pre-existing flap of the pet position, not a head effect. Nothing legitimately changes: the scenarios spawn monsters with `debug spawn` (no level id) or in Normal, so they do not exercise the rule; the rule itself is covered by the differentials above.

## Not in this PR

- Monster regeneration (stat 0x4a = life256 x DamageRegen >> 12, added by FUN_00580610 / FUN_00580810; monster-stats trace D2): d2d has none. A missing feature with an unverified trigger and rate (U), no differential to fix it against: left out.
- Pets and Revive / summoned monsters from player skills: no level id passed (their stats come from the skill code); whether game.exe's pet maker applies the area rule to them is U.
- A unique's party members' leveladd (+3, x5) is PR #157's rule, untouched.
- Levels with MonLvl 0 and classes with MonStats level 0: d2d clamps the level to 1 (game.exe keeps 0); not reachable in the populations (counted apart above), unchanged.
- Classic (non-expansion) games: d2d runs expansion data only (the area column is the `Ex` one).

## Limits

- FUN_0061a1b0 (the unit's room to level id) is hooked in the oracle: the claim is the column and the exception test, not which level a room belongs to (d2d's level is the level it is placed in). U for monsters created by game.exe paths we did not run; V read: FUN_00573cb0 has one caller, the generic unit maker FUN_00574250, so summons and clones take the same rule (grep of the full disassembly).
- Drop quality and the XP penalty read `stats.level`, so they now follow the area level too; that game.exe reads stat 12 there was traced for the XP penalty only (S, `xp-award-trace.md`); drops: U.
- Presubmit WARNs, answered: *siblings, unit kinds*: pets (`summons.cpp`, `pets.cpp`) pass no level id and are unchanged (see above); mercs and co-op partners are players, not monster units; *data-gating*: `test_boss_level.cpp` already gates on `data->patched` (the hit is a comment line); *docs-sync*: no document describes the monster start level (grep of the .md files), no doc changes; *build*: answered by the ctest run above.
- Not run: an MSVC build, a live game, the CI Windows and macOS jobs.

## How a narrow verifier checks it

1. Read FUN_00573cb0 0x573e15..0x573e70 and FUN_0061dca0 with `tools/emu/disasm.py 573e15 40`, `disasm.py 61dca0`; confirm the masks 0x6ce270 / 0x6ce280 are 4 and 0x40 and that d2d's `no_ratio` / `boss_column` are those bits (`montypes.hpp`).
2. `uv run python tools/emu/boss_level.py gen-area 100000 11 > cases.tsv`, `... game cases.tsv > game.tsv`, `build/tests/test_boss_level --boss-level cases.tsv > d2d.tsv`, `... compare cases.tsv game.tsv d2d.tsv`: expect "level field mismatches 0" and "other fields not a named residual 0"; do the same on the base (the second commit's test file with `make_monster`'s last two arguments removed) and expect about 71,000 level mismatches.
3. `nm_pop.py` on `ls_area.py ... units` output as in `tools/emu/nm_pop.py`'s docstring, base and head.
4. Look for a unit maker that should pass a level id and does not (grep `make_monster(`), and for the claim that pets stay unchanged.

## Files

`components/rules/monsters.hpp`, `components/game/ai.hpp`, `ai.cpp`, `fight/tick.cpp`, `world/objects.cpp`, `world/quests/act3.cpp`, `act4.cpp`, `act5.cpp`; tests `tests/monsters/test_monsters.cpp`, `tests/game/test_boss_level.cpp`; tool `tools/emu/boss_level.py` (a tool change needs the maintainer's CODEOWNERS review).

## Shared-file lines for the merger

In `openspec/changes/nmlevel/merge-lines.txt`.
