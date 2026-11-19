# Brief C1: Item requirements and usability in the inventory view

> status: ready to send (written 2026-11-14; design in `docs/ui/redesign-2026-11-14/`). Model: Sonnet. Agent `charreq`, branch `agent/charreq`, worktree `~/code/d2lod-wt/charreq` off current `origin/main`.

**Goal.** The inventory view says, for every item, what it requires and whether the hero and the mercenary can use it now, with numbers, from the same rule the equip path uses.

**Starts after:** nothing. **Can run next to:** S1 to S6, B1 to B8.

**Read first (three):** `docs/ui/redesign-2026-11-14/I-character-screen.md (sections 5 and 10)`; `internal/invdesk/desk.go (`ItemView`, `canWear`, `noWearSlot`)`; `internal/leveler/playerequip.go`.

**You own:** `internal/invdesk/usable.go`, `internal/invdesk/usable_test.go`, `internal/invdesk/desk.go (the view build only)`, `internal/watch/testdata/itemusable/`, `docs/inventory-ui.md (a new section)`.

**Do not touch:** the equip behaviour itself; the page (C2); `internal/merc` rules beyond reading them.

**Already known (do not redo):** `canWear`, `noWear`, `noWearSlot`, base `reqStr/reqDex/reqLvl` and the mercenary's level, Str, Dex exist. d2d's mercenary rule ignores ethereal and requirements -x% (a noted gap in the fork).

## Do
1. `ItemView` gains `req {lvl, str, dex, class[]}` and `usable {hero: {ok, reasons[]}, merc: {ok, reasons[]}}`; a reason is `{code, stat, need, have, text}` with codes from the equip vocabulary (`class`, `requirements`, `identified`, `twohand`, `slot`, `nospace`, `dead`) plus `broken` and `merctype`. `merc` is present only when a mercenary is owned. Schema number raised; every old field unchanged.
2. One validity function: `usable` must be computed by the code the Equip button's check already runs (the candidate check behind `canWear` / `noWearSlot`); extend that code to return the structured reason and the numbers instead of writing a second rule. The `noWearSlot` texts gain the numbers ("needs 47 strength (you have 38)").
3. Mercenary: his level, strength and dexterity (with his gear, as the server counts) and his type's kinds from `merc.Allowed`, aligned with `docs/research/merc-gear.md` 2.3; where that doc says U (javelins, two-handed swords, axes) refuse and say so in the doc section.
4. Requirement modifiers (ethereal, requirements -x%, affix level requirement): report the number the server will enforce. Where the bot cannot know it, set the base number and add `"approx": true`; list each such case in the doc as U with the probe that would settle it.
5. Worn items get `usable` too (a worn item under its requirement is reported). Gear set Apply reports and skips items whose `usable.hero.ok` is false, never substitutes: add the reason to the set result that exists.

## Acceptance (also the card's boxes)
- The 21-row table of I.10 as a table test bound to real rows of the game tables (run with `D2_MPQ_DIR=$HOME/d2d-lab/data`), at least Sorceress, Paladin, Amazon, Barbarian and the four mercenary types; rows the tables cannot settle are listed, not guessed
- A property test: for every pack item, `usable.hero.ok` is true exactly when `canWear` is non-empty for a wearable kind (one rule, two views)
- View goldens for both classes plus a mercenary fixture; `inventory` compat goldens unchanged
- Set apply test: an item under its requirement is skipped with the reason and nothing replaces it

## Card (create first, then set it In Progress)
```sh
python3 -I tools/backlog/new_task.py "charreq: Item requirements and usability in the inventory view" --card charreq --owns internal/invdesk/usable.go --owns internal/invdesk/usable_test.go --owns internal/invdesk/desk.go --owns internal/watch/testdata/itemusable/ --owns docs/inventory-ui.md --component bot --kind feature --size M --assignee credit-runner \
  --ac "The 21-row table of I.10 as a table test bound to real rows of the game tables (run with `D2_MPQ_DIR=$HOME/d2d-lab/data`), at least Sorceress, Paladin, Amazon, Barbarian and the four mercenary types; rows the tables cannot settle are listed, not guessed" \
  --ac "A property test: for every pack item, `usable.hero.ok` is true exactly when `canWear` is non-empty for a wearable kind (one rule, two views)" \
  --ac "View goldens for both classes plus a mercenary fixture; `inventory` compat goldens unchanged" \
  --ac "Set apply test: an item under its requirement is skipped with the reason and nothing replaces it"
```

## Every bot card (binding)
- Rules: `AGENTS.md`, `docs/agents/rules-short.md`. Own worktree and branch, commit often, never push, no deploy, offline only (fake runner, `mem.Fake`, fixtures); write the live check as a checklist in the card.
- Builds are data: no skill id, class check or build literal in Go; every new rule is tested for the Sorceress and one other class.
- Priority ladder of `docs/inventory-actions.md` holds: safety first, then player commands, then chores, then roam. A player command never bypasses a safety rule. No gear decision by the bot (`docs/gear-policy.md`).
- A new watch route follows `internal/watch/repairapi.go`: a late-bound desk, GET on the viewer, POST on the answer listener with the same gates (origin list, token, JSON, small body, unknown fields refused), 503 `off` until bound, a `schema` field in the reply, goldens for both classes.
- Gates before queueing, from inside your worktree with no `cd` in the command:
  - small box, per package: `go vet ./<your packages>/... && go test -p 1 -count=1 ./<your packages>/...`; packages that need the game tables: `D2_MPQ_DIR=$HOME/d2d-lab/data go test -p 1 -count=1 ./<pkg>/...`
  - whole repo on the worker host: `labrun <worker> -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./... && go test -race -count=1 ./internal/watch/...'` (a result under 20 s did not run)
  - `gofmt -l .` empty, `python3 -I tools/contracts/check.py`, `python3 -I tools/backlog/check.py`, `go run ./cmd/leakcheck -mode fail`, `tools/ci/check.sh`
- Then `git fetch origin && git merge origin/main`, rerun the per-package gate, tick the card boxes, fill `result:` (say not verified live), and `~/code/d2lod-ops/tools/mergeq.sh add agent/<name> --card TASK-<n>`.
- Report under 200 words: what is built, the commands run, what is U, the live checklist, last commit.
