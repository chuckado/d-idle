# PR candidate: boss and minion spawn stats (D1 narrowed, D2, D3, B)

> status: BUILT, local candidate on upstream main 72132b99 (card buildmods, 2026-11-01); PROOF incomplete: no independent verifier yet, no det state hashes (need a d2d slot); nothing pushed or posted. Rules: [monster-mods-trace.md](monster-mods-trace.md) (its D1 is corrected here). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-mods-draft.md](pr-mods-draft.md).

## Candidate

Branch `fix/boss-stats` in the worker host's PR checkout (worktree `buildmods`), on upstream main 72132b99 (= current upstream main). Authored and committed with the numeric-ID noreply form copied from fix/waypoint-hostile-window, body `Agent: Claude Code`, no Claude trailer.

| commit | claim | files |
|---|---|---|
| 98c248d3 | D1: a champion's, unique's or superunique's life, defence and experience come from the MonLvl row of its MonStats level (to-hit and damage stay at the raised level) | rules/monsters.hpp (`monster_stats`), tests/monsters/test_monsters.cpp |
| a5892044 | D2: a unique's (or superunique's) minions get its leveladd: level +3, experience x5 | rules/uniques.hpp (`boss_stats`), test_monsters.cpp |
| b4a1e5da | D3: a ghostly champion keeps its speed (d2d gave -33 %) | uniques.hpp, test_monsters.cpp, docs/research/re/monsters.md |
| 1915a9a9 | B: a berserker champion keeps the fixed leveladd: level +3, experience x5 (d2d +0, x1) | uniques.hpp, test_monsters.cpp |
| c048a356 | tooling: `tools/emu/boss_level.py`, `tests/game/test_boss_level.cpp` (`--boss-level CASES`) | 2 new files |

6 files, +336 / -12. Copies of the two tooling files in this repo: `tools/emu/boss_level.py`, `tools/emu/boss_level_test.cpp`. Overlap with our open PRs: none (#152 rules.hpp / test_game.cpp / store tools; #153 world and waypoint files; #155 fight / combat / gamedata files; diffed by file list). The stacked fix/freeze-rules edits uniques.hpp at line ~132 (a comment), a different hunk; fix/walk-velocity-v3 reads `boss_speed` (D3 changes its value for ghostly champions only).

## D1 re-derived: the trace's D1 was half wrong

The trace said d2d reads life, defence, AR, damage and XP at the boosted level and game.exe all of them at the base level. Re-read:

1. FUN_00573cb0 (stat init) asks FUN_006538a0 for mask 7 only (life, defence, experience; `push 7` at 0x573f78) at the MonStats level (`[ebx+edi*2+0xaa]`, stat 12 written first at 0x573e70). It writes stats 7 / 6 (life), 0x1f (defence), 0xd (experience); no to-hit, no damage. V (disasm read; run in the oracle).
2. FUN_005a0e40 (leveladd) only writes stat 12 += 3 and stat 13 x5; FUN_005a0e80 (champion) stat 12 -1, exp - 2 x (exp / 5). Nothing re-reads MonLvl at spawn. V (disasm).
3. A hit reads to-hit and damage per mode in FUN_005a4f50: FUN_006538a0 with mask 8 / 0x10 / 0x20 at FUN_00625480(unit, 12) = the CURRENT level (0x5a5049, 0x5a508b, 0x5a50d7), writing 0x13 / 0x15 / 0x16 into the hit's list. V (disasm; run in the oracle). Upstream's re/monsters.md says the same ("Per mode (FUN_005a4f50) A1 / A2 / S1 damage and to-hit").
4. The kill's experience reads stat 0xd (FUN_0057e480's caller pushes 0xd to the getter). V (disasm read).

So game.exe: life, defence, XP at the base level (x the boss percent), AR and damage at the raised level. d2d already reads AR and damage at the raised level (`monster_stats(.., level_add)`): right. Its life, defence and XP at the raised level: wrong. The trace's "2.9x AR, 1.3x damage" for d2d's bosses was wrong (its population trace watched only the spawn-time reads); the patch leaves AR and damage alone. Upstream notes: the boss code came in abedd2e8 ("Not proven against game.exe"); no doc or comment gives a reason for the raised level; re/monsters.md states leveladd as "level + 3, experience x 5" only. Not a documented maintainer choice.

D2: FUN_005a2120 runs each fixed mod's init with flag 1 on the boss, then FUN_0058f380(unit, mod, 0, FUN_005a2100) on its party with flag 0; FUN_005a0e40 has no flag test (V disasm). So minions get +3 and x5; their life row is the minion one (flag 0), read at the base level (D1). D3: FUN_005a0e80 compares the mod with 0x24 (ghostly) at 0x5a0ff8 and jumps to the end; the `sub ecx, 0x24; je 0x5a1062` that would store -33 (`push -0x21`) is unreachable (V disasm; upstream's re/monsters.md said "-33 %": corrected in the commit). B: FUN_005a1280 (berserker) never calls FUN_005a0e80, so the fixed +3 / x5 stand (V disasm; oracle).

## Differential (game.exe's own code vs real d2d builds)

Game side (`boss_level.py game`): on a synthetic unit, the real FUN_00573cb0 (life rolled on the unit seed by the real FUN_0045c3e0), FUN_005a2120 (or a minion's fixed-mod inits, flag 0) and FUN_005a4f50 mode 4 (A1); hooks listed in the docstring (stat store, player count 1, the expansion NM / Hell area level replaced by the MonStats level, skills, MonProp, events). d2d side: the REAL `make_boss` (boss_stats + monster_stats + its application) built from each commit with only the test file added, `Rng(seed)` so the life roll is the same draw. Fields: level, life, defence (armor % folded in as d2d does), experience, A1 to-hit and damage (to-hit / damage % added as d2d does), speed %.

Generator seed 20261101 (masked to 32 bits; the C++ mode reads the cases and masks the unit seed again), 136,624 cases = 36,624 boundary (every MonStats row with a level for the difficulty x 3 difficulties x {5 champion kinds, plain unique, plain superunique, minion, 13 single unique mods}) + 100,000 random (row, difficulty, kind, 1..1+diff (+3 superunique) unique mods, seed). cases.tsv md5 4de1e1fb518e, game.tsv md5 e8dbe935a392. Rows 0x137 / 0x138 left out (their own roll). d2d's rows all match game.exe's ids by index (0 skipped).

| build | D1 (c/u/s life, def, exp) | D2 (minion lvl, exp, AR, dmg) | D3 (ghostly vel) | B (berserker lvl, exp, AR, dmg) | other | life within 2 | life 0 in game |
|---|---|---|---|---|---|---|---|
| 72132b99 (before) | **286,774** | **109,003** | **5,939** | **26,868** | 1,022 | 10,887 | 2,660 |
| 98c248d3 (D1) | 803 | 109,003 | 5,939 | 26,868 | 1,022 | 36,621 | 2,660 |
| a5892044 (D2) | 803 | **0** | 5,939 | 26,868 | 1,022 | 36,621 | 2,660 |
| b4a1e5da (D3) | 803 | 0 | **0** | 26,868 | 1,022 | 36,621 | 2,660 |
| 1915a9a9 (B) / c048a356 (head) | 803 | 0 | 0 | 176 | 1,022 | 36,621 | 2,660 |

Counts are field mismatches. Each commit removes its own slice and touches no other. Residuals at the head, all classified (script check over every residual line: 0 unexplained):
- D1 slice 803 = S1: defence exactly half of game.exe's, every one a stone-skin unique / superunique in NM / Hell with two immunities (d2d skips the doubling then; FUN_005a1370 doubles first). Same cases before (hidden under the D1 rows). Not claimed.
- B slice 176 and other 1,022 = H1: damage of the BaseId 0x76 (willowisp) family: game.exe halves the champion / strong / berserker damage % (FUN_00463900 == 0x76; d2d's own ponytail). Not claimed.
- life within 2 points (36,621): game.exe keeps life in 256ths through the percent inits and, over 4096 points, FUN_005a0d20 divides by 100 first; d2d works in whole points. Grows from 10,887 because the D1 cases' big mismatches become rounding-only. Not claimed.
- life 0 in game.exe (2,660): rows with no life (missiles, traps, hirelings: never a boss); d2d floors at 1.

## Gameplay impact, Act 1 Normal (d2d before / game.exe; MonLvl L- columns, Normal)

| MonStats level (examples) | champion life / def / XP | unique life / def / XP | minion XP / AR / damage (game / d2d) |
|---|---|---|---|
| 1 (zombie1, fallen1, quillrat1: Blood Moor, Den) | 1.71 / 3.00 / 1.67 | 2.14 / 4.00 / 2.00 | 5 / 3.25 / 2.00 |
| 2 (fallenshaman1, brute1, skeleton1: Cold Plains, Cave 1-2; Bishibosh) | 1.67 / 2.00 / 1.50 | 1.89 / 2.50 / 1.75 | 5 / 2.92 / 1.33 |
| 5 (fallen2, quillrat2, sk_archer1: Black Marsh, caves) | 1.35 / 1.40 / 1.29 | 1.59 / 1.60 / 1.43 | 5 / 1.91 / 1.50 |
| 8 (corruptrogue3, skmage_fire1: Tamoe, Monastery) | 1.30 / 1.25 / 1.20 | 1.33 / 1.38 / 1.30 | 5 / 1.49 / 1.33 |
| 10 (Jail, Catacombs rows) | 1.26 / 1.20 / 1.17 | 1.26 / 1.30 / 1.28 | 5 / 1.37 / 1.29 |

Read: today d2d's champions and uniques have 1.3-2.1x the life, 1.2-4x the defence and give 1.2-2x the experience of game.exe's, while a unique's minions give one fifth of game.exe's experience and hit with 1.4-3.3x less AR and 1.3-2x less damage (they are 3 levels low). AR and damage of champions and uniques are unchanged by the patch. Ghostly packs walk at full champion speed instead of a third slower (+0 instead of -33 %). Berserker champions: level +3 and 5x experience (d2d +0 and 1x). A rough pack sum at level 2 (unique + 4.5 minions): game.exe 1,100 XP units vs d2d's 530 today (about 2x for unique packs); champion packs give about 0.67x of d2d's today.

Soak baselines (do not edit here): the d2d baseline XP/h of 9,172 (docs/d2d-soak.md) was measured with these deviations: champions and uniques over-strong and over-paid, unique packs under-paid (minions x1), Bishibosh-class bosses about 1.9x life. After the patch the baseline is not comparable 1:1: bosses die faster, unique packs pay about 2x, champion packs less, and minions hit harder. Re-measure the d2d baseline on the patched build before using it in a gate.

## Commands for the independent verifier (on the worker host)

```sh
cd <PR checkout>/buildmods                    # fix/boss-stats, head c048a356
# build test_boss_level and test_monsters as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
cd tools/emu
uv run python boss_level.py gen 100000 20261101 > cases.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python boss_level.py game cases.tsv > game.tsv   # about 6 min on 6 processes (split -n l/6)
D2_MPQ_DIR=<MPQ dir> D2_PATCH_INSTALLER=<LODPatch_114d.exe> <build>/tests/test_boss_level --boss-level cases.tsv 2>/dev/null | grep ' | ' > d2d.tsv
uv run python boss_level.py compare cases.tsv game.tsv d2d.tsv   # D1 803 (S1), D2 0, D3 0, B 176 (H1)
# before: a detached worktree at 72132b99 plus tests/game/test_boss_level.cpp from c048a356, same commands: D1 286,774, D2 109,003, D3 5,939, B 26,868
```

A verifier should write their own generator and re-derive FUN_00573cb0 / FUN_005a4f50 (which level each reads) before reading this doc (standard section 5). Worth attacking: whether any game.exe path re-reads MonLvl life or defence at stat 12 after spawn (FUN_005a6ff0 reads MonLvl HP at stat 12, FUN_005a2620 the Fire Enchanted death blast too: not traced), the area-level hook in NM / Hell, and the 0x137 / 0x138 rows.

## CI-class checks (head c048a356; d2d binaries built from 9c275771, the same C++ tree)

- pr_guard against upstream/main 72132b99: range 0 FAIL / 0 WARN; per commit 0 / 0 each (five commits).
- presubmit --quick: 0 FAIL, 2 WARN: siblings ("the diff touches pet": the word only; the rule is monster spawn; players, mercs and pets have no MonUMod: answer in the body), build skipped (--quick). The first run flagged ruff B905 in `compare` (`zip(..., strict=True)` since).
- ruff with upstream's rule set (F, E9, B, BLE001, PLW1510, S602, S604) on boss_level.py: clean.
- clang-tidy (repo .clang-tidy, Names and includes) over test_boss_level.cpp, test_monsters.cpp, ai.cpp (includes both headers): nothing in touched files; only the pre-existing outdoor.hpp structured-binding names (also on upstream main).
- gcc -Wshadow -Wunused -Wsign-compare (-fsyntax-only with the build's flags) on the same three TUs: 0 warnings.
- test_game (MPQ dir), test_rules, test_combat, test_protocol, test_monsters: rc 0; test_boss_level without arguments: SKIP (77).
- Seeds: masked to 32 bits in the Python generator and again in the C++ mode (which has no generator of its own).
- File length: rules/monsters.hpp is 3,910 lines on upstream main already (+6 here: not ours to split); the new files are 237 and 75 lines; uniques.hpp 404, test_monsters.cpp 332. PR size 6 files.

## Limits (for the body)

1. Spawn stats only. Not changed: the unique's own mods on its minions (game.exe runs them with flag 0: Extra Strong rows 14 / 12, Extra Fast, NM / Hell enchanted and Mana Burn rows; d2d gives minions no mods because its fight hooks read `mods`): the next candidate (D2b), not claimed.
2. Residuals left as they are and named in the body: S1 (stone skin x2 with two immunities), H1 (BaseId 0x76 damage halving), life rounding (256ths vs whole points), rows without life.
3. Hooked on the game side: the stat store, the player count (1), the expansion NM / Hell area level (the case's MonStats level instead: d2d's NM / Hell level source is not this claim), skills, MonProp, events, the element part of FUN_005a4f50 (elements not compared; d2d reads them at the raised level as FUN_005a4f50 does). The to-hit / damage % are added to game.exe's rows with d2d's formula (game.exe applies them to the hit's roll).
4. Not traced: FUN_005a6ff0 and FUN_005a2620 (both read MonLvl HP at stat 12 for their own purposes).
5. Not run: the det state hashes (need a d2d slot), the independent verifier.

## How to open

On upstream main directly (no stack). Under the bundle policy it waits for the next bundle slot (one open bundle at a time); re-run pr_guard, the four test binaries and the differential if upstream main moves. Push and PR commands are the coordinator's, with the user's go: `git -C <PR checkout>/buildmods push origin fix/boss-stats` then `gh pr create --repo psi29a/d2decomped --head shatteredeagle:fix/boss-stats --title "<title from the draft>" --body-file <draft body>`.

## Independent verifier (VERIFYMODS, card verifymods, 2026-11-01)

> Offline, read-mostly. Re-derived from the disassembly (tools/emu/disasm.py and a byte scan of every E8 call into the functions below) BEFORE reading this doc or the author's oracle; own generator and harness: tools/emu/verify_mods.py (game.exe under the emulator), tools/emu/verify_mods_mode.inc (d2d's real make_boss / monster_stats), tools/emu/verify_mods_compare.py. Builds of my own throwaway worktrees on the worker host (upstream main 72132b99, 98c248d3, a5892044, b4a1e5da, 1915a9a9, c048a356); the author's worktrees and branch untouched. Not run: the det state hashes (need a d2d slot).

### Which level game.exe reads, stat by stat (own table)

| stat | when | level | where (V: disasm read; R: also run in my oracle) |
|---|---|---|---|
| level (stat 12) | spawn | MonStats Level by difficulty (word +0xaa); in an expansion Nightmare / Hell game the area level (FUN_0061a1b0 -> FUN_0061dca0, 0x573e48), unless MonStats flags 0x6ce270 / 0x6ce280 | FUN_00573cb0 0x573e70 (V, R with the classic rule) |
| life (7 / 6), defence (0x1f), experience (0xd) | spawn, once | that same base level: FUN_006538a0 mask 7 | 0x573f7e (V, R) |
| leveladd: stat 12 + 3, exp x 5 | spawn, fixed mod 4, no flag test | - | FUN_005a0e40 (V, R) |
| champion: stat 12 - 1, exp - 2 x (exp / 5), +dmg% 0x19, +tohit% 0x77, speed | spawn, mods 16 / 36 / 37 / 38 (needs flag != 0) | - | FUN_005a0e80 (V, R) |
| hpmultiply: life += life x row % | spawn, fixed mod 2: flag 0 minion rows 1..3, champion 4..6, unique 7..9 | - | FUN_005a0dc0 / FUN_005a0d20 (V, R) |
| AR, damage (A1 / A2 / S1) | every hit | stat 12 at the hit (raised) | FUN_005a4f50: FUN_006538a0 masks 0x20 / 8 / 0x10 at FUN_00625480(unit, 12), 0x5a5049 / 0x5a508b / 0x5a50d7 (V) |
| enchanted / ghostly cold / mana burn / poison hit damage | spawn, after the fixed mods | stat 12 then (raised) | FUN_005a1080 0x5a1121, FUN_005a1990 & co (V; elements not compared) |
| minions | the leader's mods, each with flag 0, through the owner's minion list | - | FUN_005a2120 -> FUN_0058f380 (list at AI control +0x34, filled by FUN_0058f100 from FUN_005a0c00) -> FUN_005a2100 (V); minions are made BEFORE the fixed mods run (0x5a2135); FUN_005a0930 copies the leader's mod bytes whose MonUMod +7 is set into the minion (D2b) |
| champion companions | separate champions (FUN_0054e1e0 -> FUN_005b2f70, FUN_005a48c0 mod 16) | - | no minion-list link (FUN_005a0c00 returns for flag 4): no pack propagation (V) |
| ghostly speed | FUN_005a0e80 with edx = 36 | - | `cmp ecx, 0x24; je 0x5a106e` at 0x5a0ff8 runs before `sub ecx, 0x24; je 0x5a1062`: the -33 store is dead (V) |
| berserker | FUN_005a1280: life -75 % (FUN_005a0d20), +300 x DifficultyLevels +0x34 / 100 to 0x19 / 0x77, no FUN_005a0e80 | the fixed leveladd stands: +3, x5 | V, R |

The other readers of MonLvl at stat 12 (byte scan of every call to FUN_006538a0: 0x573f7e, 0x5a26ff, 0x5a3f8f, 0x5a5058 / 0x5a509a / 0x5a50e6 / 0x5a527a, 0x5a71d6, 0x5c4ec8, 0x5c57a4, plus client display 0x4a7xxx / 0x4ecxxx): FUN_005a2620 (Fire Enchanted death blast, mask 1), FUN_005a6ff0 (DT start, missile 0x75 blast, mask 1), FUN_005a3ef0 (suicideminion blast), FUN_005c4df0 (Corpse Explosion, mask 1), FUN_005c56c0 (Revive, mask 1). Each takes MonStats HP % x MonLvl HP at the RAISED level for its own damage or life; none writes the unit's own life, defence or experience. **They do not change D1** (the unit's own stats). But d2d feeds the unit's rolled life (`monster.stats.hit_points`, with the boss hp % and the player bonus) into its equivalents, so the patch CHANGES their input for boss units: `Fight::fire_blast` (deaths.cpp:177), the monster Corpse Explosion (ai.cpp:2364), Resurrect (ai.cpp:2219), the player's Corpse Explosion (already a `// ponytail:` in fight.hpp). Example, Normal unique of MonStats level 2 with Fire Enchanted: game.exe's blast input is maxHP % x HP(5); d2d's is about 3x that before the patch and about 1.6x after (unique +300 % life). Closer, still not equal: a separate candidate (feed FUN_006538a0's mask-1 max at stats.level), and a line in the body's callers list.

XP-path sweep (d2d): the kill awards `monster.stats.exp` (deaths.cpp:212 / 225 / 243) with the level penalty at `monster.stats.level` (kill_exp, combat.hpp:618), as before; nothing re-boosts experience or re-reads MonLvl after make_boss; `monster_stats` callers: make_monster (level_add 0), make_boss, summons.cpp:143 (level_add 0): for level_add 0 the patch is a no-op. No other path computes boss stats (grep boss_stats / champion / minion / exp_mult / level_add).

Upstream's reason for the raised level (step b): none. abedd2e8 ("rules: champions and uniques from game.exe's spawn code") says "Not proven against game.exe (monster unit seeds aren't emulated)"; re/monsters.md line 394 repeats it.

### Differential (own harness, 120,000 cases)

Four seeds x 30,000 (0x1001, 0x2002, 0x3003, 0x4004, each masked to 32 bits in both generators), the first 768 of each the 16 edge rows (0, 1, 2, 3, 5, 19, 156, 229, 250, 300, 400, 500, 571, 650, 700, 733: first / last rows, low and high levels where +3 clamps at MonLvl's last row) x 3 difficulties x every kind; then random row, difficulty, kind (u / s / c / m / p), unique mods from {5, 6, 8, 9, 17, 18, 25, 27, 28} (0..3), champion kind from {16, 36..39}. Field mismatches (level, life, defence, experience, velocity):

| build | D1 | D2 | D3 | B | life +-2 | no-life rows | stone skin | rows 0x137 / 0x138 |
|---|---|---|---|---|---|---|---|---|
| upstream main 72132b99 | 143,984 | 36,203 | 3,427 | 7,300 | 4,875 | 4,089 | 5,526 | 804 |
| 98c248d3 (D1) | 0 | 36,203 | 3,427 | 7,300 | 18,591 | 4,089 | 327 | 472 |
| a5892044 (D2) | C2D1 | C2D2 | C2D3 | C2B | C2R | C2N | C2S | C2X |
| b4a1e5da (D3) | C3D1 | C3D2 | C3D3 | C3B | C3R | C3N | C3S | C3X |
| 1915a9a9 (B) | C4D1 | C4D2 | C4D3 | C4B | C4R | C4N | C4S | C4X |
| c048a356 (head) | 0 | 0 | 0 | 0 | 18,591 | 4,089 | 327 | 313 |

Residuals, each shown not to depend on the claims: life +-2 (FUN_005a0d20 works in 256ths, d2d in whole points: identical from 98c248d3 on; before, these cases were inside the larger D1 mismatch); no-life rows (game.exe life 0, d2d clamps to 1 and adds the hp %: the same 4,089 in every build); stone skin (d2d keeps half of game.exe's defence for a unique with mod 28 and two immunities: `resist()` returns before setting double_defense; the same 327 in every build from 98c248d3 on, a factor 2, level-free); rows 0x137 / 0x138 (my harness does not hook their extra FUN_0045c390 roll before the life roll: harness limit, the author excludes them too). Plain monsters (kind p): 0 mismatches outside those rows in every build. The BaseId 0x76 halving (H1) is in damage, which my harness does not compare.

```sh
# worker host, PYTHONPATH = upstream tools/emu, D2_MPQ_DIR / D2_PATCH_INSTALLER as for the other oracles
python verify_mods.py 30000 0x1001 > g.txt
build/tests/test_verify_mods --verify-mods 0x1001 30000 2>/dev/null | grep ' | ' > d.txt   # verify_mods_mode.inc in a test_verify_mods.cpp wrapper
python verify_mods_compare.py g.txt d.txt
```

### CI-class checks (my runs, head c048a356)

pr_guard 0 FAIL / 0 WARN for the range and each of the five commits; presubmit --quick (ruff on PATH): 0 FAIL, 2 WARN (siblings "pet": summons.cpp:143 passes level_add 0, unchanged; build skipped); test_game (data dir), test_rules, test_combat, test_protocol, test_monsters rc 0 (with -Wno-error=maybe-uninitialized -Wno-error=array-bounds); gcc -Wshadow -Wunused -Wsign-compare on test_monsters.cpp, test_boss_level.cpp, ai.cpp: 0 warnings in touched files; clang-tidy (CI's Names and includes, tidy-pr-select over the 208 TUs that include a touched file): TIDY; the reviewer mode test_boss_level calls the real make_monster / make_boss; files 6 (monsters.hpp 3,910 lines on upstream already, +9 / -3; uniques.hpp 404; test_monsters.cpp 332; new 237 and 75); replays cleanly on upstream main (still 72132b99); #152 / #153 / #155 / #156 share no file.

### Impact tables and the draft

Recomputed from MonLvl's L- columns (Normal): every cell of the Act 1 table matches except level 10 champion life: 1.14, not 1.26. So "champions 1.3-1.7x life" (BACKLOG, bundle queue) should read 1.1-1.7x and the draft's "1.3-2.1x the life" 1.1-2.1x. The pack sums (level 2: unique pack 1,100 vs 530, about 2x; champion pack 0.67x) recompute as stated and are labelled rough (same type and Exp % assumed). The soak-baseline paragraph is accurate and not overclaiming.

Draft review (title fine, no overclaim found in the four claim sections) needs these edits before it opens: (1) the level-10 cell and the two ranges above; (2) Limits: replace "Not traced: FUN_005a6ff0 and FUN_005a2620" with the traced result above and list the d2d callers whose input now changes (fire_blast, monster Corpse Explosion, Resurrect, player Corpse Explosion); (3) say the "MonStats level" in the code comment and section 1 is the classic / Normal rule (game.exe takes the area level in an expansion Nightmare / Hell game; d2d has no area-level rule, before or after); (4) answer the siblings WARN in the body.

### Verdict

- D1 (narrow form): HOLDS WITH LIMITS: 143,984 -> 0; limits: base level = MonStats Level only in Normal / classic; the stat-12 MonLvl-life readers above change input in d2d (not in game.exe).
- D2: HOLDS WITH LIMITS: 36,203 -> 0; the leader-side owner lookup in FUN_0058f380 was read, not run (the oracle runs the per-minion handlers as FUN_005a2100 calls them); D2b not built.
- D3: HOLDS: 3,427 -> 0 (dead store confirmed).
- B: HOLDS: 7,300 -> 0.
- Claim isolation: ISOLATION.
- NOT READY to open until: the det state hashes (d2d slot) and the four draft edits above. No code fix needed.
