# PR candidate: freeze rules (C5, C3, C4)

> status: BUILT, local candidate, stacked on #155 (card decompfreeze, 2026-11-01); PROOF incomplete: no independent verifier yet, no det state hashes (need a d2d slot); nothing pushed or posted. Rules and addresses: [freeze-rules-trace.md](freeze-rules-trace.md). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-freeze-draft.md](pr-freeze-draft.md).

## Candidate

Branch `fix/freeze-rules` in the worker host's PR checkout (worktree `freeze`), on #155's head ddc8281d (= the fork's `fix/static-field`; upstream main 72132b99 + the four #155 commits). Authored and committed with the numeric-ID noreply form copied from ddc8281d, body `Agent: Claude Code`, no Claude trailer.

| commit | claim | files |
|---|---|---|
| 9ec91600 | C5: the freeze length is resisted like the chill length, a plain monster is frozen only at coldeffect < 0 for length / MonsterFreezeDivisor, no chill | combat.hpp (`Freeze`, `monster_freeze`), missiles.cpp (`Fight::strike`), gamedata.hpp / gamedata_load.cpp (`freeze_divisor`) |
| 39ef104e | C3: a MonStats boss, champion, unique or superunique is chilled for the resisted length instead | combat.hpp, missiles.cpp |
| 61620d08 | C4: Blood Raven's and the Countess's freezes are not halved (their 0x76 call is state 118) | missiles.cpp (one line out), comments in ai.cpp, ai.hpp, uniques.hpp, docs/research/re/superuniques.md, tools/emu/monsters.py docstring |
| bbf506c6 | tooling: `tools/emu/freeze_rules.py`, `tests/game/test_freeze_rules.cpp` (`--freeze-rules CASES`) | 2 new files |

Diff against ddc8281d: 11 files, +476 / -9 (C3 carries the `#include <uniques.hpp>` that clang-tidy's include cleaner asked for; it was added after the per-commit run below, by rebuilding C3 and replaying C4 and the tooling commit: 39ef104e / 61620d08 / bbf506c6; the code of each is otherwise the same as the run's e1409e7f / 9c3b8019 / 51906167). With #155's 8 files: 17 files (under 100). Longest touched file: missiles.cpp 838 lines (under 1,500). Copies of the two tooling files in this repo: `tools/emu/freeze_rules.py`, `tools/emu/freeze_rules_test.cpp`.

## Causal chain (no U link)

1. Glacial Spike's hit (SrvHitFunc 13, FUN_005aa8b0) writes auralen into record +0x34 through FUN_005a8f20 (EType frze) only when it is > 0 (V read).
2. FUN_0057c6c0 -> FUN_0057c1e0 cuts +0x34 by cold res less cold pierce, row 6 of 0x732980 (V read, V run natively).
3. FUN_0057b230: special (MonStats boss / flag 8 / hireling) -> FUN_0057af80 chill; else coldeffect < 0 -> state 1 for length / MonsterFreezeDivisor; else nothing (V read, V run natively).
4. Flag 8 on champions (FUN_005a48c0 -> FUN_005a0320), uniques (FUN_005a4940), act bosses and superuniques (FUN_005a4850(mod, 1)); minions 0x10 only (V read). d2d's matching predicate: `boss_column || boss not in {none, minion}`, the one its stun code already uses (S).
5. Blood Raven / Countess `FUN_00639db0(unit, 0x76, 1)` = state 118 (V read; V run: `freeze_rules.py state` prints `state bits set: [118, 630]; stat-list bytes changed: []; stat 0x76 reads 0`).

## Differential (game.exe natively vs real d2d builds)

Game side: game.exe's FUN_0057c6c0 under the emulator with FUN_0057c1e0, c140, c060, be00, bf80, bf10, FUN_00483360, FUN_0057af80, FUN_0057b230, FUN_0057af30, FUN_0063e9f0, FUN_005a0180, FUN_0063ee90 unhooked; hooks listed in the script docstring; DifficultyLevels from the MPQ. d2d side: the REAL `Fight::strike` (missile_blow, monster_freeze, Fight::land) built from each commit with only the test file added. Generator seed 20261101 (masked to 32 bits), 225,748 cases = 125,748 boundary (every Boss kind x boss column x half_freeze x coldeffect {-128, -100, -50, -1, 0, 1, 25, 127} x 17 resistances -300..150 x 15 freeze lengths -5..250 x 3 difficulties; pierce 1..250; running chills and freezes) + 100,000 random. cases.tsv md5 ddbd53b274ad, game.tsv md5 0168caed598e. 0 RESFAIL.

| build | mismatches | C5 (plain, no half) | C3 (special, no half) | C4 (half_freeze set) |
|---|---|---|---|---|
| ddc8281d (before) | **185,746** | 48,946 | 78,501 | 58,299 |
| 9ec91600 (C5) | 85,040 | **0** | 50,393 | 34,647 |
| 39ef104e (C3) | 32,379 | 0 | **0** | 32,379 |
| 61620d08 / bbf506c6 (C4, head; re-run on bbf506c6 itself) | **0** | 0 | 0 | **0** |

Every claim: before > 0, after 0; each commit removes its own slice and leaves the others; no residual. Game side alone, `freeze_rules.py flags` (every monster-data flag byte 0..0xff x boss column x hireling class, Normal, coldeffect -50, length 50): the 896 with flag 8, the boss column or a hireling class all give chill 50 / no freeze; the other 128 give freeze 50 / no chill.

## Commands for the independent verifier (on the worker host)

```sh
cd <PR checkout>/freeze                       # fix/freeze-rules, head bbf506c6
# build test_freeze_rules as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
cd tools/emu
uv run python freeze_rules.py gen 100000 20261101 > cases.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python freeze_rules.py game cases.tsv > game.tsv
D2_MPQ_DIR=<MPQ dir> <build>/tests/test_freeze_rules --freeze-rules cases.tsv > d2d.tsv
uv run python freeze_rules.py compare cases.tsv game.tsv d2d.tsv   # mismatches 0
uv run python freeze_rules.py flags; uv run python freeze_rules.py state
# before: a detached worktree at ddc8281d plus tests/game/test_freeze_rules.cpp from bbf506c6, same commands: 185,746
```

A verifier should write their own generator and seed and re-derive FUN_0057b230 before reading this doc (standard section 5).

## CI-class checks (head bbf506c6)

- pr_guard against upstream/main (72132b99, fetched today): range 0 FAIL / 0 WARN; per commit (#155's six and these four) 0 / 0 each.
- presubmit --quick: 0 FAIL, 4 WARN: siblings (monster only: a player's freeze is already a chill in d2d, as FUN_0057b230's player branch; hirelings are chilled in game.exe but d2d's skill missiles don't strike mercs; pets and a co-op partner as freeze targets: not traced, U; say so in the body), data-gating on #155's test_cold_length / test_static_field (not this PR; test_freeze_rules gates on `data->patched` since the first presubmit flagged it), build skipped (--quick). ruff (upstream's rule set): clean after `zip(..., strict=False)` (B905 found by the first run, fixed in the tooling commit).
- clang-tidy 23.1.4 with the repo .clang-tidy (Names and includes: identifier length, misc-include-cleaner) over missiles.cpp, test_freeze_rules.cpp, gamedata_load.cpp, ai.cpp, test_combat.cpp, test_monsters.cpp: nothing in touched lines (found and fixed: `d2d::rules::Boss` without `<uniques.hpp>` in missiles.cpp); only the pre-existing outdoor.hpp structured-binding names (also on upstream main). The first full selection (138 TUs) showed the same plus pre-existing apps/ names and the launcher's missing main.moc.
- gcc -Wshadow -Wunused -Wsign-compare full build of the tree: 0 warnings on added lines; rebuild of test_freeze_rules / test_game on bbf506c6: 0 warnings.
- test_game (MPQ dir), test_rules, test_combat, test_protocol: rc 0; test_freeze_rules without arguments: SKIP, rc 0.
- Seeds: the Python generator masks the seed to 32 bits; the C++ mode reads the generated cases and has no generator of its own.
- File length: missiles.cpp 839, combat.hpp 765, freeze_rules.py 312, test_freeze_rules.cpp 118 (all under 1,500). PR size: 11 files (17 with #155).

## Limits (for the body)

1. Only Glacial Spike's freeze (`Fight::strike`'s `freeze`, hit_func 13). Not touched: Freezing Arrow (`freezingarrowexp3` EType frze loads as cold in d2d, so it chills and never freezes; game.exe puts its ELen in +0x34: next candidate, read only), Frozen Armor (stun + chill for calc1 frames; FUN_005cac50 not traced), a player's freeze (already a chill in d2d, as FUN_0057b230's player branch).
2. Not modelled: state 0x36 (no freeze while set), the freeze's events (0xc at the end, 2 at end + 1), a length / divisor of 0 (state 1 for no frame), the 20 % state 0x6b roll of the chill. Hireling branch: no d2d caller (skill missiles don't strike mercs).
3. A freeze is still d2d's stand (`stun_until`), as before: what a frozen monster can't do (state 1) is not this claim.
4. `half_freeze` keeps its name (it now marks the state-118 call; nothing reads it for the freeze); a rename is a follow-up.
5. Hooked on the game side: the stat getters, the MonStats row (one fake row: boss byte, coldeffect), state 0x36 off, the state-list calls (they carry only the end frame).
6. Not run: the det state hashes (need a d2d slot), the independent verifier.

## How to open

Stacked on #155 (the freeze path is next to its chill commit in `Fight::strike` / combat.hpp). Recommended: after #155 merges, `git rebase --onto upstream/main ddc8281d fix/freeze-rules` (4 commits, 11 files), re-run the differential and the CI-class checks on the new head, then open as its own PR with the draft body, under the bundle policy (opens when an open PR has no CodeRabbit review). A PR on upstream main before #155 merges would need #155's chill commits as a prerequisite: not recommended. Appending to #155 is possible only with the user's go (it already has a CodeRabbit review).
