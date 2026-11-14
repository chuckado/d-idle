# Brief S4: Orders, read-only

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis4`, branch `agent/uis4`, worktree `~/code/d2lod-wt/uis4` off current `origin/main`.

**Goal.** Orders, read-only. `GET /api/commands` (`watch.commands/1`) joining the five command sources, and the panel `orders` that lists them with one state set and working Cancel.

**Starts after:** S1 merged. **Can run next to:** S2, S3, S6, the bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md`; `internal/watch/unlockorderqueue.go`; `internal/watch/repairapi.go`.

**You own:** `internal/watch/commands.go`, `internal/watch/commands_test.go`, `internal/watch/testdata/commands/`, `internal/watch/web/panels/orders/`.

**Do not touch:** `orders.go`, `http_answer.go`, `invapi.go`, `repairapi.go`, the `now`, `queue` and `activity` tiles. Register the route with one line next to `/api/orders`.

**Already known (do not redo):** Five queues exist today (A-audit.md section 3). The order queue is `GET /api/orderqueue`. `goals.Status.JourneyBlocked` exists in the bot but has no watch source yet (B6); leave `blocked[]` fed only by what the Hub has.

## Do
1. `GET /api/commands` as drafted in C.6: `rows[]` from the watch orders, the order queue entries, the repair view, the invdesk queue, the gold-farm desk and open campaign go-now requests; `blocked[]`; `plan`; `safety`; `paused`; `verbs[]`; `limits`. Row ids are namespaced (`order:12`, `queue:3`, `repair:1`, `inv:44`, `farmgold`, `journey:<chapter>`).
2. Map every source state to `queued | waiting | active | done | failed | cancelled` plus `phase` (table in C.3). Row order is the priority ladder: item actions, errands, repair, then zone orders; the row that runs at the next safe moment carries `next: true`.
3. Never hold a desk mutex while calling another desk (the TASK-278 rule): take each source's own snapshot, then join.
4. `verbs[]` lists only verbs whose desk is bound, each with `label`, `enabled`, `reason` (repair reads `trip.button`). The panel renders exactly this list, so a command without backend never shows.
5. Panel `orders` (band `side`, fit, `optIn`): the verb row rendered but inert in this slice except Cancel, the safety banner, the rows (index or state mark, text, state chip, reason line, progress when present, Cancel or Dismiss), the plan row last, the empty state of C.5 without suggestion chips. Cancel posts to the existing cancel route of the row's source.

## Acceptance (also the card's boxes)
- Handler goldens `testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json`
- A table test: every state of orders, queue entries, invdesk requests and repair trips maps to exactly one row state
- DOM goldens of `orders` for the five fixtures
- Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)
- A race test over concurrent desk updates and `/api/commands` reads passes under `-race`
- Under Standard the cell of `now` shows `orders`; under Classic nothing changed

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis4: Orders, read-only" --card uis4 --owns internal/watch/commands.go --owns internal/watch/commands_test.go --owns internal/watch/testdata/commands/ --owns internal/watch/web/panels/orders/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "Handler goldens `testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json`" \
  --ac "A table test: every state of orders, queue entries, invdesk requests and repair trips maps to exactly one row state" \
  --ac "DOM goldens of `orders` for the five fixtures" \
  --ac "Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)" \
  --ac "A race test over concurrent desk updates and `/api/commands` reads passes under `-race`" \
  --ac "Under Standard the cell of `now` shows `orders`; under Classic nothing changed"
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
