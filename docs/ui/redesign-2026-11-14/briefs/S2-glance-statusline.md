# Brief S2: The glance endpoint and the status sentence

> status: ready to send (written 2026-11-14 after the user's answers; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `uis2`, branch `agent/uis2`, worktree `~/code/d2lod-wt/uis2` off current `origin/main`.

**Goal.** The glance endpoint and the status sentence. `GET /api/glance` (`watch.glance/1`) with all four blocks, and the status band panel `statusline`.

**Starts after:** S1 merged. **Can run next to:** S4, S6, the bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/B-principles-ia.md`; `docs/ui/redesign-2026-11-14/E-visual-system.md`; `internal/watch/hero.go`.

**You own:** `internal/watch/glance.go`, `internal/watch/glance_test.go`, `internal/watch/testdata/glance/`, `internal/watch/web/panels/statusline/`.

**Do not touch:** `hero.go`, `commands.go` (S4), other panels. Register the route with one line where `/api/hero` is registered and nothing else in that file.

**Already known (do not redo):** The reply shape above is fixed so S3 can build its panels against your goldens: do not change field names after your first push of the goldens; tell the coordinator if you must.

## Do
1. `GET /api/glance`: `{schema, seq, sentence{tone, word, place, reason, why{code, now, change, next}}, needs[{id, kind, sev, text, sub, actions[{label, post, body}]}], health{life, gear, readiness, merc, supplies: {light, word, detail}}, progress{level{from, to, pct, etaSeconds}, xpPerHour, xpSeries[], chapter{n, of, title, band}, goldPerHour, tally}}`. Built in the Hub from what it already holds: the hero view, `/api/why` codes and `whychange.json`, alerts of severity `alarm`, the open question, the repair view (`trip.button`), `stats.readiness`, the campaign view, the tally. Cached like `/api/hero`; no new source is polled.
2. A source that is not wired answers `light: "unknown"` (mercenary, supplies) or is left out of `needs`; never a guess. Hub setters `SetMercSource`, `SetSuppliesSource`, `SetJourneyBlockedSource` exist and are nil by default (B5 and B6 fill them).
3. xp per hour samples: a ring of 30 one-minute points in the Hub for `xpSeries`; gold per hour from the same clock when the gold source is bound, else null.
4. The `needs` order: red first, then amber, then blue; at most 8 in the reply. `actions[].post` names an existing answer route (`/api/repair/request`, `/api/answer`, `/api/campaign/intent`); this slice adds no write route.
5. Panel `statusline` (band `status`, `optIn`): dot, state word, place, reason on one line with ellipsis; a click opens Now / Would change it / Next; `aria-live="polite"` on the sentence only. Uptime and clock go to its `title`.

## Acceptance (also the card's boxes)
- Handler goldens `testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json`
- For every code in `whychange.json` the sentence has a non-empty reason (table test)
- The reply is cached: two requests within the cache window build once (test)
- DOM goldens of `statusline` for tone ok, warn, danger, info, loading and error
- Under Standard the band shows `statusline` in place of `hero-status`; under Classic nothing changed (pictures looked at)

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis2: The glance endpoint and the status sentence" --card uis2 --owns internal/watch/glance.go --owns internal/watch/glance_test.go --owns internal/watch/testdata/glance/ --owns internal/watch/web/panels/statusline/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "Handler goldens `testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json`" \
  --ac "For every code in `whychange.json` the sentence has a non-empty reason (table test)" \
  --ac "The reply is cached: two requests within the cache window build once (test)" \
  --ac "DOM goldens of `statusline` for tone ok, warn, danger, info, loading and error" \
  --ac "Under Standard the band shows `statusline` in place of `hero-status`; under Classic nothing changed (pictures looked at)"
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
