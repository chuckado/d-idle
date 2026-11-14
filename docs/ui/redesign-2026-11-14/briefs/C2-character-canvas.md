# Brief C2: The character canvas

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `charui`, branch `agent/charui`, worktree `~/code/d2lod-wt/charui` off current `origin/main`.

**Goal.** `/watch/inventory` becomes one canvas on one cell unit: the hero's and the mercenary's dolls in one frame with gear sets and stat sheets under them; the pack over the stash with the belt, the selected item and the loot rules beside the stash; a bar with search, filters and gold, repair and sell chips. It fits 1440x900, 1920x1080 and 1366x768 without scrolling.

**Starts after:** C1 merged; S1b merged for the fit check. **Can run next to:** S2 to S7, bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 4)`; `docs/ui/redesign-2026-11-14/mockups/v2/character.html`; `docs/ui/redesign-2026-11-14/I-character-screen.md (sections 5, 6, 9, 10)`.

**You own:** `internal/watch/web/inventory.js`, `internal/watch/web/inventory.css`, `internal/watch/web/inventory-merc.js`, `internal/watch/web/inventory-merc.css`, `internal/watch/web/inventory-sets.js`, `internal/watch/web/inventory-sets.css`, `internal/watch/invui.go`, `internal/watch/testdata/charscreen/`.

**Do not touch:** `inventory-tip.js`, `inventory-actions.js`, `inventory-drag.js` (C3 owns the tooltip, menu and gestures; keep their hooks), the routes and POST bodies.

**Already known (do not redo):** The old page is a masonry of cards with about a third of its area empty (research/README.md section 2). The mercenary view and fixtures exist.

## Do
1. One CSS unit `--u` from the viewport height (twelve storage rows plus chrome), capped by the width; dolls, pack, stash and belt are grids of that unit. Two zones that touch: who wears (takes the spare width) and storage (exactly ten cells wide).
2. Who wears: the hero's doll (8 by 7 cells) and the mercenary's (only the slots in `merc.slots`; no `n/a` box) side by side under one label row; his state chip, life bar, Str and Dex in his header; Revive there when dead (when the merc desk of B3 is bound). Under both: the gear set bar (chips with the existing preview on hover, a warn mark when an item is under its requirement). Under that: the two stat sheets as tiles of label over value (hero four columns, mercenary two): level, the four attributes, life, mana, defense, damage, the four resistances, xp to the next level, gear state; unmet or negative values in the danger colour. This replaces the Hero tile of the watch screen.
3. Storage: the pack (10 by 4); under it the stash (6 by 8) and a four-cell column: the belt, the selected item box (name, one line, its actions as buttons: the same list the context menu shows), the loot rules summary (a placeholder list the loot-rules slice fills), the locked line "Stash from anywhere", the last action line.
4. The bar: search, the four filter chips (dim, never hide), then gold, stash gold, `Repair <price>` (danger outline when something is broken) and `Sell marked (n)`. No Repair card, no gold line, no Item actions form, no legend line, no item panel beside the doll.
5. Requirement display from `req` and `usable` (C1) only: dotted, muted cell with a reason badge for an unusable item, `MERC` badge when only he can use it, green or red slot outlines with the reason text while an item is selected or dragged.
6. Narrow: under 700 px two tabs (Hero and mercenary; Pack and stash); the bar wraps.

## Acceptance (also the card's boxes)
- Fit check passes for the character page at 1440x900, 1920x1080 and 1366x768 (no scroll, no slot overflow, zones cover the frame)
- DOM goldens: hero only; each mercenary fixture (none, dead, desert, ironwolf, unknown); the 21 validity cases of I.10 from C1's goldens
- No rule in the page (a test greps the scripts for requirement arithmetic and class lists)
- Every action of the old page still works (old inventory DOM and logic tests pass or are ported, listed)
- Renders at the four sizes looked at and compared with `research/iterations/it3/character-*.png`

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "charui: The character canvas" --card charui --owns internal/watch/web/inventory.js --owns internal/watch/web/inventory.css --owns internal/watch/web/inventory-merc.js --owns internal/watch/web/inventory-merc.css --owns internal/watch/web/inventory-sets.js --owns internal/watch/web/inventory-sets.css --owns internal/watch/invui.go --owns internal/watch/testdata/charscreen/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "Fit check passes for the character page at 1440x900, 1920x1080 and 1366x768 (no scroll, no slot overflow, zones cover the frame)" \
  --ac "DOM goldens: hero only; each mercenary fixture (none, dead, desert, ironwolf, unknown); the 21 validity cases of I.10 from C1's goldens" \
  --ac "No rule in the page (a test greps the scripts for requirement arithmetic and class lists)" \
  --ac "Every action of the old page still works (old inventory DOM and logic tests pass or are ported, listed)" \
  --ac "Renders at the four sizes looked at and compared with `research/iterations/it3/character-*.png`"
```

## The fit rule (binding for every cockpit and canvas slice)
- The page must not scroll and no slot may overflow at 1440x900, 1920x1080 and 1366x768; at 390x844 it follows the phone design of J. `tools/uishots` gets (in S1b) a `fit` check that loads each page at those sizes in headless Chromium on the builder and fails when `document.scrollingElement.scrollHeight > innerHeight`, when any `[data-slot]` has `scrollHeight > clientHeight` (lists marked `data-scroll` excepted), or when the slots do not cover the frame (sum of slot areas plus gaps equals the frame, within 2 px). Run it (`labrun <builder> -- tools/uishots.sh --fit`, never on the worker host that runs the games) and put its output in the card result. A slice that fails the fit check is not done.
- A slot panel has no content-driven height: it renders inside the box it is given; long text is cut with an ellipsis and the full text goes to `title` or a popover; a list longer than its box scrolls inside (`data-scroll`).
- Look at your own renders at the four sizes and write three lines of critique in the card result (dead space, alignment, noise). Compare with `docs/ui/redesign-2026-11-14/research/iterations/it3/` (the accepted look).

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
