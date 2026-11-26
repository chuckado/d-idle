---
id: TASK-936
title: 'epic-act1: Act I complete in Normal'
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
ordinal: 437000
card: "epic-act1"
owns:
  - "docs/epics/epic-act1.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Act I epic (Normal). Quests, bosses and areas below; the living status is docs/epics/epic-act1.md. Bosses: Blood Raven, The Countess, The Smith, Andariel. Blockers: Smith stand click (`step.act1.smith.stand_click_pickup` RED, matrix note 68); Tools of the Trade reward word 0x204e: the Charsi reward talk does not complete (`drill.tools.merc17` RED, `necro.act1-tools` RED, matrix note 70); Andariel live margin: a clvl 21 hero (life 280, merc 17) is refused as too hard; kill needs life 600 or the bold dial (`drill.andariel.live21*`, matrix note 69). Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every quest of Act I is defined in the quest board data (data/questboard/act1.json) with its steps mapped; gaps listed in docs/epics/epic-act1.md. Evidence: act1.json, step.act1.json verdicts.
- [ ] #2 Every Act I quest chunk is GREEN in the Proving Ground drill matrix (docs/probes/quest-drill-matrix-2026-11-23.md, data/stepgrounds/verdicts/drill.act1*.json, step.act1.json), 2 of 3 runs each.
- [ ] #3 Live end to end: a fresh hero (no cheats, no memory writes, provenance rule) completes Act I in Normal on a hidden/sandbox instance (Questy test bed style) with the bot only; record `writes after start: 0`.
- [ ] #4 No open blocker: every item in the card's Blockers list is closed or has a merged fix with a GREEN drill. Current blockers: Smith stand click (`step.act1.smith.stand_click_pickup` RED, matrix note 68); Tools of the Trade reward word 0x204e: the Charsi reward talk does not complete (`drill.tools.merc17` RED, `necro.act1-tools` RED, matrix note 70); Andariel live margin: a clvl 21 hero (life 280, merc 17) is refused as too hard; kill needs life 600 or the bold dial (`drill.andariel.live21*`, matrix note 69)
- [ ] #5 Readiness floors and danger floors defined for every dangerous leg of Act I (plandata early routes, route guard, boss readiness) and the potion/kit economy holds with no death spiral in a 3 h Questy-style run.
- [ ] #6 Docs: an act page in docs/ (docs/act1check.md, docs/boss.md, docs/act-plan.md) with the quest map and known issues, current with the board data.
- [ ] #7 Works for the default Sorceress and at least one other class (build flexibility rule); drill matrix rows exist for both.
<!-- AC:END -->
