# Candidate note (local, NOT posted): P03c, a knocking hit inside a get-hit

> status: candidate 2026-11-15 (card decomp26). Nothing built, nothing posted. Follows P03b (#548: a big hit does not restart a get-hit or a block; a hit during a knockback is allowed). Re-checked on upstream main `fb6d799d` (V read): `components/game/fight/tick.cpp` line 660 still reads `if (foe.knocked_from && !attack_mode(actor->pmode) && actor->pmode != kModeBL)`, with no get-hit test.

## The rule (game.exe, V read; the table half V run)

A knocking hit (the knockback flag of `FUN_0057cee0`, `LAB_0057d364` area) goes through `FUN_00580a70` with new mode 0x13, which asks `FUN_0057edd0` first. `FUN_0057edd0` refuses a new mode 0x13 from DT, GH, BL and DD like a get-hit (mode_switch.py column 4 / 9 results: refused from GH and BL). So a knocking hit that lands while the player is already in a get-hit or a block changes nothing: no slide, no new hold.

## What d2d does (V read)

`Fight::hurt_side` starts the slide whenever `foe.knocked_from` is set and the mode is not a swing and not a block (`tick.cpp:660`): `actor->knock` is set, `set_pmode(kModeGH, ...)` runs and `pmode_until` is moved to the slide's length. That restarts the hold and moves the player during a get-hit; game.exe refuses.

## Oracle plan

`tools/emu/mode_switch.py` already runs `FUN_0057edd0`; add the new-mode column 0x13 (19) to its `game` loop (current mode 0..19 x new mode 4, 9, 19 x queued 0, 1) and a `--knock-modes` line to `test_hit_recovery`: a hit with `knocked_from` set on a Paladin standing in each mode with a hold running, printing the mode after and whether `pmode_until` moved or `knock` was set. Compare game.exe's 0x13 answers (allowed) with d2d's `entered`. Expected before the fix: the GH and BL cells differ (d2d enters, game.exe refuses); after `&& actor->pmode != kModeGH` (a real GH, not a knockback slide: `!actor->knock`): 0 differ.

## Open (U)

Whether the 0x13 row of `FUN_0057edd0`'s table has the same current-mode gates as modes 4 and 9 (the P03b run only covered 4 and 9); the swing modes (P04 changes what a hit does inside a swing, so this should be built after it); whether a refused knockback still deals its damage and plays the struck cue (damage: yes, the hit's damage is separate from the mode switch, I).
