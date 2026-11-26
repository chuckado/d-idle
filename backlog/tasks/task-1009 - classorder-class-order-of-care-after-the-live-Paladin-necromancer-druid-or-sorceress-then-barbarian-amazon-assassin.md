---
id: TASK-1009
title: 'classorder: class order of care after the live Paladin'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T18:03:00Z'
updated_date: '2026-11-26T18:30:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 483000
result: "Class order of care set 2026-11-26 by the user: 1 necromancer, 2a druid or 2b sorceress (either), then 3 barbarian, 4 amazon, 5 assassin. Written as dated priority lines in roadmap rows R10 to R12 and R16 and in six class epics. Docs only; nothing live (not verified live)."
card: "classorder"
lane: next
owns:
  - "docs/roadmap-fable.md"
  - "docs/epics/epic-class-necromancer.md"
  - "docs/epics/epic-class-druid.md"
  - "docs/epics/epic-class-sorceress.md"
  - "docs/epics/epic-class-barbarian.md"
  - "docs/epics/epic-class-amazon.md"
  - "docs/epics/epic-class-assassin.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user decision 2026-11-26 sets the class order of care after the live Paladin. Done when: the dated priority lines are in roadmap rows R10 to R12 and R16 and in the six class epics, and the checkers pass.

Decision (user, 2026-11-26): 1) Necromancer; 2) Druid or Sorceress (either); 3) then by effort from repo evidence (plan rows, profiles, drill rows, named blockers): Barbarian (3 plan rows, 2 profiles, melee path, no blocker beyond drill rows), Amazon (3 plan rows, 1 profile, smoke only), Assassin (2 plan rows, 1 profile, trap planner missing). Evidence counted by grep in this branch; no drill rows exist for any of these classes (U for live state).

Source for TASK-999 (`classorder1`, lane waiting-owner): not present in this repo or the ops checkout when this card was written; the owner decision is recorded here instead.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Priority lines written in roadmap rows R10 to R12 and R16 and in the six class epics
- [x] #2 Class order recorded in the card description
<!-- AC:END -->
