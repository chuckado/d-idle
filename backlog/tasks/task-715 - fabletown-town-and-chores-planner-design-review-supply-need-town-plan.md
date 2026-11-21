---
id: TASK-715
title: 'fabletown: town and chores planner design review (supply need, town plan)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T03:22:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 340000
card: "fabletown"
owns:
  - "docs/design/fable-review-2026-11-21/town-planner*"
exclusive: []
contracts: []
result: "Docs only, nothing run and nothing verified live: three docs under docs/design/fable-review-2026-11-21/ (town-planner.md, appendix A current map with 16 rule conflicts, appendix B data model, 12 scenario sketches, 11 migration cards). Code claims are V by reading origin/main at 5b12ce5e7; run evidence is quoted from the probe docs (CHEATED setups, mechanics only). 13 open questions for the user. No build-specific code: the design moves class lists into build profile data."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fable design review asked by the user on 2026-11-20: one source of truth for what the hero needs before it leaves town. Done when: the three docs exist, cite file:line with V/U tags, and the checks pass. Docs only, no Go change. Consumer: TASK-219 (gates call the evaluator) and the eleven migration cards of appendix B5 (first: tp-scenarios), to be created on the user's go.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Map of the chores and need rules with file:line and the list of rule pairs that disagree (appendix A)
- [x] #2 Supply need and town plan design, bank, sell, pickup, guard and recovery policies, arbiter intent, trace (main doc)
- [x] #3 Data model with knobs as plan inputs and per-build needs (appendix B1 to B3)
- [x] #4 Proving Ground scenario ids with assertion sketches (appendix B4)
- [x] #5 Migration in small scenario-first cards with risk and rollback (appendix B5)
- [x] #6 Player-visible spec and numbered open questions with recommendations
<!-- AC:END -->
