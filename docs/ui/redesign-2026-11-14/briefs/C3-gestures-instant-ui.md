# Brief C3: Gestures and the instant-action feedback

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `chargest`, branch `agent/chargest`, worktree `~/code/d2lod-wt/chargest` off current `origin/main`.

**Goal.** Drag with validity preview, double-click, ctrl-click, alt-click, and the pending, done, refused and queued-for-town states of an item action on the character screen.

**Starts after:** C2 merged; B8 merged for the instant states (before B8 every action shows the queued state). **Can run next to:** S slices, B cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/I-character-screen.md (sections 7 and 8)`; `docs/gearsets.md (drag and drop)`; `docs/design/item-verbs.md (section 5, the codes)`.

**You own:** `internal/watch/web/inventory-actions.js`, `internal/watch/web/inventory-actions.css`, `internal/watch/web/inventory-drag.js`, `internal/watch/testdata/chargest/`.

**Do not touch:** the canvas layout (C2), the Desk and routes (B8).

**Already known (do not redo):** The drag layer exists for the doll (`inventory-drag.js`, TASK-380). `request.how` exists.

## Do
1. Drag: while dragging, pack and stash cells preview the footprint green or red, doll and mercenary slots outline from `canWear` / `noWearSlot` / `usable`; a refused drop sends nothing and shows the reason at the slot.
2. Double-click or Enter = the primary action by place (I.7); ctrl-click pack to stash and back; alt-click mark for sale; Delete = Drop with the one inline confirmation. Every gesture has its inspector button; touch uses buttons only.
3. Feedback from `view.instant` and the reply: pending (ring, item drawn at the target, never the done look), done, refused (item returns, reason at the item and in the strip, the verb's own code), queued (`how: queue`: the item stays, a clock badge, the strip says what it waits for). Stash and take in the field: the `TOWN` badge, a ghost at the target, Cancel in the strip.
4. Optimistic state is reconciled with the next view by request id; a view that disagrees wins.

## Acceptance (also the card's boxes)
- DOM goldens for each state of each action kind (pending, done, refused with each code of item-verbs section 5, queued, queued for town)
- A refused requirement never renders as done (node test: with `usable.hero.ok` false no request is sent; with a server refusal the item returns)
- Drag tests: the existing drag goldens pass; new ones for stash cells and mercenary slots
- Pictures at 1440 and 390 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "chargest: Gestures and the instant-action feedback" --card chargest --owns internal/watch/web/inventory-actions.js --owns internal/watch/web/inventory-actions.css --owns internal/watch/web/inventory-drag.js --owns internal/watch/testdata/chargest/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens for each state of each action kind (pending, done, refused with each code of item-verbs section 5, queued, queued for town)" \
  --ac "A refused requirement never renders as done (node test: with `usable.hero.ok` false no request is sent; with a server refusal the item returns)" \
  --ac "Drag tests: the existing drag goldens pass; new ones for stash cells and mercenary slots" \
  --ac "Pictures at 1440 and 390 looked at"
```

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
