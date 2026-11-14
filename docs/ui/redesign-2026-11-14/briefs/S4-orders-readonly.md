# Brief S4: Orders, read-only, as a slot panel

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis4`, branch `agent/uis4`, worktree `~/code/d2lod-wt/uis4` off current `origin/main`.

**Goal.** `GET /api/commands` (`watch.commands/1`) joining the command sources, and the slot panel `orders`: one-line rows, one state set, working Cancel.

**Starts after:** S1b merged. **Can run next to:** S2, S3, S6, the bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md (sections 3, 5, 6, 8)`; `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 2)`; `internal/watch/unlockorderqueue.go`.

**You own:** `internal/watch/commands.go`, `internal/watch/commands_test.go`, `internal/watch/testdata/commands/`, `internal/watch/web/panels/orders/`.

**Do not touch:** `orders.go`, `http_answer.go`, `invapi.go`, `repairapi.go`, the `now`, `queue`, `activity` tiles. Register the route with one line next to `/api/orders`.

**Already known (do not redo):** Five queues exist today. The order queue is `GET /api/orderqueue`; the loop is B7's `GET /api/orderloop`.

## Do
1. `GET /api/commands` as drafted in C.6: `rows[]` from the watch orders, the order queue entries, the loop (when its desk is bound), the repair view, the invdesk queue, the gold-farm desk, open campaign go-now requests; `blocked[]`; `plan`; `safety`; `paused`; `verbs[]` (only verbs whose desk is bound, with `label`, `enabled`, `reason`); `limits`. Namespaced row ids.
2. Map every source state to `queued | waiting | active | done | failed | cancelled` plus a `reason` line; row order is the priority ladder; the row that runs at the next safe moment carries `next: true`.
3. Never hold a desk mutex while calling another desk: snapshot each source, then join.
4. Panel `orders` (slot `rail`): the verb row (rendered from `verbs[]`, inert in this slice), the safety banner as one line, rows of 28 px: index or state mark, text, the reason in muted text cut with an ellipsis, a state chip, a cancel cross (posting to the row source's existing cancel route); the plan row last. The list scrolls inside the panel (`data-scroll`) when it is longer than its box; the panel's height is its content up to 45 percent of the rail.

## Acceptance (also the card's boxes)
- Handler goldens `testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json`
- A table test: every source state maps to exactly one row state
- DOM goldens for the five fixtures; with eight rows the list scrolls inside and the fit check still passes
- Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)
- A race test over concurrent desk updates and reads passes under `-race`

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis4: Orders, read-only, as a slot panel" --card uis4 --owns internal/watch/commands.go --owns internal/watch/commands_test.go --owns internal/watch/testdata/commands/ --owns internal/watch/web/panels/orders/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "Handler goldens `testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json`" \
  --ac "A table test: every source state maps to exactly one row state" \
  --ac "DOM goldens for the five fixtures; with eight rows the list scrolls inside and the fit check still passes" \
  --ac "Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)" \
  --ac "A race test over concurrent desk updates and reads passes under `-race`"
```

## The fit rule (binding for every cockpit and canvas slice)
- The page must not scroll and no slot may overflow at 1440x900, 1920x1080 and 1366x768; at 390x844 it follows the phone design of J. `tools/uishots` gets (in S1b) a `fit` check that loads each page at those sizes in headless Chromium on the builder and fails when `document.scrollingElement.scrollHeight > innerHeight`, when any `[data-slot]` has `scrollHeight > clientHeight` (lists marked `data-scroll` excepted), or when the slots do not cover the frame (sum of slot areas plus gaps equals the frame, within 2 px). Run it (`labrun <builder> -- tools/uishots.sh --fit`, never on the worker host that runs the games) and put its output in the card result. A slice that fails the fit check is not done.
- A slot panel has no content-driven height: it renders inside the box it is given; long text is cut with an ellipsis and the full text goes to `title` or a popover; a list longer than its box scrolls inside (`data-scroll`).
- Look at your own renders at the four sizes and write three lines of critique in the card result (dead space, alignment, noise). Compare with `docs/ui/redesign-2026-11-14/research/iterations/it3/` (the accepted look).

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
