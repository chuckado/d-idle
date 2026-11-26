---
id: TASK-1014
title: 'aura2d: keep the hired merc aura for the merc line and status'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T18:39:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 485000
card: "aura2d"
lane: next
owns:
  - "cmd/d2bot/mercwire.go"
  - "cmd/d2bot/mercwire_test.go"
exclusive: []
result: "Unit-tested only, not live. A merc hired by the bot keeps its aura in the merc line and status; an inherited merc (the free Rogue, a hire by hand, a restart) still shows none: the aura is not readable on d2d (U)."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the audit found the merc aura unread on d2d (merc.aura not in the dump, U). Done when: the one source that exists, the hire window text of a merc this process hired, reaches MercState.Aura, the minute merc line and the status mount.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The merc state carries the aura of the merc this process hired (hire window text); other classes get none
- [x] #2 Memory reading of the merc aura stays unverified (not in the d2d dump)
<!-- AC:END -->
