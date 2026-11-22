# Map fog for the overlay layers (mapfog)

User request 2026-11-12: the coverage squares (and other overlays) must not show unexplored ground.

## What was wrong
- Terrain was already fogged server-side (`senses.Publisher`, docs/senses-map.md): cells, exits and waypoints outside the hero's visited blocks are never sent. The overlays were not: `coverage.cells` carried the roam grid of the WHOLE level (the bot's pre-loaded map, V by reading `nav.CoverageView` and `SetCoverage`), its unseen outline squares and the roam target showed where unexplored floor is; failed-roam events and heat cells were not checked either.

## How "explored naturally" is computed
- The mask is the Publisher's own visited-block set: 4-subtile blocks within `senses.RevealRadius` (24 subtiles, U: a guess near one screen half-width, never checked against the automap reveal radius; no d2d automap reveal data is read) of every hero position the bot observed (about 2 Hz), kept per area for the session. It is built only from the hero's positions, not from the bot's level data. S: it is not the in-game automap (which reveals by room/line of sight); it is the same mask the map terrain already used, so squares and terrain agree.
- A point is explored when the block holding it is in the set (`Publisher.revealedLocked`).

## What is withheld (server side, `/api/senses`)
- `coverage.cells`: points in unexplored blocks are omitted; `coverage.target` is omitted while its block is unexplored. `seen`, `failed` and `total` still count every point (aggregates, no positions); additive key `hidden` (omitted when 0) says how many points are withheld; `drawn` is the number sent. Schema string unchanged (`watch.senses/1`), no golden changed: old pages ignore `hidden`, and the page shows "N not explored yet".
- Events (`EventsOf`, e.g. a failed roam target far away) and heat cells (`HeatOf`) are filtered the same way. The trail is the hero's own path, explored by construction. Units are what the game shows near the hero (unchanged).
- Layer order on the canvas is unchanged: terrain, heat, coverage, trail, hero, units; unexplored cells carry no terrain, so the dark background is the fog.

## What the bot still knows
Everything: the roam grid, targets and the level map stay in `nav`; only the published view is masked. Nothing about bot behaviour changed.

## Checks
- V (offline): `go test -p 1 ./internal/senses -run 'CoverageWithheld|EventsAndHeat|CoverageSpans'` (a field and a cave: fresh start shows a few squares near the hero, none beyond the radius, target hidden; walking the corridor reveals squares only along it), `go test ./internal/watch -run 'Coverage|MapPanel'`, node test `testdata/map_panel.js` (`coverSummary` with `hidden`).
- V (headless render): `tools/uishots` serve flag `-fog start|walked` serves the real Publisher over a synthetic cave with a roam grid on every walkable point (Coverage layer forced on) and Chrome on the worker host takes the picture. Start: only 4 outlines beside the hero ("0/44 seen, 40 not explored yet"); walked 80 subtiles along the corridor and down the bend: squares and green seen tint only on that stretch ("31 not explored yet"). Pictures are not committed.
- U: the live d2d run (hero, grid and roam points share a frame by code reading only).

## Walked ground and the objective (card fogtrail, 2026-11-22; user decision on map-knowledge Q5)

The viewer's Map shows the ground the hero has walked and ONE marker beyond it, the objective. It does not show the outline of the level (it would give away the layout the fog is there to hide).

- Walked ground: the Publisher's visited-block set (`seen`, 4-subtile blocks within `senses.RevealRadius` of every hero position the bot observed, about 2 Hz, on every leg: roam, quest, travel; it never depended on the roam coverage). V (offline): `go test -p 1 ./internal/senses`.
- Size: one `block` key per 4x4 subtiles per area, at most about 8400 for the largest outdoor level (480x280 subtiles), a few hundred for a cave; the whole set is a few hundred KB of Go map for a long session, and is bounded by the levels of one act (the trail ring is separate and capped by time).
- Persisted: NO. It lives in the bot process, per game: `Publisher.NewGame` (called through `nav.ForgetCoverage` when a new game starts) drops it, because a new game lays the same area id out again (a different Tower Cellar) and the old blocks would reveal the new layout. After a bot restart the map fogs again; that is the honest picture of a hero that has not been there in this run. Keeping it over a restart would need a store keyed by game seed and level (not worth it for a view).
- Objective: `objective` {area, x, y, label} in `/api/senses` (tier 1, additive, absent when the hero has no walk goal or the goal is in another area). It is `nav.Navigator.ObjectiveView`: the goal of the walk in progress (a GoTo with a goal: a boss anchor read from the engine, a quest object, a waypoint, an exit). It is NOT fogged: having it is the point (ADR 0022 reads give the hero these places on arrival). The Map draws a ring with a dot in the accent colour and says "the ring is the goal" in its note; the panel has no inner scroll box and the canvas follows the panel width, so it reads on a phone. `label` is empty for now (nav has no name for a goal).
- Not done: a label per objective kind (the quest step name would need the bot to pass it), the objective of a roam target (the roam already shows its own `coverage.target` once explored), a live look at the real viewer (U: only the stub-DOM test and the Go tests ran).
- Tests: `go test -p 1 ./internal/senses ./internal/watch ./internal/nav -run 'Objective|NewGame|ForgetCoverage'`, and `node internal/watch/testdata/map_panel.js internal/watch/web/panels/map/panel.js` (the ring is drawn for the objective of the current area only).
