# PR proof: Frozen Armor's answer to a melee striker follows game.exe's freeze rule (ARMORHIT)

> status: BUILT, local candidate `fix/armor-hit` on the worker host (base upstream main 72132b99); differential run before and after on real d2d builds; **NOT READY**: no independent verifier, no det state hashes. Nothing pushed or posted (card armorhit, 2026-11-01). Follows [sk-armor-trace.md](sk-armor-trace.md) (SKARMOR: D1, the game-side oracle) and [freeze-rules-trace.md](freeze-rules-trace.md) (the freeze rule for missiles). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-armorhit-draft.md](pr-armorhit-draft.md).

Tags: **V run** game.exe's own code run under upstream's emulator by this agent; **V read** disassembled with upstream's `tools/emu/disasm.py` by this agent (addresses given); **D** 1.14d tables as the emulator loads them; **S** d2d source at 72132b99; **I** inferred; **U** unknown.

## The user's question, answered

Is Frozen Armor's behaviour on d2d (the wearer's armor striking attackers) the same as game.exe? **No.** On upstream main, every melee striker whose blow did damage is both frozen (d2d's stand, `stun_until`) and chilled for the raw calc1 frames, whatever it is. game.exe never sets both: it resists the length by the striker's cold resistance (less the wearer's Cold Mastery pierce, none through immunity); a champion, unique, superunique or MonStats boss (Andariel, Bishibosh, Corpsefire, Blood Raven...) is only **chilled** (and not at all at coldeffect 0); a plain monster is **frozen** only at a negative coldeffect, for the length / MonsterFreezeDivisor (1 / 2 / 4 by difficulty), and is not chilled; a cold-immune striker gets nothing. On the 108,506-case differential, d2d differs from game.exe on 95,920 cases: every case where the blow did damage. With the claim commit it differs on none.

## Claim (one commit)

d2d `Fight::buff_events` (components/game/fight/states.cpp, upstream 72132b99:150-161, S): `stun_until` and `chill_until` = now + max(calc1, 1) frames for every damaging melee striker.

The commit: the length is calc1 resisted by the striker's cold resistance less the wearer's cold pierce (no pierce at 100 or more, -100 at least); a special (MonStats boss column, or a champion / unique / superunique) is chilled for it with `land`'s chill rule (nothing at coldeffect 0, / MonsterColdDivisor when negative, at least 1 tick); any other is frozen only at coldeffect < 0, for length / MonsterFreezeDivisor (a new `GameData::freeze_divisor` read from DifficultyLevels), and not chilled. `merc_buff_events` (the Act 3 merc's Frozen Armor) is NOT changed (not in the differential).

## Rule re-derived from the disassembly (V read, this agent, before reading the oracle's rule)

| Address | What it does | Tag |
|---|---|---|
| FUN_005cac50 0x5cac50..0x5cac84 | wearer non-null, striker non-null and of type 0 (player) or 1 (monster), else return 0; the striker's record (arg 3) when non-null must have +8 > 0 (signed `jle`), else return 0 | V read |
| 0x5cac86..0x5cacbb | the Skills row of the skill (stride 0x23c, count +0xba0) must exist; level < 0 -> 1 | V read |
| 0x5cacc4..0x5cacfd | memset 0x70 of a local record; +4 \|= 0x20; +0x34 = FUN_00646ca0(wearer, Skills +0x138 calc1, skill, level) (the calc is on the WEARER: Shiver / Chilling hard points as synergy); FUN_0057c6c0(ECX game, EDX wearer, striker, flag 1, record) | V read |
| FUN_0057c6c0 0x57c6c0..0x57c7a0 | not hostile (FUN_00554200) and record +0 & 0x1000 clear -> clear +4 bits 0-1 and return; the striker's room / town tests (FUN_00620bb0, FUN_0061ab00); a dying striker (FUN_005541b0) -> skip; flag 1 -> FUN_0057c1e0 | V read |
| FUN_0057c1e0 | ctx: +8 wearer, +0xc striker, +0x10 / +0x14 "is a monster that is not a hireling" for each; FUN_0057c060's percent scales only the rows whose +0x20 is set (for a player wearer it returns 100 anyway: 0x57c103 -> FUN_0063edc0(player) = 0); event 0xb; FUN_0057c140 (stat 0x99 zeroes +0x30 / +0x34, stat 0x76 halves them); then the 12 rows at 0x732980 through FUN_0057bf80, rows with +0x1c set only from a non-hireling monster | V read |
| row 6 of 0x732980 | `{+0x34, res 0x2b, max 0x2c, pierce 0x14f, absorb -1, -1, reduction index 0, +0x1c 0, +0x20 0}`: the freeze length is resisted by cold resistance less cold pierce, never scaled by FUN_0057c060, no flat absorb, processed for a player wearer | V run (table bytes read from the image) |
| FUN_0057be00 | res = stat 0x2b of the striker; pierce (stat 0x14f of the wearer) is taken off only when res < 100 or the striker is not a true monster; a true monster gets no difficulty penalty and no max-resist cap; res <= -100 -> -100 | V read (0x57be00..0x57bf02) |
| FUN_0057bf80 | field <= 0 -> 0; res >= 100 -> field 0; res != 0 -> MulDiv(field, 100 - res, 100) (FUN_00483360) | V read |
| FUN_0057b230 (EAX striker; attacker, length) | length <= 0: nothing. Player striker: FUN_0057af80. Monster: state 0x36 -> nothing; MonStats +0xc & 0x40 (FUN_0063e9f0), monster data +0x16 & 8 (FUN_005a0180) or a hireling (FUN_0063ee90) -> FUN_0057af80(length) (chill), no freeze. Else coldeffect (FUN_0057af30: MonStats +0x168 + difficulty, signed) >= 0 -> nothing; else length := length / DifficultyLevels +0x14 (MonsterFreezeDivisor, `idiv`, 0 asserts) and state 1 to frame + length (an existing state 1 only lengthened, FUN_00625330 / FUN_00625310), timer event 0xc at that end; then FUN_00540e60(striker, 2) and **event 2 at frame + length + 1** (0x57b3c4..0x57b3e9) | V read |
| FUN_0057af80 (EDI unit; attacker, length) | length <= 0: nothing. coldeffect: a monster's MonStats +0x168 + difficulty, -50 for anything else; 0 -> nothing; a monster with coldeffect < 0: length / MonsterColdDivisor (+0x18, when not 0); at least 1; state 0xb to frame + length (existing only lengthened); stats 0x43 / 0x44 / 0x45 = coldeffect on the state's list; a 20 % roll on the unit's seed for state 0x6b | V read |

That is exactly what `tools/emu/freeze_rules.py` (DECOMPFREEZE, verified by VERIFYFREEZE) and SKARMOR's oracle assume, and what the claim commit codes.

## Differential (game.exe natively vs real d2d builds)

- d2d side: `test_armor_hit --armor-hit CASES` (new binary tests/game/test_armor_hit.cpp; copy in tools/emu/armor_hit_mode.inc): a Sorceress with a frozenarmor self state at the case's level, Shiver / Chilling hard points through `Fight::skill_base`, cold pierce in `psum`, one Fallen made the case's monster (Boss kind, MonStats boss column, coldeffect, cold resistance through `boss_res`, difficulty), a `Foe` with the monster in `melee_by` (and in `damaged_by` when the case's damage is over 0), then the REAL `Fight::buff_events`; prints `id chill freeze calc1` (ticks from now to `chill_until` / `stun_until`, or `-`; d2d's own calc1).
- game side: `tools/emu/armor_hit.py game` (SKARMOR's Armor runner on `freeze_rules.Game`): game.exe's FUN_005cac50 -> FUN_0057c6c0 -> FUN_0057c1e0 / FUN_0057af80 / FUN_0057b230 natively; hooks: what freeze_rules.Game hooks (stat getters, one MonStats row with only the boss byte and coldeffect, DifficultyLevels from the MPQ, the state-list calls that carry the end frame, state 0x36 off), plus the calc (returns `frames(level, synergy)`) and the overlay. Every calc call checked to be calc1 of skill 40 (95,920 calls, 0 others).
- cases: `armor_hit.py gen 50000 0x41524d48` (seed masked to 32 bits; the C++ mode has no generator): 58,506 boundary (3 difficulties x 5 Boss kinds x boss column x coldeffect {-100, -50, -1, 0, 1, 25} x 13 resistances -100..150 x Frozen Armor level {1, 2, 5, 10, 20, 30} x synergy points {0, 1, 7, 40}; a pierce block {1, 20, 50, 99, 100, 101} x 13 resistances x 5 kinds x level {1, 20}; damage-0 cases) + 50,000 random (level 1..40, synergy 0..60, pierce 0..120, coldeffect -128..127). 108,506 cases; cases.tsv md5 prefix 521488a95640, game.tsv a707f7181af0.

| Build | mismatches | calc1 differs | RESFAIL |
|---|---|---|---|
| before: upstream 72132b99 + the test file (tooling commit's files only) | **95,920 of 108,506** | 0 | 0 |
| after: fix/armor-hit head | **0 of 108,506** | 0 | 0 |

Per attacker class (before -> after; a case is counted under its kind, "boss column" when the MonStats boss byte is set):

| class | cases | before | after |
|---|---|---|---|
| plain | 25,459 | 20,533 | 0 |
| minion | 12,523 | 10,926 | 0 |
| champion | 12,370 | 10,792 | 0 |
| unique | 12,501 | 10,888 | 0 |
| superunique | 12,579 | 10,915 | 0 |
| boss column (Andariel's kind of row) | 33,074 | 31,866 | 0 |
| cold res none or negative | 53,051 | 44,269 | 0 |
| cold-resistant (1..99) | 36,885 | 34,345 | 0 |
| cold-immune (100+) | 18,570 | 17,306 | 0 |
| Cold Mastery pierce 0 | 89,467 | 81,152 | 0 |
| Cold Mastery pierce > 0 | 19,039 | 14,768 | 0 |
| Frozen Armor level 1-5 / 6-20 / 21+ | 35,499 / 38,760 / 34,247 | 33,940 / 33,969 / 28,011 | 0 / 0 / 0 |
| Normal / Nightmare / Hell | 36,183 / 36,092 / 36,231 | 31,943 / 31,962 / 32,015 | 0 / 0 / 0 |
| blow did no damage | 12,586 | 0 | 0 |

The cases that agree before are exactly the no-damage ones (95,920 = the damaging cases). game.exe outcomes over the case mix: 17,599 freeze only, 41,515 chill only, 49,392 nothing, 0 both; d2d before: 95,920 both, 12,586 nothing. Where game.exe freezes a resistance-0 plain monster, d2d's stand is as long in Normal, 2.0x in Nightmare and 4.0x in Hell (medians). The bucket sizes describe the case mix, not the game.

Second seed: see "CI-class checks" below.

## Where the oracle and this reading could share a mistake (for the independent verifier)

1. **The freeze's end frame.** Both sides are compared on the stored end (game.exe: state 1 to frame + length; d2d: `stun_until` = now + length x 40 ms). game.exe also re-schedules the striker's event 2 to frame + length + 1 (0x57b3d1..0x57b3e9) after cancelling its pending one (FUN_00540e60); if event 2 is the monster's next AI turn (I: not traced), a frozen monster acts again one frame after d2d lets it. This is not compared and not claimed; it applies equally to d2d's Glacial Spike freeze and to the freeze-rules candidate. A verifier should trace FUN_00540e60 / event 2 before trusting the frame.
2. **A zero-length freeze.** length / MonsterFreezeDivisor = 0 (a length under 2 / 4 in Nightmare / Hell) still sets state 1 to frame + 0 and event 2 to frame + 1 in game.exe; the oracle prints `-` (end <= frame) and d2d sets nothing, so both print `-`. Hidden by the output format; not claimed.
3. **calc1.** The game side's calc is hooked to `frames(level, synergy)` (Skills.txt restated in Python); d2d's own calc1 equals it on all 108,506 cases, but FUN_00646ca0 itself is not run, so a shared misreading of the calc formula would not show.
4. **When the event fires.** d2d runs `buff_events` in the world tick after the frame's swings; game.exe fires event 2 inside the hit's resolution. Same frame assumed (I), not compared.
5. **The trigger.** d2d puts a striker in `damaged_by` when `taken.damage > 0` (any element, ai.cpp:2698, S); game.exe tests the record's +8 (physical, 256ths). The harness sets `damaged_by` from the case's damage, so this deviation (SKARMOR D5) is not tested and not claimed.
6. **Inputs the oracle fakes.** One MonStats row (boss byte and coldeffect only), state 0x36 always off, a player wearer only. The hireling-wearer path (FUN_0057c060 returns DifficultyLevels +0x38 against bosses, but row 6 has +0x20 = 0) is read, not run; `merc_buff_events` is unchanged.

## Candidate

Branch `fix/armor-hit` in the worker host's PR checkout (worktree `armorhit`), base upstream main 72132b99 (= `gh api repos/psi29a/d2decomped/commits/main` on 2026-11-01). Author and committer the numeric-ID noreply form, bodies two lines plus `Agent: Claude Code`, no Claude trailer.

| commit | what | files |
|---|---|---|
| 36cbe015 | claim: Frozen Armor freezes a striker by the freeze rule | states.cpp (`buff_events`, `#include <uniques.hpp>`), gamedata.hpp / gamedata_load.cpp (`freeze_divisor`), docs/research/re/skills.md (event 2 line completed) |
| d5bd19c1 | tooling: `tools/emu/armor_hit.py`, `tools/emu/freeze_rules.py` (byte-identical to the freeze-rules candidate's), `tests/game/test_armor_hit.cpp` | 3 new files |

7 files, 2 commits. Copies in this repo: `tools/emu/armor_hit.py`, `tools/emu/armor_hit_mode.inc` (the test file); `tools/emu/freeze_rules.py` is already here (identical).

## CI-class checks (head d5bd19c1, worker host, offline)

- Isolation: the before tree is a detached worktree at 72132b99 with only the tooling commit's three files; its d2d output equals the first before run line for line (log timestamps aside).
- Second seed `0x5eed2026` (108,506 cases, cases md5 prefix 4a333bb6aff9, game 94fec0333b37): before **95,986**, after **0**; calc1 0, RESFAIL 0; 95,986 calc calls, all calc1 of skill 40.
- pr_guard: upstream/main..head 0 FAIL / 0 WARN; each commit 0 / 0.
- presubmit --quick: 0 FAIL, 3 WARN after the skills.md line (docs-sync answered by it): siblings unit kinds (merc wearer not changed, player strikers not a d2d path: Limits 1-2), siblings hit paths (the missile freeze is the freeze-rules candidate; Shiver / Chilling unchanged), build skipped by --quick. ruff (presubmit, upstream rules incl. B905): ok on both added Python files.
- gcc -Wshadow -Wunused -Wsign-compare build of test_game, test_rules, test_combat, test_protocol, test_monsters, test_armor_hit: 0 warnings on added lines.
- clang-tidy (repo .clang-tidy, names and includes) on states.cpp, gamedata_load.cpp, test_armor_hit.cpp: 0 in touched lines (only the pre-existing outdoor.hpp / merc.hpp binding names x / y, also on upstream main). `tidy-pr-select.py` needs its build directory and was bypassed (direct run on the three TUs).
- Tests (with the MPQ dir and patch): test_monsters, test_rules, test_combat, test_protocol, test_game rc 0; test_armor_hit without arguments SKIP.
- File length: states.cpp 436 lines after the change, test_armor_hit.cpp 121, armor_hit.py 174; gamedata_load.cpp is already 1,747 on upstream main and gets one line (as the freeze-rules candidate's). test_game.cpp untouched.

Commands for the independent verifier (worker host, PR checkout `armorhit`):

```sh
cd <PR checkout>/armorhit                          # fix/armor-hit, head d5bd19c1
# build test_armor_hit as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
cd tools/emu
uv run python armor_hit.py gen 50000 0x41524d48 > cases.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python armor_hit.py game cases.tsv > game.tsv
D2_MPQ_DIR=<MPQ dir> D2_PATCH_INSTALLER=<patch> <build>/tests/test_armor_hit --armor-hit cases.tsv > d2d.tsv
uv run python armor_hit.py compare cases.tsv game.tsv d2d.tsv    # mismatches 0; before (72132b99 + the 3 tooling files): 95,920
```

A verifier should write their own generator and seed and re-derive FUN_005cac50 / FUN_0057b230 (and trace event 2, point 1 above) before reading this doc.

## Dependencies

- PR #155 (head 8c56e96a) and PR #156 (head 3f418788): `git merge-tree` clean both ways for each (no shared hunk; #155 touches gamedata.hpp / gamedata_load.cpp next to these lines: the load line is placed after `curse_divisor` so the two do not meet).
- Local candidate fix/freeze-rules-v2 (409903a6): merges clean (git keeps the identical `freeze_divisor` member and `tools/emu/freeze_rules.py` once), but the merged tree then reads MonsterFreezeDivisor twice in gamedata_load.cpp (harmless; drop one line in whichever lands second). The freeze rule is coded twice after both land (`rules::monster_freeze` there, inline in `buff_events` here): a follow-up can make `buff_events` call `monster_freeze`.

## Not covered (limits)

1. The player wearer only; the Act 3 merc's Frozen Armor (`merc_buff_events`) keeps the stun and chill.
2. Monster strikers only (a player striker, PvP / co-op, is chilled in game.exe: FUN_0057b230's player branch; d2d has no player-on-player path here).
3. Not modelled (ponytail notes in the commit): state 0x36, the chill's 20 % state 0x6b roll, the freeze's event 2 at end + 1, a zero-length freeze.
4. D2 (group-1 armors stack in d2d), D3 / D4 (Energy Shield) and D5 (the trigger) are not in this claim.
5. Not run: independent verifier, det state hashes (need a d2d slot), Windows / MSVC build.

## Bot implication

Until this is fixed (upstream or in the lab binary), every d2d soak with Frozen Armor up is optimistic in melee: each damaging hit stands the striker for calc1 frames (1.2 s at level 1) whatever it is. Against bosses and champions it should only chill them (or do nothing at coldeffect 0); in Nightmare / Hell plain monsters stand 2x / 4x too long; cold-immune strikers should get nothing. **Andariel soaks are the main case** (MonStats boss: game.exe at most chills her; d2d stops her for 1.2 s on every damaging swing), then Bishibosh, Corpsefire, Blood Raven, the Countess and champion packs. Mark melee-survival and Andariel numbers measured on d2d with the armor up "d2d, D-ARMORHIT" until the fix runs.

## Fork proposal (not done)

Under the standing fork authorisation: carry the claim commit privately as `lab-armorhit` on the private fork (on top of the current lab set, after the lab-nexthit commit), build `bin/d2d.lab-armorhit`, re-run `test_armor_hit` against the oracle on that build (expect 0), and switch Frozen Armor soaks (Andariel first) to it. Proposed only; the coordinator or the user decides.

## Shared-file lines for the merger

docs/BACKLOG.md, add: `- ARMORHIT: independent verifier + det state hashes for fix/armor-hit (Frozen Armor striker effect = game.exe freeze rule: 95,920 -> 0 of 108,506); trace FUN_00540e60 / event 2 (the freeze's end + 1 frame, all freeze paths); proposed lab-armorhit fork carry (docs/research/pr-armorhit-proof.md)`

docs/INDEX.md, research section, add: `- docs/research/pr-armorhit-proof.md / pr-armorhit-draft.md: Frozen Armor striker effect (SKARMOR D1) local upstream candidate, d2d buff_events vs game.exe FUN_005cac50 differential (ARMORHIT)`

CHANGELOG.md (unreleased, docs), add: `- docs: ARMORHIT differential: Frozen Armor's answer to a melee striker on d2d differs from game.exe on every damaging case (95,920 of 108,506); local upstream candidate fix/armor-hit brings it to 0 (docs/research/pr-armorhit-proof.md).`
