---
id: TASK-713
title: 'fablecombat: design review of the combat engine and skill delivery layer'
status: In Progress
assignee:
  - pro-runner
created_date: '2026-11-21T03:15:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 340000
card: "fablecombat"
owns:
  - "docs/design/fable-review-2026-11-21/combat-engine*"
exclusive: []
contracts: []
result: "Design review written: docs/design/fable-review-2026-11-21/combat-engine.md plus appendices a (engine today, file:line), b (delivery grammar), c (tests, read verbs, differentials), d (16 proposed cards and one parked). Docs only; NOT verified live (code read at 5b12ce5e7, no game, no worker job). Nothing build-specific was added; the doc lists the class and skill literals the design removes. 11 open questions for the owner in section 6."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fable design review (user asked 2026-11-20). Docs only, no Go or behaviour change. Done when the doc and appendices are merged.

Consumer: roadmap packages R10, R11, R12, R16 (docs/design/fable-review-2026-11-19/06-roadmap.md) and the proposed dlv-* cards of docs/design/fable-review-2026-11-21/combat-engine-d-cards.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Map of what the engine supports today per skill kind, with file:line and V/U tags
- [x] #2 Data-driven delivery grammar and engine architecture, zero class or skill literals
- [x] #3 Test strategy: simulator limits, Proving Ground combat scenarios, read verbs, differentials
- [x] #4 Prioritised rollout of 8 to 20 cards with owner paths, risk and effort
- [x] #5 Numbered open questions with a recommendation each
<!-- AC:END -->
