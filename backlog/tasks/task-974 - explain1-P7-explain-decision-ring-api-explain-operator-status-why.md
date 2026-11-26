---
id: TASK-974
title: 'explain1: P7 explain: decision ring, /api/explain, operator status --why'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:55:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-968
priority: medium
ordinal: 462000
card: "explain1"
owns:
  - "docs/design/explain-p7.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Cause class killed: K (not visible) and operator controllability. Roadmap row P7 (docs/research/fable-logic-review-2026-11-25.md 4.3, 7.1, 7.2 PR 5). Ring of the last 200 decisions (tick time, chosen hook family, reason code, standing refusals with needs); GET /api/explain; operator `status --why`; later a 'Why not?' row per chunk on the watch page. OWNER DECISION NEEDED: the ADR 0030 allowlist gains a read verb (review 7.5); do not start the operator part before the owner says yes. Risk: none (read only). Enable criteria: no knob; endpoint answers on both sandboxes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Ring and /api/explain on the sandboxes
- [ ] #2 Operator status --why only after the owner's word on the ADR 0030 allowlist
<!-- AC:END -->
