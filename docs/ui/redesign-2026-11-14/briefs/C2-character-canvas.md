# Brief C2: The character screen canvas

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `charui`, branch `agent/charui`, worktree `~/code/d2lod-wt/charui` off current `origin/main`.

**Goal.** `/watch/inventory` becomes one connected canvas: hero doll, gear set chips, mercenary strip, pack, belt, gold, inspector, and the stash tabs, with requirement display everywhere, at 1440, 900 and 390 px.

**Starts after:** S1 merged and C1 merged (its fields are your data). **Can run next to:** S2 to S7, bot cards.

**Read first (three):** `docs/ui/redesign-2026-11-14/I-character-screen.md`; `docs/ui/redesign-2026-11-14/mockups/character-screen.html`; `docs/gearsets.md`.

**You own:** `internal/watch/web/inventory.js`, `internal/watch/web/inventory.css`, `internal/watch/web/inventory-tip.js`, `internal/watch/web/inventory-tip.css`, `internal/watch/web/inventory-merc.js`, `internal/watch/web/inventory-merc.css`, `internal/watch/web/inventory-sets.js`, `internal/watch/web/inventory-sets.css`, `internal/watch/invui.go`, `internal/watch/testdata/charscreen/`.

**Do not touch:** `inventory-actions.js` and `inventory-drag.js` (C3 owns the gestures; keep their hooks working), the routes and POST bodies, `/watch` panels.

**Already known (do not redo):** The current page's problems are listed in I.1 with a picture. The mercenary view and its fixtures exist (`inventory_merc_*.json`).

## Do
1. Layout per I.3: three touching zones in one frame at 1440 (who wears | pack | other container), two columns plus a band at 900, and at 390 the pack always below a three-way switch Hero / Mercenary / Stash. No card gaps, no empty detail box, no `n/a` slot: the mercenary strip draws only the slots in `merc.slots`.
2. One tooltip and one inspector (I.4): hover shows the game's lines, the requirement rows for the hero and, when relevant, for the mercenary, and the comparison; click selects and shows the same plus the action buttons. Remove the legend line and the old detail card.
3. Requirement display (I.5) from `req` and `usable` only: red and green rows with needed and current; cell dimming and corner badge (`CLS`, `LVL`, `STR`, `DEX`, `BRK`, `UNID`, `MERC`); slot outlines from `canWear` / `noWearSlot`; the reason text under a refused slot. Never compute a rule in the page.
4. Comparison (I.6): worn item of the same slot (both rings, both hands), differing lines as facts, no verdict; first line `would need ...` when unmet.
5. Search box and chips (Usable by hero, Usable by mercenary, Unidentified, Marked sell): dim, never hide or move.
6. Gear sets as chips under the doll with the existing preview on hover; a set with an item under its requirement shows the warn mark and Apply lists what it skips.
7. Stash tabs: Stash (grid, stash gold, deposit and withdraw, the at-the-chest line, the reserved locked line "Stash from anywhere" with the text of I.8: do not build anything behind it), Cube when owned, and an empty `Loot rules` tab placeholder that the loot-rules slice fills.
8. Mercenary strip states: none, alive, dead (Revive button only when the merc desk of B3 is bound), unknown. Until B9: "Give to mercenary" and drops on his slots show the one line of I.9, not a live control.
9. The bottom strip of pending and recent item actions (the existing queue list restyled).

## Acceptance (also the card's boxes)
- DOM goldens (node) for: hero only; with each mercenary fixture (none, dead, desert, ironwolf, unknown); the 21 validity cases of I.10 rendered from C1's goldens (badge, dimming, rows, slot outline, reason text)
- No rule in the page: a test greps the page scripts for requirement arithmetic and the class list (the allowed source is the view)
- Pictures at 1440, 900, 390 looked at; at 1440 by 900 nothing scrolls; no empty area larger than one grid row inside the frame
- Every action that worked on the old page still works (the existing inventory DOM and logic tests pass or are ported, listed in the result)
- Keyboard: every item and slot reachable, the inspector buttons in reading order

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "charui: The character screen canvas" --card charui --owns internal/watch/web/inventory.js --owns internal/watch/web/inventory.css --owns internal/watch/web/inventory-tip.js --owns internal/watch/web/inventory-tip.css --owns internal/watch/web/inventory-merc.js --owns internal/watch/web/inventory-merc.css --owns internal/watch/web/inventory-sets.js --owns internal/watch/web/inventory-sets.css --owns internal/watch/invui.go --owns internal/watch/testdata/charscreen/ --component watch --kind feature --size L --assignee credit-runner \
  --ac "DOM goldens (node) for: hero only; with each mercenary fixture (none, dead, desert, ironwolf, unknown); the 21 validity cases of I.10 rendered from C1's goldens (badge, dimming, rows, slot outline, reason text)" \
  --ac "No rule in the page: a test greps the page scripts for requirement arithmetic and the class list (the allowed source is the view)" \
  --ac "Pictures at 1440, 900, 390 looked at; at 1440 by 900 nothing scrolls; no empty area larger than one grid row inside the frame" \
  --ac "Every action that worked on the old page still works (the existing inventory DOM and logic tests pass or are ported, listed in the result)" \
  --ac "Keyboard: every item and slot reachable, the inspector buttons in reading order"
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
