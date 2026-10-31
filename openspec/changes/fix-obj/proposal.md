## Why

Live drill group H (docs/probes/drills/2026-10-31-h.txt, hover-audit.md) proved two d2d primitives broken: field objects (chests, shrines, Cain's cage, the Tristram portal) were never listed by the d2d state source and `Interact` answered ErrUnsupported for objects, so the bot only walked next to them (H.obj); and `d2dctl.Pause` was a wall sleep, so under lockstep the NPC approach of merc claim and revive never saw a game tick (H.pause).

## What Changes

- internal/d2dsrc: Objects gain the placed objects within 40 cells of the player from `debug presets` (class = objects.txt id of the row, id = NPCIDBase + preset index, the index space of `cmd interact`); hidden rows, rows `npcs` lists and far rows are skipped; the list is cached per level and refetched after 50 subtiles of walking. d2d only.
- internal/d2dinput: Interact on such an id sends `cmd interact <index>` (as for NPC rows); raw unit ids of objects stay ErrUnsupported. d2d only; Game.exe's path and the shared interfaces are unchanged.
- internal/d2denv: `LateClock.Pause` replaces `d2dctl.Pause` when the backend starts: under lockstep it steps the game by the requested time (the Pacer), otherwise a wall sleep that ends with the context. d2d only.
- Delta: backend-contract ADDED requirements.
