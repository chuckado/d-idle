# Brief B2: Sell now as a player command

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botsell`, branch `agent/botsell`, worktree `~/code/d2lod-wt/botsell` off current `origin/main`.

**Goal.** Sell now as a player command. An errand that walks home at the next safe moment and sells the items the player marked.

**Starts after:** nothing. **Can run next to:** everything.

**Read first (three):** `docs/inventory-actions.md`; `internal/invdesk/stash.go`; `internal/bot/playeraction.go`.

**You own:** `internal/invdesk/sellnow.go`, `internal/invdesk/sellnow_test.go`, `internal/watch/invapi_sellnow.go`, `internal/watch/invapi_sellnow_test.go`, `cmd/d2bot/invwire.go (your lines only)`, `docs/inventory-actions.md (a new section)`.

**Do not touch:** the sell policy and pickit; no new decision about what to sell: only items the player marked `sell`.

**Already known (do not redo):** Marked items already sell on the next town visit (`sellPending`); only the trigger is missing.

## Do
1. invdesk kind `sell` (an errand like `stash`): queued -> doing -> done or failed; with nothing marked it fails at once with "nothing is marked for sale".
2. The errand trip (`errandTripStep`) takes it; in town the existing sale of marked items runs; the request ends when the sale step ran.
3. `POST /api/inventory/sellnow` on the answer listener (schema bump of `watch.inventory`, additive).

## Acceptance (also the card's boxes)
- Fake tests for both classes: in the field it goes home at a safe moment and sells the marked items; in town it sells at once; nothing marked fails with the reason
- The queue limit (8) and cancel from `queued` hold
- Route test with goldens; old inventory goldens unchanged
- Live checklist added to the doc

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botsell: Sell now as a player command" --card botsell --owns internal/invdesk/sellnow.go --owns internal/invdesk/sellnow_test.go --owns internal/watch/invapi_sellnow.go --owns internal/watch/invapi_sellnow_test.go --owns cmd/d2bot/invwire.go (your lines only) --owns docs/inventory-actions.md (a new section) --component bot --kind feature --size S --assignee credit-runner \
  --ac "Fake tests for both classes: in the field it goes home at a safe moment and sells the marked items; in town it sells at once; nothing marked fails with the reason" \
  --ac "The queue limit (8) and cancel from `queued` hold" \
  --ac "Route test with goldens; old inventory goldens unchanged" \
  --ac "Live checklist added to the doc"
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
