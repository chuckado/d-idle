# Brief B9: Mercenary item verbs in the fork, and their bot wiring

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `forkmercverbs`, branch `agent/forkmercverbs`, worktree `~/code/d2lod-wt/forkmercverbs` off current `origin/main`.

**Goal.** The player can give an item to the mercenary and take one off him, through a fork verb that wraps the game's own mercenary item command, with his own requirement rules.

**Starts after:** B8 merged (the fast lane) and C1 merged (`usable.merc`). **Can run next to:** S slices.

**Read first (three):** `docs/research/merc-gear.md (sections 2.2, 2.3, 4)`; `docs/design/item-verbs.md ("Mercenary items" and sections 3 to 6)`; `docs/ui/redesign-2026-11-14/I-character-screen.md (section 9)`.

**You own:** `internal/d2dctl/item_verbs_merc.go`, `internal/d2dctl/item_verbs_merc_test.go`, `internal/invdesk/mercverbs.go`, `internal/invdesk/mercverbs_test.go`, `cmd/d2bot/invverbs.go (your lines)`, `docs/research/merc-gear.md (status lines)`.

**Do not touch:** the hero verbs; the mercenary rules in the fork beyond calling them (report the ethereal and requirements -x% gap, do not fix it here).

**Already known (do not redo):** There is no verb and no click executor for mercenary gear today; the game refuses a gear change on a dead mercenary; d2d's rule skips ethereal and requirements -x%.

## Do
1. Fork (private checkout, branch `fork-only/itemverbs` or a branch off it as the coordinator says; exclusive resources `fork-branch:<name>` and `d2d-lab-binary` on the card): verbs `item merc-equip <id>` and `item merc-unequip <slot>`, planned on a copy with the server's own mercenary slot rule and then sent as the mercenary item command, listed in `caps` under `--item-verbs`. Refusals use the existing codes plus `merctype` and `dead`. Unit test for at least two hireling types; the headless differential against the manual command sequence.
2. Go: client calls and stub-server tests; a Desk executor for `target: "merc"` requests (equip and unequip) on the fast lane of B8; the route stops answering code `merc` when the capability is present.
3. Without the capability nothing changes: the route keeps refusing with `merc` and the page keeps its "not built yet" line.
4. A live drill written as a checklist (hire, give a helm, take it off, a wrong weapon refused, a dead mercenary refused); not run.

## Acceptance (also the card's boxes)
- Fork tests green on the worker host; the fork commit is on the private remote and named in the card
- Go tests with a stub server for both verbs and each refusal code; Desk tests for two mercenary types
- `usable.merc` (C1) and the verb agree on the 21-row table's mercenary rows, or each disagreement is listed
- Compat goldens unchanged

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "forkmercverbs: Mercenary item verbs in the fork, and their bot wiring" --card forkmercverbs --owns internal/d2dctl/item_verbs_merc.go --owns internal/d2dctl/item_verbs_merc_test.go --owns internal/invdesk/mercverbs.go --owns internal/invdesk/mercverbs_test.go --owns cmd/d2bot/invverbs.go --owns docs/research/merc-gear.md --component d2d --kind feature --size L --assignee credit-runner \
  --ac "Fork tests green on the worker host; the fork commit is on the private remote and named in the card" \
  --ac "Go tests with a stub server for both verbs and each refusal code; Desk tests for two mercenary types" \
  --ac "`usable.merc` (C1) and the verb agree on the 21-row table's mercenary rows, or each disagreement is listed" \
  --ac "Compat goldens unchanged"
```

## Fork card (binding)
- Rules: `AGENTS.md` ("The private d2d fork") and `docs/agents/rules-short.md`. Work only in the private d2d fork checkout and on the branch named above; commit and push to the fork's private remote the same day (the one exception to never push); a doc may cite only a commit that is on that remote. No patch text goes into this repo. Nothing is posted to GitHub.
- No live game. The fork's own headless tests are the proof: the unit test for two hireling types and the headless differential against the manual command sequence, as the hero verbs have.
- The Go side (client, executor, wiring) is in this repo on `agent/<name>` and follows the bot-card gates: per package `go vet ./internal/d2dctl/... ./internal/invdesk/... && go test -p 1 -count=1 ./internal/d2dctl/... ./internal/invdesk/...`; whole repo `labrun <worker> -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./... && go test -race -count=1 ./internal/watch/...'` from inside the worktree (under 20 s did not run); `gofmt -l .`, `python3 -I tools/contracts/check.py`, `python3 -I tools/backlog/check.py`, `go run ./cmd/leakcheck -mode fail`, `tools/ci/check.sh`.
- Then `git fetch origin && git merge origin/main`, tick the card, fill `result:` (say not verified live, name the fork commit), `~/code/d2lod-ops/tools/mergeq.sh add agent/<name> --card TASK-<n>`.
- Report under 200 words: fork commit, tests run, what is U, the live checklist, last commit here.
