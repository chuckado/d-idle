# Brief B4: Until-level for zone orders

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botuntil`, branch `agent/botuntil`, worktree `~/code/d2lod-wt/botuntil` off current `origin/main`.

**Goal.** Until-level for zone orders. `POST /api/orders` accepts `until: "level"` with a target level, as the order queue already does.

**Starts after:** nothing. **Can run next to:** everything except S5's edit of `http_answer.go` (two lines there; merge origin/main before queueing).

**Read first (three):** `internal/watch/orders.go`; `internal/watch/unlockorderqueue.go (`untilLevel`)`; `internal/watch/http_answer.go (`postOrder`)`.

**You own:** `internal/watch/orders.go`, `internal/watch/orders_untillevel_test.go`, `internal/watch/http_answer.go (`postOrder` only)`.

**Do not touch:** the queue sequencer; the charter flow.

**Already known (do not redo):** `postOrder` refuses it today with "until-level comes later"; the queue path has it.

## Do
1. `Order` gains `UntilLevel` (omitempty); `validateOrder` accepts it for `go` and `fight` (level above the hero's, at most the queue table's `maxUntilLevel`); the order text reads "until level N"; the order ends `done` when the hero reaches the level (reuse the queue's own check).
2. `postOrder`: `until: "level"` with `level: N`; the old error text for unknown values stays for anything else.

## Acceptance (also the card's boxes)
- Table tests for validate; an order ends at the level (hub test with the sim); the `/api/orders` reply keeps every old field (golden)
- The queue's until-level behaviour is unchanged

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botuntil: Until-level for zone orders" --card botuntil --owns internal/watch/orders.go --owns internal/watch/orders_untillevel_test.go --owns internal/watch/http_answer.go (`postOrder` only) --component watch --kind feature --size S --assignee credit-runner \
  --ac "Table tests for validate; an order ends at the level (hub test with the sim); the `/api/orders` reply keeps every old field (golden)" \
  --ac "The queue's until-level behaviour is unchanged"
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
