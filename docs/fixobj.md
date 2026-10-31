# fixobj: field objects and lockstep pauses on d2d

> status: current. Card fix-obj. Offline fix of drill group H findings (docs/probes/drills/hover-audit.md); the live proof is a drill rerun after merge.

## What was wrong (V drill H.obj, H.pause)
- Field objects were not in the d2d snapshot and `Interact` returned ErrUnsupported; the fallback `cmd move` only walks next to a chest. `cmd interact <index>` opens one (the drill's control group).
- `d2dctl.Pause` was a wall sleep; under lockstep the game does not move between hover and click, so the Kashya approach timed out (8 s wall, 0 ticks; with a stepping pause 147 ticks).

## What changed (d2d only; Game.exe and the shared interfaces are untouched)
- `d2dsrc` (items.go `fieldObjects`): `debug presets` rows within 40 cells become Objects: Class = the row's objects.txt id, id = `NPCIDBase + index`, name in Extras.Names. The index is the one `cmd interact` takes and `npcs` rows share (V: the waypoint row of fixwp). Skipped: hidden rows, indexes `npcs` already lists. Cached per level, read again after 50 subtiles (presets only change when rooms are built). Object mode stays unknown (`debug objects` has it under another index).
- `d2dinput` `Interact`: ids in the NPC range, objects included, send `cmd interact <index>`; a raw unit id of an object stays ErrUnsupported.
- `d2denv` `LateClock.Pause`, installed into `d2dctl.Pause` by `Backend.Start` after the clock is detected: lockstep (the Pacer exists) steps the game by the requested time; otherwise a wall sleep that returns on context cancel. Lockstep is detected by `clock` answering `lockstep=1`. Every d2d caller (d2bot, d2dshop) gets it; `cmd/d2drill` keeps its own.
- Not changed: `cmd/d2bot/objects.go` (its `ClickWorld` fallback is now only reached on a real error), `internal/nav/waypoint.go` (waypoints are `npcs` rows, fixwp).
- Note for the town adapter owner: `d2dtown.townInput.Interact` treats ids from PortalIDBase up as portals; field objects use the NPC range, so they are not affected, but a chest id passed to the town input would ApproachNPC (wait for a menu). No caller does.

## Tests
`TestFieldObjectsListed` (d2dsrc), `TestInteractFieldObject` (d2dinput), `TestFieldChestListedAndOperated` and `TestPauseStepsUnderLockstep` (d2denv).

## Drills to rerun after merge
H.obj (chest, shrine, cage, Tristram portal, field waypoint by the bot's own verbs), H.pause (Kashya approach with the bot's pause), merc claim and revive D8, D3.n (NPC menus). U: whether a cage or a Tristram portal row is in `debug presets` with the class the bot expects (`quest.ObjCainGibbet`, `ObjRedPortal`); the rerun checks the classes.
