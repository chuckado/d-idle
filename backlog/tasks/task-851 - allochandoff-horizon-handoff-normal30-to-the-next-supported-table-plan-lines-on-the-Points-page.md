---
id: TASK-851
title: 'allochandoff: horizon handoff normal30 to the next supported table, plan lines on the Points page'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T04:17:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 386000
card: "allochandoff"
owns:
  - "internal/points"
  - "cmd/d2bot/autoallocwire_test.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/web/points.js"
  - "docs/autoalloc.md"
  - "tools/provingground/scenarios/autoalloc.json"
exclusive: []
contracts: []
result: "Horizon handoff built: a locked table continues with the next SUPPORTED table of the same catalogue build (fromLevel, or when exhausted), logged, stored (plan, planAtLevel) and shown on the Points page; never into an experimental table. Unit tests Necromancer + Paladin, race x20 ok; catalogue test lists builds with no supported plan (lightning, poison nova, frenzy, tornado) with reasons. Verified live on the builder (lab slot, CHEATED setup): points.autoalloc-handoff and points.autoalloc-necro-clvl4 GREEN (they assert the handoff and at least one spend, the allocator paces about one point per calm 3 s)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Handoff to the next supported table, logged and stored, tests on two classes
- [x] #2 Every catalogue build has a plan or a stated reason (test)
- [x] #3 Points page plan lines
- [x] #4 Scenarios points.autoalloc-handoff and points.autoalloc-necro-clvl4 run on the builder (live check)
<!-- AC:END -->
