# Draft (local, NOT posted): "game: a saved purse or stash over its cap is 0 on entering a game"

> status: r3, **all required changes of the verifier done** (card prfix2, 2026-11-14; verifier report [../verifier-p03-p05-2026-11-14.md](../verifier-p03-p05-2026-11-14.md)); not posted, needs the user's go. Branch `fid/load-gold-clamp-r3` on the private fork remote, head `d229705a`, two commits on upstream `a2b4f686` (oracle `dda1fa57`, claim `d229705a`; the numeric noreply identity, last line `Agent: Claude Code`). Replaces `fid/load-gold-clamp-r2` (`3e679d8f` on `f5e11693`) and `-r1`. Earlier draft and proof: [../../research/pr-load-gold-draft.md](../../research/pr-load-gold-draft.md), [../../probes/fid/fid-inv-limits.md](../../probes/fid/fid-inv-limits.md). Body: [../bodies/load-gold-clamp-body.md](../bodies/load-gold-clamp-body.md).

PR title: `game: a saved purse or stash over its cap is 0 on entering a game`

## What changes

`World::take_character` keeps the saved purse (`kGold`) and stash gold (`kGoldBank`) as they are. game.exe's loader `FUN_0056b180` restores them through `FUN_0056af80` (`0x56afb4..0x56b00f`) and `FUN_0053ff00`, which sets a stat to 0 when it would pass its cap: the purse's `clvl x 10000` (`FUN_00622e70`), the stash's 2,500,000 (`FUN_00623460`). The claim commit zeroes both in `take_character` under the same two tests and adds a `test_game` block (at the cap kept, one over zeroed).

Two commits: the oracle (`tools/emu/fl_gold.py`, `test_game --load-gold`, the `tools/emu/README.md` row, and now an exit 1 plus a ctest when the case file cannot be opened), then the claim (`world/setup.cpp`, the unit block, the `docs/research/re/d2s-header.md` section "Gold on load").

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| game.exe zeroes a loaded purse over `clvl x 10000` and a stash over 2,500,000; equal values are kept | **V run** (card emurun, TASK-445, 2026-11-14, emulator image on the worker host): `fl_gold.py game` ran the real `FUN_0056b180` -> `FUN_0056af80` on the 539 cases: **the game column equals the rule in 539 of 539 cases**. This replaces the old oracle claim that rested on decomp5's 2026-11-08 run (S, output not kept); both agree | `fl_gold.py gen`, `fl_gold.py game` |
| d2d's `World::enter` kept an over-cap purse or stash | V run (card prfix2, NAS builder, release build with the game data, oracle commit `dda1fa57`): `test_game --load-gold` on the 539 cases against the rule written out (the verifier's `check.py`): **363 of 539 differ**; equal to the verifier's count and to decomp21's | `test_game --load-gold cases.tsv` |
| after the claim d2d matches | V run: the same cases on `d229705a`: **0 of 539 differ**; `test_game` rc 0 and prints `OK: a save's gold over its cap is 0 on entering a game`; `test_combat`, `test_rules` rc 0. As for P03, the literal `diff game.tsv d2d.tsv` of the emulator's file against the head's was not run (U, one command on the worker host); it follows from the two rows above | `test_game` |
| a case file that cannot be opened fails | V run: `test_game --load-gold /nonexistent.tsv` exits 1 and prints `load-gold: cannot open ...` on both commits; ctest `test_game_load_gold_missing_cases` (PASS_REGULAR_EXPRESSION, skipped without game data) | |
| lint | V run: clang-tidy-19 with the repo's `.clang-tidy` clean on `tests/game/test_game.cpp` and `world/setup.cpp`; `pr_guard.py` 0 FAIL / 0 WARN | |
| reachability | U: needs an edited or imported save; not claimed in the PR | none |

The 363 is grid arithmetic (33 of the 49 purse and stash pairs per level, 11 levels): said so in the body.

## Limits and open items

- The maintainer may not want a loader that zeroes a player's gold: the body says this is game.exe's rule (an over-cap save cannot come from play) and asks in review.
- The stash cap is a second constant next to `kStashGoldMax` in `commands.cpp`; offered in the body, not done (CodeRabbit may ask).
- Owed before posting (worker host or the user's go): the literal `diff` above, the full `tools/presubmit.py`, the whole `ctest` against the parent (the NAS builder ran the three tests of the area).
