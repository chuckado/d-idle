# Brief S5: Orders write path and the act map as a slot panel

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis5`, branch `agent/uis5`, worktree `~/code/d2lod-wt/uis5` off current `origin/main`.

**Goal.** `POST /api/commands` and `/api/commands/cancel` as a dispatcher over the existing actions, the live verb row with Undo and keys, and the slot panel `zone-map`: the act as a grid of one-line zone buttons with the zone menu.

**Starts after:** S4 merged (B4 and B7 before the Loop goldens are final). **Can run next to:** S3, S6, C2.

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md (sections 4 to 8)`; `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 2, the act map in a rail)`; `docs/ui/redesign-2026-11-14/mockups/v2/watch.html`.

**You own:** `internal/watch/commands.go`, `internal/watch/commands_post_test.go`, `internal/watch/web/panels/orders/`, `internal/watch/web/panels/zone-map/`, `internal/watch/testdata/zonemap/`.

**Do not touch:** `orders.go`, `unlockorderqueue.go`, `repairapi.go`, `invapi.go` (call their methods); `app.js` and the old `orders-map` card stay for free-grid layouts. In `http_answer.go` only the two route-table lines.

**Already known (do not redo):** The stored node coordinates use three X values and ten Y values for Act I, which is what makes the compact grid possible.

## Do
1. The dispatcher as in the first S5 brief: `{intent, args, mode}`; `go`, `fight_here`, `town` to `Hub.place`; `mode: "queue"` to the order queue (its `locked` code while closed); `loop` to the loop controller when bound; `repair`, `farm_gold`, item intents, `journey`, `pause`, `sell_now`, `revive_merc` to their desks when bound, else 503 `off`. Same gates and error codes as the old routes.
2. Verb row live with at most five buttons from `verbs[]`; every accepted command shows the Undo toast for five seconds; keys `g t r p o z /` and `Esc`.
3. Panel `zone-map` (slot `rail`): from `/api/orders/map`, the nodes as a CSS grid of buttons, three columns by ten rows for Act I (column and row from the node's stored X and Y), each one line: band stripe (below, in range, above, too hard, not orderable), name, and at the right the level range, `here`, or the queue number. Edges and the gold route line are an SVG underlay scaled to the grid. Arrow keys walk the graph. An act without map data shows the list of orderable areas in the same box.
4. The zone menu never covers a zone button or the picture: on a two-column rail it opens in the area-map tile beside the map, otherwise over the Orders tile directly above the map; the picked zone is outlined. All zones of the act are visible at every supported height: the zone rows share the tile's height, drop their second line when the tile is under about 430 px, and never go under 20 px. Content of the menu: name, range and route in one or two lines, the guard's reason when too hard, duration chips, Go now (Ask anyway), then Loop and Add to queue when open or their ladder teaser as one muted line when locked; the town node offers the town errands; the charter is paid inline.
5. A switch `Act | Area` in the tile header shows the `area-map` panel (S6) in the same box when the layout has no separate slot for it.

## Acceptance (also the card's boxes)
- Answer-route tests for each intent: gates, and the same resulting state as the old route
- DOM goldens: the grid for the fixture act with here, target, two pins, a failed mark; the menu for field, here, town, hard, locked, charter, loop locked, loop open, queue locked, queue full
- At every size of the fit rule the map tile shows all ten rows with readable names (no row under 20 px) and the open menu overlaps no zone button (node test of the two rectangles; fit check with the menu open)
- From a zone click to an accepted order is two clicks (node test)
- The old `orders-map` card still places an order in Classic

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis5: Orders write path and the act map as a slot panel" --card uis5 --owns internal/watch/commands.go --owns internal/watch/commands_post_test.go --owns internal/watch/web/panels/orders/ --owns internal/watch/web/panels/zone-map/ --owns internal/watch/testdata/zonemap/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "Answer-route tests for each intent: gates, and the same resulting state as the old route" \
  --ac "DOM goldens: the grid for the fixture act with here, target, two pins, a failed mark; the menu for field, here, town, hard, locked, charter, loop locked, loop open, queue locked, queue full" \
  --ac "At every size of the fit rule the map tile shows all ten rows with readable names (no row under 20 px) and the open menu overlaps no zone button (node test of the two rectangles; fit check with the menu open)" \
  --ac "From a zone click to an accepted order is two clicks (node test)" \
  --ac "The old `orders-map` card still places an order in Classic"
```

## The fit rule (binding for every cockpit and canvas slice)
- The page must not scroll and nothing may clip at the full-screen sizes 1920x1080, 1440x900, 1366x768 and at the browser-window sizes 1914x930, 1904x1000, 1536x730, 1280x620 (a window with its chrome is shorter than the screen: the user's first look was at 1914x930); at 390x844 it follows the phone design of J. `tools/uishots` gets (in S1b) a `fit` check that loads each page at those sizes in headless Chromium on the builder and fails when `document.scrollingElement.scrollHeight > innerHeight`, when any `[data-slot]` has `scrollHeight > clientHeight` (lists marked `data-scroll` excepted), or when the slots do not cover the frame (sum of slot areas plus gaps equals the frame, within 2 px). Run it (`labrun <builder> -- tools/uishots.sh --fit`, never on the worker host that runs the games) and put its output in the card result. A slice that fails the fit check is not done.
- A slot panel has no content-driven height: it renders inside the box it is given; long text is cut with an ellipsis and the full text goes to `title` or a popover; a list longer than its box scrolls inside (`data-scroll`).
- Look at your own renders at those sizes and write three lines of critique in the card result (dead space, alignment, noise). Compare with `docs/ui/redesign-2026-11-14/research/iterations/it3/` (the accepted look).

## Every UI slice (binding)
- Rules: `AGENTS.md`, `docs/agents/rules-short.md`. Own worktree and branch, commit often, never push, no deploy, no live game, do not touch the sandbox. Trailer `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`.
- Compatibility promise (`docs/ui-design-system.md` section 0 and the checklist in 6.2): additive only; a new feature is a new panel directory; no fixture under `internal/watch/testdata/compat/` is edited; no panel id, endpoint field, token or data attribute is renamed or removed; `go test ./internal/watch -run Compat` stays green. A stored layout must render the same cells as before your change.
- Panels: three files (`panel.json` with `version` and `optIn: true`, `panel.css`, `panel.js`), tokens only, no timer or fetch of their own, no new polling: one endpoint per panel through the shell scheduler.
- Tests: Go handler goldens for both classes (Sorceress and Paladin fixtures), node DOM goldens per panel state in `internal/watch/testdata/`, `go run ./tools/uilint` clean.
- Pictures: `labrun <worker> -- tools/uishots.sh --out <dir>` from inside your worktree (headless Chrome on the worker host); look at 390 and 1440 px and name what you saw in the card result. Do not commit pictures over 256 KB.
- Gates before queueing, from inside your worktree with no `cd` in the command:
  - small box, per package: `go vet ./internal/watch/... ./tools/uilint/... && go test -p 1 -count=1 ./internal/watch/... ./tools/uilint/...`
  - whole repo on the worker host: `labrun <worker> -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./... && go test -race -count=1 ./internal/watch/...'` (a result under 20 s did not run)
  - `gofmt -l .` empty, `python3 -I tools/contracts/check.py`, `python3 -I tools/backlog/check.py`, `go run ./cmd/leakcheck -mode fail`, `tools/ci/check.sh`
- Then `git fetch origin && git merge origin/main`, rerun the per-package gate, tick the card boxes, fill `result:` (say not verified live), and `~/code/d2lod-ops/tools/mergeq.sh add agent/<name> --card TASK-<n>`.
- Report under 200 words: what is built, the commands run, what you looked at, what is U, last commit.
