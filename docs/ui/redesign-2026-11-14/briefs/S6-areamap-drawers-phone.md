# Brief S6: The area map, the drawers, and the phone arrangement

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis6`, branch `agent/uis6`, worktree `~/code/d2lod-wt/uis6` off current `origin/main`.

**Goal.** The slot panel `area-map` (the minimap without its debug controls), the drill-downs that replace the removed tiles (log drawer, onboarding sheet), and the phone tabs.

**Starts after:** S1b merged. **Can run next to:** S2 to S5.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 2)`; `docs/ui/redesign-2026-11-14/mockups/v2/watch.html`; `internal/watch/web/panels/map/panel.js`.

**You own:** `internal/watch/web/panels/area-map/`, `internal/watch/web/panels/log/`, `internal/watch/web/panels/map/ (only to export its drawing code as a module both panels import; version raised, behaviour unchanged)`, `internal/watch/testdata/areamap/`, `internal/watch/testdata/logpanel/`.

**Do not touch:** `app.js`, `unlockui.js`, `alerts.js` and their cards; `shell.js` and `shell.css` (ask S1b's owner for a hook).

**Already known (do not redo):** `map/panel.js` is 35 KB: move its drawing functions into a module without changing them, then build `area-map` on that.

## Do
1. `area-map` (slot): the same `/api/senses` drawing as `map`, filling its box, with two controls only: zoom and a `Layers` popover (trail window, coverage, heat, events). No timeline, replay, export or legend text; the zone name and explored percent are the tile's label. The full `map` panel is untouched and stays in Classic and Debug.
2. `log` as a drawer opened from the band's `...` menu and from the sentence popover: tabs Now, Journal, Story, Alerts over the existing endpoints, one fetched at a time through a small `GET /api/log?tab=`; times in the viewer's zone. An open question is a Needs you row, not a log line.
3. Onboarding in a cockpit: a chip "Getting started 1 of 5" in the `...` menu and, until dismissed, as the first Needs you row of kind info; the existing onboarding panel opens in a sheet. The panel itself is not edited.
4. Phone: the three tabs of J.2 (`phoneTabs`); 38 px rows and buttons; the picture, the band and the tab bar stay put, only the tab's content may scroll inside itself.

## Acceptance (also the card's boxes)
- DOM goldens of `area-map` (with trail, empty, no data) and of the drawer per tab
- The `map` panel's goldens are unchanged and Classic at 1440 is pixel-identical (uishots diff)
- At 1920x1080 the rail shows both maps; at 1440x900 the switch shows one at a time; the fit check passes in both
- At 390x844 each tab fits the screen under the picture with the fixture data (render looked at)

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis6: The area map, the drawers, and the phone arrangement" --card uis6 --owns internal/watch/web/panels/area-map/ --owns internal/watch/web/panels/log/ --owns internal/watch/web/panels/map/ --owns internal/watch/testdata/areamap/ --owns internal/watch/testdata/logpanel/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens of `area-map` (with trail, empty, no data) and of the drawer per tab" \
  --ac "The `map` panel's goldens are unchanged and Classic at 1440 is pixel-identical (uishots diff)" \
  --ac "At 1920x1080 the rail shows both maps; at 1440x900 the switch shows one at a time; the fit check passes in both" \
  --ac "At 390x844 each tab fits the screen under the picture with the fixture data (render looked at)"
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
