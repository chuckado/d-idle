# Freeze rules: game.exe vs d2d (decomp item 6, second follow-up)

> status: card decompfreeze, 2026-11-01, offline. Re-derives the freeze rules from the disassembly (`tools/emu/disasm.py`) independently of docs/research/static-field-chill-trace.md (Q2) and the chill verifier section of pr-static-chill-proof.md, and decides C3 / C4 / C5. PR candidate: [pr-freeze-proof.md](pr-freeze-proof.md), draft body [pr-freeze-draft.md](pr-freeze-draft.md).

Tags: V (read with disasm.py or run natively under the emulator: say which), D (from the 1.14d data tables), S (d2d source, file:line on #155's head ddc8281d), I (inference), U (not verified).

## Summary

- **C5 CONFIRMED** (V run): game.exe cuts the freeze length by cold resistance less cold pierce like the chill length, freezes a plain monster only at a negative coldeffect for length / MonsterFreezeDivisor, and adds no chill; d2d froze for the raw length at any coldeffect and added a chill.
- **C3 CONFIRMED, wording CORRECTED** (V run + V read): the test is MonStats boss (+0xc & 0x40), monster data +0x16 & 8, or a hireling class. Champions DO carry flag 8: FUN_005a48c0 (the champion-pack maker, mod 16 `champion`) calls FUN_005a0320 (sets 8) before `| 5`; uniques get 8 in FUN_005a4940; act bosses and the Countess through FUN_005a4850(mod, 1). The verifier's "flags 4 and 0x10 were frozen" is right for those bits alone; a real champion has 0xD. So C3's original "bosses, champions, uniques, hirelings" is right; minions (0x10 only) are frozen.
- **C4 CONFIRMED** (V run + V read): Blood Raven's 0x5b1fb0 and the Countess's 0x5a45d1 / 0x5a4c00 `push 1; push 0x76; push unit; call FUN_00639db0` set **state 118** (corpse_noselect); run natively it sets bit 118 of the state bits and no stat byte, and the real stat getter reads stat 0x76 as 0. d2d halved their freeze.

## Rules (game.exe)

| Address | Rule | Tag |
|---|---|---|
| `FUN_005aa8b0` (SrvHitFunc 13, Glacial Spike) | radius = Missiles sHitPar1 (+0x4c) if > 0 else calc(Skills +0x64 aurarangecalc), at least 1; freeze = sHitPar2 (+0x50) if > 0 else calc(Skills +0x60 auralencalc), **no minimum**; fills the record (FUN_005a8c70), then only if freeze > 0 FUN_005a8f20(missile, record, freeze) | V read |
| `FUN_005a8f20` | by the missile's EType (FUN_0064b0c0) - 4, table 0x5a8f84: cold 4 -> +0x30, pois 5 -> +0x2c, stun 9 -> +0x44, burn 11 -> +0x18, **frze 12 -> +0x34**; Glacial Spike's missile row has EType `frze` (D) | V read; D |
| `FUN_0057c1e0` table 0x732980 (stride 0x2c) | row 5 = field 0x30 (chill length), row 6 = field **0x34 (freeze length)**, both resist 0x2b (cold res), max 0x2c, pierce 0x14f (cold pierce), no flat / absorb stats: the freeze length is cut exactly as the chill length (FUN_0057bf80; res >= 100: 0; -100 floor) | V read; V run (the oracle runs it natively) |
| `FUN_0057c140` | stat 0x99: cold and freeze lengths 0; else stat 0x76: both halved | V read (and earlier V run, cold_length.py) |
| `FUN_0057b230` (defender eax, attacker, length) | length <= 0: nothing. Player (type 0): chill (FUN_0057af80) with the length. Monster: state 0x36 (uninterruptable) set: nothing; MonStats boss (FUN_0063e9f0: row +0xc & [0x6ce280] = 0x40) or monster data +0x16 & 8 (FUN_005a0180) or a hireling (FUN_0063ee90: classes 0x10f, 0x152, 0x167, 0x230, 0x231): chill with the length, no freeze. Otherwise (and any non-player, non-monster unit): coldeffect >= 0 (FUN_0057af30: MonStats +0x168 + difficulty, signed; 0 for a non-monster): nothing; monster: length / DifficultyLevels MonsterFreezeDivisor (+0x14; 0 is a fatal error), no minimum; state 1 (freeze) until frame + length, only extended; events 0xc at the end and 2 at frame + length + 1 | V read; V run (the oracle and the `flags` sweep) |
| flag 8 makers | `FUN_005a0320` sets +0x16 \|= 8 (0x5a034a / 0x5a0375); callers: `FUN_005a48c0` (champion packs, called with mod 0x10 `champion` from 0x54e24c / 0x54e8fe; then \|= 5 at 0x5a4905), `FUN_005a4940` (uniques; before FUN_005a0760), `FUN_005a4850(mod, 1)` (0x5a487b; act bosses FUN_005b1cf0, superuniques FUN_005a49b0), 0x5a0b61. FUN_005a0760's own champion branch (0x5a083e \|= 4) does not call it, but its caller FUN_005a4940 already did. Minions: \|= 0x10 (0x5a47df, 0x5a0cca) with no FUN_005a0320 on that path | V read; mod 16 = `champion` (D); "champion pack" from the call sites (I) |
| `FUN_00639db0(unit, id, on)` | bounds id by the state count ([tables] +0xc4), FUN_00625a70 (unit +0x5c list, +0x58 bits: sets bit id and its changed bit), FUN_0064c040; no stat write | V read; V run (`freeze_rules.py state`: bits 118 and 630 set, 0 stat-list bytes changed, FUN_00625480(unit, 0x76) = 0) |
| state 118 sites | Blood Raven (FUN_005b1cf0 0x5b1fae..0x5b1fb3, after mods 0xc / 0x16 and quest 2); the superunique makers FUN_005a4440 (0x5a45cf, 0x5a4607) and FUN_005a49b0 (0x5a4bfe, then FUN_005a4850(0x16, 1); d2d reads that case as the Countess, ai.cpp:2965, S; which superunique each case is: not re-derived, I), 0x5a4d99. States.txt row 118 = `corpse_noselect`; ItemStatCost row 118 = `item_halffreezeduration` (D) | V read; D |

## d2d (S, ddc8281d)

- `Fight::strike` (components/game/fight/missiles.cpp:635-663) gets `freeze` only from hit_func 13 (`area(..., radius(hit_par2, auralen))`, :676; `radius()` clamps to at least 1). It halves it for `half_freeze` (:658), sets `blow.chill_ticks = max(chill, freeze)` (:659) so `land` chills any monster with coldeffect != 0 for freeze / MonsterColdDivisor, and stands the monster (`stun_until`) for the raw freeze (:662), unresisted, at any coldeffect, for bosses too.
- `half_freeze` is set for Blood Raven (`act_boss` base 0x10b, components/rules/uniques.hpp:155) and the Countess (components/game/ai.cpp:2965), copied for Diablo (act4.cpp:298); read only at missiles.cpp:658.
- Other freeze paths, not in this candidate: Freezing Arrow's `freezingarrowexp3` (EType frze) loads as cold (gamedata_load.cpp:350) so it chills by its ELen and never freezes (game.exe: +0x34, FUN_0057b230; I: not run); Frozen Armor (states.cpp:151-160, 190-197) stuns and chills its strikers for calc1 frames with no resistance, divisor or boss rule (game.exe FUN_005cac50 not traced, U); a player's freeze is a chill in d2d (combat.hpp:587-595: same as FUN_0057b230's player branch, V read).

## Deviations

| id | d2d | game.exe | tag | candidate |
|---|---|---|---|---|
| C5 | plain monster: frozen (stun) for the raw length, no MonsterFreezeDivisor, at any coldeffect, plus a chill | frozen only at coldeffect < 0, for resisted length / MonsterFreezeDivisor, no chill | V run | claim 1 |
| C3 | special monster (MonStats boss, champion, unique, superunique): frozen and chilled | chilled for the resisted length (coldeffect / MonsterColdDivisor rules of FUN_0057af80), never frozen | V run | claim 2 |
| C4 | Blood Raven and the Countess: freeze length halved | not halved (their 0x76 call is state 118) | V run + V read | claim 3 |
| R1 | freeze length from auralencalc <= 0: 1 frame (`radius()` clamp) | no freeze | V read; unreachable with 1.14d Glacial Spike (ln34 > 0, D) | none |
| R2 | length / divisor = 0: no freeze | state 1 set until the current frame | V read | none (no frame of freeze) |
| R3 | state 0x36 not modelled | no freeze or chill while it is set | V read | none |
| R4 | Freezing Arrow freeze as chill; Frozen Armor stun + chill | +0x34 freeze via FUN_0057b230; Frozen Armor U | I / U | next candidates |

Hirelings: game.exe chills them instead (V run, `flags`); d2d's skill missiles strike only `monsters`, so no merc path exists (S): no claim.

## What drills should expect (live, a throwaway sorceress with Glacial Spike)

1. A plain monster with negative coldeffect (most Act 1 monsters, D) in Normal: state 1 for auralen ticks less its cold resistance; in Nightmare / Hell half / a quarter (MonsterFreezeDivisor 1 / 2 / 4, D); no state 0xb from the spike itself.
2. A monster with coldeffect >= 0: no state 1 and no state 0xb.
3. A champion, unique, superunique or act boss: state 0xb (chill) for the resisted length (/ MonsterColdDivisor when coldeffect < 0), never state 1. Blood Raven and the Countess: the full length, not half.
4. A minion of a unique: frozen as 1.
d2d before the candidate: every monster stands for the raw length and is chilled; after: as game.exe.
