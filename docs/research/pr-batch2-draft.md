# Upstream batch 2: minmana floor and Teleport (card prbuildb2, 2026-11-02)

> status: current, private. Two fork branches built on upstream main after #177; nothing pushed to the public fork, no PR, no upstream comment. Both are on the private remote only. A final rebase onto #178 (the maintainer's batch 10, open) happens after it merges. Tags: V = run (how stated), S = secondary, U = unverified.

## 1. Branches

| branch | head | base | commits | files | size |
|---|---|---|---|---|---|
| local/batch2-minmana | 82f8d3e4 | upstream main 3e073888 (merge of #177) | 1 | 8 | +193 -11 |
| local/batch2-teleport | d2572d6c | upstream main 3e073888 | 2 | 3 | +189 -1 |

Re-rebase: `git rebase --onto <new upstream main> 3e073888 <branch>`; each branch touches only `components/rules/skills.hpp`, `components/game/fight*`, `components/game/gamedata/gamedata.hpp` and its own tests. Author and committer on every commit: the contributor's numeric noreply identity; trailer `Agent: Claude Code` only (V: `git log --format='%an <%ae> | %cn <%ce>'`).

## 2. Minmana floor (one commit)

Deviation (ledger skillfx.mana, sk-mana-trace D2): `FUN_0056c160` compares mana with the formula alone (no minmana floor); `FUN_0056bfe0` takes `max(formula, minmana)` and takes nothing when mana is short of that, or for a row with mana 0 and lvlmana 0, while the cast has already happened. d2d refused the cast under the floor and paid the floor (Might: one point). Change: `rules::mana_formula` (check), `rules::mana_paid` (payment), `Fight::cost_of` is the check and `Fight::pay` takes `mana_paid`, on all four cast paths.

Counted differential (V): `tests/game/test_sk_mana.cpp` runs the REAL `Fight::cast`, `cast_missile`, `cast_summon` and `start_swing` for each case; game.exe's side is `tools/emu/sk_mana.py` (`FUN_0056c160` then `FUN_0056bfe0` natively); helper `tools/emu/pb2_mana.sh`. A case with no cast path in this build (passives, auras, corpse skills, a weapon the test character lacks) is not compared.

| run | cases | compared | base differs | head differs |
|---|---|---|---|---|
| `gen 40000 1` (every Skills row x levels 1..60 x thresholds, plus random) | 130,840 | 76,939 | 278 | 0 |
| `gen 100000 7` | 190,840 | 105,978 | 281 | 0 |

The base differences are all "game.exe casts, d2d refuses", by skill: Teleport 74, Dopplezon 71, Bone Prison 67, Guided Arrow 66 (seed 7: 74, 72, 68, 67). Limits: Might and Summon Resist (the trace's 242 and 91 cases) have no cast path in this harness, so they are not claimed; timing (D1) and interrupted casts (D3) are not part of this commit.

Left out: pets and merc tie order (sk-mana-trace / candidates C8): shown only at rule level with a reordered harness (1,254 of 100,000), never on a real d2d build and with no effect on the Sorceress; not built.

## 3. Teleport (two commits)

Both apply to Teleport (`srvdofunc` 27) only; Dragon Flight and Leap share the branch and keep their old rules. Game side: `tools/emu/sk_tele_spot.py` (`FUN_005ca360` and the warp `FUN_00554ea0` natively on 5 x 5 rooms of 8 x 8 tiles); d2d side: `tests/game/test_tele_spot.cpp` puts the case's collision words into a copy of a real level (25 rooms of 8 x 8 tiles, walled round) and calls the REAL `Fight::spot`; helper `tools/emu/pb2_tele.sh`. The units' bit (0x1000) is taken out of the cases (d2d keeps units in the Crowd). Verifier note: `verify-teleport-room.md` holds the room rule (HOLDS, V read and V run).

`gen 100000 1`, 100,000 cases, equal to game.exe's answer (landed or refused, and the subtile):

| class | cases | base | D1 head | D2 head |
|---|---|---|---|---|
| free target in a near room | 52,539 | 52,539 | 52,539 | 52,539 |
| Teleport 0 | 5,026 | 5,026 | 5,026 | 5,026 |
| target room not near | 40,974 | 15,854 | 40,974 | 40,974 |
| target blocked (plus footprint) | 1,461 | 0 | 0 | 1,461 |
| all | 100,000 | 73,419 | 98,539 | 100,000 |

- **D1, room limit (74a17909):** the target subtile must lie in the caster's room or its near list (`missile_rooms`, the existing FUN_00463740 model): 25,120 cases where game.exe refuses and d2d landed anywhere -> 0 (V, the same number as the trace's line copy).
- **D2, nearest free spot (d2572d6c):** a blocked target lands on `Level::warp_spot`: the first square ring (to 50) with a free plus footprint, the least |dx| + |dy| in it. Tie order learned from 20,000 dense-obstacle oracle runs (ring and distance matched in 20,000 of 20,000; ties: left and right columns row by row, then top and bottom rows column by column, then corners): 1,461 cases where game.exe lands and d2d refused -> 0. A first version with row-major ties left 70 cases on another spot; the tie order is an inference from runs, not a read of `FUN_0064dea0` (U for rings above 7).
- Limits: the oracle's near list is the 3 x 3 touching rooms (a hook); d2d's near list gives the same set for 8 x 8 tile rooms (closeness under 6 tiles), but real levels with irregular rooms are not compared (U). Teleport value 2 (the clear-line test) is unchanged and not covered (oracle hooks it to clear).

Left out: D3 (a monster at the target lands next to it): units live in d2d's Crowd and the oracle only injects raw unit bits, so no honest differential on the real build; D4 (mana paid on a refused hop; it joins the pay-timing change), D5 to D7 (unit target, line mask, arrival bookkeeping): V read only.

## 4. Overlap with #178 (batch 10, head 25978757 at the time)

- `git merge-tree` of each branch with #178's head: clean, no conflict (V). #178 does not touch `mana_cost`, `Fight::pay`, `cost_of` or the Teleport branch; it edits `skills.hpp` (attack line colours), `fight.hpp`, `strike.cpp` (char panel, Throw) and `gamedata.hpp` (state colours) in other hunks.
- Throwaway merge of each branch with #178's head, built with the CI flags (g++, -Werror): both build; test_rules, test_game_rules, test_monsters, test_monster_thinks, test_monster_ai, test_combat, test_game, test_sk_ground, test_skills, #178's test_panel_states and the branch's own test all rc 0 on each merged tree (V). The differential was not re-run on the merged trees.

## 5. Parity (V, worker host, nice 19, -j2)

| check | minmana | teleport |
|---|---|---|
| g++ and clang++ with the CI flags (-Wall -Wextra -Wpedantic -Wno-missing-field-initializers -Werror) | pass | pass |
| test_rules test_game_rules test_monsters test_monster_thinks test_monster_ai test_combat test_game (with the data dir), test_sk_ground, test_skills, new test | all rc 0 | all rc 0 |
| extra -Wshadow -Wrange-loop-construct -Wconversion, syntax only, g++ and clang++, added lines | 0 (one -Wconversion found and fixed) | 0 |
| clang-tidy on the tidy-pr-select TUs (93 / 92), diagnostics on added lines | 0 after fixing one include | 0 after fixing three includes and short names |
| pr_guard per commit | 0 FAIL 0 WARN | 0 FAIL 0 WARN |
| presubmit --quick incl. attribution | 0 FAIL, 3 WARN | 0 FAIL, 3 WARN |
| MSVC /W4 and windows.h macros, by reading the diff | no `min`/`max`/`near`/`far` identifier added (`std::max` as before), no narrowing without a cast | same |

Notes: the full-project g++ build fails on three TUs unrelated to the diff (a gcc 16 `-Werror=array-bounds` / maybe-uninitialized false positive in `apps/d2d/ui.cpp`, `town/walk.cpp`, `test_protocol.cpp`, plus the sentry/crashpad dependency); the tidy selection keeps such TUs. Presubmit WARNs: siblings (merc, pet, partner use the same `Fight::pay` and cost rule through the same casts; Teleport is a player skill, the others are not changed), docs-sync (no doc says a different thing; the skill matrix lives in our repo), build (skipped by `--quick`, built by hand above). Not run: determinism scenario hashes, a live game.

## 6. Re-run

`tools/emu/run-on-worker.sh --ttl 1500 pb2_mana.sh 40000 1 BASE_BIN HEAD_BIN` and `pb2_tele.sh 100000 1 BASE_BIN HEAD_BIN...` (the bins are `test_sk_mana` and `test_tele_spot`; base = upstream main plus the test file).

## 7. Independent check

Ready for the independent Opus check: yes for both, with the limits above (D1/D2 tie order, near list on irregular rooms, Might not in the harness).
