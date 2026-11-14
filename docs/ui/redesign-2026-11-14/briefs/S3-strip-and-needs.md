# Brief S3: Needs you and the strip: Progress, Journey, Health

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis3`, branch `agent/uis3`, worktree `~/code/d2lod-wt/uis3` off current `origin/main`.

**Goal.** The slot panels `needs-you`, `progress`, `journey-strip` and `health`, all reading `/api/glance`, each inside a fixed box.

**Starts after:** S2 merged. **Can run next to:** S4, S5, S6.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 2)`; `docs/ui/redesign-2026-11-14/mockups/v2/watch.html`; `docs/ui/redesign-2026-11-14/B-principles-ia.md`.

**You own:** `internal/watch/web/panels/needs-you/`, `internal/watch/web/panels/progress/`, `internal/watch/web/panels/journey-strip/`, `internal/watch/web/panels/health/`, `internal/watch/testdata/glancepanels/`.

**Do not touch:** `glance.go` (ask for a field), the old `gear`, `tally`, `why`, `journey` panels (they stay for the free-grid layouts).

**Already known (do not redo):** Mercenary and Supplies stay `unknown` until B5. The mockup shows the wording and sizes.

## Do
1. `needs-you` (slot `rail`, first): zero to three one-line rows of 32 px: a shape by severity, the text and a muted sub text cut with an ellipsis, one or two buttons from `actions[]` posted through `ctx.post`; "+N more" as a fourth line that opens the full list in a popover. With no rows the panel takes no height (the map below grows).
2. `progress` (slot `strip`): the level meter with "N%, about M min", xp/h with a 30-point sparkline, gold/h. A rate still building shows a dash with a tooltip. A click opens a popover with the tally and marks.
3. `journey-strip` (slot `strip`): the chapters as a dotted path (done, current, to come), the next chapter with its band in one line, and the two buttons the reply names (Go now, Hold); a click on a dot opens the existing Journey panel's detail in a popover. The existing `journey` panel is not edited.
4. `health` (slot `strip`): five lights with a label and a word; `unknown` is a hollow disc with "not reported"; a click opens the detail (gear chips, resistances, belt) in a popover.
5. Notifications (user answer 11): a switch in the `needs-you` menu, off by default, desktop notification for a new red row only.
6. All four use the look of J.5: 10 px small caps label, 12.5 px body, no border.

## Acceptance (also the card's boxes)
- DOM goldens: `needs-you` empty, one red, red + amber, five rows (overflow line); `progress` measuring and full; `journey-strip` for chapter 1, 3, 9 and blocked; `health` all ok, gear broken, two unknown
- A button posts exactly the `post` and `body` the reply names (node test)
- At 1366x768, 1440x900 and 1920x1080 the strip is one row of three cells with no clipped text; the fit check passes with zero, two and five needs rows
- No number without a unit or label

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis3: Needs you and the strip: Progress, Journey, Health" --card uis3 --owns internal/watch/web/panels/needs-you/ --owns internal/watch/web/panels/progress/ --owns internal/watch/web/panels/journey-strip/ --owns internal/watch/web/panels/health/ --owns internal/watch/testdata/glancepanels/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens: `needs-you` empty, one red, red + amber, five rows (overflow line); `progress` measuring and full; `journey-strip` for chapter 1, 3, 9 and blocked; `health` all ok, gear broken, two unknown" \
  --ac "A button posts exactly the `post` and `body` the reply names (node test)" \
  --ac "At 1366x768, 1440x900 and 1920x1080 the strip is one row of three cells with no clipped text; the fit check passes with zero, two and five needs rows" \
  --ac "No number without a unit or label"
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
