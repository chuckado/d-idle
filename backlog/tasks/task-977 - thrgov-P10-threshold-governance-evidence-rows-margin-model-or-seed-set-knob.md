---
id: TASK-977
title: 'thrgov: P10 threshold governance: evidence rows, margin model or seed set, knob'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:55:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 465000
card: "thrgov"
owns:
  - "docs/process/threshold-governance.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Cause class killed: G (thresholds moved by one incident, 20) and the knob count (560 rows, +58 in six days). Roadmap row P10 (docs/research/fable-logic-review-2026-11-25.md 2.3, 7.1). Process rule: (a) a floor or line changes only with the margin model's number or a seed set, never one death; (b) rows carry `evidence: n`; (c) a fix card may not add a knob without removing one or the owner's word; (d) later: safety and supply thresholds as rows in the loot-rule grammar. Doc plus a check in tools/ (knob added without a removed one or an owner word fails) wired by the implementing agent. Does not touch docs/decisions.md or ADRs; if the owner wants it as an ADR, the coordinator records it. Risk: low. Enable criteria: check in CI report-only for one session, then failing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rule doc with the evidence-row schema and the knob budget rule
- [ ] #2 CI check (report-only first) for knob adds without a removal or owner word
<!-- AC:END -->
