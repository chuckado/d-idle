# Brief B1: Pause and resume as player commands

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botpause`, branch `agent/botpause`, worktree `~/code/d2lod-wt/botpause` off current `origin/main`.

**Goal.** Pause and resume as player commands. Pause = walk to town, then stop (user answer 6); Resume continues the plan.

**Starts after:** nothing. **Can run next to:** everything.

**Read first (three):** `docs/ui/redesign-2026-11-14/C-orders.md (sections 3 and 5)`; `docs/inventory-actions.md (priority ladder)`; `internal/bot/loop.go (`selfPause`, `safeToPause`)`.

**You own:** `internal/bot/playerpause.go`, `internal/bot/playerpause_test.go`, `internal/watch/pauseapi.go`, `internal/watch/pauseapi_test.go`, `cmd/d2bot/pausewire.go`, `docs/pause.md`.

**Do not touch:** the safety and chicken rules; `commands.go` (S4 reads your desk through `verbs`).

**Already known (do not redo):** `bot.Runner.Pause()` and `safeToPause` exist; there is no watch route for them.

## Do
1. A player pause request: at the next safe moment (the shared `repairSafeMoment` rule) the hero goes home (`Town.ToTown`, as the errand trip), finishes nothing else, and the runner enters its paused state; in town already: pause at once. Safety interrupts still run while going home. A measured run (`D2BOT_RUN`) never installs it.
2. Resume clears it; player commands given while paused are kept and run after resume; the Why line gets a code `player_pause` (add the text to `whychange.json` only if you own it in this wave, else put the line in your report).
3. `GET /api/pause` (`watch.pause/1`: `state: running | going_home | paused`, `since`, `reason`) and `POST /api/pause {on}` on the answer listener, pattern of `repairapi.go`; state persisted so a restart stays paused.

## Acceptance (also the card's boxes)
- Fake-runner tests, Sorceress and Paladin: pause in the field goes home first and then stops; pause in town stops at once; safety interrupts a going-home pause; resume continues; a restart stays paused
- A pause never leaves the hero standing in a field (test: no paused state while outside town)
- Route tests: gates, 503 until bound, goldens
- Live checklist in `docs/pause.md`

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botpause: Pause and resume as player commands" --card botpause --owns internal/bot/playerpause.go --owns internal/bot/playerpause_test.go --owns internal/watch/pauseapi.go --owns internal/watch/pauseapi_test.go --owns cmd/d2bot/pausewire.go --owns docs/pause.md --component bot --kind feature --size M --assignee credit-runner \
  --ac "Fake-runner tests, Sorceress and Paladin: pause in the field goes home first and then stops; pause in town stops at once; safety interrupts a going-home pause; resume continues; a restart stays paused" \
  --ac "A pause never leaves the hero standing in a field (test: no paused state while outside town)" \
  --ac "Route tests: gates, 503 until bound, goldens" \
  --ac "Live checklist in `docs/pause.md`"
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
