# P9: the generic step executor (card stepexec1, TASK-976)

> status: current (2026-11-26). Source: docs/research/fable-logic-review-2026-11-25.md 2.4 (classes F, H), 7.1 P9; ADR 0021 (questspec), docs/outcomes.md.

## What it is
One engine, `questspec.RunStep` (internal/questspec/exec.go), runs a step of the four verbs `interact`, `pickup`, `talk`, `use`:
APPROACH the target, ACT, WAIT for the declared fact (the step's `post` clauses, read from a fresh snapshot, never from the
executor's own memory), retry within the step's `tries` with doubling backoff. A post that already holds before the first act
is Done and touches nothing (resumption after a restart). V: unit tests `internal/questspec/exec_test.go`.

## The row is data
A step of `data/questspec/*.json` carries an optional `exec` block: `verb`, `on` (the thing acted on, by class id, never a
display name or a numeric literal in Go), `then` (a second verb per wait poll, e.g. pickup after the click), `tolerance`,
`wait_polls`, `wait_ms`, `backoff_ms`, `backoff_cap_ms`, `wake`. `Validate` rejects a bad verb, no post or no tries.

## Outcome
`questspec.Outcome` has the five kinds of internal/bot/outcome.go (Done, Progress, Refused{need, wake, why}, Failed{err},
Interrupted{err}). questspec cannot import bot, so the kinds are mirrored and the bot adapter maps them; this avoids
touching the hot outcome.go. Progress (the backend moved but has not arrived) does not use a try. A refusal is data, not a
failure. A backend error that is not one of the typed ones is a failed attempt, retried in budget.

## Adapter
`internal/bot/stepexec.go`: Backend over `TaskDeps.Objects`, `Deps.Loot` and the pack scan; one call site in
`taskActs.Interact`, behind `quests.stepExec` (default 0, `[dark-launch]`, ledger row in data/darklaunch.json).
Approach is a no-op there: the object and dialog adapters walk themselves (U: a separate approach verb is for chunks
whose adapter does not).

## Migrated
Cain scroll `search_for_cain/take_scroll` (Tree of Inifuss click and pick-up). Scenario `cain.scroll.stepexec` (Paladin and
Sorceress, same assertions as `cain.scroll`). Not run live or on the lab yet (U).

Card stepexec2: the TALK and USE adapters (`execBackend.Act`). TALK says the step's `exec.pick` to the NPC named from the class id
(the inverse of `talk.ByAct`, so no name is written in Go) through `TaskDeps.Dialog`; `taskActs.Talk` takes the executor path when
`quests.stepExec` is 1 and the data has a talk step with an `exec` row for that NPC class and pick (`execTalkStep`, data-driven).
USE right-clicks the named pack item through the object adapter (`UseItem`); no chunk uses it yet. The second migrated chunk is
Akara reading the bark: `search_for_cain/read_scroll` (pick `bark`, post `has_item(bkd)`, read from the pack), scenario
`cain.read.stepexec` (Paladin and Sorceress, same assertions as `cain.read`). V: unit tests `internal/bot/stepexec_test.go` (two
classes, must-fail when the dialog says yes and the pack keeps the scroll). U: not run on the lab. `Backend.Act` gained the `pick`
argument.

## Rule (acceptance criterion 2)
No new quest chunk may add a hand-written click-wait-pick-up handler: a new step of these four verbs is a questspec row with
an `exec` block, run by `RunStep`. Old handlers move over one at a time, each behind its scenario and a knob; the old code is
deleted after a 24 h watch with auto-revert (docs/process/dark-launch.md). Next in line: Duriel `place_staff`
(epic-act2), Akara's bark read (`talk`), the cairn stones.
