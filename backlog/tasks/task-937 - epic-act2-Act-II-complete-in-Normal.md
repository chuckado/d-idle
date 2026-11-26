---
id: TASK-937
title: 'epic-act2: Act II complete in Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:02:00Z'
labels:
  - bot
  - feature
  - T4
  - gate
dependencies: []
priority: high
ordinal: 438000
card: "epic-act2"
owns:
  - "docs/epics/epic-act2.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Act II epic (Normal). Quests, bosses and areas below; the living status is docs/epics/epic-act2.md. Bosses: Radament, Fangskin, Beetleburst-class tomb packs, The Summoner, Duriel. Blockers: Duriel: place_staff needs a hover-scan verb the backend lacks (`drill.act2.duriel` RED, matrix notes 40, 66); Act 2 flaky fight rows (Radament l30, amulet/Fangskin, Palace Cellar/Blunderbore, Sewers 1 Horror): same commit GREEN and RED (matrix 0.); Arcane Sanctuary teleport pads and Act 2 level plan (no areaxp rows; matrix notes 5, 39); Tomb doors and Maggot Lair 1 arrival nook (nav cards, matrix note 9). Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every quest of Act II is defined in the quest board data (data/questboard/act2.json) with its steps mapped; gaps listed in docs/epics/epic-act2.md. Evidence: act2.json, step.act2.json verdicts.
- [ ] #2 Every Act II quest chunk is GREEN in the Proving Ground drill matrix (docs/probes/quest-drill-matrix-2026-11-23.md, data/stepgrounds/verdicts/drill.act2*.json, step.act2.json), 2 of 3 runs each.
- [ ] #3 Live end to end: a fresh hero (no cheats, no memory writes, provenance rule) completes Act II in Normal on a hidden/sandbox instance (Questy test bed style) with the bot only; record `writes after start: 0`.
- [ ] #4 No open blocker: every item in the card's Blockers list is closed or has a merged fix with a GREEN drill. Current blockers: Duriel: place_staff needs a hover-scan verb the backend lacks (`drill.act2.duriel` RED, matrix notes 40, 66); Act 2 flaky fight rows (Radament l30, amulet/Fangskin, Palace Cellar/Blunderbore, Sewers 1 Horror): same commit GREEN and RED (matrix 0.); Arcane Sanctuary teleport pads and Act 2 level plan (no areaxp rows; matrix notes 5, 39); Tomb doors and Maggot Lair 1 arrival nook (nav cards, matrix note 9)
- [ ] #5 Readiness floors and danger floors defined for every dangerous leg of Act II (plandata early routes, route guard, boss readiness) and the potion/kit economy holds with no death spiral in a 3 h Questy-style run.
- [ ] #6 Docs: an act page in docs/ (docs/act2.md) with the quest map and known issues, current with the board data.
- [ ] #7 Works for the default Sorceress and at least one other class (build flexibility rule); drill matrix rows exist for both.
<!-- AC:END -->
