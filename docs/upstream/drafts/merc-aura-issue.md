# Issue draft (local, NOT posted): the merc's own life regeneration (stat 0x4a); friendly auras on the merc are now upstream

> status: draft 2026-11-15 (card decomp26, P10). Nothing posted. Source: `docs/research/merc2-trace.md` D1 to D3, `docs/research/mercstats-trace.md` D1 / D2. **Re-check on upstream main `fb6d799d` (V read): two of the three are no longer open.** Commit `50f7f610` ("The merc's friendly auras on it and the player, the player's on it") makes `Fight::friendly_auras` join an aura's `aurastat`s into the merc's stats (`components/game/fight/merc.cpp` line 81 area) and the Act 2 merc's aura reaches the player and itself; so D1 (friendly auras reach the merc) and D3 (the Act 2 merc's aura does nothing) of the trace are merged. What is still open is D2, the merc's life regeneration.

## Title (proposed)

The hired merc has no steady life regeneration; game.exe sets stat 0x4a at every level-up

## Body (proposed)

game.exe gives every hired merc life regeneration: stat 0x4a (74) is written as the merc's maximum life (stat 7, 256ths) divided by 2000, at creation and at every level-up (`0x572ac2..0x572af2`). The monster regeneration event (`FUN_00580610` / `FUN_00580810`) adds it to the life stat each frame once the unit is hurt.

| Item | Value | Tag |
|---|---|---|
| Stat | 0x4a := stat 7 / 2000 at creation and level-up | V run (stat value, mercstats-trace Q4: Rogue level 15 gives 19 256ths, level 50 gives 76; Desert level 20 gives 36, level 75 gives 183) |
| Adder | `FUN_00580610` / `FUN_00580810` add the stat to life per frame once hurt | V read (I for the exact gate "once hurt", `monster-stats-trace.md` Q9) |
| Size | life / 2000 per 40 ms frame = 1.25 percent of max life per second (Rogue level 15 about 0.075 of a point per frame) | derived from the stat |
| d2d | no regeneration of the merc outside potions, auras and Prayer (V read: `git grep -i regen components/game/fight/merc.cpp` is empty; the merc's healing is the potion path in `setup.cpp:267` and the aura pulse) | V read |

## Oracle plan (not built)

Game side: create a merc (`FUN_00572a40` family) or hurt a spawned one, run the regeneration event for N frames with the real stat set by the level-up code, record life per frame (`tools/emu/mc2_run.sh`, `mercstats` scripts exist). d2d side: a `Fight` test with a merc below max life and no potion or aura, step N ticks, read `merc->hit_points`. Gate: equal life after N frames for levels 1 to 82 (the stat is a function of max life only). Fix shape: add `merc_max_life() / 2000` (256ths) to the merc's regeneration in the same per-tick place as the potion heal.

## What is not proved (U)

The exact gate of the event (hurt only, or also while in a mode such as get-hit), whether stat 0x4a is also applied while the merc stands in town, and the per-frame granularity (the event may be per 25 frames rather than per frame). The effect on survival is U until the oracle runs.
