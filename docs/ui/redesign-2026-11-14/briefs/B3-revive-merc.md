# Brief B3: Revive mercenary as a player command

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botmerc`, branch `agent/botmerc`, worktree `~/code/d2lod-wt/botmerc` off current `origin/main`.

**Goal.** Revive mercenary as a player command. A request that makes the next safe moment a town trip to revive the dead mercenary when the gold is there.

**Starts after:** nothing. **Can run next to:** everything (B5 reads the same merc state: agree on the file split with its owner through the coordinator).

**Read first (three):** `docs/campaign.md (section 6a)`; `docs/inventory-actions.md (priority ladder)`; `internal/merc/ (the revive policy)`.

**You own:** `internal/merc/playerrevive.go`, `internal/merc/playerrevive_test.go`, `internal/watch/mercapi.go`, `internal/watch/mercapi_test.go`, `cmd/d2bot/mercwire.go`, `docs/merc-revive.md`.

**Do not touch:** the merc policy's own thresholds; mercenary gear (not possible yet).

**Already known (do not redo):** The policy already revives in town when gold allows; the live hero showed `merc: revive skipped (gold) price=600 gold=46` (docs/campaign.md 6a).

## Do
1. A player request: `GET /api/merc` (`watch.merc/1`: `state: alive | dead | none`, `price`, `gold`, `button {label, enabled, reason}`, the open request) and `POST /api/merc/revive`, `/cancel`. Button rules like the repair button: off with a reason when alive, none hired, or gold short ("need N more gold").
2. With enough gold the bot goes home at the next safe moment (errand trip) and the existing revive step runs first in the town plan; one request at a time; a failed revive keeps the reason and a cooldown like the repair trip.

## Acceptance (also the card's boxes)
- Fake tests for two classes: dead + enough gold -> trip and revive; gold short -> button off with the amount; alive -> off; failure -> reason and cooldown
- Route tests with goldens; 503 until bound
- Live checklist in the doc

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botmerc: Revive mercenary as a player command" --card botmerc --owns internal/merc/playerrevive.go --owns internal/merc/playerrevive_test.go --owns internal/watch/mercapi.go --owns internal/watch/mercapi_test.go --owns cmd/d2bot/mercwire.go --owns docs/merc-revive.md --component bot --kind feature --size M --assignee credit-runner \
  --ac "Fake tests for two classes: dead + enough gold -> trip and revive; gold short -> button off with the amount; alive -> off; failure -> reason and cooldown" \
  --ac "Route tests with goldens; 503 until bound" \
  --ac "Live checklist in the doc"
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
