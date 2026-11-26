---
id: TASK-938
title: 'epic-act3: Act III complete in Normal'
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
priority: medium
ordinal: 439000
card: "epic-act3"
owns:
  - "docs/epics/epic-act3.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Act III epic (Normal). Quests, bosses and areas below; the living status is docs/epics/epic-act3.md. Bosses: Mephisto (Durance of Hate), Council members, Flayer Jungle packs. Blockers: Great Marsh and Flayer Jungle fights and timeouts (`drill.act3.lam_esen`, `.blade` RED; matrix notes 24, 25, 41); Golden Bird chunk parked by the board (`drill.act3.golden_bird` BLOCKED); `drill.act3s.golden_bird` GREEN only with pre.rooms_up; Blade of the Old Religion reward word not set (`drill.act3s.blade` RED); Guardian/Mephisto travel to Durance 2 and 3 RED (`step.act3.guardian.mephisto.*`); Act 3 level plan missing (no areaxp rows, matrix note 7). Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every quest of Act III is defined in the quest board data (data/questboard/act3.json) with its steps mapped; gaps listed in docs/epics/epic-act3.md. Evidence: act3.json, step.act3.json verdicts.
- [ ] #2 Every Act III quest chunk is GREEN in the Proving Ground drill matrix (docs/probes/quest-drill-matrix-2026-11-23.md, data/stepgrounds/verdicts/drill.act3*.json, step.act3.json), 2 of 3 runs each.
- [ ] #3 Live end to end: a fresh hero (no cheats, no memory writes, provenance rule) completes Act III in Normal on a hidden/sandbox instance (Questy test bed style) with the bot only; record `writes after start: 0`.
- [ ] #4 No open blocker: every item in the card's Blockers list is closed or has a merged fix with a GREEN drill. Current blockers: Great Marsh and Flayer Jungle fights and timeouts (`drill.act3.lam_esen`, `.blade` RED; matrix notes 24, 25, 41); Golden Bird chunk parked by the board (`drill.act3.golden_bird` BLOCKED); `drill.act3s.golden_bird` GREEN only with pre.rooms_up; Blade of the Old Religion reward word not set (`drill.act3s.blade` RED); Guardian/Mephisto travel to Durance 2 and 3 RED (`step.act3.guardian.mephisto.*`); Act 3 level plan missing (no areaxp rows, matrix note 7)
- [ ] #5 Readiness floors and danger floors defined for every dangerous leg of Act III (plandata early routes, route guard, boss readiness) and the potion/kit economy holds with no death spiral in a 3 h Questy-style run.
- [ ] #6 Docs: an act page in docs/ (docs/act3.md) with the quest map and known issues, current with the board data.
- [ ] #7 Works for the default Sorceress and at least one other class (build flexibility rule); drill matrix rows exist for both.
<!-- AC:END -->
