---
id: TASK-421
title: 'lootrules: one rules engine for item decisions (design + prototype)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T15:20:00Z'
labels:
  - bot
  - research
dependencies: []
priority: medium
ordinal: 254000
card: "lootrules"
owns:
  - "docs/lootrules.md"
  - "internal/lootrules/**"
exclusive: []
contracts: []
result: "Design in docs/lootrules.md plus a stdlib-only prototype internal/lootrules (strict JSON rules, 3-valued conditions, requirement predicates for hero and merc, floors, Plan, Preview/Diff/Shadowed, Starter preset). Starter pack rules equal tidy.Decide on 432 decisions (Paladin, Sorceress, Amazon); ground half translated by hand, table test vs the real pickit needs game tables (card L2). Not wired, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the user wants a configurable, unlockable loot filter; today four tables decide about items and disagree at the edges. Done when: the design doc and a tested prototype exist. Consumer: loot-rules wiring cards L2 to L7 (docs/lootrules.md section 9).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/lootrules.md: fact model, rule language, migration, unlock, API, wiring cards
- [x] #2 internal/lootrules prototype: parse, validate, evaluate, Starter equals tidy policy on a two-class table, invariants, benchmark
- [x] #3 checks: gofmt, vet, test -p 1, contracts, backlog, leakcheck
<!-- AC:END -->
