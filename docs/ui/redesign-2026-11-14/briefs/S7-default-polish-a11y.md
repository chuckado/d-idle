# Brief S7: The cockpit becomes the default; polish and access

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis7`, branch `agent/uis7`, worktree `~/code/d2lod-wt/uis7` off current `origin/main`.

**Goal.** With every cockpit panel merged, `web/layout.json` becomes the cockpit for everyone without a stored layout; then the shared look, contrast, focus, touch targets, reduced motion, the light theme and the command palette.

**Starts after:** S3, S5 and S6 merged. **Can run next to:** C3, L1.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md`; `docs/ui/redesign-2026-11-14/E-visual-system.md`; `docs/ui/redesign-2026-11-14/research/README.md (the critique list)`.

**You own:** `internal/watch/web/layout.json`, `internal/watch/web/layouts/`, `internal/watch/web/components.css`, `internal/watch/web/panels/{statusline,needs-count,needs-you,orders,zone-map,area-map,progress,journey-strip,health,log}/panel.css`, `tools/uishots/layout-rows.json`, `tools/uishots/genlayouts.js`.

**Do not touch:** panel behaviour (report a bug instead); stored layouts.

**Already known (do not redo):** This is the slice the user looks at. It flips the default only because every slot is filled by then.

## Do
1. `web/layout.json` becomes the cockpit; the present default moves to `web/layouts/standard.json` (it exists) and stays in the Layout menu with Classic, Simple, Stream and Debug. A new hero also starts on the cockpit (it is the simple view). A one-time line: "The layout changed. The old ones are in the menu."
2. One shared look for the cockpit panels in `components.css` (J.5); contrast 4.5:1 text and 3:1 marks; the 2 px focus ring on every control including zone buttons; 38 to 44 px targets under `(pointer: coarse)`; `prefers-reduced-motion`; light theme values for every token used.
3. The `/` palette: filter the verbs and the orderable areas; Enter runs through `POST /api/commands`.
4. The numbers on the first screen, counted with the fixture, in the card result (target: about 16).

## Acceptance (also the card's boxes)
- Fit check passes for the default page at the four sizes in dark and light, with zero, two and five needs rows and with one and eight orders
- A stored layout is untouched (test); the Layout menu returns to every older layout
- Renders at the four sizes looked at and compared with `research/iterations/it3/`; the differences listed
- Tab order follows reading order; every control reachable by keyboard

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis7: The cockpit becomes the default; polish and access" --card uis7 --owns internal/watch/web/layout.json --owns internal/watch/web/layouts/ --owns internal/watch/web/components.css --owns internal/watch/web/panels/{statusline,needs-count,needs-you,orders,zone-map,area-map,progress,journey-strip,health,log}/panel.css --owns tools/uishots/layout-rows.json --owns tools/uishots/genlayouts.js --component watch --kind feature --size M --assignee credit-runner \
  --ac "Fit check passes for the default page at the four sizes in dark and light, with zero, two and five needs rows and with one and eight orders" \
  --ac "A stored layout is untouched (test); the Layout menu returns to every older layout" \
  --ac "Renders at the four sizes looked at and compared with `research/iterations/it3/`; the differences listed" \
  --ac "Tab order follows reading order; every control reachable by keyboard"
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
