# Level layout read (K1, ADR 0022)

> status: current (2026-11-20). Tags: V = verified (say how), U = unverified.

## What

The private d2d fork has a read-only control verb `debug levelmap` (fork branch `didle/level-layout-20261120`, on the private remote only; patches never go into this repo). It answers one JSON line, then `ok`: the CURRENT level's id, width and height, its rooms (`x y w h preset def`, level cells; `def` is the LvlPrest row of a preset room, -1 otherwise), its objects (`index class name x y kind state hidden`, kind from OperateFn: shrine chest tree stone cage waypoint portal door object), its warps and the player's open town portals. It exposes only what the engine already generated; nothing from other levels, no memory writes. It is on the fork's read-only allow-list, so it needs no `--lab-verbs` (U until the fork build runs the test: see below). By ADR 0022 an engine fact read this way never makes a run CHEATED; record the use in the run record (the knob value is in the knob dump).

## Bot side

- `internal/d2dctl/levelmap.go`: wire types and `Client.LevelMap`; the fake (`internal/d2dctl/fake`) answers `err` unless `Fixtures.LevelMap` is set (stock d2d behaviour).
- `internal/d2dworld`: with `Options.LayoutRead` true, `fetch` adds the objects and preset rooms the layout knows and `debug presets` does not yet (objects of rooms not built are the point) to the level's `Presets` and `RoomRows`. Nothing is sent with the knob off; an error or a missing verb leaves the level as `debug presets` made it.
- Knob `quests.layoutRead` (default 0, off; `docs/knobs.md` is generated: `go run ./tools/knobdump`). The d2dworld option reads it live (`internal/d2denv`).
- Consumer: `cmd/d2bot/objects.go` `findNear` already walks to a known preset of the wanted class (the Tree of Inifuss class 30, the cairn stones 17.., the Tristram cage class 26), so with the knob on a far quest object is walked to straight; with the knob off, or the verb absent, `find` explores as before.

## Checks

- Offline (this repo, V): `go test -p 1 ./internal/d2dctl/... ./internal/d2dworld/... ./cmd/d2bot -run LayoutRead ./internal/env/... ./tools/knobdump`. The must-fail cases: hero far from the tree, knob on gives one walk to the tree's position; knob off or verb absent gives no walk (two start spots, a Paladin and a Sorceress arrival).
- Fork side: `tests/levelmap_d2d.py` (Dark Wood and Tristram): the object list equals `debug presets` rows (same engine table walk), 0 differences; state, presets and levelmap are equal before and after repeated reads. Result is in the task card.
- The units of the layout (cells, like `debug presets`) are U until a live capture compares a room rectangle with the hero's position.
