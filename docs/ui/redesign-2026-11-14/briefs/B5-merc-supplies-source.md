# Brief B5: Mercenary and supplies for the Health lights

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botlights`, branch `agent/botlights`, worktree `~/code/d2lod-wt/botlights` off current `origin/main`.

**Goal.** Mercenary and supplies for the Health lights. Hub sources so `/api/glance` reports them instead of `unknown`.

**Starts after:** S2 merged (the setters exist). **Can run next to:** S3 to S6, other bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/B-principles-ia.md (Health lights)`; `internal/watch/glance.go`; `cmd/d2bot/readywire.go (the pattern)`.

**You own:** `cmd/d2bot/glancewire.go`, `cmd/d2bot/glancewire_test.go`.

**Do not touch:** `glance.go` beyond what the setters need (ask S2's owner or the coordinator).

**Already known (do not redo):** The belt and tome counts are in the snapshot and on `/api/inventory`.

## Do
1. Mercenary: alive, dead or none from the runner's newest snapshot (read-only).
2. Supplies: healing and mana potions on the belt and town portal scrolls against the kit plan's numbers; light green at plan, amber below, red at none. The plan numbers come from the kit data of the hero's class, never a literal.

## Acceptance (also the card's boxes)
- Tests for two classes with fake snapshots: each light value; nil snapshot -> unknown
- The glance goldens gain new files only

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botlights: Mercenary and supplies for the Health lights" --card botlights --owns cmd/d2bot/glancewire.go --owns cmd/d2bot/glancewire_test.go --component bot --kind feature --size S --assignee credit-runner \
  --ac "Tests for two classes with fake snapshots: each light value; nil snapshot -> unknown" \
  --ac "The glance goldens gain new files only"
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
