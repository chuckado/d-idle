# Brief L1: The loot rules tab

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `lootui`, branch `agent/lootui`, worktree `~/code/d2lod-wt/lootui` off current `origin/main`.

**Goal.** The player sees and, once unlocked, edits the loot rules the bot follows: two ordered stages, pickers for conditions, presets, a live preview, import and export. HOLD until the engine card has merged its schema, validate, evaluate and dry-run functions.

**Starts after:** C2 merged AND the loot-rules engine merged (another agent's card; the coordinator names it). **Can run next to:** S slices, bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/loot-rules-ux.md`; `docs/ui/redesign-2026-11-14/I-character-screen.md (section 3)`; `the engine's own doc (named by the coordinator)`.

**You own:** `internal/watch/lootrulesapi.go`, `internal/watch/lootrulesapi_test.go`, `internal/watch/web/lootrules.js`, `internal/watch/web/lootrules.css`, `internal/watch/testdata/lootrules/`, `data/unlocks/ladder.json (one row, only if no other card owns the file in your wave)`.

**Do not touch:** the engine, the rule files' format, pickit and tidy behaviour.

**Already known (do not redo):** The engine is another agent's design; this brief is on hold until it exists. If the engine lacks something loot-rules-ux.md section 5 lists, report it instead of working around it.

## Do
1. Routes: `GET /api/lootrules` (the rule list, the engine's schema of conditions and actions, the protected set, the presets, the unlock state) and `POST /api/lootrules` on the answer listener (`save`, `preset`, `import`, `undo`, and `preview` for an unsaved list: per-rule counts, matched ids, the difference to the saved list). All evaluation is the engine's; the watch side only carries it.
2. The tab (the placeholder C2 left): the two stages, rows with on and off, name, the sentence of chips, the action, the match count; add, edit (pickers generated from the schema), reorder by buttons and drag, the fixed last row, the grey protected block, the `never reached` mark.
3. Preview: selecting a rule lights its items in the pack and stash grids and lists the matched drops; while editing the counts update from `preview`; the net-change line before Save.
4. Presets with a difference view before applying; import and export as text with the engine's validation errors shown at the rule and field; one-step Undo.
5. Locked state: the tab is readable (the Starter preset with reasons), the controls are replaced by the ladder teaser. Ladder row `lootrules` as proposed in loot-rules-ux.md section 4 (the user decides the trigger).
6. The item tooltip line "Loot rule N (name): action" from the engine's evaluate.

## Acceptance (also the card's boxes)
- Handler goldens for locked, Starter, edited, invalid import, preview; both classes
- DOM goldens of the tab: locked, list, editing one rule, preview selected, error
- No rule logic in the page (the pickers come from the schema; a test adds a fake condition to the schema and the picker shows it)
- The conditions `usable by hero`, `usable by mercenary`, `usable after +N` are present when the engine offers them
- Pictures at 1440 and 390 looked at

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "lootui: The loot rules tab" --card lootui --owns internal/watch/lootrulesapi.go --owns internal/watch/lootrulesapi_test.go --owns internal/watch/web/lootrules.js --owns internal/watch/web/lootrules.css --owns internal/watch/testdata/lootrules/ --owns data/unlocks/ladder.json --component watch --kind feature --size L --assignee credit-runner \
  --ac "Handler goldens for locked, Starter, edited, invalid import, preview; both classes" \
  --ac "DOM goldens of the tab: locked, list, editing one rule, preview selected, error" \
  --ac "No rule logic in the page (the pickers come from the schema; a test adds a fake condition to the schema and the picker shows it)" \
  --ac "The conditions `usable by hero`, `usable by mercenary`, `usable after +N` are present when the engine offers them" \
  --ac "Pictures at 1440 and 390 looked at"
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
