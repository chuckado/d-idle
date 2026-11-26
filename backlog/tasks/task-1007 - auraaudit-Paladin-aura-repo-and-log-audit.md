---
id: TASK-1007
title: 'auraaudit: Paladin aura repo and log audit'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T17:55:00Z'
updated_date: '2026-11-26T18:30:00Z'
labels:
  - bot
  - research
dependencies: []
priority: medium
ordinal: 482000
card: "auraaudit"
lane: next
owns:
  - "docs/probes/paladin-aura-audit-2026-11-26.md"
exclusive: []
contracts: []
result: "Audit written, not verified live: the d2d bot confirms an aura only by the selected right skill, never by the aura state; logs show Might and Holy Fire picks but cannot show an active aura; probe proposed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the owner suspects the Paladin aura is never active on d2d. Done when: the audit doc lists the flow, live-log evidence, defects and a read-only probe. Consumer: sme-aura (docs/sme/paladin-auras-2026-11-26.md) and the aura fix card.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Audit doc written with flow, live-log evidence, merc, ranked defects, probe
<!-- AC:END -->
