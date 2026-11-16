---
id: TASK-550
title: 'survab: survival A/B harness, report generator, R3 drill plan, approval memo'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T16:30:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 290000
card: "survab"
owns:
  - "tools/survival/"
  - "docs/survival/ab-harness.md"
exclusive: []
contracts: []
result: "docs/survival/ab-harness.md (audit: S-A..S-C done, S-D bot.zoneModel MISSING; lockstep-slot A/B design, gate wording decision, R3 plan, approval memo) and tools/survival/{abreport,abrun.sh,ab.sh}; abreport tested on synthetic logs, scripts only bash -n; NOT verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Why: the S-D gate needs a paired A/B harness and report. Done when: design doc, report tool with tests, R3 plan, memo.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/survival/ab-harness.md: audit S-A..S-D, harness design, R3 plan, memo
- [x] #2 tools/survival/abreport with unit tests on synthetic minute logs
<!-- AC:END -->
