---
id: TASK-480
title: >-
  cockpitreview: Design review of the merged watch cockpit (S1b to S6, B1, B4 to
  B6) before S7 makes it the default
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T07:12:00Z'
updated_date: '2026-11-15T07:27:00Z'
labels:
  - watch
  - research
dependencies: []
priority: medium
ordinal: 268000
card: "cockpitreview"
owns:
  - "docs/ui/redesign-2026-11-14/reviews/cockpit-review.md"
  - "docs/ui/redesign-2026-11-14/reviews/cockpit/"
exclusive: []
contracts: []
result: "APPROVE WITH CHANGES, not verified live (fixture viewer on the builder only; the picture tile shows stream down in every render). docs/ui/redesign-2026-11-14/reviews/cockpit-review.md: 13 required changes before S7 flips the default (a player pause never reaches the band because GlanceSources.Paused is bound nowhere; player_pause texts proposed; Needs you overflows at 1280x620 with five needs and eight orders; Health is colour only on short windows; Hold unreachable at three desktop sizes; order titles cut before their reasons at 1440x900; the sentence redrawn and re-announced on every poll; doubled Next label; Go now cancels the open orders and Undo does not restore them; zone menu has no name and no focus handling; map legend missing; map switch covers a zone button at 1440x900 and on the phone; phone band wraps and cuts the reason), 10 should-fix items (text-3 contrast under 4.5:1 in both themes among them), wording per Why code, a 12-point S7 checklist (default choice lives in shell.js, the Layout menu has no Cockpit entry, the layout note and first-run card do not show in a cockpit), 9 live checks, 6 renders under 150 KB. 113 renders made on the builder, about 25 read. Nothing build-specific."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: S7 flips the default watch layout to the cockpit; the user's criteria (one screen at eight sizes, no dead space, everything visible at once, margin in words) need an independent read of what is merged. Done when: the review file is in the tree with a verdict, the required changes and the S7 checklist. Consumer: uis7 (S7, the cockpit becomes the default).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The merged cockpit rendered on the builder at the eight sizes, dark and light, with the needs, orders, paused, blocked, mercenary and supplies, zone menu, phone tab and drawer states, and every render group looked at
- [x] #2 Review file with a verdict, required changes by slice owner with file and line, nits, the S7 checklist and the live checks
- [x] #3 Wording reviewed for every Why code and a text proposed for player_pause
<!-- AC:END -->
