---
id: TASK-487
title: 'uis7review: design review of S7 (the cockpit as the default) before it merges'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T09:38:00Z'
labels:
  - watch
  - research
dependencies: []
priority: medium
ordinal: 271000
card: "uis7review"
owns:
  - "docs/ui/redesign-2026-11-14/reviews/s7-review.md"
  - "docs/ui/redesign-2026-11-14/reviews/s7/"
exclusive: []
contracts: []
result: "APPROVE WITH CHANGES for agent/uis7 at be9ac3373. Six required changes: the update banner covers the fixed frame and has no dismiss (and no readable button in the light theme); the one-time layout line reuses the old dismissed key; a hero question row is cut to five words and offers two of its answers; a slot panel that fails at run time leaves no visible way back; phone buttons of 26 and 28 px; the release line must name the way back. Compat matrix of 28 browser states run on the builder: every stored layout and menu choice is kept, the overlay strip is unchanged. Not verified live; fixture renders only."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: S7 flips the default watch layout for everyone without a stored layout; the user rejected the first redesign after seeing it live, so the flip gets a design review before it merges. Consumer: uis7 (TASK of branch agent/uis7). Done when: the review doc has a verdict, the required changes with file and line, the live checks and the rollback.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Compat matrix of stored browser and server layout states run against the branch, each row with its result
- [x] #2 The default page rendered on the builder at the eight fit sizes and 1024x768, dark and light, empty and full, plus the update banner, the reveal sheet, question rows and the live picture; every render looked at
- [x] #3 Verdict with required changes (file:line, why), nits, live checks and the rollback story in docs/ui/redesign-2026-11-14/reviews/s7-review.md
<!-- AC:END -->
