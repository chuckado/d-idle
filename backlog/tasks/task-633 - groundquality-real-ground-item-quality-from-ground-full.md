---
id: TASK-633
title: 'groundquality: real ground item quality from ground full'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T14:28:00Z'
labels:
  - d2d
  - bug
dependencies: []
priority: medium
ordinal: 318000
card: "groundquality"
owns:
  - "internal/d2dloot"
  - "cmd/d2bot/d2dloot.go"
  - "cmd/d2bot/d2dwire.go"
  - "docs/d2d-loot.md"
exclusive: []
contracts: []
result: "Loot classifier reads the real quality, ilvl and identified flag via ground full (feature-detected, cached, one call per pass); without the verb rings, amulets, jewels and charms are unknown, not normal. Offline tests only; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: unidentified magic/rare ground items were guessed normal from the label and hit g-junk. Done when: the classifier uses the real quality where the verb exists.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Loot classifier uses real quality from ground full when the verb exists (cached per id, one call per pass, gear/jewelry/charms only)
- [x] #2 Without the verb rings, amulets, jewels and charms are never normal (quality unknown)
- [x] #3 Offline tests incl. must-fail cases pass with -race -count=20
<!-- AC:END -->
