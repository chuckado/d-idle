---
id: TASK-940
title: 'epic-act5: Act V complete in Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:03:00Z'
labels:
  - bot
  - feature
  - T4
  - gate
dependencies: []
priority: medium
ordinal: 441000
card: "epic-act5"
owns:
  - "docs/epics/epic-act5.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Act V epic (Normal). Quests, bosses and areas below; the living status is docs/epics/epic-act5.md. Bosses: Shenk, Anya (Prison of Ice), Nihlathak, the Ancients, Baal. Blockers: Siege: the hero leaves Harrogath through the gate, reward bit not set (`drill.act5.siege` RED); Harrogath gate and Bloody Foothills nav card; Prison of Ice reward bit not set (`drill.act5.prison` RED); Ancients: statues in view, fight never starts (`drill.act5.ancients` BLOCKED); Betrayal and Baal RED (`drill.act5.betrayal`, `drill.act5.baal`); Nihlathak portal object class 60 not found; Quest words of Acts 3 to 5 are not observable in the lab (matrix note 6, section 8). Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every quest of Act V is defined in the quest board data (data/questboard/act5.json) with its steps mapped; gaps listed in docs/epics/epic-act5.md. Evidence: act5.json, step.act5.json verdicts.
- [ ] #2 Every Act V quest chunk is GREEN in the Proving Ground drill matrix (docs/probes/quest-drill-matrix-2026-11-23.md, data/stepgrounds/verdicts/drill.act5*.json, step.act5.json), 2 of 3 runs each.
- [ ] #3 Live end to end: a fresh hero (no cheats, no memory writes, provenance rule) completes Act V in Normal on a hidden/sandbox instance (Questy test bed style) with the bot only; record `writes after start: 0`.
- [ ] #4 No open blocker: every item in the card's Blockers list is closed or has a merged fix with a GREEN drill. Current blockers: Siege: the hero leaves Harrogath through the gate, reward bit not set (`drill.act5.siege` RED); Harrogath gate and Bloody Foothills nav card; Prison of Ice reward bit not set (`drill.act5.prison` RED); Ancients: statues in view, fight never starts (`drill.act5.ancients` BLOCKED); Betrayal and Baal RED (`drill.act5.betrayal`, `drill.act5.baal`); Nihlathak portal object class 60 not found; Quest words of Acts 3 to 5 are not observable in the lab (matrix note 6, section 8)
- [ ] #5 Readiness floors and danger floors defined for every dangerous leg of Act V (plandata early routes, route guard, boss readiness) and the potion/kit economy holds with no death spiral in a 3 h Questy-style run.
- [ ] #6 Docs: an act page in docs/ (docs/act5.md) with the quest map and known issues, current with the board data.
- [ ] #7 Works for the default Sorceress and at least one other class (build flexibility rule); drill matrix rows exist for both.
<!-- AC:END -->
