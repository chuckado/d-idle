# Soak stall: area ping-pong without kills (run gate1-20261101-1704)

## Symptom (V: read from logs/bot-0.log of the run, read-only)
Kills 2122 (17:43:20) to 2129 (17:49:08), then none; clvl 10, xp 71203 constant. From about 17:40 the log is a loop of about 35 s wall (about 7 virtual minutes at the logged pace ratio 13): `Entered Cold Plains` -> `poverty mode: on (... no town portal)` -> `poverty mode: farming area 2` -> `nav exit to 2` -> `Entered Blood Moor` -> `poverty mode: off (kit complete)` -> `nav exit to 3`. Every few cycles `task step level.17 blocked: explore: nav: timeout: roaming for 5m0s`. `grep -c "area cleared\|area abandoned"` after 17:40 is 0.

## Cause
1. Both fields are emptied (hostileNear 0, roam coverage finds nothing, one zombie in 6 minutes) and nothing ever calls them cleared:
   - `sweepPass` marks `cleared[area]` only when Explore returns nil (internal/bot/goal.go, the block after `r.d.Mover.Explore`). A roam that ends with `nav.ErrTimeout` (internal/nav/roam.go:182, the 5 minute explore limit) returns `explore: ...` and the area stays uncleared. V: the error text in the log is exactly that path.
   - `emptyAreaLeave` (goal.go) is the designed exit, but `bot.emptyAreaSec` is off by default (card m2-roam task 1.3 is still open; no `emptyAreaSec` in cmd/ or in the soak's plan.env, 0 "area abandoned" lines). Even when on, its timer restarts on every area change (`l.emptyArea != area`) and after any 90 s gap (`stale`), and the ping-pong did both every few minutes. U: not run with the knob on.
2. Why the bot alternates instead of sitting in one field: `resInput().HasTP` (resource.go:33, combat.PortalSource) reads false in Cold Plains and true in Blood Moor on arrival, so poverty mode (resource.PovertyCheck) flips with the area, `starveArea` picks Blood Moor, the plan picks Cold Plains, and `holdTarget` (goal.go) only holds 3 virtual minutes (10 when 3 flips fall in 10 minutes; the cycle was 7 minutes, so it never reached the longer hold). V: poverty lines per cycle. Why HasTP differs per area is UNKNOWN (the shop log says `bought Scroll of Town Portal` and the next line `back to field: no portal in the unit table`; suspect the d2d inventory dump or tome stat read, see Missing data).

Hypotheses: (a) true (no cleared mark, no next area); (b) not supported (the roam moves legitimately, there is just nothing to kill); (c) no, level.17 is only the plan step label; (d) not tested, no restart in the stall window.

## Fix (goal.go `exploreGaveUp`, runner.go `exploreFail`)
A failed Explore with no hostile near counts against the area; two failures in a row with no kill in between (per-area count, so area changes do not reset it) mark the area cleared and the pass returns nil, like `emptyAreaLeave`. The planner then takes another area or the fresh-game path (cleared areas are excluded in `target0`). Quest areas (`l.run.Area == area`) never give up. Test: `TestExploreGaveUp` (internal/bot/emptyarea_test.go), `go test -p 1 ./internal/bot` passes.
Not fixed: the HasTP flip, and the knob default (m2-roam).

## Live drill (checklist, not run)
1. Soak or d2d run at clvl about 10 with Blood Moor and Cold Plains emptied (or `bot.emptyAreaSec=0`): expect `area abandoned: exploring failed with no kill` for both within two roam timeouts, then `fresh game` or the next plan area, and kills resuming within the 10 minute progress rule.
2. Check no regression: a field with monsters far apart (roam timeout then a kill) must not be abandoned after the kill (count restarts).
3. Log `HasTP` and the portal item list at area entry to find the per-area flip.

## Missing data
A state dump of `s.Inventory.Items` (class, Where) plus `UnitStats` result for the tome/scroll at entry to Cold Plains versus Blood Moor.
