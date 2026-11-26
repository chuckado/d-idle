---
id: TASK-979
title: 'faultsoak: P12 fault-injection soak on lockstep (T5)'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:56:00Z'
labels:
  - bot
  - drill
dependencies:
  - TASK-967
  - TASK-968
priority: medium
ordinal: 467000
card: "faultsoak"
owns:
  - "docs/design/faultsoak-p12.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Cause classes killed: E, F and the recovery paths nobody walks on purpose. Roadmap row P12 and T5 (docs/research/fable-logic-review-2026-11-25.md 5.3, 7.1). Soak with injections on lockstep: empty the belt, zero the purse, drop life, fail the portal, time out the control socket (14 such timeouts in the quest log), remove an NPC from the unit list; P1 monitors are the oracle. Injections are lab verbs of the fork (ADR 0014 DIAGNOSIS; never in a measured run; recorded in the ledger). Night window ADR 0027 for T3/T4 runs. Risk: low (lab only). Enable criteria: soak runs nightly report-only; each hit becomes an incident1 scenario.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Injection list and the monitor oracle documented; first soak run recorded with writes ledger
<!-- AC:END -->
