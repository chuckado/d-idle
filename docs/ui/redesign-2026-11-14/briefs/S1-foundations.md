# Brief S1: Shell foundations and the default flip

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis1`, branch `agent/uis1`, worktree `~/code/d2lod-wt/uis1` off current `origin/main`.

**Goal.** Shell foundations and the default flip. Opt-in panels, the layout keys `status` and `fallback`, the four presets, Standard as the default for everyone without a stored layout, Simple for a new hero, the layout menu, the new tokens. No new panel.

**Starts after:** nothing (first slice). **Can run next to:** the bot cards B1, B2, B3, B4, B7.

**Read first (three):** `docs/ui/redesign-2026-11-14/D-layouts.md`; `docs/ui/redesign-2026-11-14/E-visual-system.md`; `docs/ui-design-system.md (sections 0, 6.4, 15.5)`.

**You own:** `internal/watch/web/shell.js`, `internal/watch/web/layout.js`, `internal/watch/web/shell.json`, `internal/watch/web/layout.json`, `internal/watch/web/layouts/`, `internal/watch/web/tokens.css`, `internal/watch/web/shell.css`, `internal/watch/layout.go`, `internal/watch/layout_test.go`, `tools/uilint/`, `tools/uishots/layout-rows.json`, `tools/uishots/genlayouts.js`.

**Do not touch:** any `web/panels/*` directory, `watch.html`, `app.js`, the compat fixtures.

**Already known (do not redo):** The layout drafts and their checks exist (`docs/ui/redesign-2026-11-14/layouts/gen.py`, 0 findings). Per-entry `hidden: true` already exists in `web/layout.js`. `web/layout.json` is written by `tools/uishots/genlayouts.js`, so change the generator, not only the file.

## Do
1. Panel manifest key `optIn` (boolean, optional): the shell mounts such a panel only when the layout in force names its id (placed, not `hidden`). `tools/uilint` accepts the key. No existing manifest gets it.
2. Layout top-level key `status` (optional list of ids or `{id, fallback}`): when present it chooses the status band's panels instead of the `status` list of `web/shell.json`. An entry whose id is not a mounted panel uses its `fallback`, else is skipped.
3. Layout entry key `fallback` (optional id): when the entry's `id` is not a mounted panel or pre-shell card, the shell places the `fallback` target in that cell and shows it even if the same layout lists that target as `hidden`. When the id is mounted, the fallback is ignored. `ParseLayout` and `uilint` (rule `layout-manifest`) learn both keys; an unknown id with a fallback is not a lint error.
4. Add the new ids to `reserved` in `web/shell.json`: `statusline`, `needs-count`, `needs-you`, `orders`, `health`, `progress`, `zone-map`, `log`.
5. Presets: copy today's `web/layout.json` unchanged to `web/layouts/classic.json`; add `simple.json`, `standard.json`, `stream.json`, `debug.json` from `docs/ui/redesign-2026-11-14/layouts/` (regenerate with its `gen.py`; port the rows into `tools/uishots/layout-rows.json` and `genlayouts.js` so the repo's own generator writes the same files); `web/layout.json` becomes Standard. `?layout=classic|simple|standard|stream|debug` loads each.
6. A new hero (no stored layout and `/api/onboarding` says the welcome card is not dismissed) gets Simple; everyone else without a stored layout gets Standard. A stored layout always wins and is never rewritten.
7. The band's link row becomes: Inventory, Skills, and one `...` menu holding Help, Edit layout, Theme and `Layout: Classic / Simple / Standard / Stream / Debug` (kept per browser in `localStorage`, read in try/catch). One dismissible line for users on the new default: "The layout changed. Classic is in the menu."
8. Tokens (additive, with light values): `--d-ok`, `--d-danger`, `--d-muted`, `--d-state-bg-ok|warn|danger|info`, `--d-hit` (44px), `--d-row` (36px), as in E.2.

## Acceptance (also the card's boxes)
- A test loads `testdata/compat/layout-v1-{mockup,live,command}.json` and the old `layout.json` (now classic) and asserts the same placed ids and cells as before this change
- Standard and Simple resolved against today's panel list: no unknown id, no overlap, no empty cell; every fallback of D.4 is shown in its cell (node test of the plan)
- An `optIn` panel is not mounted under a layout that does not name it (node test with a fixture manifest)
- `?layout=classic` renders the page as before S1 (uishots diff against the present golden, said in the result)
- The `...` menu switches presets without a POST; Edit layout still saves and Reset returns to the chosen preset
- uilint knows `status`, `fallback`, `optIn`; token compat test green; pictures at 390 and 1440 looked at for Standard, Simple and Classic

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis1: Shell foundations and the default flip" --card uis1 --owns internal/watch/web/shell.js --owns internal/watch/web/layout.js --owns internal/watch/web/shell.json --owns internal/watch/web/layout.json --owns internal/watch/web/layouts/ --owns internal/watch/web/tokens.css --owns internal/watch/web/shell.css --owns internal/watch/layout.go --owns internal/watch/layout_test.go --owns tools/uilint/ --owns tools/uishots/layout-rows.json --owns tools/uishots/genlayouts.js --component watch --kind feature --size L --assignee credit-runner \
  --ac "A test loads `testdata/compat/layout-v1-{mockup,live,command}.json` and the old `layout.json` (now classic) and asserts the same placed ids and cells as before this change" \
  --ac "Standard and Simple resolved against today's panel list: no unknown id, no overlap, no empty cell; every fallback of D.4 is shown in its cell (node test of the plan)" \
  --ac "An `optIn` panel is not mounted under a layout that does not name it (node test with a fixture manifest)" \
  --ac "`?layout=classic` renders the page as before S1 (uishots diff against the present golden, said in the result)" \
  --ac "The `...` menu switches presets without a POST; Edit layout still saves and Reset returns to the chosen preset" \
  --ac "uilint knows `status`, `fallback`, `optIn`; token compat test green; pictures at 390 and 1440 looked at for Standard, Simple and Classic"
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
