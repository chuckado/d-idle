---
id: TASK-973
title: 'arbmig: P6 arbiter migration: outcome exit rule, TEMPORARY intents, lost-control'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:54:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-981
priority: high
ordinal: 461000
card: "arbmig"
owns:
  - "docs/design/arbiter-migration-p6.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Cause classes killed: A (deciders undo each other, 29) and E (stale state after preemption). Roadmap row P6 (docs/research/fable-logic-review-2026-11-25.md 7.1, 7.3; ADR 0020 item 3). The owner approved the changed exit rule on 2026-11-26: a family moves by OUTCOME, not agreement: on recorded sessions (ticklog, d2arbiter -follow) and incident scenarios (incident1) the arbiter-led run has fewer P1 monitor hits than the chain-led run and no new ones; then the family moves behind its knob on one sandbox. This card rewrites the docs/arbiter.md migration table accordingly (it is a doc, not an ADR). Also: TEMPORARY intents and a lost-control hook (Baritone idea; two fields on Intent, one method on Source); sources read P2 refusals instead of re-deriving facts (also covers the quest-board intent source, boardsm 9 of the 2026-11-21 review, and town10 town.plan as source). Order of families unchanged: survive and player, chores and claim, gold farm and recovery, quest, level. Deletes after 14 days of no firing: five loop guards, chores cooldown family, forceTrip, claimSkipUntil. Depends on P2 (outcome1) and T4 (ticklog); soft on incident1. Risk: medium to high, the core. Enable criteria per family: outcome rule met, one sandbox for 3 days, owner go; bot.arbiterShadow stays 1 meanwhile. Will split into several cards at start (one per family).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/arbiter.md migration table rewritten to the outcome rule
- [ ] #2 First family (survive, player) acting on one sandbox behind its own knob with a ledger row
<!-- AC:END -->
