# epic-act2: Act II complete in Normal

> status: OPEN epic, written 2026-11-25. Tags: V verified (named run or file), S secondary, U unverified. Lab setups are CHEATED (stand-in level, gear, words): mechanics only, never evidence for the live hero.

Card: `epic-act2` (backlog/tasks). Parent bar: ADR 0016 T4 Questor and the Done-means bullet in docs/decisions.md (Current truth). Numbers are not copied here; follow the links.

## Quests (V data/questboard/act2.json)

| id (data/questboard/act2.json) | quest | kind | chunks | boss |
|---|---|---|---|---|
| `radament` | Radament's Lair | optional | radament | radament |
| `tainted_sun` | The Tainted Sun | story | amulet | fangskin |
| `horadric_staff` | The Horadric Staff | story | cube, shaft, assemble | - |
| `arcane_sanctuary` | The Arcane Sanctuary | optional | portal | - |
| `summoner` | The Summoner | optional | kill_summoner | summoner |
| `seven_tombs` | The Seven Tombs | story | duriel | duriel |
| `act2_done` | Sail east to Act 3 | unlock | sail_east | - |

## Status (S, from the verdict files)

Drill verdicts today: `drill.act2.*` 6/9 GREEN, `step.act2.*` 2/6 GREEN (data/stepgrounds/verdicts/, lab only, every setup CHEATED).

Acceptance 1 (quests defined with steps): V for the rows above, the board carries chunks and step lists for each; known gap list is in the Blockers section.

## Blockers (the real current REDs)

- Duriel: place_staff needs a hover-scan verb the backend lacks (`drill.act2.duriel` RED, matrix notes 40, 66)
- Act 2 flaky fight rows (Radament l30, amulet/Fangskin, Palace Cellar/Blunderbore, Sewers 1 Horror): same commit GREEN and RED (matrix 0.)
- Arcane Sanctuary teleport pads (matrix notes 5, 39). Act 2 level plan: DATA landed 2026-11-26 (docs/plan/act2-level-plan.md, `internal/plandata/act2plan.json`, road rows `act2_*`; entries SME-ballpark U); still missing: areaxp rows for Act 2 (lab AREAMON run) and Act 2 chapters in the spine
- Tomb doors and Maggot Lair 1 arrival nook (nav cards, matrix note 9)

## Evidence pages

- docs/probes/quest-drill-matrix-2026-11-23.md
- data/stepgrounds/verdicts/ (drill.act2*.json, step.act2.json)
- docs/act2.md

## Not yet proven (U)

Live fresh-hero run of this act with writes after start 0; readiness floors per leg; 3 h economy run; second class. Update this page when a card lands, do not hand-count here.
