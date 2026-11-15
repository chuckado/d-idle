# Draft (local, NOT posted): "A monster missile with no physical damage cannot be blocked" (P06)

> status: 2026-11-15 (card prfix6, supersedes the r3 text of decomp26). Fork branch `fid/missile-block-r4` (private fork remote), two commits on upstream main `aaa17aee`: oracle `426c87f6`, claim `058b0d43`; `-r3` (`13783a6d` / `cd314f2c`) stays as the backup, `-r1` and `-r2` are stale. The independent verifier (verifyp06) passed r3 with required changes; they are applied here, and because change 1 is a code change a SECOND verifier pass is needed before posting. Not posted. Body: [bodies/missile-block-body.md](../bodies/missile-block-body.md). P07 (a monster's melee block gate) is a second PR in the same subsystem after this one.

PR title (when unblocked): `A monster missile with no physical damage cannot be blocked`

## What changes

The missile's hit entry `FUN_005ad730` calls `FUN_0057dfb0(defender, 1, [record + 8] != 0)`. The record's physical field is rolled by `FUN_005a89a0` (`FUN_005a8910`: min and max both over 0, else 0). With can-block 0 the block roll is skipped (and its draw on the defender's seed) and the defender's other rolls run. d2d rolled the shield's block for every monster missile. New `MonStats::unblockable`, set for the MissA / MissS1 / MissC shots (`monster_shot_stats`) for a monster skill's missile (`missile_blow`: the Fallen Shaman's fire, Radament's bolt) and, after the carried skill's physical replaces the row's, in `andariel_missile` (a Death Mauler's maul stays blockable), read by `monster_blow`. The Countess's firewall is not part of the claim: d2d burns it without a block roll already.

## Claims

| Claim | Evidence | Re-check |
|---|---|---|
| can-block is `physical min > 0 and max > 0` | V run: `FUN_005a89a0` and `FUN_005ad730` under the emulator, `FUN_0057dfb0` hooked to read its third argument, on game.exe's own launch records (`mm_diff.py`, 30,000 cases, seed 20261115: 29,421 compared, 579 Blood Raven cases excluded because her record carries her weapon's damage): 4,550 with can-block 0 | `missile_block.py game` |
| d2d before | V run (NAS builder, oracle commit `426c87f6`): `test_monster_missile --missile-block` (the real `monster_shot_stats` then `monster_blow`, defender Block 75, 400 trials a case): 4,550 of 29,421 differ (d2d blocks about 75 percent where game.exe cannot); skill rows (`mm_launch.py`, 75 lines incl. the Death Mauler's maul): 42 of 60 compared differ (12 NOROW, 3 excluded) | `missile_block.py compare` (exit 1) |
| d2d after | V run (claim `058b0d43`): 0 of 29,421, 0 of 60 (the maul's 15 launches match: game.exe can-block 1); an earlier ordering (unblockable from the row before the carried physical) differs in the 15 maul cases and aborts `test_carried_missile`; the unit assert in `test_monster_missile` (min 0 or max 0 never blocks, 1..5 and 3..3 do; a skill row with no physical never blocks); `test_game`, `test_combat`, `test_rules`, `test_monster_ai`, `test_monsters`, `test_game_rules`, `test_monster_thinks`, `test_missile_blow_poison`, `test_energy_shield` rc 0 | same |

## Limits

- Only the block roll is skipped. The dodge, avoid and evade rolls and the claws' weapon block (inside `FUN_0057dd60`) are unchanged, as in game.exe.
- d2d's rng is shared, so the skipped draw changes a sequence, not the distribution: the claim is about the blocked fraction (0 against about 75 percent), not about a seed.
- 12 skill-launch cases (`shafire2`, `shafire4`, `unholybolt2`, `unholybolt3`) have no d2d row: d2d loads only the Missiles.txt rows its monsters and skills name, and `andariel_missile` returns without a missile when the row is absent (`ai.cpp:471`; the Fallen Shaman picks `shafire` + 1 + `TransLvl`). The compare lists them as NOROW and does not judge them. Whether a Fallen Shaman 2 or 4 really fires nothing in d2d is I (a `test_game` scenario would prove it). A separate PR.
- A skill missile whose owner adds damage (SrcDamage over 0) is not compared (U); only `andarielspray` (3 cases) has a record, and its can-block matches.
- P07 (the monster's own block of a melee swing) and the to-hit half (merged earlier) are not part of this.
