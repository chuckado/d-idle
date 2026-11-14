# Brief B8: Instant item actions: the fast lane

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `botinstant`, branch `agent/botinstant`, worktree `~/code/d2lod-wt/botinstant` off current `origin/main`.

**Goal.** A verb-capable item action the player gives runs at once through the item verbs instead of waiting for the bot's goal step and a safe moment; everything else stays an errand.

**Starts after:** nothing. **Can run next to:** S slices, C1, other bot cards (B9 builds on you).

**Read first (three):** `docs/ui/redesign-2026-11-14/I-character-screen.md (section 8)`; `docs/design/item-verbs.md (sections 5 to 8, 10)`; `docs/inventory-actions.md`.

**You own:** `internal/invdesk/instant.go`, `internal/invdesk/instant_test.go`, `internal/invdesk/desk.go (the request entry and the view's `instant` block only)`, `cmd/d2bot/invverbs.go`, `docs/inventory-actions.md (a new section)`.

**Do not touch:** the click path, the safe-moment predicate for non-verb actions, the knob's default, the friends release flags.

**Already known (do not redo):** The verbs, the Go client and the executors exist and are tested offline; the first release kept the safe moment so the two paths were comparable (item-verbs section 7). Whether a verb in a named run breaks the crutch rule is the user's decision (draft text in item-verbs section 10): keep the fast lane off in named runs.

## Do
1. A fast lane in the Desk: when a request's kind is verb-capable now (the backend advertises the verb in `caps`, `invdesk.verbs` is on, the hero is alive, no safety action holds the input, the bot's own click path does not hold the cursor, not a named run), execute it immediately through the existing `VerbEquipper` / `VerbMover` / `VerbTrasher`, serialised with the bot's own use of the control channel (one owner at a time; say in the doc how). The POST answers with the final request state (`done` with `how: verb`, or the refusal code and text) within a bounded wait, else `queued`.
2. Verb-capable in the field and in town: equip, unequip, pack and belt moves, identify with a scroll or tome, drop. Stash, take and gold: instant only when the hero is at the chest (the verb's own `place` rule), otherwise the request is queued for the next town visit with `wait: "town"` and runs on arrival without a new click. Sell, Cain, buy scrolls, repair, revive stay errands.
3. When the fast lane cannot run the request falls to the queue unchanged (`how: queue`). A refusal is final and uses the verb's code vocabulary; the read-back check after a verb stays (a verb that answered ok but did not move the item is reported failed).
4. The view gains `instant {equip, unequip, move, identify, drop, stash, take, gold}` (what would be instant right now) and requests gain `wait` (`""`, `safe`, `town`). Additive; schema number raised.
5. Requirements: the Desk refuses before sending when `usable` (C1, or today's candidate check) says no, with the requirement reason and numbers.

## Acceptance (also the card's boxes)
- Fake-backend tests, two classes: each verb-capable kind completes within the POST; each refusal code comes back as the request's reason; with the knob off, no caps, a named run, a busy cursor or a safety action the request is queued as before
- Stash and take: at the chest instant; in the field `wait: town`, then done on arrival; cancel while waiting
- A race test: the bot's own verb use and a player's instant request never interleave on the channel (`-race`)
- Old inventory goldens unchanged; new view goldens with the `instant` block
- The doc section says what stays an errand and why, and carries a live checklist

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "botinstant: Instant item actions: the fast lane" --card botinstant --owns internal/invdesk/instant.go --owns internal/invdesk/instant_test.go --owns internal/invdesk/desk.go --owns cmd/d2bot/invverbs.go --owns docs/inventory-actions.md --component bot --kind feature --size L --assignee credit-runner \
  --ac "Fake-backend tests, two classes: each verb-capable kind completes within the POST; each refusal code comes back as the request's reason; with the knob off, no caps, a named run, a busy cursor or a safety action the request is queued as before" \
  --ac "Stash and take: at the chest instant; in the field `wait: town`, then done on arrival; cancel while waiting" \
  --ac "A race test: the bot's own verb use and a player's instant request never interleave on the channel (`-race`)" \
  --ac "Old inventory goldens unchanged; new view goldens with the `instant` block" \
  --ac "The doc section says what stays an errand and why, and carries a live checklist"
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
