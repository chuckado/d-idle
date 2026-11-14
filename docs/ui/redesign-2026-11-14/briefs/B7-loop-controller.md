# Brief B7: The Loop order and its unlock

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botloop`, branch `agent/botloop`, worktree `~/code/d2lod-wt/botloop` off current `origin/main`.

**Goal.** The Loop order and its unlock. A loop controller (go to an area for X minutes, home, repair, repeat) and the ladder rows that open the Loop and then the queue.

**Starts after:** nothing. **Can run next to:** S1 to S4, B1 to B4 (you are the only owner of `data/unlocks/ladder.json` in this wave).

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md (section 8)`; `internal/watch/unlockorderqueue.go`; `docs/repair.md (the repair trip)`.

**You own:** `internal/watch/orderloop.go`, `internal/watch/orderloop_test.go`, `internal/watch/testdata/orderloop/`, `data/unlocks/ladder.json`, `internal/watch/orderqueue.json`, `cmd/d2bot/loopwire.go`, `docs/unlock-orderloop.md`.

**Do not touch:** `unlockorderqueue.go` logic (read it as the pattern), the repair controller, `commands.go`.

**Already known (do not redo):** Decision to flag in your report, not to make: with the strict order, a hero who cleared the Den of Evil no longer gets the first queue slot from it. The coordinator may ask you to keep the Den as an `alt` trigger (one line of data).

## Do
1. `OrderLoop` in `internal/watch`, the pattern of `OrderQueue`: fields `area`, `minutes` (5 to 240), `repair` (`damaged` | `always`), `rounds` (1 to 20) or `forever`. A round places one `go` order through `Hub.place`; when it ends `done`, it asks the repair desk for a repair when `repair` says so (an `undamaged` answer counts as done) or places `home`; when that ends, the next round. State in `<state dir>/orderloop.json`; a restart resumes at the start of the round.
2. Stop conditions with a reason: rounds done; player cancel; a player zone order with replace; death; the guard refuses the area; the repair failed or was cancelled by the bot; the `Halt` hook (Pact, low-life pause). A finite loop holds the running place and queue entries wait; `forever` with waiting entries is refused (`queue-busy`), both ways.
3. `GET /api/orderloop` (`watch.orderloop/1`: `open`, `teaser`, `state`, `round`, `rounds`, `area`, `minutes`, `repair`, `stoppedWhy`) and `POST /api/orderloop {action: start | stop, ...}` on the answer listener; 409 `locked` while the ladder row is closed.
4. Ladder: new row `orderloop` (group `ui`, trigger `seen` / `manual_repair_cycle`, texts in C.8). Emit that event when a player-requested repair ends `done` and a player zone order was given before it. Emit `loop_rounds_3` after the third finished round and make it the trigger of `orderqueue` (the strict order of the user's answer 5: single orders, then Loop, then queue); `orderqueue2` and `orderqueue3` unchanged. A row already open for a hero stays open: prove it with a test or report that it does not.
5. The loop is one row for `/api/commands` (S4 reads your view): text "Loop: <area> N min, repair, repeat (round r of R)".

## Acceptance (also the card's boxes)
- Sim tests for two classes: three rounds run go -> repair -> go; `damaged` skips the smith when nothing is damaged; each stop condition ends the loop with its reason; restart resumes the round
- Ladder tests: `orderloop` opens on the event; `orderqueue` opens after three rounds; a hero with `orderqueue` already open keeps it
- Route tests with goldens (`testdata/orderloop/view-*.json`), 409 while locked
- `docs/unlock-orderloop.md` with the tiers, the spec and a live checklist

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botloop: The Loop order and its unlock" --card botloop --owns internal/watch/orderloop.go --owns internal/watch/orderloop_test.go --owns internal/watch/testdata/orderloop/ --owns data/unlocks/ladder.json --owns internal/watch/orderqueue.json --owns cmd/d2bot/loopwire.go --owns docs/unlock-orderloop.md --component watch --kind feature --size L --assignee credit-runner \
  --ac "Sim tests for two classes: three rounds run go -> repair -> go; `damaged` skips the smith when nothing is damaged; each stop condition ends the loop with its reason; restart resumes the round" \
  --ac "Ladder tests: `orderloop` opens on the event; `orderqueue` opens after three rounds; a hero with `orderqueue` already open keeps it" \
  --ac "Route tests with goldens (`testdata/orderloop/view-*.json`), 409 while locked" \
  --ac "`docs/unlock-orderloop.md` with the tiers, the spec and a live checklist"
```

## Every bot card (binding)
- Rules: `AGENTS.md`, `docs/agents/rules-short.md`. Own worktree and branch, commit often, never push, no deploy, offline only (fake runner, `mem.Fake`, fixtures); write the live check as a checklist in the card.
- Builds are data: no skill id, class check or build literal in Go; every new rule is tested for the Sorceress and one other class.
- Priority ladder of `docs/inventory-actions.md` holds: safety first, then player commands, then chores, then roam. A player command never bypasses a safety rule. No gear decision by the bot (`docs/gear-policy.md`).
- A new watch route follows `internal/watch/repairapi.go`: a late-bound desk, GET on the viewer, POST on the answer listener with the same gates (origin list, token, JSON, small body, unknown fields refused), 503 `off` until bound, a `schema` field in the reply, goldens for both classes.
- Gates before queueing, from inside your worktree with no `cd` in the command:
  - small box, per package: `go vet ./<your packages>/... && go test -p 1 -count=1 ./<your packages>/...`; packages that need the game tables: `D2_MPQ_DIR=$HOME/d2d-lab/data go test -p 1 -count=1 ./<pkg>/...`
  - whole repo on the worker host: `labrun <worker> -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./... && go test -race -count=1 ./internal/watch/...'` (a result under 20 s did not run)
  - `gofmt -l .` empty, `python3 -I tools/contracts/check.py`, `python3 -I tools/backlog/check.py`, `go run ./cmd/leakcheck -mode fail`, `tools/ci/check.sh`
- Then `git fetch origin && git merge origin/main`, rerun the per-package gate, tick the card boxes, fill `result:` (say not verified live), and `~/code/d2lod-ops/tools/mergeq.sh add agent/<name> --card TASK-<n>`.
- Report under 200 words: what is built, the commands run, what is U, the live checklist, last commit.
