# Issue draft (local, NOT posted): a raised monster plays its ResurrectMode, d2d puts it in NU

> status: draft 2026-11-15 (card decomp26, P09). Nothing posted, no fork branch. Source: `docs/probes/fid/fid-revive-mode.md` (TASK-316, decomp9: game side V run, d2d side V read). Re-checked on upstream main `fb6d799d` (V read): `git grep -i ResurrectMode components` is empty and `components/game/ai.cpp` still ends the raise with `set_mode(game_data, raised, "NU", now_ms)` (the Resurrect / Resurrect2 branch, line 1963 area).

## Title (proposed)

A Fallen Shaman's Resurrect puts the raised monster in NU; game.exe starts its MonStats2 ResurrectMode

## Body (proposed)

When a Fallen Shaman (or any `srvdofunc 97` caster) raises a corpse, game.exe's `FUN_005ccb10` calls `FUN_005cc960` (life := max life, unit flags +0x4000000 and +0x20000, back in the room list) and then asks `FUN_005fd7b0(unit, 1)` for a mode: the byte at +0x10b of the unit class's MonStats2 record, the `ResurrectMode` column. The mode is applied in the same server step (`FUN_005a7e60` builds the request, `FUN_005a7c20` applies it), so the monster is alive and already in S1 (Fallen, Mummy) or SQ (the skeleton families); NU otherwise.

d2d (`ai.cpp`, the `Resurrect` / `Resurrect2` branch) sets `NU` for every class.

| Item | Value | Tag |
|---|---|---|
| Rule | mode := MonStats2 `ResurrectMode` (NU, S1, SQ), applied in the raise step | V run |
| Coverage | 703 spawnable MonStats rows run natively through `FUN_005fd7b0`: 638 NU, 14 S1, 51 SQ; the 578 rows with a MonStats2 row equal the column (NU / S1 / SQ = 1 / 8 / 14), 0 disagree | V run (`tools/emu/d9_revive.py`, branch decomp9) |
| Fallen S1 length | 24 frames at speed 256 (`FAS1HTH` in the AnimData), about 24 ticks if the S1 plays at rate 256 | D, rate I |
| d2d | NU at once for every class | V read |
| Size | about one second of the raised monster standing in S1; the raised Fallen is alive in both | I |

## What is not proved (U)

- Whether a monster in S1 can be targeted or struck the same way as in NU (the mode gates `FUN_0057edd0` / `FUN_0057eec0` are the player's; the monster's are not traced).
- Whether the S1 plays at rate 256 (the rate of a monster's skill mode, stat 69, was not read).
- The no-experience / no-drop meaning of flags 0x4020000: no reader with such an immediate was found in `.text` (dead end, `d9_flagscan.py`); d2d's `no_exp` / `no_drop` stays as it is.

## Oracle plan (not built)

There is no differential on the d2d side because the raise is inline in the monster loop. Build a `test_game` scenario (a Fallen Shaman, a Fallen corpse in DD within the AI's range, step until the corpse moves, read its `mode`) and compare with the `M class mode` lines of `d9_revive.py` (game.exe's own `FUN_005fd7b0` per class). Fix shape: read `ResurrectMode` from MonStats2 where the Resurrect branch calls `set_mode` (one column in `gamedata_load.cpp`, three lines in `ai.cpp`).

## Impact

Small: the raised monster is alive and hostile in both builds; only about 24 ticks of S1 animation (and what a hit does to a monster in S1, U) differ. Filing as an issue first because the fix needs the S1 targetability question answered.
