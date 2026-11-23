# Brief S8: Retire the old tiles

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis8`, branch `agent/uis8`, worktree `~/code/d2lod-wt/uis8` off current `origin/main`.

**Goal.** Retire the old tiles. HOLD: send only on the user's word that the new layouts have proven themselves (answer 8).

**Starts after:** S7 merged AND the user's go. **Can run next to:** nothing that edits `watch.html`, `app.js` or `shell.json`.

**Read first (three):** `docs/ui/redesign-2026-11-14/D-layouts.md (section 3, the migration table)`; `docs/ui-design-system.md (section 0)`; `internal/watch/compat_test.go`.

**You own:** `internal/watch/web/shell.json`, `internal/watch/web/watch.html`, `internal/watch/web/app.js`, `internal/watch/web/panels/{why,activity,gear,tally,currency}/panel.json`, `internal/watch/testdata/compat/`.

**Do not touch:** the new panels; any fixture that already exists (add new ones only).

**Already known (do not redo):** This is the one slice that removes things. If any step would break a stored layout, stop and report.

## Do
1. For `why`, `activity`, `gear`, `tally`, `currency`: `mergedInto` in the manifest and an alias in `web/shell.json` to the panel named in D.3; the directory and manifest stay (stub).
2. For `now`, `queue`, `orders-map`, `atlas`, `thoughts`, `journal`, `chronicle`, `diary`, `alerts`: alias to the new panel, remove from `compat`, keep the `data-lid` section in `watch.html` as an empty stub, remove their drawing code and timers from `app.js`, `unlockui.js`, `alerts.js` only where nothing else reads it.
3. Classic and stored layouts then show the new panel at the old tile's first position (the existing alias rule). Say in the card which tiles a Classic user will see change.
4. Add compat fixtures `layout-v1-{classic,simple,standard,stream,debug}.json`; never edit an existing one.

## Acceptance (also the card's boxes)
- Every id of every compat fixture still resolves (alias, stub or panel); `go test ./internal/watch -run Compat` green without touching an old fixture
- A stored layout naming an old tile shows one tile (the new panel) at that position, not two and not none
- `app.js` has no `setInterval` left for a retired card (list what remains and why)
- Pictures of Classic before and after, looked at, and the differences listed for the user

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis8: Retire the old tiles" --card uis8 --owns internal/watch/web/shell.json --owns internal/watch/web/watch.html --owns internal/watch/web/app.js --owns internal/watch/web/panels/{why,activity,gear,tally,currency}/panel.json --owns internal/watch/testdata/compat/ --component watch --kind chore --size L --assignee credit-runner \
  --ac "Every id of every compat fixture still resolves (alias, stub or panel); `go test ./internal/watch -run Compat` green without touching an old fixture" \
  --ac "A stored layout naming an old tile shows one tile (the new panel) at that position, not two and not none" \
  --ac "`app.js` has no `setInterval` left for a retired card (list what remains and why)" \
  --ac "Pictures of Classic before and after, looked at, and the differences listed for the user"
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
