# fixstall2: NPC approach over the map, portal entry from a pocket

> status: current for DRILLS6 findings F7 and F8. Predecessor: [fixstall.md](fixstall.md). Card `openspec/changes/fixstall2`.

## Root causes (evidence: docs/probes/drills/README.md DRILLS6, drills6-h/hb/d6pin/d6pinb logs)
- F7 (H.pause, 1 of 2 tries fails, deterministic per try): `ApproachNPC` only knows "interact walks in a straight line until something blocks" and then tries up to 3 blind side steps of 3.5 cells. The failing try ends at (25.9, 20.5) after 718 ticks (budget 28.4 s): the first camp objects around (25.5, 22.5) form an obstacle (and Kashya wanders) wider than one blind step, so the side step lands behind another blocker and the walk stops dead again. The logs hold only start and end positions and the error, no object list, so WHICH object or NPC blocks is UNVERIFIED. What is known: `debug grid 5` gives the level collision (V live, d2d-backend.md), and `ApproachNPC` never used it. The fix targets that cause: the approach had no map.
- F8 (D6.d run 2, 62 s pinned): `town.enterPortal` did `GoTo(portal)` before `Interact` whenever the portal was farther than NPCStopDist. From a collision pocket the walker cannot reach the portal (the ladder runs again, "open portal did not lead to town"). The d2d Interact for a portal is the `EnterPortal` verb (`cmd interact -2000-which`), which walks the player to the portal and moves it to the twin by itself, so the pre-walk was both unnecessary and the thing that failed. Run 1 only worked because the portal opened inside the ring.

## What changed
- `internal/d2dctl/approach_path.go`: `planAround` runs a BFS over the subtile grid (clearance 1 subtile, no corner cutting, other NPCs blocked within 4 subtiles) to the free cell nearest the NPC (ends within 6 subtiles), returned as waypoints 12 subtiles apart. `ApproachNPC` uses it on the 1st, 3rd, ... stall (`routeAround`: grid verb, legs with `cmd move`, a leg ends on arrival or 4 still polls); even stalls, no grid or no route fall back to the old blind side step.
- `internal/town/trip.go`: `WalkingInteractor` (optional input interface). `enterPortal` skips the `GoTo` when the input's Interact walks itself; `internal/d2dtown/adapter.go` implements it for portals and NPCs.
- No numeric gate changed.

## Tests (offline)
`go test -p 1 ./internal/d2dctl ./internal/town ./internal/d2dtown`: `TestPlanAroundWall`, `TestPlanAroundAvoidsOtherNPC`, `TestPlanAroundNoRoute`, `TestApproachNPCRoutesAroundTheWall` (synthetic grid: no recorded drill geometry exists in the logs), `TestEnterPortalDoesNotWalkWhenInteractWalks`.

## VERIFIED LIVE vs UNVERIFIED
- Verified live: nothing of this card.
- UNVERIFIED: that `debug grid 5` carries the camp fires/stalls (if objects are not in the grid the route equals the straight line and the fallback side steps apply as before); the clearance and standoff numbers; that the portal verb works from the pinned pocket (it should: it does not use our walker).

## Live proof recipe
On the worker host through labrun, lab build cheats3: `D2D_AGENT=fixstall2 tools/agentrun.sh fixstall2 --ttl 900 -- tools/d2dslot.sh run -n 0 -t 900 -- d2drill -group F -only H.pause -label fixstall2-h -v` three times (each run has 2 tries: want 6 of 6 opened), and the same with `-only D6.d` three times (want ErrLevelChanged escape in all, under 30 s). If H.pause still fails, log `debug grid 5` around the end position and the NPC list at the failing try to see what blocks.
