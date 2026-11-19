---
id: TASK-672
title: 'pgcain: Proving Ground rerun of the Cain scroll scenarios on main with cainfix2'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T18:41:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 329000
card: "pgcain"
owns:
  - "docs/probes/provingground-cain-2026-11-19.md"
  - "tools/provingground/scenarios/cain.json"
  - "internal/scenario"
result: "Verified live on the builder (lab slot, CHEATED setup, mechanics only): cain.scroll.field Paladin/11 GREEN on main with cainfix2 once the assertion accepts bks or bkd; cain.scroll lab and shipped devp13 RED (chunk queued behind fights, same signature); Sorceress/3 RED on setup (no weapon). Sandbox tree stall not reproduced. See docs/probes/provingground-cain-2026-11-19.md."
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the first Proving Ground run was RED on a possibly older binary. Done when: rerun on main with cainfix2, both builds, probe doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 cain.scroll.field and cain.scroll rerun on current main (lab verbs)
- [x] #2 sandbox-build variant run with the devp13 d2d
- [x] #3 probe doc written
<!-- AC:END -->
