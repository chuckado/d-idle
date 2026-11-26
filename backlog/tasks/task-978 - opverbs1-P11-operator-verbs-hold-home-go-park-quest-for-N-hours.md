---
id: TASK-978
title: 'opverbs1: P11 operator verbs: hold, home, go, park-quest-for-N-hours'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:56:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-968
priority: low
ordinal: 466000
card: "opverbs1"
owns:
  - "docs/design/opverbs-p11.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Controllability. Roadmap row P11 (docs/research/fable-logic-review-2026-11-25.md 4.1, 7.1). Operator socket gains hold, home, go and park mapped onto the existing orders. OWNER DECISION NEEDED: the ADR 0030 allowlist amendment (review 7.5); same guard rails (rate limit, audit, never in a measured run). Do not start before the owner's yes. Risk: low. Enable criteria: owner yes; audited and rate-limited verbs on a sandbox.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Verbs added only after the owner's word on the allowlist; tests for refusal in a measured run
<!-- AC:END -->
