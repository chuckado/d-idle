A knocking hit (Smite, a monster's knockback) asks `FUN_0057edd0` for new mode 0x13 first (`FUN_00580a70`, `0x580adc`); game.exe refuses it from a get-hit or a block, like a big hit's mode 4 (#548). d2d slid a player who was standing in a get-hit and moved her hold. One condition in `Fight::hurt_side` fixes it: a get-hit that is not itself a knockback slide takes no knockback.

Oracle (commit 1): `tools/emu/mode_switch.py` also asks `FUN_0057edd0` for new mode 19 (`compare-knock`), and `test_hit_recovery --knock-modes` prints what `hurt_side` does with a 1-point knocking hit in each mode. Before: 13 cells compared, 1 differs (GH), 7 excluded by name (dead modes; swing modes, which game.exe allows and d2d never interrupts: a separate change). After: 0 differ. `test_hit_recovery` gets an assert for the GH case.

Re-check: `mode_switch.py game > game.tsv`; `test_hit_recovery --knock-modes | grep '^K ' > knock.tsv`; `mode_switch.py compare-knock game.tsv knock.tsv`.
