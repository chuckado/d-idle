# Brief C3: The tooltip, the context menu, gestures and instant feedback

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `chargest`, branch `agent/chargest`, worktree `~/code/d2lod-wt/chargest` off current `origin/main`.

**Goal.** A real floating tooltip with requirements and comparison, a right-click menu and keys for every action, drag with validity preview, and the pending, done, refused and queued-for-town states.

**Starts after:** C2 merged (B8 for the instant states). **Can run next to:** S slices, bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 4)`; `docs/ui/redesign-2026-11-14/mockups/v2/character.html`; `docs/ui/redesign-2026-11-14/I-character-screen.md (sections 6 to 8)`.

**You own:** `internal/watch/web/inventory-tip.js`, `internal/watch/web/inventory-tip.css`, `internal/watch/web/inventory-actions.js`, `internal/watch/web/inventory-actions.css`, `internal/watch/web/inventory-drag.js`, `internal/watch/testdata/chargest/`.

**Do not touch:** the canvas layout (C2), the Desk and routes (B8).

**Already known (do not redo):** The page had three places that explained an item (hover, a panel beside the doll, a legend); this leaves one.

## Do
1. Tooltip: fixed width (about 270 px), anchored beside the hovered or focused cell, flipped at the screen edge, never scrolling, never a card in the layout. Content: name in its quality colour, the game's lines, where it is (worn by you, worn by the mercenary, pack, stash; queued moves), requirement rows for the hero and, for kinds a mercenary wears, for him (met and unmet marks, the gap), the comparison against the worn item of the slot as old to new values with signed deltas (both rings, both hands), one hint line. On touch the same content is a bottom sheet.
2. Context menu on right-click and on the menu key: every action with its key; an action that would be refused is listed greyed with the reason. Double-click or Enter runs the first; Ctrl-click pack and stash; Alt-click mark for sale; Delete drops after one inline confirmation.
3. Drag: footprint preview in pack and stash, slot outlines from the view; a refused drop sends nothing.
4. Feedback from `view.instant` and the reply: pending (dashed outline at the target, never the done look), done, refused (the item returns; the reason at the item and in the last action line; the verb's own code), queued (`how: queue`), queued for town (`TOWN` badge, Cancel).

## Acceptance (also the card's boxes)
- DOM goldens of the tooltip for: no requirements, each unmet kind, class item, mercenary block, comparison with deltas, both rings; of the menu per place; of each feedback state per action kind
- The tooltip never leaves the viewport and never overlaps its own cell at the four sizes (node test of the placement function)
- A refused requirement never renders as done (node test)
- The existing drag goldens pass; pictures at 1440 and 390 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "chargest: The tooltip, the context menu, gestures and instant feedback" --card chargest --owns internal/watch/web/inventory-tip.js --owns internal/watch/web/inventory-tip.css --owns internal/watch/web/inventory-actions.js --owns internal/watch/web/inventory-actions.css --owns internal/watch/web/inventory-drag.js --owns internal/watch/testdata/chargest/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens of the tooltip for: no requirements, each unmet kind, class item, mercenary block, comparison with deltas, both rings; of the menu per place; of each feedback state per action kind" \
  --ac "The tooltip never leaves the viewport and never overlaps its own cell at the four sizes (node test of the placement function)" \
  --ac "A refused requirement never renders as done (node test)" \
  --ac "The existing drag goldens pass; pictures at 1440 and 390 looked at"
```

## The fit rule (binding for every cockpit and canvas slice)
- The page must not scroll and nothing may clip at the full-screen sizes 1920x1080, 1440x900, 1366x768 and at the browser-window sizes 1914x930, 1904x1000, 1536x730, 1280x620 (a window with its chrome is shorter than the screen: the user's first look was at 1914x930); at 390x844 it follows the phone design of J. `tools/uishots` gets (in S1b) a `fit` check that loads each page at those sizes in headless Chromium on the builder and fails when `document.scrollingElement.scrollHeight > innerHeight`, when any `[data-slot]` has `scrollHeight > clientHeight` (lists marked `data-scroll` excepted), or when the slots do not cover the frame (sum of slot areas plus gaps equals the frame, within 2 px). Run it (`labrun <builder> -- tools/uishots.sh --fit`, never on the worker host that runs the games) and put its output in the card result. A slice that fails the fit check is not done.
- A slot panel has no content-driven height: it renders inside the box it is given; long text is cut with an ellipsis and the full text goes to `title` or a popover; a list longer than its box scrolls inside (`data-scroll`).
- Look at your own renders at those sizes and write three lines of critique in the card result (dead space, alignment, noise). Compare with `docs/ui/redesign-2026-11-14/research/iterations/it3/` (the accepted look).

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
