# PR draft: Frozen Armor freezes a striker by the freeze rule (ARMORHIT)

> status: DRAFT, not posted; NOT READY (independent verifier and det state hashes pending; the user's go). Evidence: [pr-armorhit-proof.md](pr-armorhit-proof.md). Branch `fix/armor-hit` on the worker host (head d5bd19c1, base upstream main 72132b99).

Title: `Frozen Armor: freeze or chill a melee striker by the freeze rule`

---

## Limits

- Only the player's Frozen Armor (`Fight::buff_events`). The Act 3 hireling's Frozen Armor (`merc_buff_events`) still stuns and chills; game.exe's rule for it is the same by reading (row 6 is not scaled for a hireling wearer) but it is not in the differential.
- Monster strikers only. A player striker is chilled in game.exe (FUN_0057b230's player branch); d2d has no player-on-player path here.
- Not modelled (ponytail notes): state 0x36 (no freeze), the chill's 20 % state 0x6b roll, the freeze's event 2 at end + 1 (FUN_005417d0: if that is the monster's next AI turn, a frozen monster acts one frame later than d2d lets it; this also holds for Glacial Spike's freeze and is not claimed), a zero-length freeze (length / MonsterFreezeDivisor = 0 sets state 1 for no frame).
- The trigger stays d2d's (`damaged_by`: any damage; game.exe tests the physical field +8), and the armors' stacking (game.exe's state group 1) is unchanged.
- The game side hooks the calc (Frozen Armor calc1 restated from Skills.txt); d2d's own calc1 agrees on every case.
- Not run: det state hashes, an MSVC build.

## What changes

game.exe's Frozen Armor event (FUN_005cac50, damagedinmelee) gives a striker whose blow did damage a hit carrying only calc1 as its freeze length (record +0x34) through FUN_0057c6c0. FUN_0057c1e0 resists that length by the striker's cold resistance less the wearer's cold pierce (row 6 of the table at 0x732980; no pierce through immunity), and FUN_0057b230 decides: a MonStats boss, champion, unique or superunique is chilled for it (FUN_0057af80: not at coldeffect 0, / MonsterColdDivisor when negative), anything else is frozen only at a negative coldeffect, for length / MonsterFreezeDivisor, and is not chilled.

d2d stood and chilled every damaging striker for the raw calc1 frames: Andariel and every champion stopped 1.2 s a hit, cold immunes too, and Nightmare / Hell freezes 2x / 4x long.

## Proof

`tests/game/test_armor_hit.cpp` (`--armor-hit CASES`) drives the real `Fight::buff_events`; `tools/emu/armor_hit.py game` runs FUN_005cac50 natively under the emulator on the same cases (Boss kinds, MonStats boss column, coldeffect, cold resistance -100..150, Cold Mastery pierce, Frozen Armor levels 1..40 with Shiver / Chilling synergy, three difficulties).

| build | mismatches (seed 0x41524d48) | seed 0x5eed2026 |
|---|---|---|
| upstream main + the test | 95,920 of 108,506 | 95,986 of 108,506 |
| this branch | 0 | 0 |

Every damaging case differed before (plain 20,533, champion 10,792, unique 10,888, superunique 10,915, minion 10,926, boss column 31,866; cold-immune 17,306).

pr_guard 0 FAIL / 0 WARN per commit; presubmit --quick 0 FAIL (siblings answered above); test_game, test_rules, test_combat, test_protocol, test_monsters pass; gcc -Wshadow and clang-tidy clean on the touched lines.

## Files

`components/game/fight/states.cpp`, `components/game/gamedata/gamedata.hpp`, `components/game/gamedata/gamedata_load.cpp` (MonsterFreezeDivisor), `docs/research/re/skills.md` (event 2 line); tooling `tests/game/test_armor_hit.cpp`, `tools/emu/armor_hit.py`, `tools/emu/freeze_rules.py` (shared with the freeze-rules change, identical).

Merges cleanly with #155 and #156 in either order.
