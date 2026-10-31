## Why

Drills D2.t and D2.g (docs/probes/drills/2026-10-31-d2.txt, finding W1): with a town portal open next to the player, neither the bot's `ToTown` nor the d2dctl `EnterPortal` verb ever entered it on a lockstep d2d, and `town.ToTown` then spent every remaining scroll on new casts. The 4 h soak shows the same signature (portal failures, 'portal did not lead' lines, 425 of 520 trips fell back to the waypoint). Cause (docs/d2d-portal-entry.md): d2d's client turns a left click into "use the portal" only when its hover pick, computed in the frame draw, names the portal; the lab's lockstep mode draws no frame, so every portal click became a walk. The bot clicked a screen point from the uncalibrated transform as well.

## What Changes

- `d2dctl.TakePortal` sends the command the client's portal click sends (`cmd interact -2000-which`); `EnterPortal` and `d2dinput.Controller.UsePortal` use it (no screen point, no hover). `d2dsrc` gives portal objects the World's portal index (`which`), not the list row.
- `town.ToTown` enters an open portal within `PortalReuseDist` on every try and reads a scroll only when none is in reach; when only an open portal was tried and it never led to town it falls back to the walk/waypoint path instead of casting.
- Delta: backend-contract and town-basics ADDED requirements.
