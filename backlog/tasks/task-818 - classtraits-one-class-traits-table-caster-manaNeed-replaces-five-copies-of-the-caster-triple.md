---
id: TASK-818
title: 'classtraits: one class traits table (caster, manaNeed) replaces five copies of the caster triple'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T19:53:00Z'
labels:
  - bot
  - chore
dependencies: []
priority: medium
ordinal: 373000
card: "classtraits"
owns:
  - "internal/d2/classtraits.go"
  - "internal/d2/classtraits.json"
  - "internal/d2/classtraits_test.go"
  - "internal/bot/task.go"
  - "internal/bot/geargate.go"
  - "internal/bot/routeguard.go"
  - "internal/resource/types.go"
  - "internal/merc/policy.go"
exclusive: []
contracts: []
result: "internal/d2/classtraits.json (caster, manaNeed per class) with Class.Caster, Class.ManaNeed, TraitsOf and TraitsByName (unknown class = error naming it); bot task, geargate, routeguard (two places), resource.ManaNeed and merc.AuraPrefs use it. Offline tests only; behaviour same for all seven classes."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 caster and manaNeed read from classtraits.json; no class list in the five places; unknown class is an error naming it
- [x] #2 Same behaviour for all seven classes (test)
<!-- AC:END -->
