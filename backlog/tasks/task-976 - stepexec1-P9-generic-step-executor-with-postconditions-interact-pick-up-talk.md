---
id: TASK-976
title: 'stepexec1: P9 generic step executor with postconditions (interact, pick up, talk'
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
ordinal: 464000
card: "stepexec1"
owns:
  - "docs/design/stepexec-p9.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Cause classes killed: H (per-quest special case, 45) and F (a fact the backend does not give). Roadmap row P9 (docs/research/fable-logic-review-2026-11-25.md 7.1; ADR 0021, 0022). One executor: approach, act, wait for the declared fact, bounded retries, progress is not failure; identity by id; new chunks may use ONLY it. Replaces cainNodes-style handlers one by one, each behind the scenario of its chunk. Declaration side exists in internal/questspec. Unblocks the Act II epic card (epic-act2: Duriel place_staff needs a verb path) and later act epics. Depends on P2. Risk: medium; one chunk per card. Enable criteria per chunk: its Proving Ground scenario green 2 of 3 runs with the executor on; old handler deleted after 7 days green.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Executor skeleton plus the Cain scroll chunk ported behind a knob with a ledger row
- [ ] #2 Rule in the card doc: no new chunk without the executor
<!-- AC:END -->
