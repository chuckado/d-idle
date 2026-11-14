# Brief S5: Orders write path and the act map

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis5`, branch `agent/uis5`, worktree `~/code/d2lod-wt/uis5` off current `origin/main`.

**Goal.** Orders write path and the act map. `POST /api/commands` and `/api/commands/cancel` as a dispatcher over the existing actions, the live verb row, the panel `zone-map` with the zone menu, Undo, and the keys.

**Starts after:** S4 merged. **Can run next to:** S3, S6, S7 preparation; B7 may land before or after (the Loop control appears only when `/api/orderloop` says open).

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md`; `docs/ui/redesign-2026-11-14/mockups/watch-orders.html`; `internal/watch/web/app.js (the order sheet, lines around `drawSheet`)`.

**You own:** `internal/watch/commands.go`, `internal/watch/commands_post_test.go`, `internal/watch/web/panels/orders/`, `internal/watch/web/panels/zone-map/`, `internal/watch/testdata/zonemap/`.

**Do not touch:** `orders.go`, `unlockorderqueue.go`, `repairapi.go`, `invapi.go` (call their methods, do not copy them); `app.js` and the `orders-map` card stay working for Classic. In `http_answer.go` add only the two route-table lines.

**Already known (do not redo):** The dispatcher is a second door to the same actions: one behaviour, shared code. The queue is an unlock (user answer 5), so `mode: queue` must respect the ladder.

## Do
1. `POST /api/commands {intent, args, mode}` on the answer listener with the existing gates. It validates and calls the same Hub and desk methods the old routes call: `go`, `fight_here`, `town` -> `Hub.place`; `mode: "queue"` -> the order queue's add (answers its `locked` code while the slot rows are closed); `repair` -> the repair desk; `farm_gold`; `identify`, `stash`, `take`, `gold`, `gear_set` -> the invdesk routes' methods; `journey` -> the campaign desk; `loop` -> the loop controller when bound, else 503 `off`. `POST /api/commands/cancel {id}` by namespaced id. Error codes are the existing ones.
2. Verb row live: Go to..., Town, Repair (label from `verbs`), Pause (only when the pause desk is bound, B1), More. Every accepted command shows the Undo toast for five seconds (Undo = cancel of the new row).
3. Panel `zone-map` (band `grid`, `optIn`) from `/api/orders/map`: nodes with the level-band stripe, here ring, target ring, numbered pins for waiting orders, the route line, a failed mark; the legend in a popover; the list view kept as the accessible alternative. Nodes are buttons; arrow keys walk the graph.
4. Zone menu (C.4): field, here, town, too hard, not orderable, charter. Duration chips incl. until level only when the order route accepts it (B4). The charter is paid inline (no `window.confirm`). Loop and Add to queue follow the unlock state: open -> controls; locked -> the ladder teaser as one muted line, never a dead button (C.8).
5. Keys `g t r p o z` and `Esc` as in C.7; off in text fields and in overlay mode.
6. Update `web/layouts/simple.json`: hide the `zone-map` tile there now that Go to opens it as a sheet (ask S1's owner if S1 is still open; otherwise it is yours for this one file).

## Acceptance (also the card's boxes)
- Answer-route tests for each intent: origin, token, JSON, unknown fields, and the same resulting state as the old route (compare with the old route's own test fixture)
- DOM goldens of the zone menu for field, here, town, hard, locked, charter, loop locked, loop open, queue locked, queue full
- From a zone click to an accepted order is two clicks (node test of the flow)
- The old `orders-map` card still places an order under Classic
- No confirmation dialog except trash, charter, revoke all and the inline too-hard step
- Pictures at 390 (sheet) and 1440 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis5: Orders write path and the act map" --card uis5 --owns internal/watch/commands.go --owns internal/watch/commands_post_test.go --owns internal/watch/web/panels/orders/ --owns internal/watch/web/panels/zone-map/ --owns internal/watch/testdata/zonemap/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "Answer-route tests for each intent: origin, token, JSON, unknown fields, and the same resulting state as the old route (compare with the old route's own test fixture)" \
  --ac "DOM goldens of the zone menu for field, here, town, hard, locked, charter, loop locked, loop open, queue locked, queue full" \
  --ac "From a zone click to an accepted order is two clicks (node test of the flow)" \
  --ac "The old `orders-map` card still places an order under Classic" \
  --ac "No confirmation dialog except trash, charter, revoke all and the inline too-hard step" \
  --ac "Pictures at 390 (sheet) and 1440 looked at"
```

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
