# Brief S6: Log, a compact minimap, and the phone layout

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis6`, branch `agent/uis6`, worktree `~/code/d2lod-wt/uis6` off current `origin/main`.

**Goal.** Log, a compact minimap, and the phone layout. Panel `log`, a compact mode of `map`, and the narrow-width polish of the new presets.

**Starts after:** S1 merged. **Can run next to:** S2, S3, S4, S5.

**Read first (three):** `docs/ui/redesign-2026-11-14/B-principles-ia.md`; `docs/ui/redesign-2026-11-14/D-layouts.md`; `internal/watch/web/panels/map/panel.js`.

**You own:** `internal/watch/web/panels/log/`, `internal/watch/web/panels/map/`, `internal/watch/testdata/logpanel/`.

**Do not touch:** `app.js`, `unlockui.js`, `alerts.js` and their cards (they stay for Classic); `shell.js` and `shell.css` belong to S1: ask for a hook rather than edit.

**Already known (do not redo):** `map/panel.js` is 35 KB: add a mode switch around the control block, do not restructure it.

## Do
1. Panel `log` (band `grid`, `optIn`, fit): tabs Now (the thoughts feed), Journal, Story (Chronicle and Diary), Alerts. Only the open tab's endpoint is read (the panel's `query()` picks it; the manifest endpoint is a small new `GET /api/log?tab=` that forwards to the existing sources, so the one-endpoint rule holds). Times in the viewer's own zone (answer 12).
2. An open question is not shown here (it is a Needs you row); the feed lines stay.
3. `map`: a compact mode chosen by the layout entry (`"mode": "compact"`, an additive entry key the panel reads from `ctx`): zoom and one `Layers` button that opens the existing switches (trail window, coverage, heat, events); timeline, replay and exports are not rendered. Without the key the panel is exactly as today (raise the manifest `version`; old layouts and Classic are unchanged). Standard, Simple and Stream use compact; Debug does not.
4. Phone and tablet: check Standard and Simple at 390 and 768; the band, the picture and the first Orders row (or its fallback) fit 844 px at 390; fix only inside your panels and report anything that needs the shell.

## Acceptance (also the card's boxes)
- DOM goldens of `log` per tab and empty; of `map` in compact mode; the existing map goldens unchanged
- Classic at 1440: the map tile is pixel-identical to before (uishots diff, said in the result)
- Only one of the four log sources is requested while a tab is open (node test of `query()`)
- Pictures at 390, 768, 1440 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis6: Log, a compact minimap, and the phone layout" --card uis6 --owns internal/watch/web/panels/log/ --owns internal/watch/web/panels/map/ --owns internal/watch/testdata/logpanel/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens of `log` per tab and empty; of `map` in compact mode; the existing map goldens unchanged" \
  --ac "Classic at 1440: the map tile is pixel-identical to before (uishots diff, said in the result)" \
  --ac "Only one of the four log sources is requested while a tab is open (node test of `query()`)" \
  --ac "Pictures at 390, 768, 1440 looked at"
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
