# Brief B6: Journey blocked on the watch side

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botblocked`, branch `agent/botblocked`, worktree `~/code/d2lod-wt/botblocked` off current `origin/main`.

**Goal.** Journey blocked on the watch side. The bot's `JourneyBlocked` reason reaches `/api/glance` (a Needs you row) and `/api/commands` (`blocked[]`).

**Starts after:** S2 merged; S4 for the second half. **Can run next to:** other bot cards.

**Read first (three):** `internal/bot/staletarget.go`; `internal/goals/runner.go (`Status.JourneyBlocked`)`; `internal/watch/bridge.go`.

**You own:** `internal/watch/bridge.go (the observe call only)`, `cmd/d2bot/glancewire_blocked.go`, `internal/watch/testdata/glance/ (new files)`.

**Do not touch:** `staletarget.go` behaviour.

**Already known (do not redo):** The Why code `journey_blocked` exists (`internal/bot/whyline.go`).

## Do
1. Feed `goals.Status.JourneyBlocked` through the bridge into the Hub source S2 declared; the row text is "Journey blocked: <reason>" with the actions Go now and Hold for the chapter in the campaign view's decision.
2. Empty string clears the row.

## Acceptance (also the card's boxes)
- Bridge test: a status with the reason yields the row and the `blocked[]` entry; clearing removes both; both classes
- New goldens only

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botblocked: Journey blocked on the watch side" --card botblocked --owns internal/watch/bridge.go (the observe call only) --owns cmd/d2bot/glancewire_blocked.go --owns internal/watch/testdata/glance/ (new files) --component watch --kind feature --size S --assignee credit-runner \
  --ac "Bridge test: a status with the reason yields the row and the `blocked[]` entry; clearing removes both; both classes" \
  --ac "New goldens only"
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
