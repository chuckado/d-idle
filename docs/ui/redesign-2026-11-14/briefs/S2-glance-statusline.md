# Brief S2: The glance endpoint and the band

> status: ready to send once the user accepts the second attempt (rewritten 2026-11-14 for J-second-attempt.md; the earlier version of this brief is void). Model: Sonnet. Agent `uis2`, branch `agent/uis2`, worktree `~/code/d2lod-wt/uis2` off current `origin/main`.

**Goal.** `GET /api/glance` (`watch.glance/1`) with sentence, needs, health, progress and journey, and the cockpit band: the status sentence with its popover, slim life and mana, the needs-you count.

**Starts after:** S1b merged. **Can run next to:** S4, S6, C2, the bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/J-second-attempt.md (section 2)`; `docs/ui/redesign-2026-11-14/B-principles-ia.md`; `internal/watch/hero.go`.

**You own:** `internal/watch/glance.go`, `internal/watch/glance_test.go`, `internal/watch/testdata/glance/`, `internal/watch/web/panels/statusline/`, `internal/watch/web/panels/needs-count/`.

**Do not touch:** `hero.go`, `commands.go` (S4), other panels. Register the route with one line where `/api/hero` is registered.

**Already known (do not redo):** The reply shape is fixed here so S3 can build against your goldens: do not rename fields after the first push.

## Do
1. `GET /api/glance`: `{schema, seq, sentence{tone, word, place, reason, why{code, now, change, next}, uptime, voice}, needs[{id, kind, sev, text, sub, actions[{label, post, body}]}], health{life, gear, readiness, merc, supplies: {light, word, detail}}, progress{level{from, to, pct, etaSeconds}, xpPerHour, xpSeries[], goldPerHour, tally}, journey{n, of, chapters[{id, state}], next{id, title, band, text}, actions[]}}`. Built in the Hub from what it already holds; cached like `/api/hero`; no new source is polled. A source that is not wired answers `light: "unknown"` or is left out. Hub setters `SetMercSource`, `SetSuppliesSource`, `SetJourneyBlockedSource` exist, nil by default (B5, B6).
2. xp per hour samples: a ring of 30 one-minute points in the Hub; gold per hour from the same clock when the gold source is bound.
3. Panel `statusline` (slot `band`): state chip, place, reason on one line with an ellipsis; a click opens the popover (Now, Would change it, Next, uptime, the voice line, a link to the log drawer). `aria-live="polite"` on the sentence only.
4. Panel `needs-count` (slot `band`): "N need you" toned by the worst severity, "all quiet" muted at zero.
5. The band in a cockpit shows: brand, `statusline`, `hero-meters` in its slim form (two 5 px bars, one number each: a `slot` mode of the existing panel, version raised, old mode unchanged), `needs-count`, then Character, Skills and the `...` menu.

## Acceptance (also the card's boxes)
- Handler goldens `testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json`
- For every code in `whychange.json` the sentence has a non-empty reason (table test)
- DOM goldens of `statusline` (ok, warn, danger, info, loading, error; long reason cut) and `needs-count` (0, 1, 3)
- The band is 44 px at the three desktop sizes and two lines on the phone; fit check passes
- Classic and the other free-grid layouts are unchanged

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "uis2: The glance endpoint and the band" --card uis2 --owns internal/watch/glance.go --owns internal/watch/glance_test.go --owns internal/watch/testdata/glance/ --owns internal/watch/web/panels/statusline/ --owns internal/watch/web/panels/needs-count/ --component watch --kind feature --size M --assignee credit-runner \
  --ac "Handler goldens `testdata/glance/view-{daily,needs,blocked,paused,dead,unknown}-{sorceress,paladin}.json`" \
  --ac "For every code in `whychange.json` the sentence has a non-empty reason (table test)" \
  --ac "DOM goldens of `statusline` (ok, warn, danger, info, loading, error; long reason cut) and `needs-count` (0, 1, 3)" \
  --ac "The band is 44 px at the three desktop sizes and two lines on the phone; fit check passes" \
  --ac "Classic and the other free-grid layouts are unchanged"
```

## The fit rule (binding for every cockpit and canvas slice)
- The page must not scroll and nothing may clip at the full-screen sizes 1920x1080, 1440x900, 1366x768 and at the browser-window sizes 1914x930, 1904x1000, 1536x730, 1280x620 (a window with its chrome is shorter than the screen: the user's first look was at 1914x930); at 390x844 it follows the phone design of J. `tools/uishots` gets (in S1b) a `fit` check that loads each page at those sizes in headless Chromium on the builder and fails when `document.scrollingElement.scrollHeight > innerHeight`, when any `[data-cockpit-slot]` or `[data-cockpit-panel]` (or a descendant) has `scrollHeight > clientHeight` (lists marked `data-scroll` excepted), or when the slots do not cover the frame (sum of slot areas plus gaps equals the frame, within 2 px). Run it (`labrun <builder> -- tools/uishots.sh --fit`, never on the worker host that runs the games) and put its output in the card result. A slice that fails the fit check is not done.
- A slot panel has no content-driven height: it renders inside the box it is given; long text is cut with an ellipsis and the full text goes to `title` or a popover; a list longer than its box scrolls inside (`data-scroll`).
- Look at your own renders at those sizes and write three lines of critique in the card result (dead space, alignment, noise). Compare with `docs/ui/redesign-2026-11-14/research/iterations/it3/` (the accepted look).

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
