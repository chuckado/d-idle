---
id: TASK-939
title: 'epic-act4: Act IV complete in Normal'
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
ordinal: 440000
card: "epic-act4"
owns:
  - "docs/epics/epic-act4.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Act IV epic (Normal). Quests, bosses and areas below; the living status is docs/epics/epic-act4.md. Bosses: Izual, Hephasto, the three seal bosses, Diablo. Blockers: Izual reward bit not set (`drill.act4.izual` RED, word 0x000c); Hellforge: no travel or Hammer step ran (`drill.act4.hellforge` RED); Seals: the stand-in dies to De Seis, Infector and Vizier (kit too weak; `step.act4.terrors_end.seals.*` RED); River of Flame and Chaos Sanctuary travel RED (the hero dies on the way). Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every quest of Act IV is defined in the quest board data (data/questboard/act4.json) with its steps mapped; gaps listed in docs/epics/epic-act4.md. Evidence: act4.json, step.act4.json verdicts.
- [ ] #2 Every Act IV quest chunk is GREEN in the Proving Ground drill matrix (docs/probes/quest-drill-matrix-2026-11-23.md, data/stepgrounds/verdicts/drill.act4*.json, step.act4.json), 2 of 3 runs each.
- [ ] #3 Live end to end: a fresh hero (no cheats, no memory writes, provenance rule) completes Act IV in Normal on a hidden/sandbox instance (Questy test bed style) with the bot only; record `writes after start: 0`.
- [ ] #4 No open blocker: every item in the card's Blockers list is closed or has a merged fix with a GREEN drill. Current blockers: Izual reward bit not set (`drill.act4.izual` RED, word 0x000c); Hellforge: no travel or Hammer step ran (`drill.act4.hellforge` RED); Seals: the stand-in dies to De Seis, Infector and Vizier (kit too weak; `step.act4.terrors_end.seals.*` RED); River of Flame and Chaos Sanctuary travel RED (the hero dies on the way)
- [ ] #5 Readiness floors and danger floors defined for every dangerous leg of Act IV (plandata early routes, route guard, boss readiness) and the potion/kit economy holds with no death spiral in a 3 h Questy-style run.
- [ ] #6 Docs: an act page in docs/ (docs/act4.md) with the quest map and known issues, current with the board data.
- [ ] #7 Works for the default Sorceress and at least one other class (build flexibility rule); drill matrix rows exist for both.
<!-- AC:END -->
