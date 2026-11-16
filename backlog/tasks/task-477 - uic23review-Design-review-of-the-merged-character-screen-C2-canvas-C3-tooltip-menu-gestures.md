---
id: TASK-477
title: 'uic23review: Design review of the merged character screen (C2 canvas, C3 tooltip, menu, gestures)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T06:22:00Z'
labels:
  - watch
  - research
dependencies: []
priority: medium
ordinal: 267000
card: "uic23review"
owns:
  - "docs/ui/redesign-2026-11-14/reviews/c2c3-review.md"
  - "docs/ui/redesign-2026-11-14/reviews/c2c3/"
exclusive: []
contracts: []
result: "APPROVE WITH CHANGES, not verified live (fixture viewer on the builder only). docs/ui/redesign-2026-11-14/reviews/c2c3-review.md: 14 required changes (phone storage tab loses belt, selected box and errands; phone hero sheet unreadable; dead zones with no mercenary, in the mercenary tiles and in the selected box; pending outline and TOWN badge lost on every repaint; grid lags an own action by up to 4 s; band clipped at 1536 px and below; light-theme tooltip contrast 2.9 to 3.5:1; long names in one-row slots; touch long press against drag; Delete then Enter runs the wrong entry; dead header with Revive overflows; fit fixtures), 8 should, nits, 9 live checks, 8 renders under 150 KB. Fit check and the watch tests pass on the builder with the default fixture. Nothing build-specific."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: the briefs README asks for a design review once C2 and C3 are merged; the first inventory page was rejected for dead space. Done when: the review doc is in the repo with a verdict and exact changes.

Consumer: uic2 (TASK-456) and uic3 (TASK-463) follow-up fix cards; S7 (the default flip) reads the required changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The merged page rendered on the builder at the eight sizes, dark and light, with tooltip, menu, selection, mercenary states, full and long data, and looked at
- [x] #2 Review doc with a verdict, required changes with file and line, nits, and the live checks
- [x] #3 Four renders embedded by path, each under 150 KB
<!-- AC:END -->
