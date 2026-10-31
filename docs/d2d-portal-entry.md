# Entering a town portal on d2d (agent enterportal, 2026-10-31)

> status: current. Card openspec/changes/enterportal. Fixes drill finding W1 (docs/probes/drills/README.md): no portal walk-in worked in the lockstep lab.

## Answer

- **What enters a portal on d2d:** the client command `Interact{-2000 - which}` (which = first column of `debug portals`, the World's portal index: 0 the cast end, 1 its twin in town). The client sends it for a left click while its hover pick names the portal. The control socket sends the same command with `cmd interact -2000-which` (`d2dctl.TakePortal`). The World walks the player to the portal, and once within 2 cells it moves the player to the twin, arriving 0.6 cells below it.
- **Why nothing entered before:** d2d computes the hover pick (`hovered_npc`: monsters, ground items, NPCs, objects, portals) only while it draws a frame (V-d2d: apps/d2d/ingame.cpp sets it in the world render; town input reads "last frame's"). The lab fork's `--lockstep` draws no frame unless `step n render` or a screenshot asks for one (V: fork apps/d2d/lockstep.cpp `draw_frame`). So the hover never names the portal, and a click on its sprite becomes `cmd::Move`: the player walks. This is a **lab artefact**, not a d2d deviation. On top of that, the bot clicked a screen point from the uncalibrated transform (`d2dinput.UsePortal`, marked U).
- **Kick ordering (V live):** after a level change the controller sends its priming click (`Kick`). With no tick between that click and `cmd interact`, the click's fresh walk dropped the portal command (D2.f in town). `prime()` now lets one tick pass after the kick.

## Live probes (lockstep slot, lab click binary, throwaway Sorceress, `--no-save`; docs/probes/enterportal/2026-10-31-ep.txt)

| Sequence | Enters | Ticks |
|---|---|---|
| `cmd move` onto the portal cell (0.6 cells away) | no | - (150 stepped) |
| hover x-3, hover x, left click on the sprite, 15 plain lockstep ticks between | no (the player walked) | - |
| hover, `step 1 render`, hover, `step 2 render`, left click | yes | 1 |
| `cmd interact -2000` next to the portal | yes | 1 |
| `cmd interact -2000` from 4.5 cells | yes (d2d walks there first) | 34 |

Cheats (logged ledger, actor `drills:<id>`): `debug clearinv`, `debug give tsc`, `debug goto 2`. Character left alive in town.

## game.exe's rule (from docs/research/town-portal-trace.md, not re-traced here)

Object 59 has OperateRange 2 and OperateFn 15 (D): a click on the object makes the player walk into range, then the operate function moves them to the twin, and the owner going back closes the pair (S). Walking onto the portal without the click does nothing, the same as in d2d (U for game.exe: not traced in the emulator).

## Fix (bot side)

- `d2dctl.TakePortal` / `PortalNPC`. `EnterPortal` now uses them, not the sprite click.
- `d2dinput.Controller.UsePortal(which)` sends `TakePortal` after `prime()`. `d2dsrc` gives the portal objects id `PortalIDBase + which` (it used the list row).
- `town.ToTown` enters an open portal within `portalReuseDist` on every try and reads a scroll only when none is in reach. When only an open portal was tried and none led to town, it walks back or takes the waypoint (scenario `TestScenario_OpenPortalBesideThePlayer`).

## Drills after the fix (docs/probes/enterportal/2026-10-31-d2.txt; the first run, before the kick tick, is `-d2-run1`)

D2.t PASS (ToTown enters in 5 ticks, no scroll spent), D2.g PASS, D2.f PASS (arrival 6 subtiles from the origin). Each ran twice with identical results.

## Not done / open

- The bot's own lockstep runs keep `d2dctl.Pause` on wall time (only cmd/d2drill steps it). Verbs that rely on Pause (EnterPortal polling, HoverClick) act on a frozen world under lockstep in cmd/d2bot. The portal path no longer needs them (UsePortal waits through the env clock).
- Other hover-dependent clicks (ground items, monsters, objects by click) have the same lockstep blind spot. The bot uses `cmd pickup` / `cmd skill` for those, so nothing else is known to be affected.
