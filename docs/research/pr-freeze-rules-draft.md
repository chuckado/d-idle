# Draft PR body: freeze rules (Glacial Spike's freeze on monsters)

> status: DRAFT for the fork branch `local/freeze-rules-v3` (four commits on upstream main 3852c904, 11 files, +482 -9; supersedes `fix/freeze-rules-v2` and the stale [pr-freeze-draft.md](pr-freeze-draft.md)). Rebuilt, re-run and re-checked on the final head on 2026-11-02 (numbers below). Not posted; needs the user's go and one independent re-check of the final head before it is opened. Evidence trail: [pr-freeze-proof.md](pr-freeze-proof.md), [freeze-rules-trace.md](freeze-rules-trace.md). Order, merge tests and CI parity: the section "Prep notes" at the end. All PRs are on hold by the user (2026-11-02); nothing is to be opened.

---

**Title:** Glacial Spike's freeze follows FUN_0057b230

### What

Glacial Spike's freeze on a monster now follows game.exe's FUN_0057b230:

- The freeze length is cut by cold resistance less cold pierce, the same way the chill length is (row 6 of the element table at 0x732980, next to row 5).
- On a MonStats boss, a champion, a unique, a superunique or an act boss (MonStats boss column, or monster-data flag 8) the freeze becomes a chill for that length. They are never frozen.
- Any other monster is frozen only when its coldeffect is negative. The freeze lasts length / MonsterFreezeDivisor, and the monster gets no chill from it.
- Blood Raven's and the Countess's freezes are no longer halved. The call d2d read as stat 0x76 is `FUN_00639db0(unit, 0x76, 1)`, the state setter (state 118, corpse_noselect); it changes no stat.

Before this, d2d stood every monster for the raw length at any coldeffect and also chilled it.

Four commits, one claim each (resist / divide / coldeffect; special monsters chill; no halving; tools). Files: `components/rules/combat.hpp` (`Freeze`, `monster_freeze`), `components/game/fight/missiles.cpp` (`Fight::strike`), `gamedata.hpp` / `gamedata_load.cpp` (`MonsterFreezeDivisor`), comments in `ai.cpp`, `ai.hpp`, `uniques.hpp`, `docs/research/re/superuniques.md`, `tools/emu/monsters.py`; new `tools/emu/freeze_rules.py` and `tests/game/test_freeze_rules.cpp`.

### Proof

`tools/emu/freeze_rules.py` runs game.exe's FUN_0057c6c0 natively under the emulator, including FUN_0057c1e0, FUN_0057af80, FUN_0057b230 and the boss, flag-8 and hireling tests. `test_freeze_rules --freeze-rules` runs the real `Fight::strike` on the same cases. 225,748 cases, seed 20261101 (125,748 boundary cases over every boss kind, boss column, half_freeze, coldeffect, resistance -300..150, freeze length -5..250 and difficulty, plus 100,000 random):

| build | cases that differ from game.exe |
|---|---|
| upstream main 3852c904 (test file added) | 185,746 of 225,748 (plain 48,946, special 78,501, half_freeze 58,299) |
| this branch | 0 of 225,748 |

```sh
cd tools/emu
uv run python freeze_rules.py gen 100000 20261101 > cases.tsv
uv run python freeze_rules.py game cases.tsv > game.tsv
../../build/tests/test_freeze_rules --freeze-rules cases.tsv > d2d.tsv
uv run python freeze_rules.py compare cases.tsv game.tsv d2d.tsv
```

`freeze_rules.py flags` sweeps every monster-data flag byte against the boss column and a hireling class: only flag 8, the boss column or a hireling turns the freeze into a chill. `freeze_rules.py state` runs the Blood Raven / Countess call natively: it sets state 118, writes no stat, and stat 0x76 then reads 0.

A second harness, written separately (its own generator and seed, 0x1F00DFACE) and re-run on this head: 146,591 cases, 0 differ in every class (plain, special, half_freeze) except 6 of 16 with a freeze length above 0x100000 (the residual above).

Det state hashes: not run.

### Limits

- Only Glacial Spike's freeze. Frozen Armor goes through the same FUN_0057b230 rule in game.exe (FUN_005cac50 sends it through FUN_0057c6c0) and is the subject of #159. Freezing Arrow, Ice Blast and item freeze are not modelled in d2d; Frozen Orb and Blizzard only chill (no freeze). All are left for separate PRs.
- Which monsters chill (read from game.exe's own spawn path, 12,830 monsters): flag word 0x0d champions, 0x2d possessed, 0x4d ghostly, 0x09 uniques, 0x0b superuniques and 0x08 act bosses are chilled; minions (0x10) and plain monsters (0x00) are frozen. Only bit 8 matters; hirelings are always chilled.
- Not modelled: state 0x36 (no freeze while it is set), the freeze's events, and a length / divisor of 0 (game.exe sets state 1 for no frame).
- Residual: a freeze length of exactly 0x100001 with a nonzero resistance differs (game.exe's MulDiv divides first above 0x100000, d2d multiplies); unreachable in play.
- A freeze is still d2d's stand (`stun_until`).
- `half_freeze` keeps its name and now only marks the state-118 call.
- The hireling branch has no d2d caller, because skill missiles don't strike mercs.
- Siblings: a player's freeze is already a chill in d2d (FUN_0057b230's player branch). Pets and a co-op partner as freeze targets are not traced.
- Gameplay change, by design: a Frost Nova or Glacial Spike now chills (does not freeze) every champion, unique, superunique and act boss, and monsters with cold resistance are frozen for a shorter time, as in the original game.

### Order with #159

This branch and #159 (Frozen Armor and the freeze resume) are independent and can merge in either order. Both add the same `tools/emu/freeze_rules.py` (byte-identical here and there) and both add `GameData::freeze_divisor` on the same lines, so git merges them without a conflict; a trial merge of this branch onto #159's head builds and passes the same suites and the same differential. If #159 lands first, nothing needs to change here. After both land, Frozen Armor's own copy of the rule in `states.cpp` could call `rules::monster_freeze`: a follow-up, not part of this PR.

---

## Prep notes (not part of the PR body)

Branch and head: private fork, `local/freeze-rules-v3`, head 21010485, pushed to the private remote only (not to the public fork, no PR). Built from upstream main 3852c904 by cherry-picking the four v2 commits with two edits so that the overlap with #159 disappears: (1) `freeze_divisor` is added on the same lines #159 uses (V: identical hunks merge to one line, checked in the merged tree: one declaration, one load line); (2) `tools/emu/freeze_rules.py` is byte-identical to #159's final head 36e7fcde (V: same blob hash; v2's copy differed only in four docstring lines), and the `compare` usage lines moved into the header of `tests/game/test_freeze_rules.cpp`.

Order, with evidence (V: `git merge-tree` and a throwaway merge, a full build of the merged tree):
- v2 as it was on #159's head: CONFLICT, add/add on `tools/emu/freeze_rules.py`, the only conflict (the other seven shared files auto-merge); the docstring is the whole difference.
- v3 with #159 (36e7fcde): clean in either order; merged tree builds with 0 warnings; test_rules, test_game_rules, test_monsters, test_monster_thinks, test_monster_ai, test_combat, test_game, test_freeze_rules, test_freeze_resume, test_armor_hit all rc 0; the freeze differential on the merged tree: 225,748 cases, 0 differ.
- Choice: (a), independent of #159; no stacking needed. Either may merge first. If #159 changes its `freeze_rules.py` or the `freeze_divisor` lines before it lands, v3 must take the same bytes (rebase and re-check merge-tree). #159's Frozen Armor keeps its own copy of the rule; a later follow-up could call `rules::monster_freeze`.

Counts (worker host, offline, no game, no d2d slot; cases md5 ddbd53b274ad, game md5 0168caed598e, equal to the earlier runs): upstream main + test file 185,746 of 225,748 differ; v3 head 0; v3 + #159 0; second harness 6 of 146,591 (0x100001 only). Suites on the v3 head with the game data dir: test_rules, test_game_rules, test_monsters, test_monster_thinks, test_monster_ai, test_combat, test_game rc 0; test_freeze_rules without arguments SKIP.

CI parity (V where run): g++ with the build's flags (-Wall -Wextra -Wpedantic -Wno-missing-field-initializers -Werror, plus -Wshadow -Wunused -Wsign-compare) full build of the test set: 0 warnings. A syntax-only pass over the four changed .cpp files with clang++ and g++ using -Wshadow -Wrange-loop-construct -Wconversion: 0 warnings on added lines (g++ -Wsign-conversion, beyond CI, flags one: test_freeze_rules.cpp line 67, `types[fallen->second]` with an int index; left as is). clang-tidy (repo .clang-tidy) on the four changed .cpp files: 0 warnings on added lines (pre-existing x / y binding names elsewhere). pr_guard: range and each of the four commits 0 FAIL / 0 WARN. presubmit --quick: 0 FAIL, 2 WARN (siblings: the diff touches monster only; build skipped by --quick). MSVC /W4 and windows.h macros (read the diff, U for a real MSVC build): the added code uses no min/max macro names, no unsigned/narrowing implicit conversions beyond `std::uint32_t(...)` casts already in the replaced lines, the `.chill_ticks =` designated initializers are C++20 and already used in the tree. Direct includes: `missiles.cpp` adds `<uniques.hpp>` for `rules::Boss`; combat.hpp needs nothing new (`std::max` already included).

U: det state hashes (need a d2d slot); an MSVC build; the final head has not had the independent re-check yet; the commit body of the tooling commit cites a base hash ("ddc8281d: 185,746 differ") that is not upstream's: change it to "upstream main" before opening (the body is already in the private branch; amending changes the head).


Agent: Claude Code
