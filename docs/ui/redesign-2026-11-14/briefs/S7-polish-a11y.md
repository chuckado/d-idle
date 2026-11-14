# Brief S7: Polish and accessibility pass over the new panels

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis7`, branch `agent/uis7`, worktree `~/code/d2lod-wt/uis7` off current `origin/main`.

**Goal.** Polish and accessibility pass over the new panels. Tile look, focus, touch targets, reduced motion, light theme, the command palette.

**Starts after:** S3, S5 and S6 merged. **Can run next to:** B cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/E-visual-system.md`; `docs/ui/redesign-2026-11-14/B-principles-ia.md`; `docs/ui-design-system.md (sections 3, 5)`.

**You own:** `internal/watch/web/components.css`, `internal/watch/web/panels/{statusline,needs-you,needs-count,health,progress,orders,zone-map,log}/panel.css`.

**Do not touch:** panel JS behaviour (report a bug instead), the old tiles' look.

**Already known (do not redo):** This is the slice the user and the reviewer look at as a whole; keep it to look and access, no new data.

## Do
1. One shared look for the eight new panels (answer 10): no border, `--d-surface-1`, small caps labels, row separators; move what S3 to S6 duplicated into `components.css` as components named in E.3.
2. Contrast (4.5:1 text, 3:1 marks), the 2 px focus ring on every control incl. map nodes, 44 px targets under `(pointer: coarse)`, `prefers-reduced-motion` (no moving dash, instant swaps), light theme values for every new token.
3. The `/` command palette (C.7): type to filter the verbs and the orderable areas; Enter runs through `POST /api/commands`. Keyboard only as an accelerator; every entry has a button elsewhere.
4. Count the numbers on the first screen of Standard and Simple with the fixture and put the two counts in the card result (targets: about 16 and about 8).

## Acceptance (also the card's boxes)
- uishots light and dark at 390 and 1440 for Standard and Simple looked at; no text under 4.5:1 (list the pairs checked)
- Tab order follows reading order; every control reachable by keyboard (node test of the focusable list per panel)
- DOM goldens of the palette: open, filtered, empty
- Classic unchanged (uishots diff)

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis7: Polish and accessibility pass over the new panels" --card uis7 --owns internal/watch/web/components.css --owns internal/watch/web/panels/{statusline,needs-you,needs-count,health,progress,orders,zone-map,log}/panel.css --component watch --kind feature --size M --assignee credit-runner \
  --ac "uishots light and dark at 390 and 1440 for Standard and Simple looked at; no text under 4.5:1 (list the pairs checked)" \
  --ac "Tab order follows reading order; every control reachable by keyboard (node test of the focusable list per panel)" \
  --ac "DOM goldens of the palette: open, filtered, empty" \
  --ac "Classic unchanged (uishots diff)"
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
