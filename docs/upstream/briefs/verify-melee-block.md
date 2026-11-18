# Verifier brief: fid/melee-block-r1 (P07, a negative result), for the independent reviewer (Opus, singly)

Claim to break: a monster's melee swing is blockable exactly as d2d's `monster_blow` blocks it, so there is NO block gate to port (unlike P06). Branch `fid/melee-block-r1` on the private fork remote, one oracle commit `3d3d6b25` (tools only), base upstream `77990456`; evidence `docs/research/decomp27-trace.md` Q1 to Q6.

1. Re-run: `mel_block.py game 400 > game.tsv`, `test_melee_block 400 | grep '^D ' > d2d.tsv`, `mel_block.py compare game.tsv d2d.tsv`: 180 cells, 0 differ, exit 0. Mutate one d2d cell (a mode, chance 75, blocked 0): exit 1.
2. Read `FUN_0057ec10` (`0x57ec10`..`0x57ed44`) and `FUN_0057dfb0` (`0x57dfb0`..`0x57e08b`): confirm can-block 1 / missile 0 at `0x57ece1`. Decide whether the hooks (`FUN_00622720`, `FUN_0057dd60`, `FUN_0057d9b0`, `FUN_00622c40`, `FUN_00554010`, `FUN_005347c0`, `FUN_00554200`, `FUN_00639df0`) hide a gate: a monster-specific test inside `FUN_0057ec10` (the attacker type, `FUN_005a4ee0` for a monster DEFENDER only) or in the block chance `FUN_00622720` for a player.
3. Find a melee-like monster hit that does NOT go through `FUN_0057ec10` and that d2d blocks: the 33 call sites of `FUN_0057ec10`, the `ResultFlags` bit 0 skip in `FUN_0056e680`, and the can-block 0 callbacks (`FUN_0056b9c0`: do-functions 38 and 53, no monster skill uses them). BearSmite (do 2, `FUN_0056f1f0`) is the open U.
4. The d2d side is rules only (`monster_blow` with `no_hit_roll`): is the `Motion` mapping (WL walking, RN running, else standing) what the fight code passes (`foe.motion()`)?
Report: verdict, counts, what you could not check. Not allowed: posting, pushing, a force push.
