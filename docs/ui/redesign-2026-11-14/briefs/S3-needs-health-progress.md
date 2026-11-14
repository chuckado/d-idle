# Brief S3: Needs you, Health and Progress

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis3`, branch `agent/uis3`, worktree `~/code/d2lod-wt/uis3` off current `origin/main`.

**Goal.** Needs you, Health and Progress. Panels `needs-you`, `needs-count`, `health`, `progress`, all reading `/api/glance`.

**Starts after:** S2 merged (its goldens are your fixtures). **Can run next to:** S4, S5, S6.

**Read first (three):** `docs/ui/redesign-2026-11-14/B-principles-ia.md`; `docs/ui/redesign-2026-11-14/E-visual-system.md`; `docs/ui/redesign-2026-11-14/mockups/watch-orders.html`.

**You own:** `internal/watch/web/panels/needs-you/`, `internal/watch/web/panels/needs-count/`, `internal/watch/web/panels/health/`, `internal/watch/web/panels/progress/`, `internal/watch/testdata/glancepanels/`.

**Do not touch:** `glance.go` (ask the coordinator for a field instead), the old `gear`, `tally`, `why`, `alerts-count` panels.

**Already known (do not redo):** The mockup shows the intended look and wording. Mercenary and Supplies stay `unknown` until B5.

## Do
1. `needs-you` (band `side`, fit): up to four alert rows (shape icon by severity, one line, a sub line, the action buttons from `actions[]` posted through `ctx.post`), "+N more" beyond four, hidden (no cell) when empty. Red rows may use `aria-live="assertive"`, the list is polite.
2. `needs-count` (band `status`): "N need you" toned by the worst severity, "all quiet" muted at zero; a click scrolls to `needs-you`.
3. `health` (fit): five lights (Life, Gear, Readiness, Mercenary, Supplies) with a word each; `unknown` is a hollow disc with the word "not reported"; a click opens the detail text under the row (one open at a time).
4. `progress` (fit): level meter with "level N in about M min", xp/h with the 30-point sparkline (inline SVG, no axis), journey "n of m" with the next chapter and band, gold/h; a dash with a tooltip while a rate is still building.
5. Notifications (user answer 11): a switch in the `needs-you` header menu, off by default, that raises a desktop notification for a new red row only; stored per browser.
6. Borderless tile look with small caps labels (answer 10) for these four panels only, in their own `panel.css`.

## Acceptance (also the card's boxes)
- DOM goldens: `needs-you` for empty, one red, red + amber + blue, six rows (overflow); `needs-count` for 0, 1, 3; `health` for all ok, gear broken, two unknown; `progress` for measuring and for full data
- A button in a `needs-you` row posts exactly the `post` and `body` the reply names (node test with a fake `ctx.post`)
- Under Standard the cells of `gear`, `why`, `tally` and `alerts-count` now show the new panels; under Classic nothing changed
- No number on these panels lacks a unit or a label; pictures at 390 and 1440 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis3: Needs you, Health and Progress" --card uis3 --owns internal/watch/web/panels/needs-you/ --owns internal/watch/web/panels/needs-count/ --owns internal/watch/web/panels/health/ --owns internal/watch/web/panels/progress/ --owns internal/watch/testdata/glancepanels/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "DOM goldens: `needs-you` for empty, one red, red + amber + blue, six rows (overflow); `needs-count` for 0, 1, 3; `health` for all ok, gear broken, two unknown; `progress` for measuring and for full data" \
  --ac "A button in a `needs-you` row posts exactly the `post` and `body` the reply names (node test with a fake `ctx.post`)" \
  --ac "Under Standard the cells of `gear`, `why`, `tally` and `alerts-count` now show the new panels; under Classic nothing changed" \
  --ac "No number on these panels lacks a unit or a label; pictures at 390 and 1440 looked at"
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
