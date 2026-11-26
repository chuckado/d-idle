---
id: TASK-980
title: 'proptest1: T3 property tests on the pure deciders (combat.Decide, supply.Solve,'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:56:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: low
ordinal: 468000
card: "proptest1"
owns:
  - "docs/design/proptest-t3.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Cause classes: B, C, G at unit-test cost. Roadmap row T3 (docs/research/fable-logic-review-2026-11-25.md 5.3). Random-input properties: a reflex always wins; more potions never worsen the verdict; no intent starves; a refusal's need, once met, admits. One package per PR (combat, supply, arbiter first; admission after admit1). Two builds as input generators. Risk: none. Enable criteria: tests in CI.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Properties for combat.Decide, supply.Solve and arbiter.Decide with two builds
<!-- AC:END -->
