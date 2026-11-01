# Independent verifier: fix/armor-hit (Frozen Armor striker by the freeze rule) (VERIFYARMORHIT, card verifyarmorhit, 2026-11-01)

> status: verifier record for the local upstream candidate `fix/armor-hit` (claim 36cbe015, tooling d5bd19c1, base upstream main 72132b99) on the worker host. Author's proof and draft: `docs/research/pr-armorhit-proof.md` / `pr-armorhit-draft.md` (merge 4fbfbf2). Offline and read-mostly; nothing pushed or posted; the author's branch untouched. Verdict in section 7.

Tags: **V read** disassembled with upstream's `tools/emu/disasm.py` (my own dumps); **V run** game.exe's code run under upstream's emulator in my own oracle; **S** d2d source; **I** inferred; **U** not traced.

Order of work: the claim diff and the disassembly first (section 1), then my own oracle and d2d harness (section 2), then the author's proof and draft (read only after the first differential).

## 1. The chain in game.exe

1. **FUN_005cac50** (ECX game; wearer, striker, the striker's record, skill, level) (V read 0x5cac50..0x5cad33): nothing unless the striker is a player (0) or a monster (1) and, when a record is given, its +8 > 0. A zeroed 0x70 record, +4 |= 0x20 (no on-hit events), +0x34 = FUN_00646ca0(wearer, Skills +0x138 = calc1, skill, level), then FUN_0057c6c0(game, wearer -> striker, flag 1, record), then the overlay Skills +0x110.
2. **The trigger** (V read): event 2 is dispatched by FUN_005c0c30 from FUN_0057c6c0 (0x57c7d8..0x57c7f3, melee: event 5 on the attacker, then 1 `damagedinmelee` on the one struck) AFTER FUN_0057c1e0 has applied the wearer's resistances and damage reduction to that record. So +8 is the PHYSICAL damage after reduction, in 256ths (record layout: +8 physical, +0x10 fire, +0x1c lightning, +0x24 cold, +0x30 chill length, +0x34 freeze length; S, upstream skills.md). d2d keeps `taken.damage > 0` (any element, points; ai.cpp:2698, S): an elemental-only or fully physically reduced blow freezes on d2d only, a sub-point physical blow in game.exe only (SKARMOR D5, unchanged by the claim, untested by both harnesses).
3. **FUN_0057c1e0** (V read 0x57c1e0..0x57c3d0): context +0 game, +4 DifficultyLevels row, +8 wearer, +0xc striker, +0x10 / +0x14 "a monster that is not a hireling" (FUN_0063ee90 = 0) for each. FUN_0057c060's percent scales only rows with +0x20 set; FUN_0057c140: target stat 0x99 zeroes +0x30 / +0x34, stat 0x76 halves them. The 12 rows at 0x732980 (my dump, V run): **row 6 = {+0x34, resist 0x2b, max 0x2c, pierce 0x14f (335, Cold Mastery), absorb -1 / -1, index 0, +0x1c 0, +0x20 0}**: not scaled, and processed for any wearer (the loop breaks at the first row with +0x1c set when the wearer is not a non-hireling monster: rows 9..11, after row 6).
4. **FUN_0057be00 / FUN_0057bf80** (V read): res = stat 0x2b; if the pierce stat row is set and (res < 100 or the target is not a non-hireling monster): res -= stat 335 of the wearer. The difficulty ResistPenalty and the 75 + max cap apply only to players and hirelings. res <= 0 is floored at -100 (for any reason, not only after pierce). Length = FUN_00483360(length, 100 - min(res, 100), 100). So pierce never passes immunity, -100 doubles the length, >= 100 gives 0.
5. **FUN_0057b230** (EAX striker; wearer, length) (V read 0x57b230..0x57b418): length <= 0: nothing. A player striker: FUN_0057af80 (chill). A monster: state 0x36 -> nothing; MonStats +0xc & [0x6ce280] (FUN_0063e9f0, the boss column), monster data +0x16 & 8 (FUN_005a0180) or a hireling class (FUN_0063ee90) -> FUN_0057af80(length), no freeze. Champions carry flag 8: FUN_005a48c0 calls FUN_005a0320 (`or [+0x16], 8`, V read 0x5a0375) then `or 5` (0x5a4905); uniques and superuniques also through FUN_005a0320 (author's trace, consistent with my read of FUN_005a0320). Else coldeffect (FUN_0057af30: MonStats +0x168 + difficulty, signed; txt_fields: ColdEffect +0x168..0x16a) >= 0 -> nothing; else length / DifficultyLevels +0x14 (MonsterFreezeDivisor, idiv truncating, 0 asserts; 1 / 2 / 4 in the 1.14d table, V run); state 1 to frame + length (an existing state 1 is only lengthened), timer 0xc at that end, FUN_00540e60(striker, type 2) removes the striker's pending AI events and FUN_005417d0 files **event 2 at frame + length + 1**.
6. **FUN_0057af80** (the chill, V read 0x57af80..0x57b162): coldeffect 0 -> nothing; a monster with coldeffect < 0: length / MonsterColdDivisor (when not 0); at least 1; state 0xb to frame + length (only lengthened); stats 0x43..0x45 = coldeffect; a 20 % roll of state 0x6b unless MonStats flag 0x13.

### The event 2 at end + 1 (the author's open question): superseded, the real resume is end + aidel

- FUN_005a7f80 (the monster timer dispatcher, V read): for event types outside 3..12 (types 0, 1, 2), if the unit has state 1 and is alive, the event is **dropped** (not refiled). Type 2's handler is 0x5b1740 (table 0x6e2490, V run dump): the monster's AI turn.
- The state-1 list's callback is 0x57b170 (set by FUN_0057b230 through FUN_00625ce0). At the list's expiry (timer 0xc at frame E = frame + length) it runs (V read 0x57b170..0x57b21e): clears state 1, then for a monster **FUN_00540e60(unit, 2) and event 2 at E + aidel**, aidel = MonStats +0x4f (+ difficulty when game +0x6a or +0x74 is set; txt_fields: aidel +0x4f, aidel(N) +0x50, aidel(H) +0x51), 15 when <= 0.
- So the end + 1 event never fires as an AI turn: E < E + 1, and at E the callback removes it. When a longer prior freeze is running, the end + 1 event fires while state 1 is on and is dropped. Either way the monster's next AI turn is **E + aidel** (13..15 frames for the common rows: zombie1 / fallen1 / skeleton1 15 / 14 / 13 by difficulty; Andariel 15 / 11 / 9). V run: my oracle runs 0x57b170 for real at the expiry and replays the recorded timers with FUN_005a7f80's drop rule.
- d2d: `stun_until` gates the monster while `now_ms < stun_until` (ai.cpp:1831, S), so it may act from frame E (if its `next_act` allows): **aidel frames earlier than game.exe, not one frame**. This holds for every d2d freeze (this claim, Glacial Spike's missiles.cpp:662, the freeze-rules candidate), so it is a shared, separate deviation, not a defect of this claim's length rule.
- A zero-length freeze (length / divisor = 0) still creates state 1 ending now and refiles the AI turn (callback at the expiry -> now + aidel, or + 1; the frame a timer filed for the current frame runs is U). d2d does nothing. 205 of my 134,400 cases.

## 2. Own differential (own generator, own oracle, own harness)

- Generator: `tools/emu/verify_armor_hit.py gen SEED N [edge]`: an LCG (`x * 1664525 + 1013904223`, masked to 32 bits, top 24 bits), seeds 0xA101 (+ 14,400 edge cases), 0xB202, 0xC303, 0xD404, 30,000 random cases each: **134,400 cases**. Fields: difficulty; kind (plain, champion, unique, superunique, minion, MonStats boss column); coldeffect (real values, 0, -128..127); cold resistance (-200..150 incl. 90..110); pierce 0..199; Frozen Armor level 1..40; Shiver and Chilling hard points separately 0..40; physical damage (0 in about a fifth); aidel (0, 13..15, 0..29); a prior freeze and a prior chill (0 or 1..80 / 1..200 frames).
- game side: `verify_armor_hit.py game`: FUN_005cac50 and everything under it run for real (section 1 list in its docstring), timers and the state-1 callback replayed. Hooked: stat get/set, the DifficultyLevels row (read from the MPQ myself), a fake MonStats row (boss byte, aidel, coldeffect), states, stat lists, timers (recorded), events, logging, the other record appliers (their fields are 0), and the calc FUN_00646ca0 (calc1 from Skills.txt's own text, ln56 x (100 + (Shiver + Chilling blvl) x par8) / 100, ln56 = 30 + 3 (lvl - 1); every call checked to be calc1 of skill 40). Output: chill end, freeze end, AI resume frame.
- d2d side: `tools/emu/test_verify_armor_hit.cpp` (built as `tests/game/test_verify_armor_hit.cpp`): the real `Fight::buff_events`, zombie1 as the striker (the author used fallen1), Shiver / Chilling points set separately, prior `stun_until` / `chill_until` set from the case. Prints chill end, stun end, calc1 (and RESFAIL if the cold resistance was not taken).
- Compared: the chill end frame and the freeze end frame per case (absolute frames: both relative to the same now), d2d's calc1 against Skills.txt's, and the AI resume frame (game.exe's replayed event 2 vs d2d's first unstunned frame = stun end, ai.cpp:1831). The author compares the two state ends only (ticks to the end), not the resume.

| build | cases with a different chill / freeze end (of 134,400) | by kind (seed 0xA101) | resume frame differs | calc1 differs | RESFAIL |
|---|---|---|---|---|---|
| upstream main 72132b99 | **109,573** (38,171 + 23,885 + 23,765 + 23,752) | plain 11,302, champion 5,425, unique 5,367, superunique 5,386, minion 5,352, boss column 5,339 | 116,092 | 0 | 0 |
| fix/armor-hit head d5bd19c1 | **0** | 0 | **52,714** (14,513 + 12,710 + 12,745 + 12,746): every freeze, game.exe later by aidel (15: 22,612; 13: 10,508; 14: 10,433; the rest the random aidel values) | 0 | 0 |

The state ends match on every case, every kind, every difficulty, with priors and with zero-damage cases (prior states only). The whole residual is the resume frame of section 1 (shared with every d2d freeze) plus the 205 zero-length freezes (in the resume count).

## 3. Isolation (each commit built alone)

Four Release builds of my harness: base (72132b99), tool (72132b99 + cherry-pick of d5bd19c1 alone), claim (36cbe015 alone), head (d5bd19c1). md5 of the case lines (log lines dropped):

| seed | base | tool alone | claim alone | head |
|---|---|---|---|---|
| 0xA101 | 03c665dd0c64 | 03c665dd0c64 | be86d4a38c5d | be86d4a38c5d |
| 0xB202 | 366530fcc9cc | 366530fcc9cc | c1d85d65f8d8 | c1d85d65f8d8 |
| 0xC303 | 2221f578fd23 | 2221f578fd23 | b90e92247f8b | b90e92247f8b |
| 0xD404 | 36bf0b0f4ef9 | 36bf0b0f4ef9 | 9a29285c0639 | 9a29285c0639 |

The tooling commit changes nothing in d2d (tool = base); the claim alone carries the whole change (claim = head).

## 4. CI-class checks (head d5bd19c1, worker host; run by my helper in a separate clone)

- pr_guard: 0 FAIL / 0 WARN on the range and on each commit (36cbe015, d5bd19c1).
- presubmit --quick: rc 0, 0 FAIL, 3 WARN: (a) `siblings: unit kinds` (monster, not player, merc, pet): answer: the Act 3 hireling's Frozen Armor (`merc_buff_events`, states.cpp) keeps the old stun + chill on every striker although game.exe runs the same FUN_005cac50 for it (the state's event, Skills.txt auraeventfunc1 2); player strikers have no d2d path; (b) `siblings: hit paths` (melee, not missile, skill): Chilling Armor (missile) and Shiver Armor unchanged, the missile freeze is the freeze-rules candidate; (c) `build: skipped (--quick)`: the full build below.
- Tooling unittests (test_tooling, test_gaps, test_pr_guard, test_presubmit, test_tidy_pr_select): pass.
- gcc `-Wshadow -Wunused -Wsign-compare` Release build of test_game, test_rules, test_combat, test_protocol, test_monsters, test_armor_hit: 0 warnings on added lines and in touched files.
- test_monsters, test_rules, test_combat, test_protocol, test_game: rc 0.
- clang-tidy via tidy-pr-select (136 TUs): 0 findings in the touched files (29 elsewhere, untouched); the added `#include <uniques.hpp>` is needed (`d2d::rules::Boss`).
- ruff 0.17.3 with presubmit's rules (F, E9, B, BLE001, PLW1510, S602, S604) and `--select B905` on armor_hit.py and freeze_rules.py: clean.
- Replay: upstream main is still 72132b99 (`gh api repos/psi29a/d2decomped/commits/main`, 2026-11-01): both commits apply as they are.
- Merges, both orders, textual: clean with #155 (shares gamedata.hpp, gamedata_load.cpp), #156, #157, #158 (shares gamedata.hpp, gamedata_load.cpp, re/skills.md) and the local fix/freeze-rules-v2 (head 409903a6; shares gamedata.hpp, gamedata_load.cpp, freeze_rules.py). With freeze-rules-v2 the identical `freeze_divisor` member collapses into one, but **gamedata_load.cpp gets the `MonsterFreezeDivisor` load line twice** (after `cold_divisor` and after `curse_divisor`): harmless, builds (test_game, test_armor_hit rc 0), but a reviewer will see it.
- Size: claim 4 files +35 / -8; tooling 3 files +607.

## 5. Det state hashes

SEE_DET_SECTION

## 6. Code review notes (S unless tagged)

1. MEDIUM: `merc_buff_events` (states.cpp, the Act 3 hireling's Frozen Armor) still stuns and chills every damaging striker for raw calc1 frames, no resistance, no divisor, champions and bosses included. game.exe runs the same event function for any unit in the state; the claim fixes one of two callers of the same rule. Recommend: the same block (shared helper) for the merc wearer in this PR, or say it in the body (the draft does: Limits 1). Note the hireling wearer's pierce is its own stat 335 (none) and row 6 is not scaled (section 1.3).
2. MEDIUM: the draft's limit "the freeze's event 2 at end + 1 ... a frozen monster acts one frame later than d2d lets it" is wrong in size: the expiry callback refiles the AI turn at end + aidel (section 1), so game.exe's monster acts aidel frames (13..15 for most rows) after the freeze ends, d2d's at the end. Shared with every d2d freeze; a separate candidate (stun end -> `next_act = max(next_act, stun_until + aidel)` for a freeze; Glacial Spike and freeze-rules-v2 have it too).
3. LOW: `monsters[monster_index].type >= 0` guard is new and silent: a striker without a MonStats row gets nothing (game.exe would read the row through FUN_0057af30 and get -50 for a missing row, i.e. frozen): unreachable for real monsters; fine.
4. LOW: the `// ponytail:` comment says "the freeze's AI event at its end + 1 (FUN_005417d0 2) ... aren't kept": change to "the AI turn at the freeze's end + aidel (FUN_0057b170 refiles event 2; the end + 1 event of FUN_0057b230 is removed by it) isn't kept".
5. LOW: duplicate `MonsterFreezeDivisor` load line after a merge with freeze-rules-v2 (section 4): land freeze-rules-v2 first and drop the two gamedata hunks here, or the reverse.

## 7. Verdict

- **C1 (the length is calc1 resisted by the striker's cold resistance less the wearer's Cold Mastery pierce; no pierce through immunity; floor -100; cold immunes get nothing): HOLDS** (V read, V run): 0 of 134,400 end frames differ; upstream main 109,573.
- **C2 (plain monsters frozen only at coldeffect < 0, for length / MonsterFreezeDivisor, not chilled): HOLDS WITH LIMITS**: state end equal on every case; the AI turn after the freeze is end + aidel in game.exe (end + 1 is superseded, section 1) and end in d2d (52,714 cases, a shared deviation of all d2d freezes, not introduced here); a zero-length freeze (205 cases) still interrupts the AI in game.exe.
- **C3 (champions, uniques, superuniques, MonStats bosses at most chilled; nothing at coldeffect 0; / MonsterColdDivisor when negative): HOLDS** (flag 8 traced for champions; 0 differ by kind).
- **C4 (MonsterFreezeDivisor added and loaded): HOLDS** (1 / 2 / 4 read from the MPQ in my oracle; d2d's result equal).
- **C5 (skills.md line): HOLDS** (matches section 1).
- Not claimed and confirmed open: D5 trigger (physical after reduction vs any element), the merc wearer, D2 (armors stacking), state 0x36, the 0x6b roll.
- Isolation: HOLDS (tool = base, claim = head on all four seeds).
- CI-class: clean (0/0 pr_guard per commit, presubmit 0 FAIL / 3 WARN answered above, tests rc 0, gcc and clang-tidy 0 in touched files, ruff clean); merges clean with #155..#158 and freeze-rules-v2 (duplicate load line with the latter).

READINESS_LINE

Draft edits (pr-armorhit-draft.md):
1. Limits, "Not modelled": replace "the freeze's event 2 at end + 1 (... acts one frame later ...)" with "the AI turn after a freeze: game.exe's freeze-end callback (FUN_0057b170) refiles the monster's AI event at the end + MonStats aidel (13..15 frames for most monsters; the end + 1 event FUN_0057b230 files is removed by it), d2d lets the monster act at the end; this holds for every d2d freeze and is not claimed".
2. Limits: keep the merc wearer and D5 lines; add "a zero-length freeze (length / MonsterFreezeDivisor = 0) still interrupts the monster's AI in game.exe".
3. Proof table: add the verifier's run: "independent oracle and harness, 134,400 cases on 4 seeds: upstream main 109,573 differ, this branch 0 (chill and freeze end frames)" and "det state hashes: SEE_DET_SHORT".
4. Merges: "cleanly with #155, #156, #157 and #158 in either order" (shares gamedata files with #155 and #158).
5. Not run: drop "det state hashes"; keep "an MSVC build".
6. If freeze-rules-v2 is posted first: rebase and drop the duplicate `freeze_divisor` member / load line.

## Commands (worker host)

```sh
# worktrees: base 72132b99, tool 72132b99 + cherry-pick d5bd19c1, claim 36cbe015, head d5bd19c1; each gets tests/game/test_verify_armor_hit.cpp
python verify_armor_hit.py gen 0xA101 30000 edge > c-a.tsv            # 0xB202 / 0xC303 / 0xD404 without edge
PYTHONPATH=<upstream>/tools/emu python verify_armor_hit.py game c-a.tsv > g-a.tsv
build/tests/test_verify_armor_hit --varmor c-a.tsv > d-a.tsv
python verify_armor_hit.py compare c-a.tsv g-a.tsv d-a.tsv
```

## Shared-file lines for the merger

- docs/BACKLOG.md: `- ARMORHIT: verifier (VERIFYARMORHIT, docs/research/pr-armorhit-verifier.md): claim holds (0 of 134,400 end frames); draft limit edits listed there; new shared deviation: a frozen monster's next AI turn is freeze end + aidel in game.exe (FUN_0057b170), d2d acts at the end (all freezes)`
- docs/INDEX.md, research: `- docs/research/pr-armorhit-verifier.md: independent verifier of fix/armor-hit (freeze chain and resume frame traced, own differential, det hashes)`
- CHANGELOG.md: none.
- GLOSSARY.md: none missing.
