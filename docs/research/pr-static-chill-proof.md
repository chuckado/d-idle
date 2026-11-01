# PR proof: a cold missile chills whatever damage it deals (addition to #155)

> status: BUILT, local candidate, independent verifier run (card verifychill, 2026-10-31: HOLDS WITH LIMITS, section at the end); PROOF incomplete only for the det state hashes (need a d2d slot); nothing pushed or posted (card chillpatch, 2026-10-31). Follows docs/research/static-field-chill-trace.md (Q2, C1 / C2, which this doc re-derives and partly corrects). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-static-chill-draft.md](pr-static-chill-draft.md).

Tags: **V** = run in game.exe's own code under the emulator or read with upstream's `tools/emu/disasm.py` (said which); **D** = patched 1.14d tables read with mpq-cat (docs/process/mpq-data-reading.md); **S** = source read (d2d file:line at the candidate's base 6270d162); **I** = inferred; **U** = unknown.

## Claim (C1)

d2d's `missile_blow` (components/rules/combat.hpp:457 at 6270d162) sets a cold missile's `chill_ticks` only `if (resisted_damage > 0)`, i.e. when the cold left at least one whole point after resistance. game.exe chills whatever the damage came to: FUN_0057c6c0 hands the record's resisted cold length (+0x30) to FUN_0057af80 with no test on the cold damage (+0x24). The commit drops the whole-point gate; the length, the resistance, the pierce and `Fight::land`'s chill rules are unchanged (they already match). One behaviour, one commit.

## Rule re-derived from the disassembly (independently of cold_length.py)

| Address | What it does | Tag |
|---|---|---|
| FUN_0057c6c0 0x57c793..0x57c7a0 | calls FUN_0057c1e0 (resistances) when its 2nd argument is not 0 | V read |
| FUN_0057c6c0 0x57c852..0x57c874 | life -= record +0x4c (dead under 0x100) | V read |
| FUN_0057c6c0 0x57c8a0..0x57c8a8 | `push [rec+0x30]; push [ebp-8]; call FUN_0057af80`: no test on +0x24 or on the life between the subtraction and the call | V read |
| FUN_0057c1e0 0x57c304 | FUN_0057c140 before the element loop | V read |
| FUN_0057c140 | defender stat 0x99 (FUN_00625500) not 0: +0x30 and +0x34 = 0; else stat 0x76 not 0: both `sar 1` (signed halving) | V read |
| table 0x732980 row 5 | dwords: offset 0x30, res 0x2b, maxres 0x2c, pierce 0x14f, absorb -1 / -1, reduction index 0, monster-only 0, damage-% 0: the length goes through cold resistance less cold pierce, no flat reduction (index 0 is the zero slot [ebp-0x18]), no absorb, no damage % | V (bytes read from game.exe with pefile) |
| FUN_0057c1e0 loop 0x57c380 | FUN_0057bf80 per row; rows with the monster-only flag (9..11) stop the loop for a non-monster attacker; row 5 is before them | V read |
| FUN_0057be00 | res = defender stat; pierce (attacker stat) off when res < 100 or the defender is not a plain monster; no difficulty penalty for a monster defender; at or under 0: at least -100; a monster gets no max-resistance cap | V read |
| FUN_0057bf80 | value <= 0: 0; res capped at 100; `MulDiv(v, 100 - res, 100)` (res 100: 0); untouched at res 0 | V read |
| FUN_0057af80 | length <= 0: nothing. Monster with a MonStats row: coldeffect = signed byte row +0x168 + difficulty, 0: nothing; any other unit: -50. Monster with coldeffect < 0: `length / DifficultyLevels +0x18` (idiv, when not 0). Length under 1: 1. End = game frame +0xa8 + length. No chill state (FUN_006256b0 state 0xb): a new one (FUN_006251f0, then stats 0x43 / 0x44 / 0x45 += coldeffect, FUN_00623f50); running: lengthened only if the new end is later (FUN_00625330 / FUN_00625310). Then a 20 % roll on the target's own seed sets state 0x6b (monsters only) | V read, and run in the oracle |
| MonsterColdDivisor | 1 / 2 / 4 (Normal / Nightmare / Hell) | D |

d2d (S): `missile_blow` uses `resisted(damage.elen, res)` with the same res (`res - pierce` when res < 100, at least -100); `Fight::land` (strike.cpp:429-437) has the coldeffect 0 / divisor for negative / at least 1 / only lengthened rules.

## Boundary decisions, per case

| Case | game.exe | d2d before | Verdict | Tag |
|---|---|---|---|---|
| resisted cold under 1 point, length > 0 | chilled | not chilled (C1) | **deviation, fixed** | V (differential) |
| resistance >= 100 | length 0, no chill | no chill | same | V (0 of the res >= 100 cases differ) |
| negative resistance | length x (100 - res) / 100, res at least -100 | same | same | V |
| pierce | off when res < 100, floor -100 | same | same | V |
| running chill longer than the new one | kept | kept | same | V (prior cases) |
| running chill and a C1 hit | lengthened | untouched (C1) | deviation, fixed (627 of the before mismatches) | V |
| minimum 1 tick, divisor only for coldeffect < 0, coldeffect 0 never | as read | same (land) | same | V |
| stats 0x99 / 0x76 on the target (C2) | zero / halve | not read for a monster | **not a reachable deviation**: see C2 | V read + D |
| freeze to chill on bosses, champions, uniques, hirelings (FUN_0057b230) | chill instead of freeze | freeze path not compared | out of this claim: candidate C3 | V read; d2d U |

### C2: corrected, no patch

game.exe reads 0x99 / 0x76 on the defender (V read). No 1.14d monster carries either stat: no MonProp, MonUMod, Skills or States row names `nofreeze`, `half-freeze`, `item_cannotbefrozen` or `item_halffreezeduration` (only ItemStatCost and Properties define them) (D); game.exe's other `push 0x99` sites (0x4f6804, 0x4fb45a, 0x50803e, 0x50ba9c, 0x639f7a) are string / UI lookups, not stat writes, and every `push 0x76` on a monster (0x5a45d1, 0x5a4607, 0x5a4c00, 0x5a4d99, 0x5b1fb0, ...) feeds FUN_00639db0 / FUN_00639df0, the state setter / tester (it bounds the id by the state count at data tables +0xc4 and calls FUN_00625a70), i.e. **state 118 `corpse_noselect`**, not the stat (V read; state name D). Player targets already get both stats in d2d (`chill_length`, combat.hpp:589, S). So the trace's C2 mismatches (stat set on a monster) are inputs 1.14d never produces: no claim. The trace's figure "12,804 of 81,559 with neither stat" is C1 alone and stands (I: same rule; this doc's own count is below).

### Found on the way (candidates, not in the claim)

- **C4** (V read, not proven by a differential): d2d's `Monster::half_freeze` for Blood Raven (`act_boss` base 0x10b, components/rules/uniques.hpp) and the Countess (components/game/ai.cpp:2965, "FUN_005a49b0 case 6: stat 0x76") reads that FUN_00639db0(unit, 0x76, 1) call as stat 0x76; it sets state 118 `corpse_noselect`. d2d halves their freezes (missiles.cpp:658); game.exe does not.
- **C3** (V read on game.exe, d2d U): FUN_0057b230 gives a MonStats boss, a champion / unique (FUN_005a0180 flag 8) or a hireling a chill (FUN_0057af80) instead of the freeze; d2d's freeze (missiles.cpp:659-662) adds chill ticks and a stun to any monster.
- player_blow's cold (combat.hpp:542, 548, 558) keeps the same whole-point gate; game.exe's melee record goes through the same FUN_0057c6c0 (I), but its length source (stat 56, the 50-tick conversion minimum) is not re-derived here: own claim later.

## Differential (real d2d builds both sides)

- game side: `tools/emu/chill_length.py` (upstream tools/emu, in the tooling commit): game.exe's FUN_0057c6c0 natively, including FUN_0057c1e0 / FUN_0057c140 / FUN_0057be00 / FUN_0057bf80 and FUN_0057af80 itself (life_apply.py's hook on it removed); prints the ticks to the chill's end or `none`. Hooks listed in its docstring (stat getters / setter, unit-kind predicates, the difficulty row with MonsterColdDivisor 1 / 2 / 4, events, logging, the state-list calls that only record the end, the speed stats and the 0x6b roll).
- d2d side: `test_game --cold-length SEED N` (copy in tools/emu/chill_length_mode.inc): the REAL `missile_blow` on the REAL `Fight::target_of` of a fallen1 in a `World`, then the REAL `Fight::land`; the case's coldeffect written into the loaded MonStats row and restored; res through `boss_res[4]` (every case checks `target.res[4] == res`, 0 RESFAIL).
- cases (seed 1): 21,312 boundary (3 difficulties x coldeffect {-50, -1, 0, 25} x 15 resistances -150..150 x 9 rolls 1..1024 256ths x 12 lengths 0..375; running chills {1, 37, 1000} ticks; pierce {1, 40, 100} at res {50, 99, 100, 150}) and 100,000 random (roll to 0x4000, length 0..500, res -150..150, pierce 0..60, coldeffect mostly real values or -100..100, running chill 30 %).

| Build | Mismatches of 121,312 | Boundary | Random |
|---|---|---|---|
| before: 6270d162 (upstream 72132b99 + the two #155 commits) + the mode | **14,239** (13,612 not chilled, 627 running chill not lengthened; 0 at res >= 100, 0 at coldeffect 0, 0 at length 0) | 4,973 of 21,312 | 9,266 of 100,000 |
| after: fix/static-field-chill 79c9a496 | **0** | 0 | 0 |

game.txt md5 prefix 310ffcbf207a.

## Commands for the independent verifier (on the worker host, PR checkout `chill`)

```sh
cd <PR checkout>/chill                     # fix/static-field-chill, head 79c9a496
# build test_game as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
cd tools/emu && D2_MPQ_DIR=<dir with bin/game.exe> uv run python chill_length.py 100000 1 > g.txt
D2_MPQ_DIR=<MPQ dir> <build>/tests/test_game --cold-length 1 100000 | grep -E '^[BR][0-9]' | diff g.txt -   # empty
# before: a detached worktree at 6270d162 with the tooling commit's tests/game/test_game.cpp, same two commands: 14,239 lines differ
```

## CI-class checks (fix/static-field-chill 79c9a496)

- pr_guard against upstream/main (72132b99, fetched today: still the tip): 0 FAIL, 0 WARN.
- presubmit --quick: 0 FAIL, 3 WARN (siblings: players / mercs / pets; docs-sync; build skipped), ruff clean on 2 files. Answers for the body: players are chilled through `chill_length` (already reads both stats), mercs' and pets' cold missiles go through the same `missile_blow` and get the fix; no .md describes this rule.
- clang-tidy 23.1.4 with the repo .clang-tidy over test_game.cpp and test_combat.cpp (both include combat.hpp): nothing in touched files (only the pre-existing outdoor.hpp structured-binding names, also on upstream main).
- gcc -Wshadow -Wunused -Wsign-compare over test_game.cpp, test_combat.cpp, strike.cpp, missiles.cpp: 0 warnings on added lines (two would-be shadows of main's `rng` / `made` were named `blow_rng` / `fallen_made` before the commit).
- test_game (with the MPQ dir), test_rules, test_combat, test_protocol: rc 0.
- Merge check against the walk-velocity candidate (fix/walk-velocity-v2): `git merge-tree` clean; it touches combat.hpp:62 (Fighter) and test_game.cpp after `character.panel`, this candidate combat.hpp:423-460 and test_game.cpp after `--tome-use`. Disjoint behaviour: walk-velocity patches the chill's SPEED (coldeffect into stat 0x43, floor 25); this claim only whether and how long the chill runs.
- PR size: #155 has 7 files; this adds 1 new file (tools/emu/chill_length.py; combat.hpp and test_game.cpp are already in it): 8 files, 4 commits.

## Callers

`missile_blow` callers (S): missiles.cpp:656, 790; states.cpp:165, 227, 254, 352; pets.cpp:120, 161, 942; strike.cpp:673, 679, 771; tick.cpp:403 (shrine). Only `etype == 2` blows change: they now carry `chill_ticks` when the resisted cold is under a point; `Fight::land` applies them as before. ai.cpp's `missile_blow` is a different function (ai.hpp:654).

## Limits

1. Only skill missiles (`missile_blow`); the melee / weapon cold in `player_blow` keeps the whole-point gate (ponytail in the code).
2. Record flags 0x100 / 0x200 / 0x400 (FUN_0057c1e0: positive resistance ignored) and FUN_0057c6c0 with its 2nd argument 0 (no resistance step) are not exercised.
3. The 20 % state 0x6b roll consumes the target's seed in game.exe on every chill call, now on more hits; d2d has no such roll (existing ponytail in land).
4. Hooked on the game side: see the oracle's docstring; MonStats is a one-row table holding the case's coldeffect.
5. The length a skill puts in the record (ELen, FUN_00644f20) is input, proven elsewhere (skill-damage-trace).
6. Not run: the det state hashes (need a d2d slot), the independent verifier.

## How it would join #155

The branch is a fast-forward of #155's head (`git merge-base --is-ancestor origin/fix/static-field fix/static-field-chill`: yes). With the user's go, before CodeRabbit's first review of #155: `git -C <PR checkout>/chill push origin fix/static-field-chill:fix/static-field` (a normal push, no force), then edit the PR title and body as in the draft. Commits authored and committed with the numeric-ID noreply form copied from fix/static-field, body line `Agent: Claude Code`, no trailer.

## Independent verifier (card verifychill, 2026-10-31)

Role: the verifier of docs/process/upstream-pr-standard.md section 5. Offline, no game, no d2d slot. Candidate `fix/static-field-chill` (claim 46ef6c75, tooling 79c9a496) in a throwaway worktree of the PR checkout on the worker host; the author's branches and worktree were not touched. The rule was re-derived from `disasm.py` before this doc, the draft or `chill_length.py` were read.

**Verdict: HOLDS WITH LIMITS. READY to append to #155 after: the det state-hash run (needs a d2d slot), the title and body changes below, and the user's go.**

### (a) Rule re-derived from the disassembly (V read with disasm.py; table bytes with pefile)

- FUN_0057c6c0: the early returns (FUN_00554200 / record +0 & 0x1000, FUN_00620bb0, FUN_0061ab00 with the MonStats +0xd test, FUN_00457490(class, 0xf) for a monster defender, FUN_005541b0) all come before the resistance call at 0x57c7a0, and none reads the damage. From 0x57c7a5 to the call at 0x57c8a8 the only branches are the event flags (0x57c7a9..0x57c7f3, they rejoin at 0x57c7f8), the attacker-is-monster test at 0x57c830 (rejoins 0x57c844) and `jle 0x57c879` around the life step (rejoins at 0x57c879). Nothing between the resistance and `push [rec+0x30]; push attacker; call FUN_0057af80` tests +0x24 or the life: CONFIRMED. The life step runs first, so even a killing blow is chilled. The freeze (FUN_0057b230, +0x34) follows at 0x57c8b7.
- FUN_0057c1e0: context +0x10 / +0x14 = attacker / defender "is a monster and not a hireling" (FUN_0063ee90 is the hireling-class test: 271, 338, 359, 560, 561). FUN_0057c140 first (defender stat 0x99: +0x30 and +0x34 to 0; else 0x76: both halved, signed). Table 0x732980 row 5 = offset 0x30, res 0x2b, maxres 0x2c, pierce 0x14f, no absorb (-1 / -1), reduction slot 0, hireling-only 0, damage-% 0; the hireling-only rows 9..11 come after it, so row 5 always runs. CONFIRMED as the author has it.
- FUN_0057be00 for a plain monster defender: pierce taken off unless res >= 100; no difficulty penalty, no 75 + maxres cap; at or under 0: at least -100. FUN_0057bf80: value <= 0 gives 0; res >= 100 gives 0; else FUN_00483360(v, 100 - res, 100) (divides first above 0x100000).
- FUN_0057af80: length <= 0 nothing; coldeffect = MonStats +0x168 + difficulty (signed), 0 nothing; a non-monster -50; monster with coldeffect < 0: / DifficultyLevels +0x18 (MonsterColdDivisor); at least 1; only ever lengthened (FUN_00625330 / FUN_00625310); a positive coldeffect chills too (no division).
- FUN_0057b230 (freeze): a player gets FUN_0057af80 (chill); a monster with state 0x36 nothing; a MonStats boss (FUN_0063e9f0, +0xc & [0x6ce280]), monster data +0x16 & **8** (FUN_005a0180) or a hireling (FUN_0063ee90) gets FUN_0057af80 with the freeze length; any other monster is frozen only if its coldeffect is negative (FUN_0057af30), length / MonsterFreezeDivisor (+0x14), and gets no chill.
- Boundaries: resisted length 0 (res >= 100, or e.g. length 1 at res 1): no chill; resisted length 1 at a negative coldeffect and divisor 2 or 4: 0 after the division, raised to 1 tick; negative resistance lengthens (to x2 at -100); pierce at res >= 100 does nothing for a monster.

### (b) Own differential (tools/emu/verify_chill.py, tools/emu/verify_chill_mode.inc)

- game side: game.exe's FUN_0057c6c0 natively with FUN_0057c1e0 / c140 / be00 / bf80 / bf10 / FUN_00483360 / FUN_0057af80 / FUN_0057b230 / FUN_0057af30 / FUN_0063ee90 / FUN_0063e9f0 / FUN_005a0180 unhooked (the author's oracle also runs FUN_0057af80; mine also runs the freeze and the boss, unique and hireling tests natively); hooks listed in the script's docstring; DifficultyLevels read from the MPQ.
- d2d side: phase 0 the REAL `missile_blow` + `Fight::land`; phase 1 the REAL `Fight::strike` with a freeze (the hit_func 13 path), on real builds of 6270d162 (#155's head, = the fork's `fix/static-field`) and 79c9a496, each with only the mode pasted.
- Cases (seed 20261031): 201,757 = 101,757 boundary + 100,000 random; rolls -256..0x10000 (256ths), lengths -50..0xFFFFFF, res -300..300, pierce 0..300, coldeffect -128..127, running chills, monster-data flags 0 / 4 / 8 / 0xA / 0x10, MonStats boss bit, hireling classes (game only), freeze 0..200; 0 RESFAIL. game.tsv md5 616349712bdb.

| Class | before (6270d162) | after (79c9a496) |
|---|---|---|
| C1: resisted cold under one point, length > 0 (the claim) | **34,635** | **0** |
| C3: freeze on a MonStats boss / flag 8 / hireling: game.exe chills, d2d stuns and chills (phase 1) | 4,282 | 4,282 |
| C5: freeze on any other monster: game.exe freezes only (coldeffect < 0, length resisted and / MonsterFreezeDivisor), d2d stuns for the raw length and chills (phase 1) | 2,836 | 2,836 |
| R-ehi: roll max <= 0: `missile_blow` returns before the cold, game.exe chills on the length alone | 17,912 | 17,912 |
| R-big: length >= 0x100000: `resisted()` multiplies in int (overflow) where FUN_00483360 divides first | 41 | 29 |
| any other | 0 | **0** |

Every before mismatch outside C3 / C5 / R-ehi / R-big passes the whole-point test (resisted fine < 256 with roll > 0), and none is left after. C3, C5, R-ehi and R-big are not touched by the commit and are not part of the claim. The author's own commands also reproduce: `chill_length.py 100000 1` vs `test_game --cold-length 1 100000` on 79c9a496: 121,312 lines, 0 differ; the mode calls the real `missile_blow` and `Fight::land` (read in test_game.cpp), not a transcription.

Reachability of the residuals: R-ehi needs a cold length with no cold damage: no Skills.txt missile skill has it (Vengeance has cold ELen 30 and no EMin / EMax, but it is a melee skill: `player_blow`); Missiles.txt `rogue3` (cold, ELen 50, no damage) is the one row of that shape; whether it reaches `missile_blow` in d2d is not traced. R-big needs a length of 0x100000 frames (about 11.6 hours): unreachable.

### (c) Other cold-chill paths in d2d (at 79c9a496)

| Path | Covered by the commit | Note |
|---|---|---|
| `missile_blow` callers: missiles.cpp:656 (skill missiles), 790 (enemy auras incl. Holy Freeze); states.cpp:165 (Shiver Armor), 227, 352; pets.cpp:120, 161, 942; strike.cpp:673, 679, 771; tick.cpp:403 (shrine potion) | yes | game.exe path of each caller not traced; all reach FUN_0057c6c0 only by inference (I) |
| combat.hpp:545 `player_blow` skill cold (`swing.cold_pct`, Vengeance) | no | `cold > 0` gate; length `max(cold_len, 1)` (game.exe: resisted length 0 is no chill) |
| combat.hpp:551 `player_blow` gear cold (`fighter.elem[2]`) | no | `element_damage > 0` gate |
| combat.hpp:560-561 converted cold (`conv_type == 2`) | no | gate, and a 50-tick floor |
| combat.hpp:618 `monster_blow` on the player | n/a (player target) | gates on the raw roll > 0, not the resisted damage; `chill_length` reads 0x99 / 0x76 |
| deaths.cpp:164 suicide explosion on the party | no | `cold > 0` gate; the length is `resisted(element, res)` with `element` = aip[0], the element code: looks wrong (U), own candidate |
| missiles.cpp:657-662 freeze | no | C3, C4 (half_freeze), C5 |
| states.cpp:161, 195 Frozen Armor (stun + chill on the striker) | no | not cold damage; not compared |

The ponytail "player_blow's cold still chills only on a whole point" is accurate (three sites); it does not say that Vengeance, a real skill, is one of them.

### (d) CI-class checks on 79c9a496

- One claim per commit: 46ef6c75 changes only combat.hpp (the gate and its comment); 79c9a496 only tooling. pr_guard 0 FAIL / 0 WARN against upstream/main for 46ef6c75, against 46ef6c75 for 79c9a496, against 6270d162, and for the whole branch.
- test_game (MPQ dir), test_rules, test_combat, test_protocol (`-Wno-error=maybe-uninitialized -Wno-error=array-bounds`): rc 0.
- presubmit --quick with ruff 0.17.3 on PATH: against upstream/main 0 FAIL, 3 WARN (siblings, docs-sync, build skipped); against 6270d162 0 FAIL, 2 WARN. ruff check of chill_length.py clean (ruff format would reformat it, as 64 upstream tools/emu files; CI does not check format).
- gcc -Wshadow -Wunused -Wsign-compare over test_game.cpp, test_combat.cpp, strike.cpp, missiles.cpp: 0 warnings in the whole files.
- clang-tidy with the repo's .clang-tidy over test_game.cpp and test_combat.cpp: only the 4 pre-existing components/drlg/outdoor.hpp binding names (also on upstream main), nothing on an added line, nothing in a touched file.
- Size: 8 files against upstream/main 72132b99 (still its tip). Fast-forward of #155's current head (fork `fix/static-field` = 6270d162): yes. `git merge-tree` with fix/walk-velocity-v2 (10338345): clean.

### (e) Limits and the draft

- Not run: the det state hashes (gate 4, needs a slot). Record flags 0x100 / 0x200 / 0x400 and FUN_0057c6c0 without the resistance step not exercised. The 0x6b roll is hooked (it does not change the chill). Hireling targets only on the game side (d2d's cold missiles do not hit mercs). Stats 0x99 / 0x76: re-checked (D): Properties codes `nofreeze` / `half-freeze`; no MonProp, MonUMod, States, Skills, SuperUniques, MonStats, MonStats2 or Hireling cell names them or their stats.
- Title "Static Field and chill length as game.exe does" overclaims: the commit does not change the length, and only skill missiles. Proposed: `game: Static Field and cold-missile chill as game.exe does (differential proof vs game.exe)`.
- Body: "chills the monster whatever damage it deals" holds when the element's max is over 0 (R-ehi; say so). "The length ... were already right": add "for lengths under 0x100000 frames". Limits: add that Vengeance is among the `player_blow` paths that keep the gate, and that freeze handling (C3, C5) is not in the claim. Keep the existing ponytail on pierce ("isn't traced"), or update it: FUN_0057be00 skips the pierce at res >= 100 only for a plain-monster defender (V read here).
- C3 wording in this doc and BACKLOG: game.exe tests monster-data flag **8** (FUN_005a0180 edx 8). In the common D2 layout 8 is unique (superuniques carry it too) and champions carry 4; in my run flags 4 and 0x10 were frozen, only 8 chilled. "champions" in C3 is not verified; say "MonStats bosses, flag-8 (unique) monsters and hirelings". New candidates from this run (reading only, not claims): C5 (freeze length unresisted, no MonsterFreezeDivisor, chill added, coldeffect >= 0 still frozen), deaths.cpp:164 length.
