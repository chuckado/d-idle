# PR body (final): Frozen Armor freezes or chills a melee striker by the freeze rule (ARMORFINISH)

> status: READY, NOT posted, not opened (the user's go is pending; six PRs are already open). Evidence: [pr-armorhit-addendum.md](pr-armorhit-addendum.md), [pr-armorhit-proof.md](pr-armorhit-proof.md), verifier record `pr-armorhit-verifier.md`. Everything below the rule is the text to post. Branch `fix/armor-hit` (claim 36cbe015, tooling d5bd19c1, base upstream main 72132b99).

**Title:** `Frozen Armor: freeze or chill a melee striker by the freeze rule`

---

## Limits

- Only the player's Frozen Armor (`Fight::buff_events`). The Act 3 hireling's Frozen Armor (`merc_buff_events`) still stuns and chills every striker. game.exe runs the same event function for it, so the rule is the same by reading, but it is not in either differential.
- Monster strikers only. game.exe chills a player striker; d2d has no player-on-player path here.
- The AI turn after a freeze is not modelled, and this is bigger than one frame. game.exe's freeze-end callback (FUN_0057b170) refiles the monster's AI event at the end plus MonStats `aidel` (13 to 15 frames for most rows); the event at end + 1 that FUN_0057b230 files is removed by it, and FUN_005a7f80 drops AI events while the freeze is on. d2d lets the monster act at the end. This holds for every d2d freeze (Glacial Spike too), is not introduced here, and is not claimed. On the verifier's 134,400 cases the freeze and chill end frames match, and the AI resume frame differs in 52,714 of them (by `aidel`).
- A zero-length freeze (length / MonsterFreezeDivisor = 0) still interrupts the monster's AI in game.exe; d2d does nothing.
- Not modelled: state 0x36 (no freeze), the chill's 20 percent state 0x6b roll.
- The trigger stays d2d's (`damaged_by`: any damage; game.exe tests the physical field after reduction), and the armors' stacking (game.exe's state group 1) is unchanged.
- The game side hooks the calc (Frozen Armor calc1 restated from Skills.txt); d2d's own calc1 agrees on every case.
- Not run: an MSVC build. Det state hashes were run (14 of 14 identical) but no scenario wears Frozen Armor, so they show only that nothing outside the changed branch moved.

## What

game.exe's Frozen Armor event (FUN_005cac50, damagedinmelee) gives a striker whose blow did damage a hit carrying only calc1 as its freeze length (record +0x34) through FUN_0057c6c0. FUN_0057c1e0 resists that length by the striker's cold resistance less the wearer's cold pierce (row 6 of the table at 0x732980; no pierce through immunity), and FUN_0057b230 decides: a MonStats boss, champion, unique or superunique is chilled for it (FUN_0057af80: nothing at coldeffect 0, divided by MonsterColdDivisor when negative); anything else is frozen only at a negative coldeffect, for length / MonsterFreezeDivisor, and is not chilled.

d2d stood and chilled every damaging striker for the raw calc1 frames: Andariel and every champion stopped for 1.2 s a hit, cold immunes too, and Nightmare / Hell froze 2x / 4x too long.

## Change

- `Fight::buff_events` follows that rule: resisted length, specials chilled with `land`'s chill rule, plain monsters frozen at coldeffect < 0 for length / MonsterFreezeDivisor and not chilled.
- `GameData::freeze_divisor`, read from DifficultyLevels (MonsterFreezeDivisor).
- `docs/research/re/skills.md`: the event 2 line completed.
- Tooling: `tests/game/test_armor_hit.cpp`, `tools/emu/armor_hit.py`, `tools/emu/freeze_rules.py` (the last is shared with the freeze rules change and identical). Files under `tools/` need the CODEOWNERS review.

## Evidence

Differentials against game.exe's own code (FUN_005cac50 and everything under it run natively under the emulator), d2d side through the real `Fight::buff_events`:

| harness | cases | upstream main | this branch |
|---|---|---|---|
| author's, 2 seeds (cold resistance -100..150, pierce, Frozen Armor 1..40 with synergy, 3 difficulties) | 108,506 each | 95,920 / 95,986 differ | 0 / 0 |
| independent verifier's own generator, oracle and harness, 4 seeds, chill and freeze end frames | 134,400 | 109,573 differ | 0 |

Isolation: the tooling commit changes nothing in d2d; the claim commit alone carries the whole change. Every kind (plain, champion, unique, superunique, minion, boss column) and difficulty agrees after the change.

```sh
cd tools/emu
uv run python armor_hit.py gen 50000 0x41524d48 > cases.tsv
uv run python armor_hit.py game cases.tsv > game.tsv
build/tests/test_armor_hit --armor-hit cases.tsv > d2d.tsv
uv run python armor_hit.py compare cases.tsv game.tsv d2d.tsv     # mismatches 0
```

## Checklist

pr_guard 0 FAIL / 0 WARN per commit; presubmit --quick 0 FAIL (warnings answered above); test_game, test_rules, test_combat, test_protocol, test_monsters pass; gcc -Wshadow -Wunused -Wsign-compare and clang-tidy clean on the touched lines; ruff clean; det state hashes identical on the 14 scenarios. 7 files, 2 commits.

Merges cleanly with #155, #156, #157 and #158 in either order (it shares gamedata.hpp and gamedata_load.cpp with #155 and #158). The local freeze rules candidate reads MonsterFreezeDivisor too: whichever lands second drops its copy of the member and the load line.

Agent: Claude Code
