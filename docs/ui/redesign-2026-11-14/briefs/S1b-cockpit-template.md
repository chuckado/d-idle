# Brief S1b: The cockpit template in the shell, and the fit check

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis1b`, branch `agent/uis1b`, worktree `~/code/d2lod-wt/uis1b` off current `origin/main`.

**Goal.** The shell can lay a page out as a viewport-filling cockpit from a layout file (`template: "cockpit/1"` with named slots), next to the free grid that stays as it is; and `tools/uishots` can prove a page fits.

**Starts after:** nothing (S1 is merged). **Can run next to:** the bot cards; C2.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (sections 2, 3)`; `docs/ui/redesign-2026-11-14/mockups/v2/watch.html`; `docs/ui-design-system.md (sections 0, 6.4, 15.5)`.

**You own:** `internal/watch/web/shell.js`, `internal/watch/web/shell.css`, `internal/watch/web/layout.js`, `internal/watch/web/layoutedit.js`, `internal/watch/web/shell.json`, `internal/watch/web/layouts/cockpit.json`, `internal/watch/layout.go`, `internal/watch/layout_test.go`, `tools/uilint/`, `tools/uishots/`, `tools/uishots.sh`.

**Do not touch:** `web/layout.json` (the default does not change in this slice), any panel directory, the compat fixtures, the free-grid behaviour.

**Already known (do not redo):** The first attempt's free-grid presets cannot fit a screen: content-height rows give ragged bottoms. Fallbacks to old tiles (S1) made the live page the old tiles in new places; the cockpit has none.

## Do
1. Layout keys (additive, ignored by older readers): `template` (`cockpit/1`), `slots` (`band`, `stage`, `strip`, `rail`, `railWide {left, right}`, `phoneTabs`), `sizes` (`strip`, `railMin`). `ParseLayout` and `uilint` learn them; a layout without `template` is laid out exactly as today.
2. Cockpit layout in `shell.css` / `shell.js` as in the mockup: band 44 px; stage height = viewport height - band - strip - gaps, width = height x 4/3 capped so the rail keeps `railMin`; strip of three cells; rail of three rows (auto, auto, 1fr); from 1700 px the rail is two columns (`railWide`); under 700 px the phone arrangement with tabs. The page does not scroll. Each slot element carries `data-slot`.
3. Panel manifest key `slot: true`: the panel renders into a fixed box. The shell mounts in a slot only panels that declare it; a slot whose panel does not exist is left out and its neighbours take the room (no fallback to old tiles in a cockpit).
4. The editor on a cockpit layout: tiles are dragged onto another slot to swap; the strip height and the rail width have drag handles with limits; a slot's panel can be changed from a list of slot panels; Save, Reset and Done as today. Free placement remains only in free-grid layouts.
5. `web/layouts/cockpit.json` with the slots of J.3 and old-style `breakpoints` as the fallback for older shells; `?layout=cockpit` loads it. It is not the default yet.
6. `tools/uishots`: the `fit` check described below, for `/watch?layout=cockpit` and (later) the character page; four sizes; runs on the builder.

## Acceptance (also the card's boxes)
- Every existing layout (stored, Classic, Standard, Simple, Stream, Debug, the three old presets) renders the same cells as before (the S1 compat test still passes, plus a new one for `template`-less files)
- `?layout=cockpit` with the panels that exist today (video, hero-meters) fills the viewport and does not scroll at the four sizes; empty slots are left out without holes (fit check output in the result)
- Editor on the cockpit: swap two slots, move both dividers, save, reload: the same arrangement (node test of the plan; pictures looked at)
- uilint accepts the new keys and rejects an unknown slot name or a non-slot panel in a slot
- The fit check fails on a deliberately overflowing fixture page and passes on the cockpit

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis1b: The cockpit template in the shell, and the fit check" --card uis1b --owns internal/watch/web/shell.js --owns internal/watch/web/shell.css --owns internal/watch/web/layout.js --owns internal/watch/web/layoutedit.js --owns internal/watch/web/shell.json --owns internal/watch/web/layouts/cockpit.json --owns internal/watch/layout.go --owns internal/watch/layout_test.go --owns tools/uilint/ --owns tools/uishots/ --owns tools/uishots.sh --component watch --kind feature --size L --assignee credit-runner \
  --ac "Every existing layout (stored, Classic, Standard, Simple, Stream, Debug, the three old presets) renders the same cells as before (the S1 compat test still passes, plus a new one for `template`-less files)" \
  --ac "`?layout=cockpit` with the panels that exist today (video, hero-meters) fills the viewport and does not scroll at the four sizes; empty slots are left out without holes (fit check output in the result)" \
  --ac "Editor on the cockpit: swap two slots, move both dividers, save, reload: the same arrangement (node test of the plan; pictures looked at)" \
  --ac "uilint accepts the new keys and rejects an unknown slot name or a non-slot panel in a slot" \
  --ac "The fit check fails on a deliberately overflowing fixture page and passes on the cockpit"
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
