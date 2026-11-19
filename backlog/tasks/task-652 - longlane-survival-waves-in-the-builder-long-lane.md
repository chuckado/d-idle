---
id: TASK-652
title: 'longlane: survival waves in the builder long lane'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T20:47:00Z'
updated_date: '2026-11-18T21:01:00Z'
labels:
  - ops
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "longlane"
owns:
  - "tools/survival/wavelane.sh"
exclusive: []
contracts: []
result: "Builder long lane (150 min, own lock, nice 12, memory guard) built, installed and tested on the builder (private ops repo commit 7c432d6); wavelane.sh launches survival waves there with a time-cap check. Verified live on the builder with sleep jobs and a 5 s timeout drill; the real wave.sh is on agent/survhold2 and was not touched."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S. Builder long lane (150 min) for survival A/B waves; the builder side lives in the private ops repo.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 wavelane.sh launches a wave in the long lane with a time-cap check
- [x] #2 long lane installed and tested on the builder (ops repo commit 7c432d6)
<!-- AC:END -->
