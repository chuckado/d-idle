# Incident to scenario (report-only tier)

> status: current (card incident1, Fable logic review 2026-11-25 T2 and P8; ADR 0020 amended 2026-11-25). TOOLS ONLY: nothing here is in the merge gate. Tags: V = a test in this branch, U = unverified.

## What it is
At each death the bot writes the recorded state of that moment next to its death log: `<state dir>/incidents/<unix>-death.json` (`death-bug` when the autopsy hit its potion-bug test). The newest 50 files are kept. The code is `internal/bot/incident.go`, called from `recordDeath` (V `TestDeathWritesIncident`, two classes). It only writes a file; it changes nothing the hero does. Monitor hits (invariant monitor, review P1) join the same writer when that card lands: it is one function.

The file carries the scenario `pre` (start field, area id, level, gold, stash, difficulty, life percent at the start of the 60 s window) and the facts a `pre` cannot hold (belt and pack potion counts, portal scrolls, hostiles near, killer, summary, bug flag). It is tagged `SETUP`: a RECORDED state. Under the test-character provenance rule a row built from it tests mechanics from that state and is never evidence about game behaviour, balance or parity.

## Replay as a regression row
```
go run ./cmd/d2scenario incident [-class C] [-seed N] [-out DIR] FILE.json...
go run ./cmd/d2scenario plan -tier nightly -dir DIR          # dry run, no game
go run ./cmd/d2scenario run -tier nightly -dir DIR -out RES  # on the lockstep slot of the builder (not done by agents)
```
`incident` turns each file into a nightly behaviour scenario (`internal/scenario/incident.go`, V `TestIncidentScenario`): the hero starts in the area with the recorded level, gold, belt (lowest potion tier: the tier is not recorded, U) and life; the bot plays the campaign goal for 15 virtual minutes; the row is red when `character died` appears again (and, for a potion-bug death, when a `BUG: ... healing potions in the belt` autopsy line appears), plus the standing `no_writes`. The units that killed the hero are not recorded, so a row proves the state of the hero, not the situation; add `pre.spawn` by hand to build it (the three rows below do).

## The first three rows (`tools/provingground/incidents/questy-2026-11-25.json`)
Source: Fable logic review 2026-11-25 section 3.1 and 3.4 (V log lines quoted there). Level 8 and the gold are U (the log fixes the area, the belt and the pack). Both Sorceress and Paladin are in the matrix.

| row | the death | what the row needs to go green |
|---|---|---|
| `incident.questy-0251-chores-cold-plains` | 02:51Z Cold Plains, trip in its chores stage, 12 hostiles, 3 belt potions, `lastSafety=chicken`, `bug=true` | the reflex lane (review P3 c, card reflex1): a drink without cancelling the action. The trip stage cannot be declared (U), so the row proves the reflex, not the trip |
| `incident.questy-0319-stony-field-no-potions` | 03:19Z Stony Field quest walk with 0 potions | the early road and kit gate (P4, ADR 0025): refuse the step with a need, do not walk it |
| `incident.questy-0056-dark-wood-treehead` | 00:56Z Dark Wood, town portal return next to the Treehead pack, belt 0 | leave a unique pack with an empty belt; the pack members and arrival cell are U (a super unique stands beside the hero) |

V: the file loads and plans for both classes (`go run ./cmd/d2scenario plan -tier nightly -dir tools/provingground/incidents`, `TestQuestyIncidentRows`). U: no lockstep run was done (no game in this card); the expected colour of each row (RED before its fix) is a prediction. U: the super unique name `treehead` must exist in the fork's `debug spawn super` list.

## Tier status
Report-only. The rows live in their own directory, so `d2scenario list/plan/validate` on the default directory and the merge gate never see them. Moving them into the nightly run, then the gate, needs the owner's decision (ADR 0020 amendment) after they have run for a stretch.

## Ready-to-run checklist for the builder (live)
1. `d2scenario run -tier nightly -dir tools/provingground/incidents -out RES` on the builder in its night window (docs/process/proving-ground.md); read `RES/results.json` with `d2scenario report`.
2. Record each hero in docs/probes/test-characters.md as SETUP/CHEATED position (the run declares it).
3. After reflex1 and the kit gate merge, the first and second rows should turn green; if not, the row is the failing evidence for the card.
