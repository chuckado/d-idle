# PR candidate: freeze rules (C5, C3, C4)

> status: BUILT, local candidate, stacked on #155 (card decompfreeze, 2026-11-01); independent verifier done (card verifyfreeze, 2026-11-01: C5 / C3 / C4 HOLD WITH LIMITS, NOT READY until the commit bodies are cut to three lines and the det state hashes run; section "Independent verifier" below); nothing pushed or posted. Rules and addresses: [freeze-rules-trace.md](freeze-rules-trace.md). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-freeze-draft.md](pr-freeze-draft.md).

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

## Independent verifier (card verifyfreeze, 2026-11-01)

Agent VERIFYFREEZE (Opus), not the author. Offline: no game, no d2d slot, nothing pushed or posted, the author's branch and worktree untouched (own throwaway worktrees on the worker host, removed at the end). The rules were re-derived from the disassembly (`disasm.py`, a full `--all` sweep) before this doc or the trace was read; harnesses are new: `tools/emu/verify_freeze.py` (game side, generator, `--compare`, `--game-only`), `tools/emu/verify_freeze_mode.inc` (copied to `tests/game/test_verify_freeze.cpp` in each build; `--verify-freeze SEED N`, its own copy of the generator, seed masked to 32 bits in both), `tools/emu/verify_freeze_spawn.py` (native spawn run).

### Re-derived (V read with disasm.py, V run natively)

- FUN_0057c6c0 has the only call of FUN_0057b230 in game.exe (0x57c8b7), right after FUN_0057af80 (0x57c8a8): chill with record +0x30, then freeze with +0x34, the attacker as the stack argument, the struck unit in EAX.
- Resist: FUN_0057c1e0 runs FUN_0057c140 (stat 0x99 zeroes +0x30 and +0x34, stat 0x76 halves both) and then the 12-row table at 0x732980 through FUN_0057bf80: row 5 is `+0x30, res 0x2b, max 0x2c, pierce 0x14f`, row 6 is the same for `+0x34`. The freeze length is resisted exactly as the chill length is.
- FUN_0057b230: nothing at length <= 0; a player (type 0) goes to FUN_0057af80; a monster with state 0x36 (uninterruptable, FUN_00639df0) gets nothing; a monster whose MonStats row (FUN_00451f80) has `+0xc & [0x6ce280]` (= 0x40, the boss column), or whose monster data has `+0x16 & 8` (FUN_005a0180, EDX 8), or whose class is a hireling (FUN_0063ee90: 271, 338, 359, 560, 561) goes to FUN_0057af80 with the same length (chill, no freeze); anything else needs coldeffect < 0 (FUN_0057af30: MonStats +0x168 + difficulty, signed; -50 for non-monsters), divides by DifficultyLevels +0x14 (MonsterFreezeDivisor; 0 asserts at 0x57b2fd) and sets state 1 for frame + length (existing state only lengthened). The freeze path never calls FUN_0057af80.
- FUN_0057af80 (context): nothing at coldeffect 0, length / MonsterColdDivisor (+0x18) at coldeffect < 0, at least 1, only lengthened; then a 20 % state 0x6b (shatter) roll on the unit seed. A special monster's freeze therefore rolls the shatter state a second time in game.exe (the land() ponytail on state 107 covers it).
- FUN_00639db0 is the state setter (bounds by the state count at +0xc4, FUN_00625a70, FUN_0064c040); FUN_00639df0 the getter. states.txt rows: 1 freeze, 11 cold, 54 uninterruptable, 107 shatter, 118 corpse_noselect. `push 0x76` before `call FUN_00639db0` at 0x5a45d1, 0x5a4607, 0x5a4c00, 0x5a4d99, 0x5b1fb0 and others: state 118. No `push 0x76` / `push 0x99` reaches a stat setter in the sweep.
- Spawn flags: FUN_005a0320 always ORs 8 into +0x16; FUN_005a48c0 (champion mod) calls it and then ORs 5, so a champion carries 0x0d, not 4 alone.

### Champions: the contradiction resolved (V run, native spawn path)

`verify_freeze_spawn.py 4 0x51ed`: game.exe's own population (FUN_0052d0f0 over every room) of Cold Plains, Stony Field, the Den, Burial Grounds, Tower Cellar 5, Catacombs 4, Rocky Waste, Duriel's lair, Spider Forest, Durance 3 and the Chaos Sanctuary, 4 seeds x 3 difficulties, 12,830 monsters; each spawned unit then run through the REAL FUN_0057b230 (coldeffect forced to -50 for the call):

| flag word (+0x16) | who (MonUMod list) | units | FUN_0057b230 |
|---|---|---|---|
| 0x00 | plain | 11,035 | freeze |
| 0x10 | minion | 1,272 | freeze |
| 0x09 | unique | 207 | chill |
| 0x0b | superunique (12 of them the Countess, state 118 set) | 84 | chill |
| 0x0d | champion (mod 16) | 160 | chill |
| 0x2d | possessed champion (mod 38) | 10 | chill |
| 0x4d | ghostly champion (mod 36) | 14 | chill |
| 0x08 | act boss: Andariel 156, Duriel 211, Mephisto 242 (36), Blood Raven 267 (12, state 118 set) | 48 | chill |

Every flag-8 unit chills, every unit without it freezes; no unit had state 0x36 or a prior state. `verify_freeze.py --game-only`: all 256 flag bytes x boss column x coldeffect {-50, 0, 25} x 3 difficulties give the same outcome as `flags & 8` (4,608 of 4,608); the five hireling classes chill at any flags; state 0x36 blocks both; state 118 changes nothing; a real stat 0x76 halves the freeze (50 -> 25) and stat 0x99 removes it. The earlier verifier note (verify-chill.md, "flags 4 and 0x10 were frozen") was right about those synthetic bytes but wrong to read "champion = 4": the spawn path gives champions 0x0d. The candidate's C3 wording (champions chilled, minions frozen) holds.

### Own differential (V run)

`verify_freeze.py 120000 0x1F00DFACE` (masked to 0xF00DFACE on both sides) against `test_verify_freeze --verify-freeze 0x1F00DFACE 120000` built from each commit with only the test file added: 146,591 cases (26,591 boundary: difficulty x coldeffect {-50, -1, 0, 1, 25} x seven kinds with the flag words above x boss column x half_freeze for the Countess / act-boss kinds x resistance -150..101 x freeze length 0..375; pierce 1..100; running chills and freezes; divisors 1..7; lengths around 0x100000; roll 0; plus 120,000 random). 0 RESFAIL, inputs identical line by line.

| build | all | C5 (plain) | C3 (special) | C4 (half_freeze) | length > 0x100000 |
|---|---|---|---|---|---|
| ddc8281d (#155 head, before) | 124,662 | 27,309 / 32,207 | 74,031 / 86,391 | 23,306 / 27,977 | 16 / 16 |
| 9ec91600 (C5) | 70,683 | **0** | 53,751 | 16,921 | 11 |
| 39ef104e (C3) | 14,550 | 0 | **0** | 14,544 | 6 |
| 61620d08 (C4) | 6 | 0 | 0 | **0** | 6 |
| bbf506c6 (head) | 6 | 0 | 0 | 0 | 6 |

Isolation: each commit takes its own slice to 0 and no slice grows; C5 also clears part of the C3 / C4 slices because its resistance step is shared (a special's chill uses the resisted length), not because it changes their rule. Residual (6, not a claim): freeze length 0x100001 with a resistance that isn't 0: game.exe's MulDiv (FUN_00483360) divides first above 0x100000, d2d's `resisted` multiplies (2,097,000 vs 2,097,154 at -100); unreachable (a freeze length is frames from a skill calc), the same class as #155's chill residual. The author's own flow (`freeze_rules.py gen 100000 20261101`, game, `test_freeze_rules`) re-run here on the head: 225,748 cases, 0 mismatches, 0 RESFAIL. `test_freeze_rules --freeze-rules` calls the real `Fight::strike` (read: tests/game/test_freeze_rules.cpp builds a World and a fallen1 and calls `world.fight.strike(missile, 0, now, freeze)`); it reads the author's case file, it has no generator of its own.

### Other freeze paths (sweep)

| path | game.exe | d2d | covered by the PR |
|---|---|---|---|
| Glacial Spike (missile `glacialspike`, SrvHitFunc 13) | +0x34 -> FUN_0057c6c0 -> FUN_0057b230 | `area(..., freeze)` -> `Fight::strike` -> `monster_freeze` | yes (the claim) |
| `monglacialspike`, `hellmeteordown` (also SrvHitFunc 13) | same | not launched at monsters by d2d | no caller |
| Frozen Armor (player, merc) | FUN_005cac50 builds a record, +0x34 = FUN_00646ca0, calls FUN_0057c6c0 with resist on (0x5cacfd): the same resist / special / coldeffect / divisor rules | `buff_events` and `merc_buff_events` (fight/states.cpp) stand and chill any striker for calc1 frames, raw | **no**: same gap, not touched (the draft's "FUN_005cac50 not traced" is now traced: it does reach FUN_0057b230) |
| Freezing Arrow (`freezingarrowexp3`), Ice Blast, item "freezes target" | reach +0x34 per the trace (not re-traced here) | chill only / not modelled | no |
| Frozen Orb, Blizzard, Ice Bolt, Frost Nova | chill (+0x30) only | chill | n/a (no freeze) |
| player / merc as target | FUN_0057b230's player branch chills; hirelings chill | player: chill; mercs: no skill-missile hits | hireling branch has no d2d caller |

The candidate's wording stays true as long as it says Glacial Spike: the C5 subject does; the C3 subject ("a special monster is chilled, not frozen") and the body bullets are read in that context, but Frozen Armor still freezes specials in d2d, so the body should keep "Glacial Spike's freeze" in every bullet and name Frozen Armor under Limits as reaching the same game.exe function.

### CI-class checks (V run, clean throwaway worktree of bbf506c6)

- pr_guard: upstream/main..head 0 FAIL / 0 WARN; ddc8281d..head 0 / 0; each of the four commits 0 / 0.
- presubmit --quick (ruff on PATH): rc 0, 0 FAIL, 4 WARN (siblings: answer in Limits; data-gating on #155's test_cold_length / test_static_field, not this PR; build skipped).
- gcc -Wshadow -Wunused -Wsign-compare full build: 0 warnings on added lines. test_rules, test_combat, test_protocol, test_game (MPQ dir): rc 0; test_freeze_rules without arguments: SKIP (rc 77, the skip code).
- clang-tidy 23.1.4, repo .clang-tidy, on missiles.cpp, gamedata_load.cpp, ai.cpp, test_freeze_rules.cpp, test_combat.cpp at the head and on missiles.cpp / ai.cpp at each claim commit: 0 on added lines (only the pre-existing outdoor.hpp binding names x / y). The include fix sits in C3 as the author says (C5 needs only combat.hpp).
- Size: 11 files on #155's (17 with #155, under 100); new files 312 and 118 lines; touched pre-existing files over 1,500 lines (ai.cpp 3,050, gamedata_load.cpp 1,749) get one line each.
- Fast-forward: the fork's `fix/static-field` moved to 8c56e96a (SPOT155: test_static_field.cpp, static_field.py). The branch is not a fast-forward of it. Cherry-picking ddc8281d..bbf506c6 onto 8c56e96a is clean (no shared file), pr_guard 0 / 0 on the result; the tree equals bbf506c6 plus #155's fixup. After #155 merges, `git rebase --onto upstream/main ddc8281d fix/freeze-rules` replays only the four commits either way.
- Commit format: author / committer the user's noreply form, `Agent: Claude Code`, no trailer; **but the bodies are 9 to 20 lines** (standard section 8: at most three). C3's body also has "d2ds" / "dont".

### Verdict

- **C5 HOLDS WITH LIMITS**: 27,309 -> 0 (plain monsters), resisted as row 6, coldeffect < 0, MonsterFreezeDivisor, no chill. Limits: Glacial Spike only; state 0x36, the freeze events and a 0 frame freeze not modelled (ponytail, honest); freeze still `stun_until`; divisor 0 asserts in game.exe; lengths over 0x100000.
- **C3 HOLDS WITH LIMITS**: 74,031 -> 0; champions (0x0d, possessed 0x2d, ghostly 0x4d), uniques, superuniques and act bosses carry flag 8 and chill; minions (0x10) freeze (native spawn run above). Limits: hireling branch game-side only; the second shatter roll; Frozen Armor not covered.
- **C4 HOLDS**: 23,306 -> 0; the 0x76 call is FUN_00639db0, state 118, no stat (native: state 118 changes nothing, a real stat 0x76 would halve).
- **NOT READY to open.** Remaining: (1) cut every commit body to at most three lines (and fix "d2ds" / "dont"), then re-run pr_guard per commit; (2) det state hashes (need a d2d slot; not run here: the live slot is busy); (3) body: say Frozen Armor reaches the same FUN_0057b230 in game.exe and stays unpatched in d2d, keep "Glacial Spike's" on the C3 bullet; (4) after #155 merges: `git rebase --onto upstream/main ddc8281d fix/freeze-rules`, re-run the differential and pr_guard on the new head; (5) the user's go.
